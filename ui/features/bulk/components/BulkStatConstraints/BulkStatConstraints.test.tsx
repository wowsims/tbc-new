import { BulkStatConstraint, BulkStatConstraintOp } from '@generated/proto/api';
import { Class, PseudoStat, Stat } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { bulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore } from '@sim/state/sim_store';
import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { setBulkStatConstraints } from '../../model/settings';
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
