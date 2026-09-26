import { ErrorOutcomeType, Raid as RaidProto, ReforgeOptimizeMode, ReforgeOptimizeRequest, ReforgeSettings } from '@generated/proto/api';
import { EquipmentSpec } from '@generated/proto/common';
import { SimRequest } from '@worker/types';

import { SimSignals } from '../sim_signal_manager';
import { generateRequestId, WorkerPool } from '../workers/worker_pool';

// windows-1252 ('latin1') maps all 256 byte values to distinct code points, so this is an
// injective byte-to-string encoding - one exact-length string, no intermediate array.
const reforgeGearKeyDecoder = new TextDecoder('latin1');
export const reforgeGearKey = (gear: EquipmentSpec): string => {
	return reforgeGearKeyDecoder.decode(EquipmentSpec.toBinary(gear));
};

// The outcome of one solve: the gear, or null when it failed; infeasibleStatConstraints
// when the batch's stat constraints cannot be met by any gem choice for that gear.
export type ReforgeGearSolve = {
	gear: EquipmentSpec | null;
	infeasibleStatConstraints: boolean;
};

export const optimizeReforgeGear = async (
	baseRaid: RaidProto,
	templateRequest: ReforgeOptimizeRequest,
	gear: EquipmentSpec,
	includeGems: boolean,
	workerPool: WorkerPool,
	signals: SimSignals,
	mode = ReforgeOptimizeMode.ReforgeOptimizeModeSingle,
): Promise<ReforgeGearSolve> => {
	const failed: ReforgeGearSolve = { gear: null, infeasibleStatConstraints: false };
	const reforgeRequest = makeReforgeRequest(baseRaid, templateRequest, gear, includeGems, mode);
	if (!reforgeRequest) {
		return failed;
	}

	try {
		const result = await workerPool.reforgeOptimizeAsync(reforgeRequest, signals);
		if (result.infeasibleStatConstraints) {
			return { gear: null, infeasibleStatConstraints: true };
		}
		if (result.error) {
			if (result.error.type != ErrorOutcomeType.ErrorOutcomeAborted) {
				console.warn(`[Reforge] Optimization failed includeGems=${includeGems}: ${result.error.message}`);
			}
			return failed;
		}

		return { gear: result.optimizedGear ? EquipmentSpec.clone(result.optimizedGear) : null, infeasibleStatConstraints: false };
	} catch (error) {
		if (!signals.abort.isTriggered()) {
			console.warn(`[Reforge] Optimization failed includeGems=${includeGems}`, error);
		}
		return failed;
	}
};

export const makeReforgeRequest = (
	baseRaid: RaidProto,
	templateRequest: ReforgeOptimizeRequest,
	gear: EquipmentSpec,
	includeGems: boolean,
	mode = ReforgeOptimizeMode.ReforgeOptimizeModeSingle,
): ReforgeOptimizeRequest | null => {
	const raid = RaidProto.clone(baseRaid);
	const player = raid.parties[0]?.players[0];
	if (!player) {
		return null;
	}

	player.equipment = EquipmentSpec.clone(gear);
	const reforgeRequest = ReforgeOptimizeRequest.clone(templateRequest);
	reforgeRequest.requestId = generateRequestId(SimRequest.reforgeOptimizeAsync);
	reforgeRequest.raid = raid;
	reforgeRequest.mode = mode;
	reforgeRequest.settings = ReforgeSettings.clone(reforgeRequest.settings ?? ReforgeSettings.create());
	if (!includeGems) {
		reforgeRequest.gemOptions = [];
	}
	return reforgeRequest;
};
