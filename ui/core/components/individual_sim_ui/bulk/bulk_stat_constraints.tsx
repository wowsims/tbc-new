import { ref } from 'tsx-vanilla';

import i18n from '../../../../i18n/config';
import { IndividualSimUI } from '../../../individual_sim_ui';
import { BulkStatConstraint, BulkStatConstraintOp } from '../../../proto/api';
import { PseudoStat, Stat } from '../../../proto/common';
import { displayStatOrder, UnitStat } from '../../../proto_utils/stats';
import { TypedEvent } from '../../../typed_event';
import { Component } from '../../component';
import { BulkTab } from '../bulk_tab';
import { newStatConstraint, STAT_CONSTRAINT_OP_SYMBOLS } from './stat_constraints';

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
// TODO: localize these if the feature graduates from a prototype.
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
const SELECTABLE_STATS: UnitStat[] = displayStatOrder.flatMap(unitStat =>
	unitStat.equalsStat(Stat.StatDefenseRating) ? [unitStat, CRIT_REDUCTION] : [unitStat],
);

const unitStatOptionValue = (unitStat: UnitStat): string => (unitStat.isPseudoStat() ? `p${unitStat.getPseudoStat()}` : `s${unitStat.getStat()}`);
const unitStatFromOptionValue = (value: string): UnitStat =>
	value.startsWith('p') ? UnitStat.fromPseudoStat(Number(value.slice(1)) as PseudoStat) : UnitStat.fromStat(Number(value.slice(1)) as Stat);

const OPS: BulkStatConstraintOp[] = [
	BulkStatConstraintOp.BulkStatConstraintOpGreaterThan,
	BulkStatConstraintOp.BulkStatConstraintOpGreaterThanOrEqual,
	BulkStatConstraintOp.BulkStatConstraintOpEqual,
	BulkStatConstraintOp.BulkStatConstraintOpLessThanOrEqual,
	BulkStatConstraintOp.BulkStatConstraintOpLessThan,
];

interface ConstraintRow {
	elem: HTMLElement;
	statSelect: HTMLSelectElement;
	opSelect: HTMLSelectElement;
	valueInput: HTMLInputElement;
}

// Table of stat constraints for the batch sim settings panel. Each row is a
// stat dropdown, an operator dropdown and a numeric threshold. The picker owns
// no state: it reads from and writes to the BulkTab, and re-renders when the
// tab's settings change.
export default class BulkStatConstraintsPicker extends Component {
	private readonly simUI: IndividualSimUI<any>;
	private readonly bulkTab: BulkTab;
	private readonly listElem: HTMLElement;
	private readonly emptyNotice: HTMLElement;
	private rows: ConstraintRow[] = [];

	constructor(parent: HTMLElement, simUI: IndividualSimUI<any>, bulkTab: BulkTab) {
		super(parent, 'bulk-stat-constraints');
		this.simUI = simUI;
		this.bulkTab = bulkTab;

		const listRef = ref<HTMLDivElement>();
		const emptyNoticeRef = ref<HTMLDivElement>();
		const addBtnRef = ref<HTMLButtonElement>();

		this.rootElem.appendChild(
			<>
				<h6 className="mb-2">{i18n.t('bulk_tab.settings.stat_constraints.label')}</h6>
				<div className="fs-content mb-2">{i18n.t('bulk_tab.settings.stat_constraints.tooltip')}</div>
				<div className="fs-content text-muted mb-2" ref={emptyNoticeRef}>
					{i18n.t('bulk_tab.settings.stat_constraints.empty')}
				</div>
				<div className="bulk-stat-constraints__list" ref={listRef} />
				<button className="btn btn-sm btn-outline-primary bulk-stat-constraints__add" ref={addBtnRef}>
					<i className="fas fa-plus me-1" />
					{i18n.t('bulk_tab.settings.stat_constraints.add')}
				</button>
			</>,
		);

		this.listElem = listRef.value!;
		this.emptyNotice = emptyNoticeRef.value!;

		addBtnRef.value!.addEventListener('click', () => {
			this.bulkTab.setStatConstraints([...this.bulkTab.statConstraints, newStatConstraint()]);
		});

		const changeEvent = this.bulkTab.settingsChangedEmitter.on(() => this.render());
		this.addOnDisposeCallback(() => changeEvent.dispose());

		this.render();
	}

