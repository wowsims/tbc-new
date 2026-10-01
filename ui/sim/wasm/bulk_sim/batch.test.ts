// The candidate batch reuses one sim request across candidates. Each candidate's weapon stones must
// be adjusted from the base character's, not from the candidate simmed before it: the rule drops a
// stone for a hand without a sharp or blunt weapon and has nothing to restore it from.
import { BulkSimRequest, Player, Raid, RaidSimRequest, RaidSimResult, SimOptions } from '@generated/proto/api';
import { ConsumesSpec, EquipmentSpec, ItemSlot } from '@generated/proto/common';
import { describe, expect, it, vi } from 'vitest';

import { ADAMANTITE_SHARPENING_STONE_ID, adjustWeaponImbueId } from '../../proto/utils';
import type { SimSignals } from '../../sim_signal_manager';
import type { WorkerPool } from '../../workers/worker_pool';
import { runBulkSimCandidateBatchOnWorkers } from './batch';

// Weapon types without an item database: an off-hand item id of SWORD is a sharp weapon, anything
// else (a shield) takes no stone. The imbue rule itself is the real one.
const SWORD = 100;
vi.mock('../../proto/database', () => ({
	Database: {
		getSync: () => ({
			lookupEquipmentSpec: (gear: EquipmentSpec) => ({
				adjustImbues: (consumes: ConsumesSpec) => {
					const ohImbueId = adjustWeaponImbueId(consumes.ohImbueId, gear.items[ItemSlot.ItemSlotOffHand]?.id === SWORD, false);
					return ohImbueId === consumes.ohImbueId ? consumes : ConsumesSpec.clone({ ...consumes, ohImbueId });
				},
			}),
		}),
	},
}));

const signals = { abort: { isTriggered: () => false } } as unknown as SimSignals;

const gearWithOffHand = (id: number) => {
	const gear = EquipmentSpec.create({ items: Array.from({ length: ItemSlot.ItemSlotRanged + 1 }, () => ({})) });
	gear.items[ItemSlot.ItemSlotOffHand].id = id;
	return gear;
};

describe('runBulkSimCandidateBatchOnWorkers', () => {
	it("adjusts each candidate's weapon stone from the base character's", async () => {
		const offHands: number[] = [];
		const workerPool = {
			getNumWorkers: () => 1, // One at a time, so the shield candidate is simmed first.
			raidSimAsync: vi.fn(async (request: RaidSimRequest) => {
				offHands.push(request.raid!.parties[0].players[0].consumables!.ohImbueId);
				return RaidSimResult.create();
			}),
		} as unknown as WorkerPool;

		await runBulkSimCandidateBatchOnWorkers(
			BulkSimRequest.create({
				baseRequest: RaidSimRequest.create({
					raid: Raid.create({
						parties: [{ players: [Player.create({ consumables: ConsumesSpec.create({ ohImbueId: ADAMANTITE_SHARPENING_STONE_ID }) })] }],
					}),
					simOptions: SimOptions.create({ iterations: 10 }),
				}),
			}),
			[
				{ index: 0, gear: gearWithOffHand(200) },
				{ index: 1, gear: gearWithOffHand(SWORD) },
			],
			10,
			workerPool,
			signals,
			{ completedSimsBase: 0, completedIterationsBase: 0, emitter: { report: vi.fn() } },
		);

		expect(offHands).toEqual([0, ADAMANTITE_SHARPENING_STONE_ID]);
	});
});
