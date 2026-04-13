# Hunter Expected-Damage APL Scheduling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the old Hunter `adaptiveRotation` opportunity-cost damage model (sim/hunter/rotation.go, old TBC codebase) into the APL system so the Hunter rotation picks the highest-expected-DPS option at each decision point independently of weapon speed, closing the ~14 DPS gap observed when swapping weapons of identical stats but different speeds (2.8s vs 3.0s crossbow reproduction).

**Architecture:** Three independent, incrementally-valuable changes:

1. **Phase 1** — Add `ExpectedInitialDamage` to every damaging Hunter spell (Steady/Multi/Arcane/Aimed/Raptor/Serpent/Scorpid/Kill Command). Uses averaged weapon damage (no RNG) and otherwise mirrors `ApplyEffects`.
2. **Phase 2** — Expose `ExpectedInitialDamage` to APL via a new generic value node `APLValueSpellExpectedDamage`. Prove the pipeline by letting users write the verbose opportunity-cost comparison directly in their APL (Hunter + any future spec).
3. **Phase 3** — Add a generic APL action `APLActionCastByExpectedDamage` that takes a list of candidate spells, evaluates each one's opportunity-cost-adjusted score, and casts the best. This encapsulates the `adaptiveRotation` math so users don't write it by hand. Then rewrite the Hunter preset APL to use it.

Each phase ships independently and is testable with the fixed-seed repro (`2_8-speed.json` vs `3_0-speed.json` — but equalize stats per tomorrow's notes; the user's `3_0-speed.json` offsets are wrong — see Background).

**Tech Stack:** Go (sim/core + sim/hunter), Protobuf (proto/apl.proto), TypeScript (ui/core/components/individual_sim_ui/apl_values.ts), i18n JSON (assets/locales/en/translation.json).

---

## Background & Reference Context

Read before starting — the plan assumes this context:

1. **The reference implementation is `/home/hillerstorm/src/tbc/sim/hunter/rotation.go`** (the OLD tbc repo, NOT tbc-new). Study `adaptiveRotation` (line 139) carefully — it's the damage model we're porting. Key ideas:
   - For each option, score = `avgDmg[option] − Σ(DPS_of_delayed × delay_time_on_delayed)`
   - Delays are pairwise (casting Steady delays next auto; weaving delays next GCD *and* next auto; etc.)
   - Averages come from a **presim** pass that runs `lazyRotation` first, then captures `avgShootDmg`, `avgSteadyDmg`, etc. See lines 382-404.
   - In the new APL world we replace the presim with live `ExpectedInitialDamage` calls (picks up current buffs — strictly better than frozen presim averages).

