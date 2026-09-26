import { BulkSettings, ProgressMetrics } from '@generated/proto/api';
import i18n from '@i18n/config';
import type { BulkSimReforgeCacheProgress } from '@sim/bulk/reforge_cache';
import { BulkResults, BulkSimProgressConfig, TopGearResult } from '@sim/bulk/types';
import { dedupeGearSets } from '@sim/bulk/utils';
import type { Player } from '@sim/player/player';
import { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';
import { bulkState, patchBulkState } from '@sim/settings/bulk_settings';
import type { ReforgeOptimizeConfig } from '@sim/sim';
import type { IndividualSimHost } from '@sim/sim_host';
import { RequestTypes } from '@sim/sim_signal_manager';
import { toastManager } from '@ui-kit/Toast';

import { trackEvent } from '../../../tracking/utils';
import { runCoreBulkSim as runCoreBulkSimImpl } from './core_sim';
import { BulkProgress, candidateGearProgress, constraintsProgress, simProgress } from './progress';
import { createBulkSettingsProto } from './settings';
import { buildTieChains } from './tie_chains';

/**
 * A batch in flight, per player. None of it belongs in the store: an abort controller and a
 * promise are not state a selector can compare, and the progress ticks are kept out on purpose,
 * so a tick renders the one leaf that shows it instead of every reader of the slice.
 */
interface BulkRun {
	simStart: number;
	isCancelling: boolean;
	abortController: AbortController | null;
	abortPromise: Promise<void> | null;
	combinationsRequestVersion: number;
	candidateBuildStartedAt: number | undefined;
	cacheRestoreStartedAt: number | undefined;
	progress: BulkProgress | null;
	listeners: Set<(progress: BulkProgress) => void>;
}

const runs = new WeakMap<Player<any>, BulkRun>();

const runOf = (player: Player<any>): BulkRun => {
	let run = runs.get(player);
	if (!run) {
		run = {
			simStart: 0,
			isCancelling: false,
			abortController: null,
			abortPromise: null,
			combinationsRequestVersion: 0,
			candidateBuildStartedAt: undefined,
			cacheRestoreStartedAt: undefined,
			progress: null,
			listeners: new Set(),
		};
		runs.set(player, run);
	}
	return run;
};

export const bulkProgress = (player: Player<any>): BulkProgress | null => runOf(player).progress;

export const subscribeBulkProgress = (player: Player<any>, listener: (progress: BulkProgress) => void): (() => void) => {
	const { listeners } = runOf(player);
	listeners.add(listener);
	return () => {
		listeners.delete(listener);
	};
};

const emitProgress = (player: Player<any>, next: BulkProgress | null) => {
	if (!next) return;
	const run = runOf(player);
	run.progress = next;
	for (const listener of run.listeners) listener(next);
};

const setCandidateGearProgress = (
	player: Player<any>,
	input: { completed?: number; total?: number; title?: string; stage?: string; startedAt?: number } = {},
) => {
	emitProgress(player, candidateGearProgress({ ...input, now: new Date().getTime() }));
};

const setSimProgress = (player: Player<any>, metrics: ProgressMetrics, config: BulkSimProgressConfig) => {
	emitProgress(player, simProgress(metrics, config, runOf(player).simStart, new Date().getTime()));
};

const calculateBulkCombinations = async (host: IndividualSimHost<any>) => {
	const { sim, player } = host;
	try {
		const bulkSettings = createBulkSettingsProto(player);
		const combinationCountResult = await sim.getBulkCombinationCount(bulkSettings);

		if (combinationCountResult.error) {
			throw new Error(combinationCountResult.error.message || 'Failed to calculate bulk combinations');
		}

		patchBulkState(player, { combinations: combinationCountResult.combinations, iterations: combinationCountResult.iterations });
	} catch (e) {
		host.handleCrash(e);
	}
};

export const refreshBulkCombinations = async (host: IndividualSimHost<any>) => {
	const run = runOf(host.player);
	const requestVersion = ++run.combinationsRequestVersion;
	patchBulkState(host.player, { combinationsPending: true });
	await calculateBulkCombinations(host);
	if (requestVersion !== run.combinationsRequestVersion) {
		return;
	}
	patchBulkState(host.player, { combinationsPending: false });
};

const throwIfBulkAborted = (player: Player<any>, signal: AbortSignal) => {
	if (signal.aborted || runOf(player).isCancelling) {
		throw new Error('Bulk Sim Aborted');
	}
};

const runWithBulkAbort = async <T>(player: Player<any>, promise: Promise<T>, signal: AbortSignal): Promise<T> => {
	throwIfBulkAborted(player, signal);

	let abortHandler: (() => void) | null = null;
	const abortPromise = new Promise<never>((_, reject) => {
		abortHandler = () => reject(new Error('Bulk Sim Aborted'));
		signal.addEventListener('abort', abortHandler, { once: true });
	});

	try {
		return await Promise.race([promise, abortPromise]);
	} finally {
		if (abortHandler) {
			signal.removeEventListener('abort', abortHandler);
		}
	}
};

const setCacheRestoreProgress = (player: Player<any>, cacheProgress: BulkSimReforgeCacheProgress) => {
	const run = runOf(player);
	const isCandidateBuildStage = cacheProgress.stage === 'candidate-build';
	if (isCandidateBuildStage) {
		run.candidateBuildStartedAt ??= new Date().getTime();
	} else {
		run.cacheRestoreStartedAt ??= new Date().getTime();
	}
	setCandidateGearProgress(player, {
		completed: cacheProgress.processedCandidates,
		total: cacheProgress.totalCandidates,
		title: isCandidateBuildStage ? i18n.t('bulk_tab.progress.building_candidate_gear_sets') : i18n.t('bulk_tab.progress.restoring_reforges_from_cache'),
		stage: isCandidateBuildStage ? 'preparing' : 'reforging',
		startedAt: isCandidateBuildStage ? run.candidateBuildStartedAt : run.cacheRestoreStartedAt,
	});
};

const runCoreBulkSim = async (
	host: IndividualSimHost<any>,
	gearSets: Gear[],
	signal: AbortSignal,
	reforgeConfig?: ReforgeOptimizeConfig,
	bulkSettings?: BulkSettings,
): ReturnType<typeof runCoreBulkSimImpl> => {
	const { player } = host;
	return runCoreBulkSimImpl(
		{
			simUI: host,
			throwIfBulkAborted: abortSignal => throwIfBulkAborted(player, abortSignal),
			runWithBulkAbort: (promise, abortSignal) => runWithBulkAbort(player, promise, abortSignal),
			setSimProgress: (metrics, config) => setSimProgress(player, metrics, config),
			setCacheRestoreProgress: cacheProgress => setCacheRestoreProgress(player, cacheProgress),
			setConstraintsProgress: (checked, total) => emitProgress(player, constraintsProgress(checked, total)),
			debugOptimisationRound: (message, data) => console.debug(`[bulk-core] ${message}`, data ?? ''),
		},
		gearSets,
		signal,
		reforgeConfig,
		bulkSettings,
	);
};

const abortBulkSimWork = async (host: IndividualSimHost<any>) => {
	const run = runOf(host.player);
	if (run.abortPromise) {
		return run.abortPromise;
	}

	const abortController = run.abortController;
	if (!abortController) return;

	run.abortController = null;
	if (!abortController.signal.aborted) {
		abortController.abort();
	}

	run.abortPromise = (async () => {
		// Narrower than `All`: cancelling a batch must not also cancel a stat-weights run, which is
		// the whole point of bulk having its own type. Reforge stays in because the batch's own
		// pre-pass registers under it.
		const abortTasks: Promise<unknown>[] = [host.sim.signalManager.abortType(RequestTypes.BulkSim | RequestTypes.ReforgeOptimize)];
		if (host.reforger) {
			abortTasks.push(host.reforger.abortReforgeOptimization());
		}
		await Promise.all(abortTasks);
	})();

	try {
		await run.abortPromise;
	} finally {
		run.abortPromise = null;
	}
};

export const cancelBulkBatch = async (host: IndividualSimHost<any>) => {
	const run = runOf(host.player);
	if (!bulkState(host.player).isRunning || run.isCancelling) return;

	run.isCancelling = true;
	await abortBulkSimWork(host);
};

export const runBulkBatch = async (host: IndividualSimHost<any>) => {
	const { sim, player } = host;
	const run = runOf(player);
	if (bulkState(player).isRunning) return;

	trackEvent({
		action: 'sim',
		category: 'simulate',
		label: 'batch',
		value: bulkState(player).combinations,
	});

	run.isCancelling = false;
	run.candidateBuildStartedAt = undefined;
	run.cacheRestoreStartedAt = undefined;
	run.progress = null;
	patchBulkState(player, { isRunning: true, started: true });
	run.abortController = new AbortController();
	run.abortPromise = null;
	const abortSignal = run.abortController.signal;

	await sim.waitForInit();
	const useNativeBulkSim = sim.isNative ?? false;
	const bulkSettings = createBulkSettingsProto(player);
	// A native server builds the candidates from the settings. The wasm path is handed gear sets
	// instead and is sent only the stat constraints, which its pipeline reads from the request.
	const wasmBulkSettings = bulkSettings.statConstraints.length ? BulkSettings.create({ statConstraints: bulkSettings.statConstraints }) : undefined;
	const requestBulkSettings = useNativeBulkSim ? bulkSettings : wasmBulkSettings;
	let candidateGearSets: Gear[] = [];
	let results: BulkResults | null = null;
	let runError: unknown = null;

	try {
		// Bulk owns its own request type now, so this has to name both: a batch still replaces an
		// in-flight single sim, and it still replaces a previous batch, which naming one would drop.
		await sim.signalManager.abortType(RequestTypes.IndividualSim | RequestTypes.BulkSim);
		run.simStart = new Date().getTime();
		const baseGear = player.getGear();
		patchBulkState(player, { runGear: baseGear, results: null });

		await refreshBulkCombinations(host);

		if (!useNativeBulkSim) {
			setCandidateGearProgress(player);
			const bulkCandidatesResult = await sim.getBulkCandidates(bulkSettings);
			if (bulkCandidatesResult.error) {
				throw new Error(bulkCandidatesResult.error.message || 'Failed to build bulk candidates');
			}
			candidateGearSets = bulkCandidatesResult.candidates
				.filter(candidate => !!candidate.gear)
				.map(candidate => sim.db.lookupEquipmentSpec(candidate.gear!));
			patchBulkState(player, { combinations: bulkCandidatesResult.combinations });
		}

		const reforgeConfig = host.reforger ? host.reforger.getReforgeOptimizeConfig(baseGear) : undefined;
		// With the gem optimizer every candidate must be submitted (its gems differentiate
		// otherwise identical gear); without it duplicates are culled up front.
		const gearSets = reforgeConfig ? candidateGearSets : dedupeGearSets(candidateGearSets, [baseGear]);

		run.simStart = new Date().getTime();
		const bulkSimResult = await runCoreBulkSim(host, gearSets, abortSignal, reforgeConfig, requestBulkSettings);
		const { referenceDpsMetrics, topGearResults, skippedByConstraints } = bulkSimResult;

		const originalGearKey = getGearIdentityKey(baseGear.asSpec());
		const rankedResults = topGearResults.filter(result => getGearIdentityKey(result.gear.asSpec()) !== originalGearKey);
		const originalGearResults: TopGearResult = {
			gear: baseGear,
			dpsMetrics: referenceDpsMetrics,
		};

		rankedResults.push(originalGearResults);
		rankedResults.sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg);

		const resultIterations = Math.max(1, sim.getIterations());
		results = {
			chains: buildTieChains(rankedResults, originalGearResults, resultIterations),
			originalGearResults,
			iterations: resultIterations,
			skippedByConstraints,
			combinations: bulkState(player).combinations,
		};
	} catch (error) {
		runError = error;
		console.error(error);
		const errorMessage = error instanceof Error ? error.message : typeof error === 'string' ? error : undefined;
		if (!run.isCancelling && errorMessage) {
			toastManager.add({
				variant: 'error',
				body: errorMessage,
			});
		}
	} finally {
		const wasCancelling = run.isCancelling;
		if (wasCancelling || runError) {
			await abortBulkSimWork(host);
		}
		await player.setGearAsync(bulkState(player).runGear!);
		if (wasCancelling) {
			toastManager.add({
				variant: 'error',
				body: i18n.t('bulk_tab.notifications.bulk_sim_cancelled'),
			});
		}
		run.isCancelling = false;
		patchBulkState(player, results ? { isRunning: false, results } : { isRunning: false });
	}
};
