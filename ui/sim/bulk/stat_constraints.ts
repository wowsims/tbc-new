import { BulkStatConstraint, BulkStatConstraintOp } from '@generated/proto/api';
import { Stat, UnitStats } from '@generated/proto/common';

// Pure batch-sim stat constraint logic. Kept free of UI and i18n imports so it
// is shared by the wasm batch pipeline and the picker (features/bulk/components/BulkStatConstraints).

export const STAT_CONSTRAINT_OP_SYMBOLS: Record<BulkStatConstraintOp, string> = {
	[BulkStatConstraintOp.BulkStatConstraintOpGreaterThan]: '>',
	[BulkStatConstraintOp.BulkStatConstraintOpGreaterThanOrEqual]: '≥',
	[BulkStatConstraintOp.BulkStatConstraintOpEqual]: '=',
	[BulkStatConstraintOp.BulkStatConstraintOpLessThanOrEqual]: '≤',
	[BulkStatConstraintOp.BulkStatConstraintOpLessThan]: '<',
};

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
		case BulkStatConstraintOp.BulkStatConstraintOpEqual:
			return statValue === constraint.value;
		case BulkStatConstraintOp.BulkStatConstraintOpLessThanOrEqual:
			return statValue <= constraint.value;
		case BulkStatConstraintOp.BulkStatConstraintOpLessThan:
			return statValue < constraint.value;
	}
};

// Reads the constrained Stat or PseudoStat out of a final-stats proto, as
// returned by the server's compute-stats API. A constraint with no target
// (not producible from the UI) reads as 0.
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

export const finalStatsPassConstraints = (constraints: BulkStatConstraint[], finalStats: UnitStats): boolean =>
	constraints.every(constraint => statConstraintPasses(constraint, constraintStatValue(constraint, finalStats)));
