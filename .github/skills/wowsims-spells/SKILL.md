---
name: wowsims-spells
description: 'Use when working on WoWSims TBC spell data: registering or downranking a spell, reading cost, cast time, GCD, cooldown, range, missile speed or damage from the generated rank tables, regenerating those tables from the client database, or reconciling a sim number against what the DBC says.'
argument-hint: 'Describe the spell, rank, or generated-table task to work on.'
---

# WoWSims TBC Spell Data Guide

## Scope

- The generated rank tables in sim/<class>/spell_ranks_auto_gen.go — 807 families, 3327 rows, all nine classes.
- Reading them from a spell config: sim/common/shared/spell_ranks.go.
- Regenerating them: tools/database/gen_spell_ranks.go, tools/database/spellranks.go.
- Reconciling a sim number that disagrees with the client data.

- Reading a talent's per-rank values, which come from the same tables.

Not in scope: item and enchant data (gen_db proper), and talent _trees_ — the JSON, protos and TS configs under ui/sim/talents are already generated from the Talent table by tools/database/gen_protos.go.

## Architecture

- sim/common/shared/spell_ranks.go — hand-written. SpellRank, the SpellRankValue union, the table accessors. The only hand-written piece of the pipeline.
- sim/<class>/spell_ranks_auto_gen.go — generated, checked in, one `genRanks` global per class with a field per family.
- tools/database/spellranks.go — the derivation rule and the row loader, shared by the generator and the regeneration check so the two cannot drift.
- tools/database/gen_spell_ranks.go — ladder discovery, role assignment, rendering.
- tools/database/gen_spellranks/ — the standalone binary. Not a mode of gen_db: gen_db imports the sim and the sim reads these tables, so a stale generated file stopped the generator that would fix it from compiling.
- sim/common/shared/spell_rank_talents.go — hand-written. The rank-indexed readers and the SPELLMOD names.
- sim/common/shared/spell_rank_enums_auto_gen.go — generated. The A_ and E_ names the tables use, mirrored from tools/database/dbc/enums.go by parsing it with go/ast, because sim must not import tools.
- docs/spell_ranks.md — the usage guide, with worked examples.

## Using a rank

```go
var exorcismRank = genRanks.Exorcism.BySpellID(27138)

ManaCost:         core.ManaCostOptions{FlatCost: exorcismRank.Cost},
BonusCoefficient: exorcismRank.Direct.BonusCoefficient(),
MinRange:         exorcismRank.MinRange,
MaxRange:         exorcismRank.MaxRange,
MissileSpeed:     exorcismRank.MissileSpeed,
Cast: core.CastConfig{
    DefaultCast: core.Cast{GCD: exorcismRank.GCD, CastTime: exorcismRank.CastTime},
    CD:          core.Cooldown{Timer: paladin.getExorcismTimer(), Duration: exorcismRank.Cooldown},
},
ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
    spell.CalcAndDealDamage(sim, target, exorcismRank.Direct.Damage(sim), spell.OutcomeMagicHitAndCrit)
},
```

A rank's value is discriminated by shape — SpellRankFlat, SpellRankRange, SpellRankPeriodic — behind a sealed interface, so a flat number has no Max to misread and only a periodic value carries a tick schedule. `Damage(sim)` rolls where the client rolls and returns the amount unchanged where it does not; `Range()` gives both ends.

Role fields: Direct, Heal, Periodic, Energize. Each is one effect. A rank can carry nothing in a role — Lay on Hands rank 1 restores no mana where ranks 2-4 do — so the `SpellRankMin/Max/Coef/APCoef` helpers read nil as zero where a direct field access would panic.

## Using a talent

A talent is read by the points spent in it, not registered at a rank it has. All three readers answer
the identity at rank 0 — untaken — where `ByRank` would panic, so no `if rank > 0` guard is needed:

```go
genRanks.Moonfury.FractionAt(rank)        // 0.10 at 5/5 — the client's 10, over 100
genRanks.NaturesReach.ValueAt(rank)       // 20 at 2/2  — the client's number as it stands
genRanks.LivingSpirit.MultiplierAt(rank)  // 1.15 at 5/5 — 1 + the fraction
```

`MultiplierAt` takes its sign from the data: Improved Righteous Fury states its reduction as -2/-4/-6,
so rank 3 gives 0.94 and nobody writes the minus.

A talent with several effects must name one, and `ValueAt` panics rather than guess:

```go
genRanks.ImprovedRighteousFury.
    Effect(shared.A_ADD_PCT_MODIFIER, shared.SPELLMOD_ALL_EFFECTS).MultiplierAt(rank)  // 1.50 threat
genRanks.ImprovedRighteousFury.
    Effect(shared.A_ADD_FLAT_MODIFIER, shared.SPELLMOD_EFFECT2).MultiplierAt(rank)     // 0.94 taken
```