2. **The repro case** (verified earlier today):
   - `2_8-speed.json` → 2997.48 DPS
   - `3_0-speed.json` (user's file, wrong bonusStats) → 2981.29 DPS
   - `3_0-equalized.json` (bonusStats corrected to `stats[2]=-24, [18]=-34, [21]=-16, [30]=-13`) → 2983.52 DPS
   - Gap = ~14 DPS driven purely by APL weave-timing thresholds reacting differently to weapon speed.
   - Without melee weaving (`Melee weave = false`, `distanceFromTarget = 30`): 3.0 > 2.8 as expected → confirms core sim is fine; issue is APL tuning.

3. **Existing APL value nodes** live in `sim/core/apl_values_*.go`. Each value node has a Go type, proto message, factory in `newAPLValue*`, registration in the `APLValue` oneof in `proto/apl.proto`, and a UI input-builder in `ui/core/components/individual_sim_ui/apl_values.ts`. `APLValueSpellCastTime` at `sim/core/apl_values_spell.go:92-114` is the cleanest template.

4. **Existing APL actions** live in `sim/core/apl_actions_*.go`. Registration happens in `sim/core/apl_action.go:181` (`newAPLActionImpl`). `APLActionCastSpell` at `sim/core/apl_actions_casting.go:10-40` is a template for a simple action; `APLActionMultidot` at line 135 is a template for an action that evaluates multiple candidates before casting.

5. **Build commands** (from earlier today):
   - `protoc -I=./proto --go_out=./sim/core ./proto/*.proto` — regenerate Go proto after editing `.proto` files
   - `go build -tags=with_db -o /tmp/wowsimcli ./cmd/wowsimcli` — build the CLI sim
   - `go test --tags=with_db ./sim/hunter/...` — run Hunter tests (from memory: `--tags=with_db` is required in this repo)
   - `go test --tags=with_db ./sim/core/...` — run core tests
   - `npx tsc --noEmit` — typecheck UI
   - Proto regeneration for UI: `npx protoc --ts_opt generate_dependencies --ts_out ui/core/proto --proto_path proto proto/api.proto` (covered by `make proto`)

6. **Reproduction harness** for DPS comparisons:
   ```bash
   /tmp/wowsimcli sim --infile 2_8-speed.json --outfile /tmp/out_2_8.json
   /tmp/wowsimcli sim --infile /tmp/3_0-equalized.json --outfile /tmp/out_3_0_eq.json
   jq '.raidMetrics.dps.avg' /tmp/out_2_8.json /tmp/out_3_0_eq.json
   ```

---

## File Structure

**Phase 1** (new `ExpectedInitialDamage` impls, no new files):
- Modify: `sim/hunter/steady_shot.go`, `sim/hunter/multi_shot.go`, `sim/hunter/arcane_shot.go`, `sim/hunter/raptor_strike.go`, `sim/hunter/serpent_sting.go`, `sim/hunter/scorpid_sting.go`, `sim/hunter/talents.go` (Aimed Shot lives here — verify), `sim/hunter/kill_command.go`.

**Phase 2** (new APL value node):
- Modify: `proto/apl.proto` (add message + oneof entry)
- Modify: `sim/core/apl_values_spell.go` (add `APLValueSpellExpectedDamage`)
- Modify: `sim/core/apl_value.go` (case in `newAPLValue`)
- Modify: `ui/core/components/individual_sim_ui/apl_values.ts` (input builder)
- Modify: `assets/locales/en/translation.json` (label + tooltip)
- Create: `sim/core/apl_values_spell_expected_damage_test.go` (focused unit test with a mocked spell)

**Phase 3** (new APL action + Hunter preset update):
- Modify: `proto/apl.proto` (add `APLActionCastByExpectedDamage` message + oneof entry)
- Create: `sim/core/apl_actions_cast_by_expected_damage.go`
- Modify: `sim/core/apl_action.go` (register action in `newAPLActionImpl`)
- Modify: `ui/core/components/individual_sim_ui/apl_actions.ts` (UI builder)
- Modify: `assets/locales/en/translation.json` (action label + tooltip)
- Modify: `ui/hunter/dps/apls/*.apl.json` (update the Hunter preset APL to use the new action)
- Create: `sim/hunter/rotation_expected_damage_test.go` (compares 2.8 vs 3.0 DPS, asserts gap < threshold)

---

## Phase 1 — Add `ExpectedInitialDamage` to Hunter damaging spells

**Why this ships value on its own:** Even without Phase 2/3, several core systems (multidot target priority, future expected-DPS debug metrics) can consume these. And it's a prerequisite for Phase 2.

**Pattern (study this before writing):** `sim/core/attack.go:456-466` shows how the MH auto-attack config sets `ExpectedInitialDamage` alongside `ApplyEffects`. Key rule: `ExpectedInitialDamage` must NOT call `sim.RandomFloat()` — use `weapon.AverageDamage()` or `CalculateAverageWeaponDamage()` in place of `weapon.BaseDamage(sim)` / `CalculateWeaponDamage(sim, ...)`, and use `spell.OutcomeExpectedMeleeWhite` / `spell.OutcomeExpectedRanged*` variants where available. Returns `*SpellResult` from `spell.CalcDamage`.

### Task 1.1: Add `ExpectedInitialDamage` to Steady Shot

**Files:**
- Modify: `sim/hunter/steady_shot.go`
- Test: `sim/hunter/steady_shot_expected_damage_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `sim/hunter/steady_shot_expected_damage_test.go`:

```go
package hunter

import (
	"testing"

	"github.com/wowsims/tbc/sim/core"
)

func TestSteadyShotExpectedInitialDamage(t *testing.T) {
	core.RunTestSuite(t, "SteadyShotExpectedDamage", func(t *testing.T) {
		sim := core.NewIndividualSim(DefaultHunterOptions())
		sim.Reset(core.DefaultSimSignals())
		hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
		target := sim.Encounter.Targets[0]

		expected := hunter.SteadyShot.ExpectedInitialDamage(sim, &target.Unit)
		if expected <= 0 {
			t.Fatalf("SteadyShot.ExpectedInitialDamage returned %v, expected > 0", expected)
		}
	})
}
```

NOTE: if `core.NewIndividualSim` / `DefaultHunterOptions` signatures differ — search for an existing hunter test (e.g. `sim/hunter/hunter_test.go`) and copy its setup pattern verbatim. The point of the test is: build a configured hunter, call `spell.ExpectedInitialDamage(sim, target)`, assert > 0 and ≠ NaN.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test --tags=with_db ./sim/hunter/ -run TestSteadyShotExpectedInitialDamage -v`
Expected: PANIC or test fail — the underlying `expectedInitialDamageInternal` is nil because `ExpectedInitialDamage` isn't set on SteadyShot's config.

- [ ] **Step 3: Implement `ExpectedInitialDamage` on Steady Shot**

Modify `sim/hunter/steady_shot.go` — add the `ExpectedInitialDamage` field to the `SpellConfig` (inside `RegisterRangedSpell(...)` call, after `ApplyEffects`). The formula mirrors `ApplyEffects` but uses `AverageDamage()` instead of `BaseDamage(sim)` and `OutcomeExpectedRangedHitAndCrit` (if it exists) otherwise use `OutcomeExpectedMagicHitAndCrit`:

```go
ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
    weaponDamage := hunter.AutoAttacks.Ranged().AverageDamage() - hunter.AmmoDamageBonus

    if ranged := hunter.Ranged(); ranged != nil && ranged.Enchant.EffectID == 2722 {
        weaponDamage -= 10
    } else if ranged != nil && ranged.Enchant.EffectID == 2723 {
        weaponDamage -= 12
    }
    if hunter.Consumables.OhImbueId == 34340 || (hunter.Consumables.MhImbueId == 34340 && !hunter.windFuryEnabled) {
        weaponDamage -= 12
    }

    baseDamage := 0.2*spell.RangedAttackPower(target) +
        weaponDamage*2.8/hunter.AutoAttacks.Ranged().SwingSpeed +
        hunter.talonOfAlarBonus() +
        150

    return spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMagicHitAndCrit)
},
```

**Important:** search `sim/core/` for `OutcomeExpected` to find the correct expected-outcome function. Candidates: `OutcomeExpectedMeleeWhite`, `OutcomeExpectedRangedHitAndCrit`, `OutcomeExpectedMagicHitAndCrit`. Pick the one whose non-expected sibling is used in `ApplyEffects`. If none matches (`OutcomeRangedHitAndCrit` has no `OutcomeExpected*` sibling), create one in `sim/core/spell_result.go` following the `OutcomeExpectedMeleeWhite` pattern — but first check whether the existing druid ExpectedInitialDamage impls (`sim/druid/shred.go:53`) give a hint.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test --tags=with_db ./sim/hunter/ -run TestSteadyShotExpectedInitialDamage -v`
Expected: PASS.

- [ ] **Step 5: Verify `ApplyEffects` and `ExpectedInitialDamage` agree on average**

Add a second subtest in the same file that compares averaged actual damage over many random rolls to `ExpectedInitialDamage`:

```go
func TestSteadyShotExpectedMatchesAverage(t *testing.T) {
    // Setup as above.
    // Run 10000 iterations of the actual cast via sim.RandomFloat-driven path,
    // average them, then compare to spell.ExpectedInitialDamage within 1%.
}
```

If this is hard to wire up without a full sim, skip and rely on the integration test in Task 3.6 instead. Don't block on this.

- [ ] **Step 6: Commit**

```bash
git add sim/hunter/steady_shot.go sim/hunter/steady_shot_expected_damage_test.go
git commit -m "hunter: add ExpectedInitialDamage to Steady Shot"
```

### Task 1.2: Add `ExpectedInitialDamage` to Multi-Shot

**Files:**
- Modify: `sim/hunter/multi_shot.go`

- [ ] **Step 1: Implement**

After `ApplyEffects` in the `RegisterRangedSpell` call:

```go
ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
    baseDamage := spell.RangedAttackPower(target)*0.2 +
        hunter.AutoAttacks.Ranged().AverageDamage() +
        hunter.talonOfAlarBonus() +
        205

    // CalcAoeDamage returns a slice; take the primary target's result.
    results := spell.CalcAoeDamage(sim, baseDamage, spell.OutcomeExpectedMagicHitAndCrit)
    for _, r := range results {
        if r.Target == target {
            return r
        }
    }
    return results[0]
},
```

**Note:** Multi-Shot is AoE. If the current `ExpectedInitialDamage` type only supports a single-target result, return the primary-target result only. If `CalcAoeDamage` doesn't have an `OutcomeExpected*` equivalent, compute the single-target damage with `spell.CalcDamage(...)` instead — we only need a scalar for the scheduling decision.

- [ ] **Step 2: Run build to confirm it compiles**

Run: `go build --tags=with_db ./sim/hunter/...`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add sim/hunter/multi_shot.go
git commit -m "hunter: add ExpectedInitialDamage to Multi-Shot"
```

