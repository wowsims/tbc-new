---
name: wowsims-bulk-sim
description: 'Use when working on WoWSims TBC Bulk Sim: local/server bulk sim, browser WASM concurrent bulk sim, BulkSimRequest/BulkSimResult protos, candidate generation, staged simulation, the finalist tie-breaker stage, deterministic seeds, IndexedDB gem-optimizer caching, tie-group result display, progress, or abort behavior.'
argument-hint: 'Describe the TBC Bulk Sim candidate flow, staging/finalist behavior, cache behavior, or validation task.'
---

# WoWSims TBC Bulk Sim Guide

## Scope
- Bulk candidate generation and staged simulation flow (low/medium/high + finalist).
- Bulk gem-optimizer pre-pass integration and cache behavior ("reforge" naming is inherited from MoP; TBC has no reforging — the optimizer chooses gems/socket bonuses only).
- Deterministic content-derived seeding and statistical result display.
- Progress, abort semantics, and local vs WASM orchestration.

## Architecture
- Shared messages: proto/api.proto (BulkSimRequest, BulkSimResult, BulkGearResult, ReforgeOptimizeRequest). BulkSimStage order: Reforge, Low, Medium, High, Finalist, Complete (Complete last).
- Core staged runner: sim/core/bulk/bulk_sim.go; per-stage logic + finalist stage: sim/core/bulk/stage.go; paired statistics: sim/core/bulk/statistics.go.
- Local/server optimizer pre-pass wrapper: sim/web/bulk.go; endpoint wiring: sim/web/main.go.
- Frontend orchestration, content seed, cache partitioning: ui/core/sim.ts (runBulkSim).
- Bulk utilities and cache helpers: ui/core/components/individual_sim_ui/bulk/utils.ts.
- Results display (tie groups, ±CI, change icons): ui/core/components/individual_sim_ui/bulk_tab.tsx + bulk/bulk_sim_results_renderer.tsx + ui/core/components/gear_change_icon.tsx (gem markers + Wowhead cover link; no reforge glyph in TBC).
- Generic cache storage: ui/core/reforge_cache.ts (sync hashString from ui/core/utils.ts; values are plain `equipmentSpec:`-prefixed proto JSON; age-index-ranged prune).
- Browser WASM path: ui/core/wasm/bulk_sim/ (index, stage, batch, merge, statistics, progress, carry_over, estimate, types) — a deliberate line-for-line mirror of sim/core/bulk.
- Generated constants (never hand-edit; `go run ./tools/database/gen_db -gen=go-to-ts`): ui/core/components/individual_sim_ui/bulk/constants_auto_gen.ts (slot maps) and ui/core/wasm/bulk_sim/constants_auto_gen.ts (tuning constants + stage ladder), emitted by tools/database/gen_bulksim_constants.ts.go from the exported consts in sim/core/bulk.

## TBC-Specific Differences From MoP
- No reforging: the optimizer is gem/socket-bonus only; EquippedItem has no reforge field; no reforge glyph anywhere in the UI.
- No Titan's Grip / dual-wield-2H capability; a two-hander never occupies the off-hand.
- The Ranged slot is a real bulk slot that iterates like any non-weapon slot (bulkSimNonWeaponOrder is built by filtering, not slicing).
- No required-set-bonus combination matcher (matchingCombinations == rawCombinations).
- Single locale (en); no fr translation file.

## Determinism and Statistics
- Content-derived seed: runBulkSim hashes baseline gear + BulkSettings + cache-relevant optimizer config into simOptions.randomSeed (utils hashString; per-part digests combined). Same setup → bit-identical results; any change → fresh sample. An explicit fixed RNG seed takes precedence; lastUsedRngSeed is updated to the seed actually used.
- Candidates share seed sequences; comparisons between candidates use PAIRED errors (bulkSimPairedDpsError over AllValues), far tighter than per-result stdev.
- Finalist stage (runBulkSimFinalistStage / runConcurrentBulkSimFinalistStage): after the high stage, the top `topResults` candidates + baseline get lockstep extra iterations until every adjacent pair separates under a paired z-test at Z95 (bulkSimZ95 / Z_95), or the budget (BulkSimFinalistMaxExtraIterationMultiplier × high-stage iterations) is spent. Its returned results ARE the refined, DPS-sorted display set.
- Shipped per top result: paired_error_to_next_result and paired_error_to_baseline (computed before AllValues are stripped; 0 = could not pair).
- FE display: adjacent results inside the paired tie threshold render in one labeled tie group (unpaired zTest fallback when paired data is absent); each row shows a ±95% CI (stDevToConf95 — the unpaired per-row error, a different quantity from the pairwise grouping test).
- The single significance threshold lives in ui/core/utils.ts (Z_95, zTest) and sim/core/bulk/statistics.go (bulkSimZ95); keep them equal.