	// Rows are only rebuilt when the count changes. Edits to an existing row
	// sync values in place, so the control that fired the change (and may
	// still be focused) is never torn out from under the browser mid-event.
	private render() {
		const constraints = this.bulkTab.statConstraints;
		this.emptyNotice.hidden = constraints.length > 0;

		if (this.rows.length !== constraints.length) {
			this.rows = constraints.map((_, idx) => this.buildRow(idx));
			this.listElem.replaceChildren(...this.rows.map(row => row.elem));
		}

		constraints.forEach((constraint, idx) => {
			const row = this.rows[idx];
			row.statSelect.value = unitStatOptionValue(constraintUnitStat(constraint));
			row.statSelect.title = row.statSelect.selectedOptions[0]?.title ?? '';
			row.opSelect.value = String(constraint.op);
			if (document.activeElement !== row.valueInput) {
				row.valueInput.value = String(constraint.value);
			}
		});
	}

	private buildRow(idx: number): ConstraintRow {
		const statSelectRef = ref<HTMLSelectElement>();
		const opSelectRef = ref<HTMLSelectElement>();
		const valueInputRef = ref<HTMLInputElement>();
		const removeBtnRef = ref<HTMLButtonElement>();
		const playerClass = this.simUI.player.getClass();

		const fullStatName = (unitStat: UnitStat) => {
			const name = unitStat.getFullName(playerClass);
			return unitStat.equals(CRIT_REDUCTION) ? `${name}. ${i18n.t('bulk_tab.settings.stat_constraints.crit_reduction_hint')}` : name;
		};
		const shortStatName = (unitStat: UnitStat) =>
			(unitStat.isPseudoStat() ? PSEUDO_STAT_SHORT_LABELS[unitStat.getPseudoStat()] : STAT_SHORT_LABELS[unitStat.getStat()]) ??
			unitStat.getShortName(playerClass);

		const row = (
			<div className="bulk-stat-constraints__row">
				<select className="form-select form-select-sm bulk-stat-constraints__stat" ref={statSelectRef}>
					{SELECTABLE_STATS.map(stat => (
						<option value={unitStatOptionValue(stat)} title={fullStatName(stat)}>
							{shortStatName(stat)}
						</option>
					))}
				</select>
				<select className="form-select form-select-sm bulk-stat-constraints__op" ref={opSelectRef}>
					{OPS.map(op => (
						<option value={String(op)}>{STAT_CONSTRAINT_OP_SYMBOLS[op]}</option>
					))}
				</select>
				<input className="form-control form-control-sm bulk-stat-constraints__value" type="number" step="any" ref={valueInputRef} />
				<button
					className="btn btn-sm btn-link link-danger bulk-stat-constraints__remove"
					ref={removeBtnRef}
					title={i18n.t('bulk_tab.settings.stat_constraints.remove')}>
					<i className="fas fa-times" />
				</button>
			</div>
		) as HTMLElement;

		const update = (patch: Partial<BulkStatConstraint>) => {
			const next = this.bulkTab.statConstraints.slice();
			// A detached row (removed while its input was focused) can still fire.
			if (idx >= next.length) return;
			next[idx] = BulkStatConstraint.create({ ...next[idx], ...patch });
			this.bulkTab.setStatConstraints(next, TypedEvent.nextEventID());
		};

		statSelectRef.value!.addEventListener('change', () => {
			const next = this.bulkTab.statConstraints.slice();
			if (idx >= next.length) return;
			next[idx] = withConstraintUnitStat(next[idx], unitStatFromOptionValue(statSelectRef.value!.value));
			this.bulkTab.setStatConstraints(next, TypedEvent.nextEventID());
		});
		opSelectRef.value!.addEventListener('change', () => update({ op: Number(opSelectRef.value!.value) as BulkStatConstraintOp }));
		valueInputRef.value!.addEventListener('change', () => {
			const parsed = Number(valueInputRef.value!.value);
			update({ value: Number.isFinite(parsed) ? parsed : 0 });
		});
		removeBtnRef.value!.addEventListener('click', () => {
			this.bulkTab.setStatConstraints(this.bulkTab.statConstraints.filter((_, i) => i !== idx));
		});

		return { elem: row, statSelect: statSelectRef.value!, opSelect: opSelectRef.value!, valueInput: valueInputRef.value! };
	}
}
