// A candidate the gem optimizer cannot bring within the batch's stat constraints must keep its own
// gear: the optimizer only knows the gems in its pool, so its verdict says nothing about the gems the
// candidate already has. The final-stats check after the pre-pass is what decides whether it is
// simmed; the pre-pass must neither drop nor flag it.
import { BulkGearCandidate, BulkSimRequest, BulkStatConstraint, Raid, RaidSimRequest, ReforgeOptimizeRequest } from '@generated/proto/api';
import { EquipmentSpec, ItemSpec, Stat } from '@generated/proto/common';
import { describe, expect, it, vi } from 'vitest';

import type { SimSignals } from '../../sim_signal_manager';
import type { WorkerPool } from '../../workers/worker_pool';
import { optimizeReforgeCandidates } from './reforge';

const optimizeReforgeGear = vi.hoisted(() => vi.fn());
vi.mock('../reforge_optimizer', async importOriginal => ({ ...(await importOriginal<typeof import('../reforge_optimizer')>()), optimizeReforgeGear }));

const gear = (gemId: number) => EquipmentSpec.create({ items: [ItemSpec.create({ id: 100, gems: [gemId] })] });
const signals = { abort: { isTriggered: () => false } } as unknown as SimSignals;
const workerPool = { getNumWorkers: () => 2 } as unknown as WorkerPool;

describe('optimizeReforgeCandidates', () => {
	it('keeps the candidate its own gear when no gem choice meets the stat constraints', async () => {
		optimizeReforgeGear.mockResolvedValue({ gear: null, infeasibleStatConstraints: true });
		const candidate = BulkGearCandidate.create({ index: 3, gear: gear(24033) });
		const request = BulkSimRequest.create({
			baseRequest: RaidSimRequest.create({ raid: Raid.create({ parties: [{ players: [{ equipment: gear(24030) }] }] }) }),
			candidates: [candidate],
			reforgeRequest: ReforgeOptimizeRequest.create({ gemOptions: [{ id: 24030 }] }),
			bulkSettings: { statConstraints: [BulkStatConstraint.create({ unitStat: { oneofKind: 'stat', stat: Stat.StatStamina }, value: 1000 })] },
		});

		const { request: next, aborted } = await optimizeReforgeCandidates(request, workerPool, vi.fn(), signals);

		expect(aborted).toBe(false);
		expect(next.candidates).toEqual([BulkGearCandidate.create({ index: 3, gear: gear(24033) })]);
		// Retrying without gems cannot help when the constraints are what failed.
		expect(optimizeReforgeGear).toHaveBeenCalledTimes(1);
		// Not reported as optimized: those entries go into the 14-day gem cache.
		expect(next.optimizedCandidates).toEqual([]);
	});

	// The solver reads the constraints from the gem optimizer's own request, the copy the gem cache
	// key is made from, so the two cannot drift apart. The batch settings' copy is a fallback for a
	// request that set none there.
	describe('the stat constraints the solver is given', () => {
		const staminaAtLeast = (value: number) => BulkStatConstraint.create({ unitStat: { oneofKind: 'stat', stat: Stat.StatStamina }, value });
		const solverConstraints = async (reforgeConstraints: BulkStatConstraint[], settingsConstraints: BulkStatConstraint[]) => {
			optimizeReforgeGear.mockReset().mockResolvedValue({ gear: gear(24030), infeasibleStatConstraints: false });
			await optimizeReforgeCandidates(
				BulkSimRequest.create({
					baseRequest: RaidSimRequest.create({ raid: Raid.create({ parties: [{ players: [{ equipment: gear(24030) }] }] }) }),
					candidates: [BulkGearCandidate.create({ index: 0, gear: gear(24033) })],
					reforgeRequest: ReforgeOptimizeRequest.create({ gemOptions: [{ id: 24030 }], statConstraints: reforgeConstraints }),
					bulkSettings: { statConstraints: settingsConstraints },
				}),
				workerPool,
				vi.fn(),
				signals,
			);
			return (optimizeReforgeGear.mock.calls[0][1] as ReforgeOptimizeRequest).statConstraints;
		};

		it("are the gem optimizer request's own", async () => {
			expect(await solverConstraints([staminaAtLeast(500)], [staminaAtLeast(1000)])).toEqual([staminaAtLeast(500)]);
		});

		it("fall back to the batch settings' when the request has none", async () => {
			expect(await solverConstraints([], [staminaAtLeast(1000)])).toEqual([staminaAtLeast(1000)]);
		});
	});
});
