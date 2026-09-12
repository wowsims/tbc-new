import { BulkStatConstraint, BulkStatConstraintOp } from '../../../proto/api';
import { Stat, UnitStats } from '../../../proto/common';

// Pure batch-sim stat constraint logic. Kept free of UI and i18n imports so it
// can be unit tested under Node; the picker UI lives in bulk_stat_constraints.tsx.

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

// Runs tasks with at most `concurrency` in flight, settling each in place.
// (core/utils has promisePool, but importing it drags in browser globals.)
const runWithConcurrency = async <R>(tasks: Array<() => Promise<R>>, concurrency: number): Promise<PromiseSettledResult<R>[]> => {
	const results: PromiseSettledResult<R>[] = new Array(tasks.length);
	let nextIdx = 0;
	const worker = async () => {
		while (nextIdx < tasks.length) {
			const idx = nextIdx++;
			try {
				results[idx] = { status: 'fulfilled', value: await tasks[idx]() };
			} catch (reason) {
				results[idx] = { status: 'rejected', reason };
			}
		}
	};
	await Promise.all(Array.from({ length: Math.max(1, Math.min(concurrency, tasks.length)) }, worker));
	return results;
};

export interface StatConstraintFilterOptions {
	concurrency: number;
	onProgress?: (checked: number, total: number) => void;
}

export interface StatConstraintFilterResult<T> {
	// Candidates that satisfy every constraint, in their original order.
	passing: T[];
	skipped: number;
}

// Keeps the candidates whose final stats satisfy every constraint. Final stats
// are fetched through getFinalStats with the given concurrency; the first
// fetch failure is rethrown. With no constraints, nothing is fetched.
export const filterByStatConstraints = async <T>(
	candidates: T[],
	constraints: BulkStatConstraint[],
	getFinalStats: (candidate: T) => Promise<UnitStats>,
	options: StatConstraintFilterOptions,
): Promise<StatConstraintFilterResult<T>> => {
	if (constraints.length === 0 || candidates.length === 0) {
		return { passing: candidates.slice(), skipped: 0 };
	}

	let checked = 0;
	options.onProgress?.(checked, candidates.length);

	const tasks = candidates.map(candidate => async () => {
		const finalStats = await getFinalStats(candidate);
		checked += 1;
		options.onProgress?.(checked, candidates.length);
		return finalStatsPassConstraints(constraints, finalStats);
	});

	const settled = await runWithConcurrency(tasks, options.concurrency);
	const rejected = settled.find(result => result.status === 'rejected');
	if (rejected && rejected.status === 'rejected') {
		throw rejected.reason;
	}

	const passing = candidates.filter((_, idx) => {
		const result = settled[idx];
		return result.status === 'fulfilled' && result.value;
	});
	return { passing, skipped: candidates.length - passing.length };
};
