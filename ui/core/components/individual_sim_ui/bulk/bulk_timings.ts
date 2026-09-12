// Wall-clock accounting for a batch sim run, kept free of UI imports so it can
// be unit tested under Node. Times are milliseconds from performance.now().

export type BulkPhase = 'build' | 'gems' | 'constraints' | 'baselineSim' | 'candidateSims';

export const BULK_PHASES: BulkPhase[] = ['build', 'gems', 'constraints', 'baselineSim', 'candidateSims'];

export interface PhaseTiming {
	ms: number;
	// Items processed in the phase (candidates built, gear sets optimized, ...).
	count: number;
}

export interface BackendCallTiming {
	count: number;
	// Total wall time the UI spent awaiting these calls.
	wallMs: number;
}

export interface SimCallTiming extends BackendCallTiming {
	// Progress payloads received before the final one. Each of these was
	// followed by one poll sleep in the worker before the result was fetched.
	pollsBeforeFinal: number;
}

export interface BulkTimingReport {
	totalMs: number;
	combinations: number;
	iterationsPerSim: number;
	phases: Record<BulkPhase, PhaseTiming>;
	sims: SimCallTiming;
	statComputations: BackendCallTiming;
}

export class BulkSimTimings {
	private readonly now: () => number;
	private startedAt: number | null = null;
	private finishedAt: number | null = null;
	private phaseStartedAt: Partial<Record<BulkPhase, number>> = {};
	private readonly phases: Record<BulkPhase, PhaseTiming> = {
		build: { ms: 0, count: 0 },
		gems: { ms: 0, count: 0 },
		constraints: { ms: 0, count: 0 },
		baselineSim: { ms: 0, count: 0 },
		candidateSims: { ms: 0, count: 0 },
	};
	private readonly sims: SimCallTiming = { count: 0, wallMs: 0, pollsBeforeFinal: 0 };
	private readonly statComputations: BackendCallTiming = { count: 0, wallMs: 0 };
	combinations = 0;
	iterationsPerSim = 0;

	constructor(now: () => number = () => performance.now()) {
		this.now = now;
	}

	start() {
		this.startedAt = this.now();
	}

	finish() {
		this.finishedAt = this.now();
	}

	startPhase(phase: BulkPhase) {
		this.phaseStartedAt[phase] = this.now();
	}

	// Ends a phase, adding its elapsed time and item count. A phase may be
	// started and ended more than once; the totals accumulate.
	endPhase(phase: BulkPhase, count = 0) {
		const startedAt = this.phaseStartedAt[phase];
		if (startedAt === undefined) return;
		this.phases[phase].ms += this.now() - startedAt;
		this.phases[phase].count += count;
		delete this.phaseStartedAt[phase];
	}

	// Times one sim request. The callback receives a function to call once per
	// progress payload, which is how the poll count is derived.
	async timeSim<T>(run: (onProgressPayload: () => void) => Promise<T>): Promise<T> {
		const startedAt = this.now();
		let payloads = 0;
		try {
			return await run(() => {
				payloads += 1;
			});
		} finally {
			this.sims.count += 1;
			this.sims.wallMs += this.now() - startedAt;
			this.sims.pollsBeforeFinal += Math.max(0, payloads - 1);
		}
	}

	async timeStatComputation<T>(run: () => Promise<T>): Promise<T> {
		const startedAt = this.now();
		try {
			return await run();
		} finally {
			this.statComputations.count += 1;
			this.statComputations.wallMs += this.now() - startedAt;
		}
	}

	report(): BulkTimingReport {
		const end = this.finishedAt ?? this.now();
		return {
			totalMs: this.startedAt === null ? 0 : end - this.startedAt,
			combinations: this.combinations,
			iterationsPerSim: this.iterationsPerSim,
			phases: Object.fromEntries(BULK_PHASES.map(phase => [phase, { ...this.phases[phase] }])) as Record<BulkPhase, PhaseTiming>,
			sims: { ...this.sims },
			statComputations: { ...this.statComputations },
		};
	}
}

export const formatDurationMs = (ms: number): string => {
	if (ms < 1000) return `${Math.round(ms)} ms`;
	if (ms < 60_000) return `${(ms / 1000).toFixed(2)} s`;
	const minutes = Math.floor(ms / 60_000);
	const seconds = (ms - minutes * 60_000) / 1000;
	return `${minutes} min ${seconds.toFixed(1)} s`;
};