### Task 1.3: Add `ExpectedInitialDamage` to Arcane Shot

**Files:**
- Modify: `sim/hunter/arcane_shot.go`

- [ ] **Step 1: Implement**

```go
ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
    baseDamage := spell.RangedAttackPower(target)*0.15 +
        hunter.talonOfAlarBonus() +
        273
    return spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMagicHitAndCrit)
},
```

- [ ] **Step 2: Build and commit**

```bash
go build --tags=with_db ./sim/hunter/...
git add sim/hunter/arcane_shot.go
git commit -m "hunter: add ExpectedInitialDamage to Arcane Shot"
```

### Task 1.4: Add `ExpectedInitialDamage` to Raptor Strike

**Files:**
- Modify: `sim/hunter/raptor_strike.go`

- [ ] **Step 1: Implement**

Raptor Strike uses MH weapon damage. Study `sim/core/attack.go:462` for the MH pattern. Add to SpellConfig:

```go
ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
    baseDamage := spell.Unit.AutoAttacks.MH().CalculateAverageWeaponDamage(spell.MeleeAttackPower(target)) + 170
    return spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMeleeWhite)
},
```

Note the outcome function — Raptor Strike's `ApplyEffects` uses `OutcomeMeleeSpecialHitAndCrit`. Check `sim/core/spell_result.go` for the matching `OutcomeExpectedMeleeSpecialHitAndCrit`. If missing, use `OutcomeExpectedMeleeWhite` as a reasonable proxy.

- [ ] **Step 2: Build and commit**

```bash
go build --tags=with_db ./sim/hunter/...
git add sim/hunter/raptor_strike.go
git commit -m "hunter: add ExpectedInitialDamage to Raptor Strike"
```

### Task 1.5: Add `ExpectedInitialDamage` to Serpent Sting, Scorpid Sting, Aimed Shot

**Files:**
- Modify: `sim/hunter/serpent_sting.go`, `sim/hunter/scorpid_sting.go`, `sim/hunter/talents.go` (Aimed Shot)

- [ ] **Step 1: Open each file and study the `ApplyEffects`**

Read each spell's `ApplyEffects`. For DoTs (Serpent Sting, Scorpid Sting), set `ExpectedTickDamage` in addition to (or instead of) `ExpectedInitialDamage`. Search the druid DoT implementations (`grep -l 'ExpectedTickDamage' sim/`) for the pattern.

- [ ] **Step 2: Implement for each**

For each spell, write a mirror of its `ApplyEffects` using averaged-weapon-damage calls. Use the judgment applied in 1.1-1.4. If a spell uses a non-RNG damage formula (flat + RAP coefficient + no weapon dmg), `ExpectedInitialDamage` returns essentially the same number as one `ApplyEffects` call — still write it explicitly so the APL value has a hook.

- [ ] **Step 3: Build and test**

```bash
go build --tags=with_db ./sim/hunter/...
go test --tags=with_db ./sim/hunter/...
```

Expected: everything compiles; tests pass.

