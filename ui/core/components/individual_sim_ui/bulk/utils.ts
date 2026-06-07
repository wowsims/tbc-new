import type { Player } from '../../../player';
import { ReforgeOptimizer } from '../../suggest_reforges_action';
import { ReforgeGearCache } from '../../../reforge_cache';
import { BulkGearCandidate, BulkSimResult, BulkSimStage, DistributionMetrics, ReforgeOptimizeRequest } from '../../../proto/api';
import { Debuffs, PartyBuffs, RaidBuffs } from '../../../proto/common';
import { ItemSlot } from '../../../proto/common';
import { Database } from '../../../proto_utils/database';
import { Gear } from '../../../proto_utils/gear';
import { sleep } from '../../../utils';
import {
	BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS,
	BULK_SIM_ITEM_SLOT_TO_SINGLE_ITEM_SLOT,
	BulkSimItemSlot,
	ITEM_SLOT_TO_BULK_SIM_ITEM_SLOT,
} from './constants_auto_gen';

export { BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS, BULK_SIM_ITEM_SLOT_TO_SINGLE_ITEM_SLOT, BulkSimItemSlot, ITEM_SLOT_TO_BULK_SIM_ITEM_SLOT };

export const getBulkItemSlotFromSlot = (slot: ItemSlot, canDualWield: boolean): BulkSimItemSlot => {
	if (canDualWield && [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand].includes(slot)) {
		return BulkSimItemSlot.ItemSlotHandWeapon;
	}
	return ITEM_SLOT_TO_BULK_SIM_ITEM_SLOT.get(slot)!;
};

export const dedupeGearSets = (gearSets: Gear[], existingGearSets: Gear[] = []): Gear[] => {
	const seenGearKeys = new Set<string>(existingGearSets.map(gear => gear.getGearKey()));
	return gearSets.filter(gear => {
		const gearKey = gear.getGearKey();
		if (seenGearKeys.has(gearKey)) return false;
		seenGearKeys.add(gearKey);
		return true;
	});
};

export type BulkSimReforgeCacheProgress = {
	processedCandidates?: number;
	totalCandidates?: number;
	restoredCandidates?: number;
	current: number;
	total: number;
	message?: string;
};

type BulkSimReforgeCacheData = {
	cache: ReforgeGearCache;
	candidates: BulkGearCandidate[];
	optimizedCandidates: BulkGearCandidate[];
	cachedOptimizedGearSets: Gear[];
	cacheKeysByCandidateIndex: Map<number, string>;
};

type BulkSimReforgeCacheContext = {
	player: Player<any>;
	gearSets: Gear[];
	db: Database;
	reforgeRequest: ReforgeOptimizeRequest;
	raidBuffs: RaidBuffs;
	partyBuffs: PartyBuffs | undefined;
	debuffs: Debuffs;
	onProgress?: (progress: BulkSimReforgeCacheProgress) => void;
	signal?: AbortSignal;
};

const BULK_CACHE_LOOKUP_BATCH_SIZE = 2000;
const BULK_CACHE_PROGRESS_CHECK_MODULO = 64;
const BULK_CACHE_YIELD_BUDGET_MS = 16;

