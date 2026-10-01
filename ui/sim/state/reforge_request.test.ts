// The gem cache key covers every batch stat constraint. One the gems cannot move still decides
// whether a gear set's solve is infeasible, and an infeasible gear set keeps its own gems instead of
// the optimizer's; the cache keys a gear set without its gems, so only the key can tell the two apart.
import { BulkStatConstraint, BulkStatConstraintOp, ReforgeGemOption, ReforgeOptimizeRequest, ReforgeSettings } from '@generated/proto/api';
import { GemColor, Stat } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { cacheRelevantReforgeRequest } from './reforge_request';

const STATS_LEN = Stat.StatShadowResistance + 1;
const gem = (id: number, statValues: Partial<Record<Stat, number>>) => {
	const values = new Array(STATS_LEN).fill(0);
	for (const [stat, value] of Object.entries(statValues)) values[Number(stat)] = value;
	return ReforgeGemOption.create({ id, color: GemColor.GemColorPrismatic, stats: values });
};
const VOID_SPHERE = gem(22459, {
	[Stat.StatArcaneResistance]: 4,
	[Stat.StatFireResistance]: 4,
	[Stat.StatFrostResistance]: 4,
	[Stat.StatNatureResistance]: 4,
	[Stat.StatShadowResistance]: 4,
});
const SOLID_STAR = gem(24033, { [Stat.StatStamina]: 12 });

const atLeast = (stat: Stat, value: number) =>
	BulkStatConstraint.create({ unitStat: { oneofKind: 'stat', stat }, op: BulkStatConstraintOp.BulkStatConstraintOpGreaterThanOrEqual, value });

const key = (constraint: BulkStatConstraint, { epStats = [Stat.StatStamina], gemOptions = [VOID_SPHERE, SOLID_STAR] } = {}) =>
	ReforgeOptimizeRequest.toJsonString(
		cacheRelevantReforgeRequest(
			ReforgeOptimizeRequest.create({ settings: ReforgeSettings.create({ epStats }), gemOptions, statConstraints: [constraint] }),
		),
	);

describe('cacheRelevantReforgeRequest', () => {
	it('keeps a resistance constraint the spec does not gem for', () => {
		expect(key(atLeast(Stat.StatFireResistance, 175))).not.toBe(key(atLeast(Stat.StatFireResistance, 180)));
	});

	it('keeps a resistance constraint no gem in the pool carries', () => {
		const options = { epStats: [Stat.StatStamina, Stat.StatFireResistance], gemOptions: [SOLID_STAR] };
		expect(key(atLeast(Stat.StatFireResistance, 175), options)).not.toBe(key(atLeast(Stat.StatFireResistance, 180), options));
	});

	it('tells a batch with a resistance constraint from the same batch without one', () => {
		const constrained = key(atLeast(Stat.StatShadowResistance, 200));
		const unconstrained = ReforgeOptimizeRequest.toJsonString(
			cacheRelevantReforgeRequest(
				ReforgeOptimizeRequest.create({ settings: ReforgeSettings.create({ epStats: [Stat.StatStamina] }), gemOptions: [VOID_SPHERE, SOLID_STAR] }),
			),
		);
		expect(constrained).not.toBe(unconstrained);
	});

	it('keeps a resistance constraint a gem can meet', () => {
		const options = { epStats: [Stat.StatStamina, Stat.StatFireResistance] };
		expect(key(atLeast(Stat.StatFireResistance, 175), options)).not.toBe(key(atLeast(Stat.StatFireResistance, 180), options));
	});

	it('keeps every other constraint', () => {
		expect(key(atLeast(Stat.StatStamina, 500))).not.toBe(key(atLeast(Stat.StatStamina, 600)));
	});
});