- [ ] **Step 4: Commit**

```bash
git add sim/hunter/serpent_sting.go sim/hunter/scorpid_sting.go sim/hunter/talents.go
git commit -m "hunter: add ExpectedInitialDamage to stings and Aimed Shot"
```

### Task 1.6: Kill Command — decide whether to implement

**Files:** `sim/hunter/kill_command.go`

- [ ] **Step 1: Read Kill Command's `ApplyEffects`**

It delegates damage to the pet (`hunter.Pet.KillCommand.Cast`). The expected damage lives on the pet's spell, not the hunter's.

- [ ] **Step 2: Decide**

Options: (a) set `ExpectedInitialDamage` on the hunter-side spell to return the pet's `KillCommand.ExpectedInitialDamage`, or (b) skip — Kill Command is auto-cast in the current APL preset and doesn't need to participate in expected-damage scheduling.

Recommended: **skip for Phase 1**. Add a TODO comment in the file:

```go
// TODO(expected-damage-apl): add ExpectedInitialDamage that returns the pet's
// KillCommand expected damage, if we want Kill Command to participate in
// castByExpectedDamage scheduling. See docs/superpowers/plans/2026-04-13-hunter-expected-damage-apl.md.
```

- [ ] **Step 3: Commit**

```bash
git add sim/hunter/kill_command.go
git commit -m "hunter: TODO note on Kill Command expected damage"
```

### Task 1.7: Phase 1 smoke test — baseline didn't regress

- [ ] **Step 1: Run the repro sim and confirm DPS unchanged**

```bash
go build --tags=with_db -o /tmp/wowsimcli ./cmd/wowsimcli
/tmp/wowsimcli sim --infile 2_8-speed.json --outfile /tmp/out_2_8.json
jq '.raidMetrics.dps.avg' /tmp/out_2_8.json
```

Expected: within ±2 DPS of 2997.48 (no APL behavior changed; only new callback was added).

- [ ] **Step 2: If it did regress**, something in an `ExpectedInitialDamage` impl has a side effect it shouldn't. Review — `ExpectedInitialDamage` MUST be pure (no stat mutations, no timer advancement). Fix and re-run.

---

## Phase 2 — Expose `ExpectedInitialDamage` via new APL value node `spell_expected_damage`

### Task 2.1: Add proto message and oneof entry

**Files:**
- Modify: `proto/apl.proto`

- [ ] **Step 1: Add the proto message**

In `proto/apl.proto`, after `message APLValueSpellCastTime { ... }` (line ~582), add:

```proto
message APLValueSpellExpectedDamage {
    ActionID spell_id = 1;
}
```

- [ ] **Step 2: Wire into the `APLValue` oneof**

In the `APLValue` message's oneof (line 107-247), in the "Spell values" block near line 168-178, add:

```proto
        APLValueSpellExpectedDamage spell_expected_damage = 130;
```

**Note:** use the next available tag number after scanning the file. Tags I saw used up through 129 earlier today; confirm by grepping `= \d+;` inside the APLValue oneof before committing.

- [ ] **Step 3: Regenerate Go proto**

Run: `make proto` (or `protoc -I=./proto --go_out=./sim/core ./proto/*.proto && npx protoc --ts_opt generate_dependencies --ts_out ui/core/proto --proto_path proto proto/api.proto && npx protoc --ts_out ui/core/proto --proto_path proto proto/test.proto && npx protoc --ts_out ui/core/proto --proto_path proto proto/ui.proto`)

- [ ] **Step 4: Commit**

```bash
git add proto/apl.proto sim/core/proto/ ui/core/proto/
git commit -m "proto: add APLValueSpellExpectedDamage"
```

### Task 2.2: Implement `APLValueSpellExpectedDamage` in Go

**Files:**
- Modify: `sim/core/apl_values_spell.go`
- Modify: `sim/core/apl_value.go` (the factory dispatcher — find it via `grep -n newValueSpellCastTime sim/core/apl_value.go`)

- [ ] **Step 1: Write the failing test**

Create `sim/core/apl_values_spell_expected_damage_test.go`:

```go
package core

import (
	"testing"
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
)

func TestAPLValueSpellExpectedDamage(t *testing.T) {
	// Build a minimal unit with a spell whose ExpectedInitialDamage returns 1234.
	// Construct APLValueSpellExpectedDamage via rot.newValueSpellExpectedDamage.
	// Assert GetFloat(sim) == 1234.
	t.Skip("TODO: mock unit + spell for this test")
}
```

(OK to use `t.Skip` initially — the real validation is the integration test in Phase 3. This placeholder ensures the test file compiles and is discoverable.)

- [ ] **Step 2: Implement the value node in `sim/core/apl_values_spell.go`**

Add after `APLValueSpellCastTime`:

```go
type APLValueSpellExpectedDamage struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellExpectedDamage(config *proto.APLValueSpellExpectedDamage, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellExpectedDamage{
		spell: spell,
	}
}

func (value *APLValueSpellExpectedDamage) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}

func (value *APLValueSpellExpectedDamage) GetFloat(sim *Simulation) float64 {
	target := value.spell.Unit.CurrentTarget
	if target == nil {
		return 0
	}
	return value.spell.ExpectedInitialDamage(sim, target)
}

func (value *APLValueSpellExpectedDamage) String() string {
	return fmt.Sprintf("Expected Damage(%s)", value.spell.ActionID)
}
```

**Important:** `spell.ExpectedInitialDamage` on a spell that didn't set the callback will nil-panic. Guard it:

