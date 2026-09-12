import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import { BulkSimTimings, formatDurationMs } from './bulk_timings';

// A controllable clock: tick() advances time by the given amount.
const fakeClock = () => {
	let t = 0;
	return { now: () => t, tick: (ms: number) => (t += ms) };
};

describe('BulkSimTimings', () => {
	it('accumulates phase durations and counts across repeated phases', () => {
		const clock = fakeClock();
		const timings = new BulkSimTimings(clock.now);
		timings.start();
		timings.startPhase('build');
		clock.tick(120);
		timings.endPhase('build', 10);
		timings.startPhase('build');
		clock.tick(30);
		timings.endPhase('build', 5);
		clock.tick(1000);
		timings.finish();
		const report = timings.report();
		assert.deepEqual(report.phases.build, { ms: 150, count: 15 });
		assert.deepEqual(report.phases.gems, { ms: 0, count: 0 });
		assert.equal(report.totalMs, 1150);
	});

	it('ignores ending a phase that was never started', () => {
		const timings = new BulkSimTimings(fakeClock().now);
		timings.endPhase('gems', 3);
		assert.deepEqual(timings.report().phases.gems, { ms: 0, count: 0 });
	});

	it('counts progress payloads before the final one', async () => {
		const clock = fakeClock();
		const timings = new BulkSimTimings(clock.now);
		// Three payloads: two intermediate polls, each followed by a sleep, then the final.
		await timings.timeSim(async onPayload => {
			onPayload();
			clock.tick(510);
			onPayload();
			clock.tick(510);
			onPayload();
			return 'done';
		});
		// A sim that finished by its first poll slept for nothing.
		await timings.timeSim(async onPayload => {
			clock.tick(5);
			onPayload();
			return 'done';
		});
		const report = timings.report();
		assert.equal(report.sims.count, 2);
		assert.equal(report.sims.wallMs, 1025);
		assert.equal(report.sims.pollsBeforeFinal, 2);
	});

	it('still records a sim that throws, and rethrows', async () => {
		const clock = fakeClock();
		const timings = new BulkSimTimings(clock.now);
		await assert.rejects(
			timings.timeSim(async () => {
				clock.tick(10);
				throw new Error('Bulk Sim Aborted');
			}),
			/Bulk Sim Aborted/,
		);
		assert.equal(timings.report().sims.count, 1);
		assert.equal(timings.report().sims.wallMs, 10);
	});

	it('totals stat computations', async () => {
		const clock = fakeClock();
		const timings = new BulkSimTimings(clock.now);
		for (const ms of [3, 4, 5]) {
			await timings.timeStatComputation(async () => {
				clock.tick(ms);
			});
		}
		assert.deepEqual(timings.report().statComputations, { count: 3, wallMs: 12 });
	});

	it('returns a snapshot that does not alias internal state', () => {
		const timings = new BulkSimTimings(fakeClock().now);
		const report = timings.report();
		report.phases.build.count = 99;
		assert.equal(timings.report().phases.build.count, 0);
	});
});

describe('formatDurationMs', () => {
	it('picks units by magnitude', () => {
		assert.equal(formatDurationMs(42.4), '42 ms');
		assert.equal(formatDurationMs(1500), '1.50 s');
		assert.equal(formatDurationMs(125_300), '2 min 5.3 s');
	});
});
