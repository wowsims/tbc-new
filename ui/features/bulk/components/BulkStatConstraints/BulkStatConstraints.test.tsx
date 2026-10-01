import { BulkStatConstraint, BulkStatConstraintOp } from '@generated/proto/api';
import { Class, PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { bulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore } from '@sim/state/sim_store';
import { fireEvent, render } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import englishTranslations from '../../../../../assets/locales/en/translation.json';
import { setBulkStatConstraints } from '../../model/settings';
import { SELECTABLE_STATS } from '../../model/stat_constraint_options';
import { BulkStatConstraints } from './BulkStatConstraints';

const STORE_KEY = 5;

const mount = () => {
	const store = createSimStore();
	const player = { sim: { store }, storeKey: STORE_KEY, getClass: () => Class.ClassWarrior } as unknown as Player<any>;
	seedBulkSettings(player);
	const host = { player, sim: player.sim } as never;
	const { container, getByTestId } = render(
		<SimHostProvider host={host}>
			<BulkStatConstraints />
		</SimHostProvider>,
	);
	const rows = () => Array.from(container.querySelectorAll('[data-testid="bulk-stat-constraints-row"]'));
	const controls = (row: Element) => ({
		stat: row.querySelectorAll('select')[0],
		op: row.querySelectorAll('select')[1],
		value: row.querySelector('input')!,
		remove: row.querySelector('button')!,
	});
	return { player, rows, controls, add: () => fireEvent.click(getByTestId('bulk-stat-constraints-add')) };
};

const settingsVersion = (player: Player<any>) => bulkState(player).v.settings;

describe('BulkStatConstraints', () => {
	it('adds a row, edits each of its three controls, and removes it', () => {
		const { player, rows, controls, add } = mount();
		expect(rows()).toHaveLength(0);

		add();
		expect(rows()).toHaveLength(1);

		fireEvent.change(controls(rows()[0]).stat, { target: { value: `s${Stat.StatFireResistance}` } });
		fireEvent.change(controls(rows()[0]).op, { target: { value: String(BulkStatConstraintOp.BulkStatConstraintOpGreaterThan) } });
		fireEvent.change(controls(rows()[0]).value, { target: { value: '175' } });

		expect(bulkState(player).statConstraints).toEqual([
			BulkStatConstraint.create({
				unitStat: { oneofKind: 'stat', stat: Stat.StatFireResistance },
				op: BulkStatConstraintOp.BulkStatConstraintOpGreaterThan,
				value: 175,
			}),
		]);

		fireEvent.click(controls(rows()[0]).remove);
		expect(rows()).toHaveLength(0);
		expect(bulkState(player).statConstraints).toEqual([]);
	});

	// Labels come from the translated stat names the rest of the UI uses, not a table of our own.
	it('explains a Defense constraint when it is selected', () => {
		const { rows, controls, add } = mount();
		add();

		fireEvent.change(controls(rows()[0]).stat, { target: { value: `s${Stat.StatDefenseRating}` } });

		expect(controls(rows()[0]).stat.title).toBe(i18n.t('bulk_tab.settings.stat_constraints.defense_hint'));
	});

	// Operators read as the rotation editor's comparisons do, from the same translations.
	it("labels the operators with the rotation editor's comparison labels", () => {
		const { rows, controls, add } = mount();
		add();

		const labels = Array.from(controls(rows()[0]).op.options).map(option => option.text);
		expect(labels).toEqual(
			['greater_than', 'greater_than_or_equal', 'less_than_or_equal', 'less_than'].map(key => i18n.t(`rotation_tab.apl.operators.${key}`)),
		);
	});

	it('offers crit reduction, a pseudo stat, and stores it as one', () => {
		const { player, rows, controls, add } = mount();
		add();

		fireEvent.change(controls(rows()[0]).stat, { target: { value: `p${PseudoStat.PseudoStatReducedCritTakenPercent}` } });

		expect(bulkState(player).statConstraints[0].unitStat).toEqual({
			oneofKind: 'pseudoStat',
			pseudoStat: PseudoStat.PseudoStatReducedCritTakenPercent,
		});
	});

	it('commits the value as it is typed, so nothing is left to commit on blur', () => {
		const { player, rows, controls, add } = mount();
		add();

		fireEvent.change(controls(rows()[0]).value, { target: { value: '5.6' } });
		const afterTyping = settingsVersion(player);
		fireEvent.blur(controls(rows()[0]).value);

		expect(bulkState(player).statConstraints[0].value).toBe(5.6);
		expect(settingsVersion(player)).toBe(afterTyping);
	});

	// A number field reports an empty value while it is cleared and while it holds only "-", so an
	// empty value is a draft, not a zero: the field keeps it and the stored value is left alone.
	it('can be cleared to type a new threshold, and reverts if left empty', () => {
		const { player, rows, controls, add } = mount();
		add();
		fireEvent.change(controls(rows()[0]).value, { target: { value: '175' } });

		fireEvent.change(controls(rows()[0]).value, { target: { value: '' } });
		expect(controls(rows()[0]).value.value).toBe('');
		expect(bulkState(player).statConstraints[0].value).toBe(175);

		fireEvent.change(controls(rows()[0]).value, { target: { value: '180' } });
		expect(bulkState(player).statConstraints[0].value).toBe(180);

		fireEvent.change(controls(rows()[0]).value, { target: { value: '' } });
		fireEvent.blur(controls(rows()[0]).value);
		expect(controls(rows()[0]).value.value).toBe('180');
		expect(bulkState(player).statConstraints[0].value).toBe(180);
	});

	it('takes a negative threshold', () => {
		const { player, rows, controls, add } = mount();
		add();

		fireEvent.change(controls(rows()[0]).value, { target: { value: '-5' } });

		expect(bulkState(player).statConstraints[0].value).toBe(-5);
		expect(controls(rows()[0]).value.value).toBe('-5');
	});

	// A row's field follows its constraint when the rows change under it.
	it('shows the right value after an earlier row is removed', () => {
		const { rows, controls, add } = mount();
		add();
		add();
		fireEvent.change(controls(rows()[0]).value, { target: { value: '10' } });
		fireEvent.change(controls(rows()[1]).value, { target: { value: '20' } });

		fireEvent.click(controls(rows()[0]).remove);

		expect(rows()).toHaveLength(1);
		expect(controls(rows()[0]).value.value).toBe('20');
	});

	// Every settings emit refreshes the combination count, which disables Simulate while it loads.
	it('does not emit a settings change when nothing changed', () => {
		const { player, add } = mount();
		add();
		const before = settingsVersion(player);

		setBulkStatConstraints(
			player,
			bulkState(player).statConstraints.map(constraint => BulkStatConstraint.clone(constraint)),
		);

		expect(settingsVersion(player)).toBe(before);
	});
});

// With the real English names, where a short name and a full name actually differ.
describe('BulkStatConstraints labels', () => {
	beforeAll(() => {
		i18n.addResourceBundle('en', 'translation', englishTranslations, true, true);
	});
	afterAll(() => {
		i18n.removeResourceBundle('en', 'translation');
	});

	it('labels every stat with its shared short name, and Defense as a rating', () => {
		const { rows, controls, add } = mount();
		add();

		const labels = Array.from(controls(rows()[0]).stat.options).map(option => option.text);
		expect(labels).toEqual(
			SELECTABLE_STATS.map(unitStat =>
				// The stats panel shows Defense as "rating (skill)", so the short name "Defense" would not
				// say which of the two numbers the constraint compares.
				unitStat.equalsStat(Stat.StatDefenseRating) ? unitStat.getFullName(Class.ClassWarrior) : unitStat.getShortName(Class.ClassWarrior),
			),
		);
		expect(labels).toContain('Defense Rating');
		expect(labels).toContain('Fire Resistance');
	});
});
