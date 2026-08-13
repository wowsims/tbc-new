---
name: wowsims-tbc-bulk-sim-handoff
description: 'Use when continuing, debugging, validating, or modifying WoWSims TBC Bulk Sim candidate generation, bulk settings, and bulk/reforge integration points.'
argument-hint: 'Describe the TBC Bulk Sim candidate flow, settings, or integration issue to continue.'
---

# WoWSims TBC Bulk Sim Handoff

## Scope

- Bulk candidate generation and staged simulation flow.
- Bulk reforge pre-pass integration and cache behavior.
- Progress, abort semantics, and local vs WASM orchestration.

## Architecture

- Shared messages: `proto/api.proto` (`BulkSimRequest`, `BulkSimResult`, `BulkSettings`).
- Candidate generation: `sim/bulk/candidates.go`.
- Local/server reforge pre-pass wrapper: `sim/web/bulk.go`.
- Web endpoint wiring: `sim/web/main.go`.
- Frontend orchestration and cache partitioning: `ui/core/sim.ts`.
- Bulk utilities and cache helpers: `ui/core/components/individual_sim_ui/bulk/utils.ts`.
- Generic cache storage: `ui/core/reforge_cache.ts`.
- Browser WASM path: `ui/core/wasm/bulk_sim.ts`.

## Core Invariants

- Baseline gear source is `base_request.raid.parties[0].players[0].equipment`.
- Candidate identity remains stable through `BulkGearCandidate.index`.
- With `reforge_request` enabled:
	- Cache hits go to `optimized_candidates`.
	- Work-to-optimize goes to `candidates`.
- Before staged sim starts:
	- Merge cache hits with newly optimized candidates.
	- Clear `request.ReforgeRequest`.
- Dedup for sim input must exclude baseline-equivalent gear and duplicate optimized gear.
- Keep full optimized candidate outputs for cache writing so every input key can be persisted.

## Settings Boundaries

- `BulkSettings` controls bulk-tab constraints.
- `ReforgeSettings` controls Suggest Reforges constraints.
- Never merge or alias these domains in request shaping or serialization.

## Local/Server Reforge Flow

- Candidate generation runs unless request is fully cache-restored.
- Reforge progress emits `BulkSimStageReforge` before low/medium/high stages.
- Per-candidate reforge failure falls back to original candidate gear.
- Abort returns partial optimized candidates that already completed.

## Frontend/WASM Flow

- WASM reforge is frontend-orchestrated with per-gear optimizer calls.
- Cache entries store optimized output gear values, not metadata-only records.
- Cache key is input identity; output is restored from cached gear value.

## Candidate Counting and Filtering

- `rawCombinations` is the mixed-radix candidate index space.
- `combinations` is the filtered runnable count.
- Required set-bonus matching scans raw combinations, then filters runnable candidates.
- User-visible progress/counts should reflect filtered candidate totals.

## Performance Guardrails

- Avoid per-candidate allocations in hot loops.
- Prefer preallocated imperative loops in candidate/cache helpers.
- Keep reforge candidate-cache lookup path read-friendly and hash reuse-aware.
- Throttle progress emission frequency to reduce lock/contention overhead.

## Logging Expectations

- Candidate generation logs started and completed with duration and counts.
- Reforge stage logs one started event and one completion summary.

## Validation Commands

```bash
make proto
npm run type-check
go test -count=1 ./sim/core/bulk ./sim/web
```

For reforge-integration changes:

```bash
go test -count=1 ./sim/core/reforge_optimizer ./sim/web
```

## Fast Search Aids

```bash
rg -n "BulkSimReforge|reforge_request|optimized_candidates|BulkSimStageReforge" proto sim ui
rg -n "bulkSimAsync|/bulkSimAsync" sim/web
```

## Common Pitfalls

- Conflating `BulkSettings` and `ReforgeSettings`.
- Porting MoP-only assumptions directly into TBC request/proto shapes.
- Writing cache metadata without optimized gear payload.
- Dropping partial optimized candidates on abort paths.
