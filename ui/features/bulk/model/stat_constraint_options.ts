import { BulkStatConstraint, BulkStatConstraintOp } from '@generated/proto/api';
import { Class, PseudoStat, Stat } from '@generated/proto/common';
import { displayStatOrder, UnitStat } from '@sim/proto/stats';

// What the stat constraints picker offers: the selectable stats in character sheet order,
// their abbreviated labels, the operators, and the option-value encoding of a stat.

// The Stat or PseudoStat a constraint applies to. Legacy rows saved before the
// oneof existed carry a bare stat, which the oneof still decodes.
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

// Abbreviated stat labels so the three controls fit on one line in the
// narrow settings sidebar. The full name is shown as the dropdown's tooltip.
const STAT_SHORT_LABELS: Partial<Record<Stat, string>> = {
	[Stat.StatHealth]: 'Health',
	[Stat.StatMana]: 'Mana',
	[Stat.StatArmor]: 'Armor',
	[Stat.StatBonusArmor]: 'Bonus Armor',
	[Stat.StatStamina]: 'Stam',
	[Stat.StatStrength]: 'Str',
	[Stat.StatAgility]: 'Agi',
	[Stat.StatIntellect]: 'Int',
	[Stat.StatSpirit]: 'Spirit',
	[Stat.StatHealingPower]: 'Healing',
	[Stat.StatSpellDamage]: 'Spell Dmg',
	[Stat.StatArcaneDamage]: 'Arcane Dmg',
	[Stat.StatFireDamage]: 'Fire Dmg',
	[Stat.StatFrostDamage]: 'Frost Dmg',
	[Stat.StatHolyDamage]: 'Holy Dmg',
	[Stat.StatNatureDamage]: 'Nature Dmg',
	[Stat.StatShadowDamage]: 'Shadow Dmg',
	[Stat.StatSpellHitRating]: 'Spell Hit',
	[Stat.StatSpellCritRating]: 'Spell Crit',
	[Stat.StatSpellHasteRating]: 'Spell Haste',
	[Stat.StatSpellPenetration]: 'Spell Pen',
	[Stat.StatMP5]: 'MP5',
	[Stat.StatAttackPower]: 'AP',
	[Stat.StatRangedAttackPower]: 'RAP',
	[Stat.StatFeralAttackPower]: 'Feral AP',
	[Stat.StatMeleeHitRating]: 'Hit',
	[Stat.StatMeleeCritRating]: 'Crit',
	[Stat.StatMeleeHasteRating]: 'Haste',
	[Stat.StatArmorPenetration]: 'ArP',
	[Stat.StatExpertiseRating]: 'Expertise',
	[Stat.StatDefenseRating]: 'Defense',
	[Stat.StatBlockRating]: 'Block',
	[Stat.StatBlockValue]: 'Block Val',
	[Stat.StatDodgeRating]: 'Dodge',
	[Stat.StatParryRating]: 'Parry',
	[Stat.StatResilienceRating]: 'Resil',
	[Stat.StatArcaneResistance]: 'Arcane Res',
	[Stat.StatFireResistance]: 'Fire Res',
	[Stat.StatFrostResistance]: 'Frost Res',
	[Stat.StatNatureResistance]: 'Nature Res',
	[Stat.StatShadowResistance]: 'Shadow Res',
	[Stat.StatPhysicalDamage]: 'Phys Dmg',
};

const PSEUDO_STAT_SHORT_LABELS: Partial<Record<PseudoStat, string>> = {
	[PseudoStat.PseudoStatSpellHitPercent]: 'Spell Hit %',
	[PseudoStat.PseudoStatSchoolHitPercentArcane]: 'Arcane Hit %',
	[PseudoStat.PseudoStatSchoolHitPercentFire]: 'Fire Hit %',
	[PseudoStat.PseudoStatSchoolHitPercentFrost]: 'Frost Hit %',
	[PseudoStat.PseudoStatSchoolHitPercentHoly]: 'Holy Hit %',
	[PseudoStat.PseudoStatSchoolHitPercentNature]: 'Nature Hit %',
	[PseudoStat.PseudoStatSchoolHitPercentShadow]: 'Shadow Hit %',
	[PseudoStat.PseudoStatSpellCritPercent]: 'Spell Crit %',
	[PseudoStat.PseudoStatSpellHastePercent]: 'Spell Haste %',
	[PseudoStat.PseudoStatMeleeHitPercent]: 'Hit %',
	[PseudoStat.PseudoStatMeleeCritPercent]: 'Crit %',
	[PseudoStat.PseudoStatMeleeHastePercent]: 'Haste %',
	[PseudoStat.PseudoStatRangedHitPercent]: 'Ranged Hit %',
	[PseudoStat.PseudoStatRangedCritPercent]: 'Ranged Crit %',
	[PseudoStat.PseudoStatRangedHastePercent]: 'Ranged Haste %',
	[PseudoStat.PseudoStatBlockPercent]: 'Block %',
	[PseudoStat.PseudoStatDodgePercent]: 'Dodge %',
	[PseudoStat.PseudoStatParryPercent]: 'Parry %',
	[PseudoStat.PseudoStatReducedCritTakenPercent]: 'Crit Reduction',
};

// Crit reduction is not a character-sheet stat (the sheet shows it as "Crit
// Immunity"), so it is added here explicitly, right after Defense.
const CRIT_REDUCTION = UnitStat.fromPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent);

// Stats offered in the dropdown, in the same order the character sheet uses.
export const SELECTABLE_STATS: UnitStat[] = displayStatOrder.flatMap(unitStat =>
	unitStat.equalsStat(Stat.StatDefenseRating) ? [unitStat, CRIT_REDUCTION] : [unitStat],
);

export const unitStatOptionValue = (unitStat: UnitStat): string => (unitStat.isPseudoStat() ? `p${unitStat.getPseudoStat()}` : `s${unitStat.getStat()}`);
export const unitStatFromOptionValue = (value: string): UnitStat =>
	value.startsWith('p') ? UnitStat.fromPseudoStat(Number(value.slice(1)) as PseudoStat) : UnitStat.fromStat(Number(value.slice(1)) as Stat);

export const STAT_CONSTRAINT_OPS: BulkStatConstraintOp[] = [
	BulkStatConstraintOp.BulkStatConstraintOpGreaterThan,
	BulkStatConstraintOp.BulkStatConstraintOpGreaterThanOrEqual,
	BulkStatConstraintOp.BulkStatConstraintOpEqual,
	BulkStatConstraintOp.BulkStatConstraintOpLessThanOrEqual,
	BulkStatConstraintOp.BulkStatConstraintOpLessThan,
];

// The abbreviated label where there is one, else the character sheet's short name.
export const unitStatShortLabel = (unitStat: UnitStat, playerClass: Class): string =>
	(unitStat.isPseudoStat() ? PSEUDO_STAT_SHORT_LABELS[unitStat.getPseudoStat()] : STAT_SHORT_LABELS[unitStat.getStat()]) ??
	unitStat.getShortName(playerClass);
