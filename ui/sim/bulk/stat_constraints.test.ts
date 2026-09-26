import { BulkSettings, BulkStatConstraint, BulkStatConstraintOp } from '@generated/proto/api';
import { PseudoStat, Stat, UnitStats } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { constraintStatValue, finalStatsPassConstraints, newStatConstraint, statConstraintPasses } from './stat_constraints';

const Op = BulkStatConstraintOp;

const statConstraint = (stat: Stat, op: BulkStatConstraintOp, value: number) => BulkStatConstraint.create({ unitStat: { oneofKind: 'stat', stat }, op, value });

const pseudoStatConstraint = (pseudoStat: PseudoStat, op: BulkStatConstraintOp, value: number) =>
	BulkStatConstraint.create({ unitStat: { oneofKind: 'pseudoStat', pseudoStat }, op, value });

const finalStats = (stats: Partial<Record<Stat, number>> = {}, pseudoStats: Partial<Record<PseudoStat, number>> = {}): UnitStats => {
	const proto = UnitStats.create({
		stats: new Array(Stat.StatPhysicalDamage + 1).fill(0),
		pseudoStats: new Array(PseudoStat.PseudoStatReducedCritTakenPercent + 1).fill(0),
	});
	for (const [stat, value] of Object.entries(stats)) proto.stats[Number(stat)] = value;
	for (const [pseudoStat, value] of Object.entries(pseudoStats)) proto.pseudoStats[Number(pseudoStat)] = value;
	return proto;
};

describe('statConstraintPasses', () => {
	const cases: Array<[BulkStatConstraintOp, number, boolean]> = [
		[Op.BulkStatConstraintOpGreaterThan, 176, true],
		[Op.BulkStatConstraintOpGreaterThan, 175, false],
		[Op.BulkStatConstraintOpGreaterThanOrEqual, 175, true],
		[Op.BulkStatConstraintOpGreaterThanOrEqual, 174, false],
		[Op.BulkStatConstraintOpEqual, 175, true],
		[Op.BulkStatConstraintOpEqual, 174, false],
		[Op.BulkStatConstraintOpLessThanOrEqual, 175, true],
		[Op.BulkStatConstraintOpLessThanOrEqual, 176, false],
		[Op.BulkStatConstraintOpLessThan, 174, true],
		[Op.BulkStatConstraintOpLessThan, 175, false],
	];

	for (const [op, value, expected] of cases) {
		it(`${Op[op]} with threshold 175 and value ${value} -> ${expected}`, () => {
			expect(statConstraintPasses(statConstraint(Stat.StatFireResistance, op, 175), value)).toBe(expected);
		});
	}
});

describe('constraintStatValue', () => {
	it('reads a Stat from the stats array', () => {
		const constraint = statConstraint(Stat.StatFireResistance, Op.BulkStatConstraintOpGreaterThan, 0);
		expect(constraintStatValue(constraint, finalStats({ [Stat.StatFireResistance]: 93 }))).toBe(93);
	});

	it('reads a PseudoStat from the pseudoStats array', () => {
		const constraint = pseudoStatConstraint(PseudoStat.PseudoStatReducedCritTakenPercent, Op.BulkStatConstraintOpGreaterThanOrEqual, 5.6);
		expect(constraintStatValue(constraint, finalStats({}, { [PseudoStat.PseudoStatReducedCritTakenPercent]: 5.98 }))).toBe(5.98);
	});

	it('reads 0 for a missing value or an unset target', () => {
		expect(constraintStatValue(statConstraint(Stat.StatFireResistance, Op.BulkStatConstraintOpGreaterThan, 0), UnitStats.create())).toBe(0);
		expect(constraintStatValue(BulkStatConstraint.create(), finalStats({ [Stat.StatStamina]: 100 }))).toBe(0);
	});

	it('decodes rows saved before the target became a oneof', () => {
		const settings = BulkSettings.fromJsonString('{"statConstraints":[{"stat":37,"value":85}]}', { ignoreUnknownFields: true });
		const [constraint] = settings.statConstraints;
		expect(constraint.unitStat).toEqual({ oneofKind: 'stat', stat: Stat.StatFireResistance });
		expect(constraint.op).toBe(Op.BulkStatConstraintOpGreaterThan);
		expect(constraintStatValue(constraint, finalStats({ [Stat.StatFireResistance]: 93 }))).toBe(93);
	});

	it('round-trips a PseudoStat row through JSON', () => {
		const original = pseudoStatConstraint(PseudoStat.PseudoStatReducedCritTakenPercent, Op.BulkStatConstraintOpGreaterThanOrEqual, 5.6);
		const decoded = BulkStatConstraint.fromJsonString(BulkStatConstraint.toJsonString(original, { enumAsInteger: true }));
		expect(decoded).toEqual(original);
	});
});

describe('finalStatsPassConstraints', () => {
	it('requires every constraint to pass', () => {
		const constraints = [
			statConstraint(Stat.StatFireResistance, Op.BulkStatConstraintOpGreaterThan, 175),
			pseudoStatConstraint(PseudoStat.PseudoStatReducedCritTakenPercent, Op.BulkStatConstraintOpGreaterThanOrEqual, 5.6),
		];
		const critImmune = { [PseudoStat.PseudoStatReducedCritTakenPercent]: 5.6 };
		expect(finalStatsPassConstraints(constraints, finalStats({ [Stat.StatFireResistance]: 200 }, critImmune))).toBe(true);
		expect(finalStatsPassConstraints(constraints, finalStats({ [Stat.StatFireResistance]: 175 }, critImmune))).toBe(false);
		expect(
			finalStatsPassConstraints(constraints, finalStats({ [Stat.StatFireResistance]: 200 }, { [PseudoStat.PseudoStatReducedCritTakenPercent]: 5.2 })),
		).toBe(false);
	});

	it('passes with no constraints', () => {
		expect(finalStatsPassConstraints([], UnitStats.create())).toBe(true);
	});

	it('creates a new constraint that any gear set passes', () => {
		expect(finalStatsPassConstraints([newStatConstraint()], UnitStats.create())).toBe(true);
	});
});
