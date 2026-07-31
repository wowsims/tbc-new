---
name: wowsims-tbc-reforge-optimizer-handoff
description: 'Use when continuing, debugging, validating, or modifying the WoWSims TBC backend reforge optimizer, /reforgeOptimize endpoint, gem/socket/cap logic, or meta-gem activation behavior.'
argument-hint: 'Describe the TBC reforge optimizer bug, fixture, or behavior to continue.'
---

# WoWSims TBC Reforge Optimizer Handoff

## Scope

- Core optimizer behavior in `sim/core/reforge_optimizer`.
- Endpoint and worker integration for Suggest Reforges.
- HiGHS-backed MIP solving across reforge, gems, and socket bonus choices.
- Relative stat cap handling and exact post-solve validation.

## Architecture

- Main flow: `sim/core/reforge_optimizer/optimizer.go`.
- MIP model and constraints: `sim/core/reforge_optimizer/solver.go`.
- Choice/cap/stat support: `choices.go`, `caps.go`, `search.go`, `stats.go`.
- Gear apply + regem minimization: `gear.go`.
- Gem/socket handling: `gems.go` and meta-gem constraints helpers.
- Relative cap logic: `relative_stat_cap.go`.
- HiGHS bridges:
    - Go non-browser: `highswasm.go`.
    - Browser wasm: `highs_js.go`.
- Frontend caller: `ui/core/components/suggest_reforges_action.tsx`.

## Core Invariants

- Final correctness must be validated with exact `core.ComputeStats` results.
- Solver objective/deltas are guidance, not final correctness authority.
- Optimizer includes reforge + gem + socket bonus surfaces when enabled.
- Socket bonus feasibility is modeled explicitly with link constraints.
- On HiGHS failure, return an error; do not silently downgrade solver behavior.
- Verbose optimizer diagnostics remain behind `ReforgeOptimizeRequest.debug`.

## Cap and Breakpoint Rules

- Validate and normalize cap settings before solve.
- Enforce hard caps and breakpoint-derived limits as MIP constraints.
- If exact post-check violates constraints, tighten existing rows and re-solve.
- Keep soft-cap scoring piecewise with pre-cap and post-cap EP behavior.

## Relative Stat Cap Rules

- Model forced-vs-constrained requirements as explicit linear constraints.
- Preserve raw Crit/Haste/Mastery deltas for feasibility checks.
- Validate relative cap results with exact final stats and tighten rows if needed.
- Windwalker forced Mastery constrains Mastery vs Crit and Mastery vs Haste; avoid unnecessary cross constraints.

## Gem and Meta Rules

- Meta gems are not regular swap targets.
- Regem minimization must preserve/restore meta socket correctness.
- Keep gem order stable in `EquipmentSpec` output.
- Preserve class-specific stat dependency semantics in gem scoring.

## Integration Contract

- Backend endpoint path remains `/reforgeOptimize`.
- Worker API call remains `reforgeOptimize`.
- Request/response proto source remains `ui.proto` (`ReforgeOptimizeRequest`, `ReforgeOptimizeResult`, `ReforgeSettings`).
- Bulk integration uses bulk mode on the same optimizer request domain; avoid duplicate reforge config types.

## Performance Guardrails

- Keep model-building and hot-path helpers allocation-aware.
- Keep selected-choice legality checks lightweight.
- Avoid debug timers/tracing overhead unless debug is enabled.

## Validation Commands

```bash
go test -count=1 ./sim/core/reforge_optimizer
npm run type-check
```

For integration changes touching bulk or endpoint behavior:

```bash
go test -count=1 ./sim/core/reforge_optimizer ./sim/web
```

## Fast Search Aids

```bash
rg -n "ReforgeOptimizeRequest|relativeStatCap|softCap|breakpoint|HiGHS" proto sim ui
rg -n "reforgeOptimize|/reforgeOptimize" sim ui
```

## Common Pitfalls

- Accepting solver-feasible output without exact-stat verification.
- Letting post-processing invalidate meta-gem or socket constraints.
- Introducing frontend/backend drift in gem/socket-force behavior.
- Porting MoP-specific assumptions into TBC proto or stat semantics.
