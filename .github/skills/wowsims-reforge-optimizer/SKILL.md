---
name: wowsims-reforge-optimizer
description: 'Use when working on the WoWSims TBC gear optimizer ("reforge optimizer" — the name is inherited from MoP; TBC has no reforging, it optimizes gems/socket bonuses), the /reforgeOptimizeAsync endpoint, HiGHS solver integration, gem/socket/cap logic, or softcap breakpoints.'
argument-hint: 'Describe the optimizer bug, fixture, or behavior to work on.'
---

# WoWSims TBC Gear Optimizer Guide

## Scope
- Core optimizer behavior in sim/core/reforge_optimizer (gem + socket-bonus choices; no reforging exists in TBC — the "reforge" naming is legacy from the MoP port).
- Endpoint path /reforgeOptimizeAsync and worker integration.
- HiGHS-backed MIP solving for gem + socket-bonus choices under stat caps.

## Architecture
Files (all under sim/core/reforge_optimizer/):
- Main flow + optimizer state: optimizer.go.
- LP model build — decision variables, gem options, objective/cap coefficient split: model.go.
- Byte-exact deterministic CPLEX LP text serialization for HiGHS: lp.go (rows are <=/>= only; term/sign formatting goes through appendLPTerm — the byte layout is load-bearing for HiGHS tie-breaking).
- Solve + cap-refinement loop: solver.go.
- Soft caps / gap-to-cap conversion + cap-detection helpers: caps.go.
- Stat/UnitStat math: reforge_stats.go.
- Gear/gem sim wrappers, applyLPSolution, regem minimization: gear.go.
- Meta-gem activation constraints: meta_gem_constraints.go.
- Relative cap modeling: relative_stat_cap.go.
- Spec/profession predicates: utils.go.
- HiGHS bridge:
  - Go non-browser (embedded, pooled wazero): highswasm.go. (Do NOT rename to highs_wasm.go — the `_wasm.go` suffix is an implicit GOARCH=wasm constraint that breaks the native build.)
  - Browser wasm: highs_js.go.
- Frontend caller: ui/core/components/suggest_reforges_action.tsx. The IndexedDB cache key contract lives in cacheRelevantReforgeRequest there — any new field that affects a solve must be reflected, and any irrelevant field excluded, or caches go silently stale / get needlessly busted. Hashing is synchronous via hashString in ui/core/utils.ts.

## Core Invariants
- Final correctness must be validated with exact core.ComputeStats results; solver deltas are guidance only.
- Optimizer includes gem + socket bonus surfaces when enabled.
- Socket bonus feasibility is modeled explicitly with link constraints.
- On HiGHS failure, return an error; do not silently downgrade solver behavior.
- Verbose optimizer diagnostics remain behind ReforgeOptimizeRequest.debug.
- Deterministic LP text: modelToLPFormat is byte-exact and stable, so HiGHS tie-breaking among equal-objective solutions stays reproducible.

## Cap and Breakpoint Rules
- Validate and normalize cap settings before solve.
- Enforce hard caps and breakpoint-derived limits as MIP constraints.
- If exact post-check violates constraints, tighten existing rows and re-solve.
- Keep soft-cap scoring piecewise with pre-cap and post-cap EP behavior.

## Gem and Meta Rules
- Meta gems are not regular swap targets.
- Regem minimization must preserve/restore meta socket correctness (meta activation conditions modeled in meta_gem_constraints.go).
- Keep gem order stable in EquipmentSpec output.
- Preserve class-specific stat dependency semantics in gem scoring.

## Bulk Sim Integration Contract
- Bulk Sim uses ReforgeOptimizeRequest mode for bulk operations; avoid duplicate optimizer config types.
- Bulk pre-pass may call the optimizer twice per candidate when includeGems fallback is enabled.
- Cache keys represent input identity; cached values are optimized output gear (plain `equipmentSpec:` proto JSON).
- Bulk runs derive their RNG seed from content including the cache-relevant optimizer config — see the wowsims-bulk-sim skill for the staged/finalist pipeline.

## Performance Guardrails
- Keep model-building and hot-path helpers allocation-aware.
- Keep selected-choice legality checks lightweight.
- Avoid debug timers/tracing overhead unless debug is enabled.

## Validation Commands
```bash
go test -tags with_db -count=1 ./sim/core/reforge_optimizer
npm run type-check
```
Fixtures require the `with_db` tag; they live in sim/core/reforge_optimizer/test-fixtures/ (gem-limits, gem-pool-narrow, gem-pool-wide, meta-gem-comparative, soft-caps-multi, stat-conversions, tank-caps, tank-soft-caps). Never update fixtures without being asked.

For integration changes touching bulk or endpoint behavior:
```bash
go test -tags with_db -count=1 ./sim/core/reforge_optimizer ./sim/web
```

## Fast Search Aids
```bash
rg -n "ReforgeOptimizeRequest|softCap|breakpoint|HiGHS" proto sim ui
rg -n "reforgeOptimizeAsync|ReforgeOptimize" sim ui
```

## Common Pitfalls
- Accepting solver-feasible output without exact-stat verification.
- Letting post-processing invalidate meta-gem or socket constraints.
- Introducing frontend/backend drift in gem/socket-force behavior.
- Porting MoP-specific assumptions into TBC (reforging, Rune of Re-Origination, Windwalker/monk specs, amp trinkets).