```go
func (value *APLValueSpellExpectedDamage) GetFloat(sim *Simulation) float64 {
    target := value.spell.Unit.CurrentTarget
    if target == nil || value.spell.expectedInitialDamageInternal == nil {
        return 0
    }
    return value.spell.ExpectedInitialDamage(sim, target)
}
```

Note: `expectedInitialDamageInternal` is unexported. Either add an exported `HasExpectedInitialDamage() bool` method on `Spell`, or make the guard inside the `Spell.ExpectedInitialDamage` wrapper itself:

```go
// in sim/core/spell.go around line 736
func (spell *Spell) ExpectedInitialDamage(sim *Simulation, target *Unit) float64 {
    if spell.expectedInitialDamageInternal == nil {
        return 0
    }
    result := spell.expectedInitialDamageInternal(sim, target, spell, false)
    spell.finalizeExpectedDamage(result)
    return result.Damage
}
```

Prefer the guard inside `Spell.ExpectedInitialDamage` — centralizes the check and fixes other callers.

- [ ] **Step 3: Register in the factory**

Find `newAPLValue` dispatcher (likely `sim/core/apl_value.go` — grep for `newValueSpellCastTime` usage). Add the case:

```go
case *proto.APLValue_SpellExpectedDamage:
    return rot.newValueSpellExpectedDamage(config.GetSpellExpectedDamage(), uuid)
```

- [ ] **Step 4: Build**

```bash
go build --tags=with_db ./...
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add sim/core/apl_values_spell.go sim/core/apl_value.go sim/core/spell.go sim/core/apl_values_spell_expected_damage_test.go
git commit -m "core: add APLValueSpellExpectedDamage value node"
```

### Task 2.3: Add UI input builder and i18n strings

**Files:**
- Modify: `ui/core/components/individual_sim_ui/apl_values.ts`
- Modify: `assets/locales/en/translation.json`

- [ ] **Step 1: Add i18n entries**

In `assets/locales/en/translation.json`, locate the `rotation_tab.apl.values` block (find it via `grep -n "cast_time" assets/locales/en/translation.json`). Add:

```json
"expected_damage": {
    "label": "Expected Damage",
    "tooltip": "Expected damage of a single cast of the given spell, using average weapon damage rolls. Useful for comparing cast candidates when scheduling around auto-attacks."
}
```

Place it adjacent to `cast_time`.

- [ ] **Step 2: Add input builder**

In `ui/core/components/individual_sim_ui/apl_values.ts`, near line 1034 (where `spellCastTime` is defined), import the new proto type (add to the existing import block at top of file):

```ts
APLValueSpellExpectedDamage,
```

Then add the builder entry after `spellCastTime`:

```ts
spellExpectedDamage: inputBuilder({
    label: i18n.t('rotation_tab.apl.values.expected_damage.label'),
    submenu: ['spell'],
    shortDescription: i18n.t('rotation_tab.apl.values.expected_damage.tooltip'),
    newValue: APLValueSpellExpectedDamage.create,
    fields: [AplHelpers.actionIdFieldConfig('spellId', 'castable_spells', '')],
}),
```

- [ ] **Step 3: Typecheck**

```bash
npx tsc --noEmit
```

Expected: no errors. If `APLValueSpellExpectedDamage` is not found, regenerate TS protos: `make proto`.

- [ ] **Step 4: Commit**

```bash
git add ui/core/components/individual_sim_ui/apl_values.ts assets/locales/en/translation.json
git commit -m "ui: add APL value spellExpectedDamage input"
```

### Task 2.4: Phase 2 smoke test — write a hand-rolled expected-damage APL

- [ ] **Step 1: Create a test APL JSON**

Copy `/tmp/3_0-equalized.json` to `/tmp/3_0-expected.json`. Replace the Multi-Shot weave-branch condition (priority item ~6) with one that uses `spellExpectedDamage`:

```
if Use Multi-Shot
   AND Melee weave
   AND spellExpectedDamage(27021) > spellExpectedDamage(34120)
   → cast Multi-Shot
```

Use a Python script like we did today to rewrite the JSON. Verify the APL parses by running the sim.

- [ ] **Step 2: Run the sim**

```bash
/tmp/wowsimcli sim --infile /tmp/3_0-expected.json --outfile /tmp/out_3_0_expected.json
jq '.raidMetrics.dps.avg' /tmp/out_3_0_expected.json
```

Expected: finishes without error, DPS is in the 2970-2990 range. The exact number depends on the rewritten APL — we're just proving the plumbing works end-to-end.

- [ ] **Step 3: Document the result in the plan file**

Add a comment at the top of this plan noting the observed DPS, so tomorrow's session has the baseline.

- [ ] **Step 4: Commit**

```bash
git add docs/superpowers/plans/2026-04-13-hunter-expected-damage-apl.md
git commit -m "docs: phase 2 smoke-test result"
```

---

## Phase 3 — New `castByExpectedDamage` APL action + Hunter preset rewrite

This is the payoff: ports `adaptiveRotation`'s opportunity-cost math as a reusable APL primitive.

### Task 3.1: Design the proto message

**Files:**
- Modify: `proto/apl.proto`

- [ ] **Step 1: Add the action message**

After `APLActionCastSpell`:

```proto
message APLActionCastByExpectedDamage {
    // Candidate spells to evaluate. The action picks the one with the highest
    // opportunity-cost-adjusted expected damage at decision time, and casts it.
    // Spells must have ExpectedInitialDamage populated.
    repeated APLActionCastSpell candidates = 1;

    // Optional: auto-attack type whose DPS is used as the "delay cost" multiplier.
    // When a candidate would delay the next auto of this type, the delay
    // cost (auto DPS × delay) is subtracted from the candidate's score.
    // If unset, defaults to RangedAuto.
    AutoType delay_cost_auto = 2;
}
```

