import { BulkStatConstraint, BulkStatConstraintOp } from '@generated/proto/api';
import { Stat, UnitStats } from '@generated/proto/common';

// Pure batch-sim stat constraint logic. Kept free of UI and i18n imports so it
// is shared by the wasm batch pipeline and the picker (features/bulk/components/BulkStatConstraints).

export const newStatConstraint = (): BulkStatConstraint =>
	BulkStatConstraint.create({
		unitStat: { oneofKind: 'stat', stat: Stat.StatStamina },
		op: BulkStatConstraintOp.BulkStatConstraintOpGreaterThanOrEqual,
		value: 0,
	});

// Evaluates a single constraint against a stat value.
export const statConstraintPasses = (constraint: BulkStatConstraint, statValue: number): boolean => {
	switch (constraint.op) {
		case BulkStatConstraintOp.BulkStatConstraintOpGreaterThan:
			return statValue > constraint.value;
		case BulkStatConstraintOp.BulkStatConstraintOpGreaterThanOrEqual:
			return statValue >= constraint.value;
		case BulkStatConstraintOp.BulkStatConstraintOpLessThanOrEqual:
			return statValue <= constraint.value;
		case BulkStatConstraintOp.BulkStatConstraintOpLessThan:
			return statValue < constraint.value;
	}
};

// Reads the constrained Stat or PseudoStat out of a final-stats proto, as
// returned by the server's compute-stats API.
export const constraintStatValue = (constraint: BulkStatConstraint, finalStats: UnitStats): number => {
	switch (constraint.unitStat.oneofKind) {
		case 'stat':
			return finalStats.stats[constraint.unitStat.stat] ?? 0;
		case 'pseudoStat':
			return finalStats.pseudoStats[constraint.unitStat.pseudoStat] ?? 0;
		default:
			return 0;
	}
};

// A constraint with no stat set (not producible from the UI) is skipped, as on the server and in
// the gem optimizer.
export const finalStatsPassConstraints = (constraints: BulkStatConstraint[], finalStats: UnitStats): boolean =>
	constraints.every(
		constraint => constraint.unitStat.oneofKind === undefined || statConstraintPasses(constraint, constraintStatValue(constraint, finalStats)),
	);
