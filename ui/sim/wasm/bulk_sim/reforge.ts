import { BulkGearCandidate, BulkSimRequest, BulkSimStage, ReforgeOptimizeMode, ReforgeOptimizeRequest } from '@generated/proto/api';
import { EquipmentSpec } from '@generated/proto/common';
import { queue } from 'async';

import { SimSignals } from '../../sim_signal_manager';
import { formatDurationSeconds } from '../../utils/format';
import { WorkerPool, WorkerProgressCallback } from '../../workers/worker_pool';
import { optimizeReforgeGear, reforgeGearKey, ReforgeGearSolve } from '../reforge_optimizer';
import { makeBulkSimStageProgressEmitter } from './progress';
import { BulkSimReforgeCandidateTask } from './types';

export const optimizeReforgeCandidates = async (
	request: BulkSimRequest,
	workerPool: WorkerPool,
	onProgress: WorkerProgressCallback,
	signals: SimSignals,
	onReforgeCandidateOptimized?: (candidate: BulkGearCandidate, optimizedGear: EquipmentSpec) => void | Promise<void>,
): Promise<{ request: BulkSimRequest; aborted: boolean }> => {
	if (!request.reforgeRequest || !request.baseRequest?.raid) {
		return { request, aborted: false };
	}
	// The batch's stat constraints become rows of every candidate's model.
	const reforgeRequest = ReforgeOptimizeRequest.create({
		...request.reforgeRequest,
		statConstraints: request.bulkSettings?.statConstraints ?? [],
	});

	const candidates = request.candidates.filter(candidate => candidate.gear);
	const optimizedCandidates: BulkGearCandidate[] = request.optimizedCandidates;
	if (!candidates.length) {
		return {
			request: withReforgedCandidates(request, dedupeBulkSimReforgeCandidates(request, optimizedCandidates), []),
			aborted: false,
		};
	}

	const startedAt = new Date().getTime();
	const concurrency = Math.max(1, Math.min(workerPool.getNumWorkers(), candidates.length));
	console.log(`[Bulk Sim] Reforge optimization started candidates=${candidates.length} concurrency=${concurrency} wasm=true`);
	const emitter = makeBulkSimStageProgressEmitter(onProgress, BulkSimStage.BulkSimStageReforge, candidates.length, candidates.length);
	emitter.report(0, 0, 0);

	const gearCache = makeBulkSimReforgeGearCache(request, reforgeRequest, workerPool, signals);
	const collector = makeBulkSimReforgeCollector(request.baseRequest.raid.parties[0]?.players[0]?.equipment, candidates.length);
	// TBC's reforger is gem-driven, so the gem-inclusive model is the one with gem options.
	const includeGems = reforgeRequest.gemOptions.length > 0;

	const reforgeQueue = queue<BulkSimReforgeCandidateTask, Error>(async ({ candidate, position }) => {
		if (signals.abort.isTriggered()) return;
		if (!candidate.gear) return;

		// Retry without gems: a gem-inclusive model can be infeasible where the
		// reforge-only one is not. Not when the stat constraints are what made it
		// infeasible: no gem choice meets them, so the candidate is ruled out.
		const gearKey = reforgeGearKey(candidate.gear);
		let solve = await gearCache.optimize(candidate.gear, gearKey, includeGems);
		if (!solve.gear && !solve.infeasibleStatConstraints && !signals.abort.isTriggered() && includeGems) {
			solve = await gearCache.optimize(candidate.gear, gearKey, false);
		}
		let optimizedGear = solve.gear;
		const optimizedSuccessfully = !!optimizedGear;
		if (!optimizedGear) {
			if (signals.abort.isTriggered()) return;
			// No gem choice from the pool meets the stat constraints: that says nothing about the
			// gems the candidate already has, so it keeps them, like a failed solve, and the
			// final-stats check decides. Not logged: with a constraint, that can be most candidates.
			if (!solve.infeasibleStatConstraints) {
				console.warn(`[Bulk Sim] Reforge optimization failed for candidate ${candidate.index}; using original gear`);
			}
			optimizedGear = candidate.gear;
		}

		const completedCandidates = collector.record(candidate, optimizedGear, position);
		if (optimizedSuccessfully) {
			await onReforgeCandidateOptimized?.(candidate, optimizedGear);
		}
		emitter.report(completedCandidates, completedCandidates, 0);
	}, concurrency);

	const queueErrorPromise = reforgeQueue.error();
	candidates.forEach((candidate, position) => reforgeQueue.push({ candidate, position }));
	await Promise.race([reforgeQueue.drain(), queueErrorPromise]);
	if (signals.abort.isTriggered()) {
		return { request: buildBulkSimReforgeRequest(request, optimizedCandidates, collector.completed(), true), aborted: true };
	}

	const outputCandidates = optimizedCandidates.length + collector.completed().length;
	console.log(
		`[Bulk Sim] Reforge optimization completed candidates=${collector.completedCount()} outputCandidates=${outputCandidates} total=${formatDurationSeconds((new Date().getTime() - startedAt) / 1000, { showMilliseconds: true, millisecondDigits: 2 })}`,
	);

	return {
		request: buildBulkSimReforgeRequest(request, optimizedCandidates, collector.completed(), false),
		aborted: false,
	};
};