(`AutoType` already exists — it's used by `APLValueAutoTimeToNext`. Confirm by grepping.)

- [ ] **Step 2: Register in the `APLAction` oneof**

In `message APLAction { oneof action { ... } }`, add:

```proto
    APLActionCastByExpectedDamage cast_by_expected_damage = 30;
```

(tag number: next available — confirm by scanning the oneof)

- [ ] **Step 3: Regenerate proto**

```bash
make proto
```

- [ ] **Step 4: Commit**

```bash
git add proto/apl.proto sim/core/proto/ ui/core/proto/
git commit -m "proto: add APLActionCastByExpectedDamage"
```

### Task 3.2: Implement the action in Go

**Files:**
- Create: `sim/core/apl_actions_cast_by_expected_damage.go`
- Modify: `sim/core/apl_action.go`

- [ ] **Step 1: Write the action**

Create `sim/core/apl_actions_cast_by_expected_damage.go`:

```go
package core

import (
	"fmt"
	"strings"

	"github.com/wowsims/tbc/sim/core/proto"
)

type APLActionCastByExpectedDamage struct {
	defaultAPLActionImpl
	unit       *Unit
	candidates []*Spell
	autoType   proto.AutoType
	nextSpell  *Spell
}

func (rot *APLRotation) newActionCastByExpectedDamage(config *proto.APLActionCastByExpectedDamage) APLActionImpl {
	candidates := make([]*Spell, 0, len(config.Candidates))
	for _, c := range config.Candidates {
		spell := rot.GetAPLSpell(c.SpellId)
		if spell != nil {
			candidates = append(candidates, spell)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	autoType := config.DelayCostAuto
	if autoType == proto.AutoType_AutoTypeUnknown {
		autoType = proto.AutoType_RangedAuto
	}
	return &APLActionCastByExpectedDamage{
		unit:       rot.unit,
		candidates: candidates,
		autoType:   autoType,
	}
}

func (action *APLActionCastByExpectedDamage) Reset(*Simulation) {
	action.nextSpell = nil
}

func (action *APLActionCastByExpectedDamage) IsReady(sim *Simulation) bool {
	target := action.unit.CurrentTarget
	if target == nil {
		return false
	}
	bestScore := 0.0
	var best *Spell = nil

	autoDPS := action.unit.AutoAttacks.ExpectedDPS(action.autoType) // NEW helper, see Task 3.3
	autoTimeToNext := action.unit.AutoAttacks.TimeToNext(sim, action.autoType).Seconds()
	gcdTimeToReady := max(0, (action.unit.NextGCDAt() - sim.CurrentTime).Seconds())

	for _, spell := range action.candidates {
		if !spell.CanCast(sim, target) {
			continue
		}
		dmg := spell.ExpectedInitialDamage(sim, target)
		if dmg <= 0 {
			continue
		}
		// Delay imposed on the next auto: gcdTimeToReady + spell.CastTime() − autoTimeToNext.
		castTime := spell.CastTime().Seconds()
		delay := math.Max(0, gcdTimeToReady+castTime-autoTimeToNext)
		score := dmg - autoDPS*delay
		if best == nil || score > bestScore {
			best = spell
			bestScore = score
		}
	}
	if best == nil {
		return false
	}
	action.nextSpell = best
	return true
}

func (action *APLActionCastByExpectedDamage) Execute(sim *Simulation) {
	action.nextSpell.Cast(sim, action.unit.CurrentTarget)
}

func (action *APLActionCastByExpectedDamage) String() string {
	names := make([]string, 0, len(action.candidates))
	for _, s := range action.candidates {
		names = append(names, s.ActionID.String())
	}
	return fmt.Sprintf("CastByExpectedDamage[%s]", strings.Join(names, ", "))
}
```

- [ ] **Step 2: Register the action in `sim/core/apl_action.go`**

In `newAPLActionImpl` (line 192), add:

```go
case *proto.APLAction_CastByExpectedDamage:
    return rot.newActionCastByExpectedDamage(config.GetCastByExpectedDamage())
```

- [ ] **Step 3: Build — will fail, missing helper**

```bash
go build --tags=with_db ./...
```

Expected: `unit.AutoAttacks.ExpectedDPS` and `unit.AutoAttacks.TimeToNext` don't exist yet. That's Task 3.3.

### Task 3.3: Add `AutoAttacks.ExpectedDPS` and `TimeToNext(autoType)` helpers

**Files:**
- Modify: `sim/core/attack.go`

- [ ] **Step 1: Implement `ExpectedDPS`**

Near the bottom of `sim/core/attack.go`, add:

```go
// ExpectedDPS returns the current expected DPS contribution of the given auto-attack type.
// Used by APL action scheduling to cost opportunity delays.
func (aa *AutoAttacks) ExpectedDPS(autoType proto.AutoType) float64 {
    var wa *WeaponAttack
    var swingSpeed time.Duration
    switch autoType {
    case proto.AutoType_MainhandAuto:
        wa = &aa.mh
        swingSpeed = aa.MainhandSwingSpeed()
    case proto.AutoType_OffhandAuto:
        wa = &aa.oh
        swingSpeed = aa.OffhandSwingSpeed()
    case proto.AutoType_RangedAuto:
        wa = &aa.ranged
        swingSpeed = aa.RangedSwingSpeed()
    default:
        return 0
    }
    if wa.spell == nil || wa.spell.expectedInitialDamageInternal == nil || swingSpeed == 0 {
        return 0
    }
    target := aa.character.CurrentTarget
    if target == nil {
        return 0
    }
    avgDmg := wa.spell.ExpectedInitialDamage(nil /* OK since impl doesn't use sim */, target)
    return avgDmg / swingSpeed.Seconds()
}
```

**Note:** ranged-auto's `ExpectedInitialDamage` isn't currently set — see `sim/core/attack.go:488-529`. Do add one there in this task (it should use `CalculateAverageWeaponDamage` just like MH at line 462):

```go
unit.AutoAttacks.ranged.config.ExpectedInitialDamage = func(sim *Simulation, target *Unit, spell *Spell, _ bool) *SpellResult {
    baseDamage := spell.Unit.AutoAttacks.Ranged().CalculateAverageWeaponDamage(spell.RangedAttackPower(target))
    return spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedRangedHitAndCrit)
}
```

If `OutcomeExpectedRangedHitAndCrit` doesn't exist, add it in `sim/core/spell_result.go` mirroring `OutcomeExpectedMeleeWhite`. Look at that function's body for the template.

- [ ] **Step 2: Implement `TimeToNext`**

`AutoTimeToNext` already exists as an APL value — find its backing function (grep `AutoTimeToNext\b` in `sim/core/apl_values_auto_attacks.go`). If there's a helper `AutoAttacks.TimeToNext(sim, autoType)`, reuse it. If not, extract from the value node into an exported method:

```go
func (aa *AutoAttacks) TimeToNext(sim *Simulation, autoType proto.AutoType) time.Duration {
    var swingAt time.Duration
    switch autoType {
    case proto.AutoType_MainhandAuto:
        swingAt = aa.mh.swingAt
    case proto.AutoType_OffhandAuto:
        swingAt = aa.oh.swingAt
    case proto.AutoType_RangedAuto:
        swingAt = aa.ranged.swingAt
    default:
        return 0
    }
    return max(0, swingAt-sim.CurrentTime)
}
```

- [ ] **Step 3: Build**

```bash
go build --tags=with_db ./...
```

Expected: success.

- [ ] **Step 4: Commit**

```bash
git add sim/core/attack.go sim/core/apl_actions_cast_by_expected_damage.go sim/core/apl_action.go sim/core/spell_result.go
git commit -m "core: add castByExpectedDamage action with ExpectedDPS helper"
```

### Task 3.4: Add UI builder for the action

**Files:**
- Modify: `ui/core/components/individual_sim_ui/apl_actions.ts` (find via `grep -rn 'castSpell.*inputBuilder' ui/core/`)
- Modify: `assets/locales/en/translation.json`

- [ ] **Step 1: Add i18n entries**

In `assets/locales/en/translation.json`, in the `rotation_tab.apl.actions` block, add:

```json
"cast_by_expected_damage": {
    "label": "Cast by Expected Damage",
    "tooltip": "Evaluates each candidate spell's expected damage, subtracts the opportunity cost of delaying the next auto-attack of the chosen type, and casts the highest-scoring candidate. Requires candidates to have ExpectedInitialDamage defined."
}
```

- [ ] **Step 2: Add the UI builder**

In `apl_actions.ts`, model off the existing `castSpell` action builder. The UI should render a list of `{spellId}` candidates (reuse `APLActionCastSpell`'s spell-picker) and an `AutoType` dropdown defaulted to `RangedAuto`.

Exact code varies based on the existing file patterns — open `apl_actions.ts`, locate the `castSpell` entry, clone it, and adapt. The fields list should be something like:

```ts
fields: [
    AplHelpers.listFieldConfig({
        itemLabel: 'Candidates',
        fieldName: 'candidates',
        newItem: APLActionCastSpell.create,
        itemFields: [
            AplHelpers.actionIdFieldConfig('spellId', 'castable_spells', ''),
        ],
    }),
    AplHelpers.enumFieldConfig('delayCostAuto', AutoType, /* default */ AutoType.RangedAuto),
]
```

- [ ] **Step 3: Typecheck and commit**

```bash
npx tsc --noEmit
git add ui/core/components/individual_sim_ui/apl_actions.ts assets/locales/en/translation.json
git commit -m "ui: add castByExpectedDamage action builder"
```

### Task 3.5: Write integration test asserting 2.8-vs-3.0 DPS parity

**Files:**
- Create: `sim/hunter/rotation_expected_damage_test.go`

- [ ] **Step 1: Write the test**

```go
package hunter

import (
	"testing"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

// TestExpectedDamageAPLCloseGap verifies that swapping between a 2.8-speed and
// 3.0-speed weapon with identical stats produces DPS within 5 of each other,
// when using the castByExpectedDamage-based rotation.
//
// See docs/superpowers/plans/2026-04-13-hunter-expected-damage-apl.md.
func TestExpectedDamageAPLCloseGap(t *testing.T) {
	// Load /home/hillerstorm/src/tbc-new/.claude/worktrees/quiet-beaming-prism/2_8-speed.json
	// and /tmp/3_0-equalized.json as RaidSimRequests. Replace their rotation
	// priorityList with a single castByExpectedDamage action over
	// [SteadyShot, MultiShot, ArcaneShot]. Run each at a fixed seed.
	// Assert: |dps28 - dps30| < 5.
	t.Skip("TODO: wire up RaidSimRequest loading in test harness")
}
```

(This test is placeholder — the real validation is manual until the UI/JSON path is proven. That's OK for Phase 3; the integration check happens via Task 3.6.)

- [ ] **Step 2: Commit**

```bash
git add sim/hunter/rotation_expected_damage_test.go
git commit -m "hunter: placeholder test for expected-damage APL gap closure"
```

### Task 3.6: Manual validation against the repro

- [ ] **Step 1: Build the sim with all changes**

```bash
make proto && go build --tags=with_db -o /tmp/wowsimcli ./cmd/wowsimcli
```

- [ ] **Step 2: Generate test APL JSONs using the new action**

Write a Python script (`/tmp/make_expected_apl.py`) that:

1. Loads `/home/hillerstorm/src/tbc-new/.claude/worktrees/quiet-beaming-prism/2_8-speed.json` and `/tmp/3_0-equalized.json`.
2. Replaces the `rotation.priorityList` with a minimal priority:
   - Mana management group
   - autocastOtherCooldowns
   - Weave group (unchanged)
   - castByExpectedDamage([SteadyShot=34120, MultiShot=27021, ArcaneShot=27019], delayCostAuto=RangedAuto)
3. Keeps `valueVariables` and `groups` intact (we still want weaving).
4. Writes `/tmp/2_8_exp.json` and `/tmp/3_0_exp.json`.

- [ ] **Step 3: Run the sims**

```bash
/tmp/wowsimcli sim --infile /tmp/2_8_exp.json --outfile /tmp/out_2_8_exp.json
/tmp/wowsimcli sim --infile /tmp/3_0_exp.json --outfile /tmp/out_3_0_exp.json
echo "2.8 expected-damage:  $(jq '.raidMetrics.dps.avg' /tmp/out_2_8_exp.json)"
echo "3.0 expected-damage:  $(jq '.raidMetrics.dps.avg' /tmp/out_3_0_exp.json)"
echo "2.8 original:         $(jq '.raidMetrics.dps.avg' /tmp/out_2_8.json)"
echo "3.0 equalized:        $(jq '.raidMetrics.dps.avg' /tmp/out_3_0_eq.json)"
```

**Success criteria:**
- Both expected-damage runs are within 5 DPS of each other (gap closed).
- Neither run is more than 20 DPS below the 2.8 original (no massive overall regression).
- Bonus: total DPS ≥ 2.8 original (we beat the old hand-tuned APL at the tuned speed).

- [ ] **Step 4: Record results in the plan**

Append results to this plan's footer (a "Results" section with the DPS numbers and the date).

- [ ] **Step 5: Commit results**

```bash
git add docs/superpowers/plans/2026-04-13-hunter-expected-damage-apl.md
git commit -m "docs: record expected-damage APL validation results"
```

### Task 3.7: Update the Hunter DPS preset APL

**Files:**
- Modify: `ui/hunter/dps/apls/*.apl.json` (find via `ls ui/hunter/dps/apls/`)

- [ ] **Step 1: Identify the default APL**

```bash
ls ui/hunter/dps/apls/
```

There's likely a `bm.apl.json`, `mm.apl.json`, `sv.apl.json` or similar. Open each and identify which corresponds to BM (matches what's in `2_8-speed.json` / `3_0-speed.json`).

- [ ] **Step 2: Rewrite the priority list**

Replace the complex weave-branch Multi/Arcane/Steady priority with the new `castByExpectedDamage` action, keeping Mana management, autocastOtherCooldowns, Kill Command, and Weave unchanged.

- [ ] **Step 3: Regenerate DPS RegressionTests for Hunter BM**

```bash
go test --tags=with_db ./sim/hunter/... -update
```

(The `-update` flag refreshes golden DPS numbers.)

- [ ] **Step 4: Review the diff**

```bash
git diff sim/hunter/testdata/
```

Sanity-check that no regression > 50 DPS appears in any fixture. If it does, investigate before committing.

- [ ] **Step 5: Commit**

```bash
git add ui/hunter/dps/apls/ sim/hunter/testdata/
git commit -m "hunter: use castByExpectedDamage in BM preset APL"
```

---

## Phase Completion Criteria

| Phase | Ships what | Verifies how |
|---|---|---|
| **Phase 1** | Every damaging Hunter spell has `ExpectedInitialDamage` | Baseline sim DPS unchanged (±2); all hunter tests pass |
| **Phase 2** | `spellExpectedDamage` APL value node usable in conditions | Sim runs with hand-written expected-damage condition without error |
| **Phase 3** | `castByExpectedDamage` APL action; Hunter preset updated | 2.8-vs-3.0 DPS gap < 5 DPS with identical stats; hunter tests pass after golden update |

## Known Open Questions (address during execution)

1. **Weave scheduling**: The old `adaptiveRotation` also scores the "Weave" option. This plan's `castByExpectedDamage` only handles cast candidates, not moves. If the weave-vs-cast decision itself turns out to be the bigger contributor to the 14 DPS gap, we'll need a separate `castOrMoveByExpectedDamage` action. Verify this during Task 3.6 — if the gap *doesn't* close to < 5 DPS, the decision is the next place to look.
2. **Kill Command participation**: Task 1.6 skips it. If KC shows up as an opportunity to schedule better, revisit.
3. **Wyvern Sting / Black Arrow**: Wotlk/Cata abilities not present in this TBC codebase — don't touch.

---

## Results (fill in after execution)

Recorded DPS values from Task 3.6 validation:

- 2.8 original (baseline): TBD
- 3.0 equalized (baseline): TBD
- 2.8 with castByExpectedDamage: TBD
- 3.0 with castByExpectedDamage: TBD
- Gap closed: TBD
