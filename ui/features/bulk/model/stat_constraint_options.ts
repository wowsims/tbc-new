import { BulkStatConstraint, BulkStatConstraintOp } from '@generated/proto/api';
import { Class, PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { displayStatOrder, UnitStat } from '@sim/proto/stats';

// What the stat constraints picker offers: the selectable stats in character sheet order, the
// operators, and the option-value encoding of a stat. Stats and operators are labelled with the
// translated names the rest of the UI uses (see statConstraintLabel and statConstraintOpLabel).

// The Stat or PseudoStat a constraint applies to. A row with no stat set, which the picker never
// produces, is shown as the default stat; the checks and the gem optimizer skip it.
export const constraintUnitStat = (constraint: BulkStatConstraint): UnitStat => {
	if (constraint.unitStat.oneofKind === 'pseudoStat') {
		return UnitStat.fromPseudoStat(constraint.unitStat.pseudoStat);
	}
	return UnitStat.fromStat(constraint.unitStat.oneofKind === 'stat' ? constraint.unitStat.stat : Stat.StatStamina);
};

export const withConstraintUnitStat = (constraint: BulkStatConstraint, unitStat: UnitStat): BulkStatConstraint => {
	const next = BulkStatConstraint.clone(constraint);
	next.unitStat = unitStat.isPseudoStat()
		? { oneofKind: 'pseudoStat', pseudoStat: unitStat.getPseudoStat() }
		: { oneofKind: 'stat', stat: unitStat.getStat() };
	return next;
};

// Crit reduction is not a character-sheet stat (the sheet shows it as "Crit
// Immunity"), so it is added here explicitly, right after Defense.
const CRIT_REDUCTION = UnitStat.fromPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent);

// Stats offered in the dropdown, in the same order the character sheet uses.
export const SELECTABLE_STATS: UnitStat[] = displayStatOrder.flatMap(unitStat =>
	unitStat.equalsStat(Stat.StatDefenseRating) ? [unitStat, CRIT_REDUCTION] : [unitStat],
);

// A stat's label in the picker: the short name the rest of the UI uses, except Defense. The stats
// panel shows Defense as "rating (skill)", so its short name would not say which of the two numbers
// a constraint compares; the full name, "Defense Rating", does.
export const statConstraintLabel = (unitStat: UnitStat, playerClass: Class): string =>
	unitStat.equalsStat(Stat.StatDefenseRating) ? unitStat.getFullName(playerClass) : unitStat.getShortName(playerClass);

export const unitStatOptionValue = (unitStat: UnitStat): string => (unitStat.isPseudoStat() ? `p${unitStat.getPseudoStat()}` : `s${unitStat.getStat()}`);
export const unitStatFromOptionValue = (value: string): UnitStat =>
	value.startsWith('p') ? UnitStat.fromPseudoStat(Number(value.slice(1)) as PseudoStat) : UnitStat.fromStat(Number(value.slice(1)) as Stat);

export const STAT_CONSTRAINT_OPS: BulkStatConstraintOp[] = [
	BulkStatConstraintOp.BulkStatConstraintOpGreaterThan,
	BulkStatConstraintOp.BulkStatConstraintOpGreaterThanOrEqual,
	BulkStatConstraintOp.BulkStatConstraintOpLessThanOrEqual,
	BulkStatConstraintOp.BulkStatConstraintOpLessThan,
];

// Labelled like the rotation editor's comparison operators (COMPARISON_OPERATORS in
// ui/features/apl/model/field_specs.ts), so both read alike.
const STAT_CONSTRAINT_OP_I18N_KEYS: Record<BulkStatConstraintOp, string> = {
	[BulkStatConstraintOp.BulkStatConstraintOpGreaterThan]: 'greater_than',
	[BulkStatConstraintOp.BulkStatConstraintOpGreaterThanOrEqual]: 'greater_than_or_equal',
	[BulkStatConstraintOp.BulkStatConstraintOpLessThanOrEqual]: 'less_than_or_equal',
	[BulkStatConstraintOp.BulkStatConstraintOpLessThan]: 'less_than',
};

export const statConstraintOpLabel = (op: BulkStatConstraintOp): string => i18n.t(`rotation_tab.apl.operators.${STAT_CONSTRAINT_OP_I18N_KEYS[op]}`);