const dedupeBulkSimReforgeCandidates = (request: BulkSimRequest, candidates: BulkGearCandidate[]): BulkGearCandidate[] => {
	const seenGearKeys = new Set<string>();
	const baselineGear = request.baseRequest?.raid?.parties[0]?.players[0]?.equipment;
	if (baselineGear) {
		seenGearKeys.add(reforgeGearKey(baselineGear));
	}

	const deduped: BulkGearCandidate[] = [];
	for (const candidate of candidates) {
		if (!candidate.gear) continue;

		const gearKey = reforgeGearKey(candidate.gear);
		if (seenGearKeys.has(gearKey)) continue;

		seenGearKeys.add(gearKey);
		deduped.push(candidate);
	}
	return deduped;
};

type BulkSimReforgeGearCache = {
	optimize: (gear: EquipmentSpec, gearKey: string, includeGems: boolean) => Promise<ReforgeGearSolve>;
};

const cloneSolve = (solve: ReforgeGearSolve): ReforgeGearSolve => ({
	gear: solve.gear ? EquipmentSpec.clone(solve.gear) : null,
	infeasibleStatConstraints: solve.infeasibleStatConstraints,
});

// Memoizes solves by gear key, and lets a second caller await an in-flight solve for the same
// key instead of starting a duplicate one. Every hand-out is cloned so a caller mutating the
// gear it received cannot corrupt the entry other candidates read.
const makeBulkSimReforgeGearCache = (
	request: BulkSimRequest,
	templateRequest: ReforgeOptimizeRequest,
	workerPool: WorkerPool,
	signals: SimSignals,
): BulkSimReforgeGearCache => {
	const optimizedGearByKey = new Map<string, ReforgeGearSolve>();
	const inFlightOptimizedGearByKey = new Map<string, Promise<ReforgeGearSolve>>();
	const failed: ReforgeGearSolve = { gear: null, infeasibleStatConstraints: false };

	return {
		optimize: async (gear: EquipmentSpec, gearKey: string, includeGems: boolean): Promise<ReforgeGearSolve> => {
			const cacheKey = `${gearKey}:${includeGems ? 1 : 0}`;
			const cachedSolve = optimizedGearByKey.get(cacheKey);
			if (cachedSolve) {
				return cloneSolve(cachedSolve);
			}
			const inFlightSolve = inFlightOptimizedGearByKey.get(cacheKey);
			if (inFlightSolve) {
				return cloneSolve(await inFlightSolve);
			}

			const baseRaid = request.baseRequest?.raid;
			if (!baseRaid) {
				optimizedGearByKey.set(cacheKey, failed);
				return failed;
			}

			const optimizePromise = optimizeReforgeGear(
				baseRaid,
				templateRequest,
				gear,
				includeGems,
				workerPool,
				signals,
				ReforgeOptimizeMode.ReforgeOptimizeModeBulk,
			);
			inFlightOptimizedGearByKey.set(cacheKey, optimizePromise);
			try {
				const solve = await optimizePromise;
				optimizedGearByKey.set(cacheKey, cloneSolve(solve));
				return solve;
			} finally {
				inFlightOptimizedGearByKey.delete(cacheKey);
			}
		},
	};
};

// Collects optimized candidates, keeping the first occurrence of each distinct gear set
// (the baseline counts as already seen) and preserving input order.
const makeBulkSimReforgeCollector = (baselineGear: EquipmentSpec | undefined, candidateCount: number) => {
	const seenGearKeys = new Set<string>();
	if (baselineGear) {
		seenGearKeys.add(reforgeGearKey(baselineGear));
	}
	const completedByPosition: Array<BulkGearCandidate | undefined> = new Array(candidateCount);
	let completedCandidates = 0;

	return {
		record: (candidate: BulkGearCandidate, optimizedGear: EquipmentSpec, position: number): number => {
			const gearKey = reforgeGearKey(optimizedGear);
			if (!seenGearKeys.has(gearKey)) {
				seenGearKeys.add(gearKey);
				completedByPosition[position] = BulkGearCandidate.create({ index: candidate.index, gear: optimizedGear });
			}
			completedCandidates++;
			return completedCandidates;
		},
		completed: (): BulkGearCandidate[] => completedByPosition.filter((candidate): candidate is BulkGearCandidate => !!candidate),
		completedCount: (): number => completedCandidates,
	};
};

// The request the bulk sim proceeds with: reforged candidates, deduplicated, with the
// reforge pre-pass removed. An aborted run also reports them so the frontend can still
// write its cache entries.
const buildBulkSimReforgeRequest = (
	request: BulkSimRequest,
	optimizedCandidates: BulkGearCandidate[],
	completedCandidates: BulkGearCandidate[],
	aborted: boolean,
): BulkSimRequest => {
	const partialOptimizedCandidates = dedupeBulkSimReforgeCandidates(request, [...optimizedCandidates, ...completedCandidates]);
	return withReforgedCandidates(request, partialOptimizedCandidates, aborted ? partialOptimizedCandidates : []);
};

const withReforgedCandidates = (request: BulkSimRequest, candidates: BulkGearCandidate[], optimizedCandidates: BulkGearCandidate[]): BulkSimRequest => ({
	...request,
	candidates,
	optimizedCandidates,
	reforgeRequest: undefined,
});
