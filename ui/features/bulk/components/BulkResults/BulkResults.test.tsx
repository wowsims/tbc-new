import i18n from '@i18n/config';
import type { TopGearResult } from '@sim/bulk/types';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { patchBulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore } from '@sim/state/sim_store';
import { render } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import englishTranslations from '../../../../../assets/locales/en/translation.json';

vi.mock('./BulkResultRow', () => ({
	BulkResultRow: ({ iterations }: { iterations: number }) => <div data-testid="row" data-iterations={iterations} />,
}));

const { BulkResults } = await import('./BulkResults');

const STORE_KEY = 9;
const RUN_ITERATIONS = 30000;

const result = (avg: number) => ({ gear: {}, dpsMetrics: { avg, stdev: 1 } }) as unknown as TopGearResult;

const mount = (skippedByConstraints = 0, checkedByConstraints = 0) => {
	const store = createSimStore();
	const player = { sim: { store, getIterations: () => 100 }, storeKey: STORE_KEY } as unknown as Player<any>;
	seedBulkSettings(player);
	const baseline = result(1000);
	patchBulkState(player, {
		started: true,
		results: {
			chains: [[result(1100)], [baseline]],
			originalGearResults: baseline,
			iterations: RUN_ITERATIONS,
			skippedByConstraints,
			checkedByConstraints,
		},
	});
	const host = { player, sim: player.sim } as never;
	return render(
		<SimHostProvider host={host}>
			<BulkResults />
		</SimHostProvider>,
	);
};

describe('BulkResults', () => {
	it('gives every row the iteration count the run was scored with, not the sidebar value', () => {
		const { getAllByTestId } = mount();
		const rows = getAllByTestId('row');

		expect(rows).toHaveLength(2);
		expect(rows.map(row => row.getAttribute('data-iterations'))).toEqual([String(RUN_ITERATIONS), String(RUN_ITERATIONS)]);
	});

	it('says how many combinations the stat constraints skipped, and nothing when none were', () => {
		expect(mount().queryByTestId('bulk-results-constraints-note')).toBeNull();
		expect(mount(3).getByTestId('bulk-results-constraints-note')).not.toBeNull();
	});
});

// With the real English strings, so the note's numbers and wording are checked together.
describe('BulkResults constraints note', () => {
	beforeAll(() => {
		i18n.addResourceBundle('en', 'translation', englishTranslations, true, true);
	});
	afterAll(() => {
		i18n.removeResourceBundle('en', 'translation');
	});

	// Both numbers count the same thing: the gear sets the check examined. Duplicates and the gear
	// already equipped (always shown as Current Gear) are removed before it, so the sidebar's
	// combination count is not the total here.
	it('counts skipped gear sets out of those the check examined', () => {
		expect(mount(2, 5).getByTestId('bulk-results-constraints-note').textContent).toBe('2 of 5 gear sets skipped for failing a stat constraint.');
	});
});