## Core Invariants
- Baseline gear source is base_request.raid.parties[0].players[0].equipment.
- Candidate identity remains stable through BulkGearCandidate.index.
- With reforge_request enabled:
  - Cache hits go to optimized_candidates.
  - Work-to-optimize goes to candidates.
- Before staged sim starts:
  - Merge cache hits with newly optimized candidates.
  - Clear request.ReforgeRequest.
- Dedup for sim input must exclude baseline-equivalent gear and duplicate optimized gear.
- Keep full optimized candidate outputs for cache writing so every input key can be persisted.
- Spec lookup goes through core.PlayerProtoToSpecSafe; eligible-slot logic through core.EligibleSlotsForItem / core.ItemTypeToSlotsMap (single source, shared with item swap); entry-point validation via newGeneratorFromRequest.

## Settings Boundaries
- BulkSettings controls bulk-tab constraints.
- ReforgeSettings controls Suggest Gems constraints.
- Never merge or alias these domains in request shaping or serialization.

## Local/Server Optimizer Flow
- Candidate generation runs unless request is fully cache-restored.
- Optimizer progress emits BulkSimStageReforge before low/medium/high stages.
- Per-candidate optimizer failure falls back to original candidate gear.
- Abort returns partial optimized candidates that already completed.

## Frontend/WASM Flow
- WASM optimizer path is frontend-orchestrated with per-gear optimizer calls.
- Cache entries store optimized output gear; cache key is input-identity hash (sync, from utils hashString).
- Incremental cache writes are batched (setGearMany); the final write only covers keys not already written incrementally.
- Results tab: rows build into a DocumentFragment and attach once; changed items render the shared gear-change icon (gem markers + Wowhead cover link) instead of a border; a slot-change renders the item plainly. Starting a new run returns to the Setup tab.

## Candidate Counting and Filtering
- rawCombinations is the mixed-radix candidate index space.
- combinations is the filtered runnable count.
- User-visible progress/counts should reflect filtered candidate totals.

## Performance Guardrails
- Avoid per-candidate allocations in hot loops.
- Prefer preallocated imperative loops in candidate/cache helpers.
- Keep optimizer candidate-cache lookup path read-friendly and hash reuse-aware.
- Progress emission is throttled (BulkSimProgressThrottle Go-side, 100ms mirror in wasm progress emitter).
- When stripping AllValues for proto output, detach the slice before cloning (see bulkSimCandidateResultToProto) — never clone megabytes just to discard them.

## Logging Expectations
- Candidate generation logs started and completed with duration and counts.
- Optimizer stage logs one started event and one completion summary.
- Every stage (finalist included) reports metrics through the shared stage-metrics builder, so observed error and concurrency are always populated.

## Validation Commands
```bash
make proto
npm run type-check
go test -count=1 ./sim/core/bulk ./sim/web
```

For optimizer-integration changes:
```bash
go test -count=1 ./sim/core/reforge_optimizer ./sim/web
```

Reproducibility check: run the same BulkSimRequest twice through BulkSimAsync with the same seed — results must be bit-identical.

## Fast Search Aids
```bash
rg -n "BulkSimReforge|reforge_request|optimized_candidates|BulkSimStageReforge|Finalist" proto sim ui
rg -n "bulkSimAsync|/bulkSimAsync" sim/web
rg -n "pairedErrorToNextResult|bulkSimPairedDpsError|Z_95|bulkSimZ95" sim ui
```

## Common Pitfalls
- Conflating BulkSettings and ReforgeSettings.
- Porting MoP-only assumptions directly into TBC request/proto shapes (reforging, Titan's Grip, required set bonuses, fr locale).
- Writing cache metadata without optimized gear payload.
- Dropping partial optimized candidates on abort paths.