export async function getBulkSimReforgeCacheData({
	player,
	gearSets,
	db,
	reforgeRequest,
	raidBuffs,
	partyBuffs,
	debuffs,
	onProgress,
	signal,
}: BulkSimReforgeCacheContext): Promise<BulkSimReforgeCacheData> {
	throwIfAborted(signal);

	const cache = ReforgeGearCache.get(player.getPlayerSpec());
	const configHash = await ReforgeOptimizer.getConfigHash({ player, reforgeRequest, raidBuffs, partyBuffs, debuffs });
	const totalCandidates = gearSets.length;
	onProgress?.({
		processedCandidates: 0,
		totalCandidates,
		restoredCandidates: 0,
		current: 0,
		total: totalCandidates,
	});

	let lastYieldAt = performance.now();

	const candidates: BulkGearCandidate[] = [];
	const optimizedCandidates: BulkGearCandidate[] = [];
	const cachedOptimizedGearSets: Gear[] = [];
	const cacheKeysByCandidateIndex = new Map<number, string>();
	const pendingEntries: Array<{ index: number; gear: Gear; cacheKey: string }> = [];

	let processedCandidates = 0;
	let restoredCandidates = 0;

	const flushPendingEntries = async () => {
		if (!pendingEntries.length) {
			return;
		}

		const cachedGearByKey = await cache.getMany(
			pendingEntries.map(entry => entry.cacheKey),
			signal,
		);
		for (const entry of pendingEntries) {
			throwIfAborted(signal);
			const cachedGear = cachedGearByKey.get(entry.cacheKey);
			if (cachedGear) {
				optimizedCandidates.push(BulkGearCandidate.create({ index: entry.index, gear: cachedGear }));
				cachedOptimizedGearSets.push(db.lookupEquipmentSpec(cachedGear));
				restoredCandidates++;
			} else {
				candidates.push(BulkGearCandidate.create({ index: entry.index, gear: entry.gear.asSpec() }));
				cacheKeysByCandidateIndex.set(entry.index, entry.cacheKey);
			}

			processedCandidates++;
			if (processedCandidates % BULK_CACHE_PROGRESS_CHECK_MODULO === 0 || processedCandidates === totalCandidates) {
				const now = performance.now();
				if (processedCandidates === totalCandidates || now - lastYieldAt >= BULK_CACHE_YIELD_BUDGET_MS) {
					onProgress?.({
						processedCandidates,
						totalCandidates,
						restoredCandidates,
						current: processedCandidates,
						total: totalCandidates,
						message: restoredCandidates > 0 ? `Restored ${restoredCandidates}` : undefined,
					});
					await sleep(0);
					lastYieldAt = performance.now();
				}
			}
		}

		pendingEntries.length = 0;
	};

	for (let i = 0; i < totalCandidates; i++) {
		throwIfAborted(signal);
		const gear = gearSets[i];
		const cacheKey = await ReforgeGearCache.getKey(gear.getGearKey(), configHash);
		pendingEntries.push({ index: i, gear, cacheKey });

		if (pendingEntries.length >= BULK_CACHE_LOOKUP_BATCH_SIZE || i + 1 === totalCandidates) {
			await flushPendingEntries();
		}
	}

	return { cache, candidates, optimizedCandidates, cachedOptimizedGearSets, cacheKeysByCandidateIndex };
}

export async function writeBulkSimReforgeCacheResults(optimizedCandidates: BulkGearCandidate[], cacheData: BulkSimReforgeCacheData): Promise<void> {
	const cacheEntries = optimizedCandidates.flatMap(candidate => {
		const cacheKey = cacheData.cacheKeysByCandidateIndex.get(candidate.index);
		if (!cacheKey || !candidate.gear) return [];
		return [{ key: cacheKey, optimizedGear: candidate.gear }];
	});
	await cacheData.cache.setGearMany(cacheEntries);
}

export const bulkSimStageToOptimisationStage = (bulkStage: BulkSimStage): 'low' | 'medium' | 'high' | undefined => {
	switch (bulkStage) {
		case BulkSimStage.BulkSimStageLow:
			return 'low';
		case BulkSimStage.BulkSimStageMedium:
			return 'medium';
		case BulkSimStage.BulkSimStageHigh:
			return 'high';
		default:
			return undefined;
	}
};

export const cleanBulkDpsMetrics = (metrics: DistributionMetrics): DistributionMetrics => ({
	...metrics,
	hist: [],
	allValues: [],
});

export const getCoreBulkSimTrackingMetrics = (result: BulkSimResult): Record<string, string | number> => {
	const metrics: Record<string, string | number> = {
		totalSeconds: result.timings?.totalSeconds ?? 0,
		simmingSeconds: result.timings?.simmingSeconds ?? 0,
		stageCount: result.stageMetrics.length,
	};

	for (const stage of result.stageMetrics) {
		const stageName = BulkSimStage[stage.stage] ?? stage.stage.toString();
		metrics[`${stageName}_inputs`] = stage.inputGearSets;
		metrics[`${stageName}_survivors`] = stage.survivors;
		metrics[`${stageName}_iterations`] = stage.iterations;
		metrics[`${stageName}_duration_seconds`] = stage.durationSeconds;
	}

	return metrics;
};

export const throwIfAborted = (signal?: AbortSignal, errorMessage = 'Bulk Sim Aborted'): void => {
	if (signal?.aborted) {
		throw new Error(errorMessage);
	}
};