`Effect` also panics when _two_ effects match: 186 ranked spells carry a duplicate aura/misc pair, so
those index into `Effects` instead. `Index` is stored rather than inferred from position — 8 rows have
a gap.

## Regenerating

```
go run ./tools/database/gen_spellranks     # then the diff must be empty on a second run
go test --tags=with_db ./tools/database/   # the regeneration check
```

Needs tools/database/wowsims.db, which is gitignored and built by `make db` from a local WoW client. The gate skips when it is absent, so CI and a fresh clone are unaffected.

## Traps

- **`--tags=with_db` is required** on any sim test run, or it panics with `No DB data for enchant with id: 2613`. That panic is the missing tag, not a broken fixture.
- **Prefer `BySpellID` to `HighestRank`.** Fireball 38692 and Frostbolt 38697 are rank 14 entries at level 70 whose SkillLineAbility rows and attributes are indistinguishable from the real rank 13s, so the generator cannot filter them. Reaching for the top rank silently casts a spell the game never grants — it cost 1.2% DPS and 380 assertions when it happened.
- **`HighestRank()` is not the last element.** Flamestrike is declared rank 7 then rank 6. Declaration order is registration order and must not be sorted: it decides which spell `GetSpell` returns where two share an ActionID, as the six Seal of Command procs do.
- **Rage is stored in tenths.** The client tracks a 0-1000 bar where the UI shows 0-100, so Heroic Strike reads 150 against the sim's 15. `NormalizePowerCost` divides it through; mana, energy and focus need no conversion.
- **DBC array columns are 0-based.** `EffectMiscValue_0` is the first element. Reading `_1` gets the second and is usually wrong.
- **A sub-second GCD must be named twice.** `core.Cast.GCDMin` overrides the global one-second floor, so Hammer of Wrath and Shadowfury set both `GCD` and `GCDMin` or `GCDTime` clamps them back up.
- **Missile speed is untested.** `TravelTime` is distance over speed and every golden runs at distance 0, so no golden defends any missile speed in the sim.
- **`make update-tests` deletes every .results before promoting .tmp.** A test that cannot run locally — `TestProtoVersioning` needs npx and network — produces no .tmp and loses its golden outright. Check `git status -- '*.results'` for a `D` afterwards. Prefer promoting individual goldens with `cp`.
- **Never delete a generated file before regenerating it.** gen_db imports the sim, so a class package missing its table does not compile and neither does the generator that would recreate it.
- **CI runs `npm run fmt` from the merge with master.** Master widened it to `npx oxfmt . --check`, so docs/ and tools/ are checked even on a branch whose own script only covers ui/.
- **Never pick a talent's effect by which one matches the number.** Survival of the Fittest states +1/2/3% to all stats and -1/-2/-3% crit taken; both ladders fit, and an automatic pass attached the stat effect to the crit-taken call site. What the call site does decides it, and the mod's `Kind` usually says so — Improved Moonfire's two mods are `SpellMod_DamageDone_Flat` and `SpellMod_BonusCrit_Percent`.
- **A per-point literal is usually right, so migrating one buys provenance, not accuracy.** Of 125 percent talents, 122 scale linearly. The ones that do not are why the table is worth reading: Improved Righteous Fury is 16/33/50, not 16/32/48, and shaman Elemental Weapons is 7/14/20, not 7/14/21.
- **The `SPELLMOD_*` names are read off the talents that use them**, not mirrored from the client, which ships no list. Each carries its evidence in a trailing comment and the four thinnest are marked; audit before trusting.
- **A proc chance of 100 is not always a 100% roll.** `SpellAuraOptions.ProcChance` reads 100 for Flurry and Enrage because they fire on a crit rather than a chance; the number the sim wants is elsewhere.
- **`rtk`-wrapped `go test` exits 0 with failing tests.** Read the summary line, not the exit code.

## When a sim number disagrees with the client

The generated table is a second opinion on every number the sim already had. When they differ, decide rather than assume — and when a disagreement is left standing, say so at the call site.

Evidence that the sim is wrong: the value equals a _different_ rank's (Shield Slam 30356 is rank 6 but dealt rank 5's 381-399); the value equals a sibling spell's (Holy Shock rank 3's max was the heal's min); the spread breaks the die-sides ladder.

Evidence that the sim is right: a uniform offset across unrelated abilities usually means the sim models something the DBC base excludes — every hunter cast time is exactly +500ms, a ranged-weapon component, not three transcription slips. A coefficient read from a different spell is not a disagreement either: Blizzard's 0.119 belongs to the tick spell 42208, not the parent.

## Verifying a change

```
go test --tags=with_db ./sim/...
git status --porcelain -- '*.results'      # empty unless a number was meant to move
go run ./tools/database/gen_spellranks     # regenerate; the diff must be empty
```

A port that was meant to be mechanical and moves a golden is a wrong port, not a new baseline.
