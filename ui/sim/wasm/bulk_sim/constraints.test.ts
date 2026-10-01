// The browser's final-stats check judges constraints on the values the stats panel shows, which
// include the raid debuffs the panel attributes to the character (see debuff_stats.ts).
import {
	BulkSimRequest,
	BulkStatConstraint,
	BulkStatConstraintOp,
	ComputeStatsRequest,
	ComputeStatsResult,
	Player,
	Raid,
	RaidSimRequest,
} from '@generated/proto/api';
import { Class, ConsumesSpec, Debuffs, EquipmentSpec, ItemSlot, PseudoStat, Stat, TristateEffect, UnitStats } from '@generated/proto/common';
import { describe, expect, it, vi } from 'vitest';

import { ADAMANTITE_SHARPENING_STONE_ID, adjustWeaponImbueId } from '../../proto/utils';
import type { SimSignals } from '../../sim_signal_manager';
import type { WorkerPool } from '../../workers/worker_pool';
import { filterBulkSimCandidatesByConstraints } from './constraints';

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

const RAW_SPELL_CRIT = 20;
const RAW_ATTACK_POWER = 1000;
const OWN_AGILITY = 700;
const CONFIGURED_AGILITY = 100;
const signals = { abort: { isTriggered: () => false } } as unknown as SimSignals;

const finalStats = () => {
	const stats = UnitStats.create({ stats: new Array(50).fill(0), pseudoStats: new Array(40).fill(0) });
	stats.pseudoStats[PseudoStat.PseudoStatSpellCritPercent] = RAW_SPELL_CRIT;
	stats.stats[Stat.StatAttackPower] = RAW_ATTACK_POWER;
	stats.stats[Stat.StatAgility] = OWN_AGILITY;
	return stats;
};
const workerPool = {
	getNumWorkers: () => 2,
	computeStats: vi.fn(async () => ComputeStatsResult.create({ raidStats: { parties: [{ players: [{ finalStats: finalStats() }] }] } })),
} as unknown as WorkerPool;

const filter = (player: Partial<Player>, debuffs: Partial<Debuffs>, constraint: Partial<BulkStatConstraint>) =>
	filterBulkSimCandidatesByConstraints(
		BulkSimRequest.create({
			baseRequest: RaidSimRequest.create({
				raid: Raid.create({ parties: [{ players: [Player.create(player)] }], debuffs: Debuffs.create(debuffs) }),
			}),
			bulkSettings: { statConstraints: [BulkStatConstraint.create({ op: BulkStatConstraintOp.BulkStatConstraintOpGreaterThanOrEqual, ...constraint })] },
		}),
		[{ index: 0, gear: EquipmentSpec.create() }],
		workerPool,
		vi.fn(),
		signals,
	);

const check = (value: number) =>
	filter(
		{},
		{ improvedSealOfTheCrusader: TristateEffect.TristateEffectImproved },
		{ unitStat: { oneofKind: 'pseudoStat', pseudoStat: PseudoStat.PseudoStatSpellCritPercent }, value },
	);

// The Survival tree's 21st talent is Expose Weakness (proto field 62 = 21 + 20 + 21).
const EXPOSE_WEAKNESS_TALENTS = '--' + '0'.repeat(20) + '1';
const attackPowerCheck = (player: Partial<Player>, value: number) =>
	filter(
		player,
		{ exposeWeaknessUptime: 1, exposeWeaknessHunterAgility: CONFIGURED_AGILITY },
		{ unitStat: { oneofKind: 'stat', stat: Stat.StatAttackPower }, value },
	);

describe('filterBulkSimCandidatesByConstraints', () => {
	it('counts the crit Improved Seal of the Crusader adds, as the stats panel does', async () => {
		expect((await check(RAW_SPELL_CRIT + 3)).skipped).toBe(0);
	});

	it('still drops a candidate below what the panel shows', async () => {
		expect((await check(RAW_SPELL_CRIT + 3.5)).skipped).toBe(1);
	});

	// The panel credits a hunter who applies Expose Weakness with their own talent their own agility.
	it("credits a talented hunter's own agility for Expose Weakness, as the panel does", async () => {
		const hunter = { class: Class.ClassHunter, talentsString: EXPOSE_WEAKNESS_TALENTS };
		expect((await attackPowerCheck(hunter, RAW_ATTACK_POWER + OWN_AGILITY * 0.25)).skipped).toBe(0);
		expect((await attackPowerCheck(hunter, RAW_ATTACK_POWER + OWN_AGILITY * 0.25 + 1)).skipped).toBe(1);
	});

	it('credits the configured agility to anyone else', async () => {
		const untalentedHunter = { class: Class.ClassHunter, talentsString: '' };
		const mage = { class: Class.ClassMage, talentsString: EXPOSE_WEAKNESS_TALENTS };
		for (const player of [untalentedHunter, mage]) {
			expect((await attackPowerCheck(player, RAW_ATTACK_POWER + CONFIGURED_AGILITY * 0.25)).skipped).toBe(0);
			expect((await attackPowerCheck(player, RAW_ATTACK_POWER + CONFIGURED_AGILITY * 0.25 + 1)).skipped).toBe(1);
		}
	});

	// The check reads only the final stats, so it asks for them without the rotation: building it,
	// and the prepull run that goes with it, changes no stat and would be paid for every gear set.
	it('computes the stats without building the rotation', async () => {
		vi.mocked(workerPool.computeStats).mockClear();
		await check(0);
		expect(workerPool.computeStats).toHaveBeenCalledTimes(1);
		expect(vi.mocked(workerPool.computeStats).mock.calls[0][0].skipRotation).toBe(true);
	});

	// Each candidate's weapon stone is derived from the base character's, not from the candidate
	// checked before it: a stone dropped for a shield must come back for the next dual-wield set.
	it("adjusts each candidate's weapon stone from the base character's", async () => {
		const offHands: number[] = [];
		const pool = {
			getNumWorkers: () => 1, // One at a time, so the shield candidate is checked first.
			computeStats: vi.fn(async (request: ComputeStatsRequest) => {
				offHands.push(request.raid!.parties[0].players[0].consumables!.ohImbueId);
				return ComputeStatsResult.create({ raidStats: { parties: [{ players: [{ finalStats: finalStats() }] }] } });
			}),
		} as unknown as WorkerPool;
		const gearWithOffHand = (id: number) => {
			const gear = EquipmentSpec.create({ items: Array.from({ length: ItemSlot.ItemSlotRanged + 1 }, () => ({})) });
			gear.items[ItemSlot.ItemSlotOffHand].id = id;
			return gear;
		};

		await filterBulkSimCandidatesByConstraints(
			BulkSimRequest.create({
				baseRequest: RaidSimRequest.create({
					raid: Raid.create({
						parties: [{ players: [Player.create({ consumables: ConsumesSpec.create({ ohImbueId: ADAMANTITE_SHARPENING_STONE_ID }) })] }],
					}),
				}),
				bulkSettings: { statConstraints: [BulkStatConstraint.create({ unitStat: { oneofKind: 'stat', stat: Stat.StatAttackPower }, value: 0 })] },
			}),
			[
				{ index: 0, gear: gearWithOffHand(200) },
				{ index: 1, gear: gearWithOffHand(SWORD) },
			],
			pool,
			vi.fn(),
			signals,
		);

		expect(offHands).toEqual([0, ADAMANTITE_SHARPENING_STONE_ID]);
	});
});
