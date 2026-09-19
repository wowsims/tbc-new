# Spell Ranks

Every ranked spell in the game has a table generated from the client database, checked in at
`sim/<class>/spell_ranks_auto_gen.go`. A spell reads its numbers from that table instead of carrying
hand-transcribed literals.

- [Using a rank](#using-a-rank)
- [The value shapes](#the-value-shapes)
- [Talents](#talents)
- [Worked examples](#worked-examples)
- [Attack power](#attack-power)
- [Regenerating](#regenerating)
- [Traps](#traps)

## Using a rank

Each class package has exactly one generated global, `genRanks`, with a field per spell family:

```go
genRanks.Exorcism          // the whole ladder, ranks 1-7
genRanks.Fireball          // ranks 1-14
```

Pick the rank your spell registers **by spell ID**:

```go
var exorcismRanks = genRanks.Exorcism.BySpellID(27138)
```

That is the identity the sim already uses everywhere - `ActionID`, saved APLs and the icon database all
key on the spell ID - so it cannot drift onto a different rank, and a regeneration that drops the ID
fails loudly instead of quietly substituting another.

The other accessors:

|                    |                                                                                               |
| ------------------ | --------------------------------------------------------------------------------------------- |
| `BySpellID(27138)` | the rank registered under that spell ID. Prefer this.                                         |
| `ByRank(6)`        | the rank numbered 6                                                                           |
| `Ranks(6, 8)`      | a subset, **in the order given**, which is registration order                                 |
| `HighestRank()`    | the highest rank _in the data_, which is not always one the game grants - see [Traps](#traps) |
| `RegisterAll(f)`   | calls `f` once per rank, in declaration order                                                 |

## The value shapes

A rank's value is discriminated by shape, so a variant only carries fields that mean something for it:

```go
shared.SpellRankFlat     {Value, Coef, APCoef}                          // a mana restore, a talent's number
shared.SpellRankRange    {Min, Max, Coef, APCoef}                       // damage or healing the client rolls
shared.SpellRankPeriodic {Tick, TickMax, TickLength, NumberOfTicks, Coef, APCoef} // a tick and its schedule
```

They sit on the four roles a rank can carry, any of which may be nil:

```go
rank.Direct     // Effect = SCHOOL_DAMAGE
rank.Heal       // Effect = HEAL
rank.Periodic   // a periodic aura
rank.Energize   // Effect = ENERGIZE, e.g. Lay on Hands' mana restore
```

Asking what a value is worth on this cast is a single call, because the question means something for
all three shapes - a range rolls between its ends, a flat value and a tick are already the answer. The
method is named for what it produces rather than for how, so a static ability does not read as if it
rolled:

```go
baseDamage := rank.Direct.Damage(sim)     // instead of CalcAndRollDamageRange(sim, min, max)
tickDamage := rank.Periodic.Damage(sim)   // a tick is the answer unless the client rolls it
```

The coefficients are methods, named for the `core.SpellConfig` fields they feed:

```go
BonusCoefficient: rank.Direct.BonusCoefficient(),     // spell power
                  rank.Periodic.BonusCoefficient(),   // same on a tick
                  rank.Direct.APBonusCoefficient(),   // attack power
```

`Range()` gives both ends of a value at once:

```go
low, high := rank.Direct.Range()   // equal for a flat value or a tick
```

The methods assume the role is there. Where it may not be - `Energize` is nil on Lay on Hands rank 1 -
use the package helpers instead, which read a nil value as zero:

```go
shared.SpellRankMin(rank.Energize)     // 0 rather than a panic
shared.SpellRankMax(rank.Direct)
shared.SpellRankCoef(rank.Periodic)
shared.SpellRankAPCoef(rank.Direct)
```

Tick length and count live only on the periodic shape, so assert for them:

```go
p := rank.Periodic.(shared.SpellRankPeriodic)
p.TickLength     // time.Duration, feeds core.DotConfig.TickLength
p.NumberOfTicks  // duration over the tick length, feeds core.DotConfig.NumberOfTicks
```

A rank also carries what the client knows about casting it:

|                               |                                                                                                                                            |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `Cost`                        | in the units the sim uses - see the rage trap below                                                                                        |
| `CastTime`, `GCD`, `Cooldown` | zero for a channel, whose duration carries it                                                                                              |
| `MinRange`, `MaxRange`        | `core` gates the cast on both; zero means ungated. `MinRange` is the dead zone on a charge, and is nonzero on only 212 spells in the build |
| `MissileSpeed`                | yards per second, which `core` turns into the delay before the damage lands. Zero is an instant hit                                        |

`MissileSpeed` is the one to be careful with: giving a spell a speed it did not have delays its damage
and moves goldens, so check the sim is not already modelling it elsewhere. Arcane Missiles is the case
to know - the channel carries no speed because the missile spell does, and that one is not a ranked row.

## Reaching a single effect

The role fields describe one effect each, which is all a castable spell needs. A talent routinely
carries two or three that the sim reads separately, and only one of them can be `Direct`:

```go
irf := genRanks.ImprovedRighteousFury.ByRank(3)

irf.Effect(shared.A_ADD_PCT_MODIFIER, 8).Value    //  50  threat bonus
irf.Effect(shared.A_ADD_FLAT_MODIFIER, 12).Value  //  -6  damage taken
irf.Effects[1].Value                              //  -6  the same effect, by index
```

The aura and effect names are generated into `sim/common/shared/spell_rank_enums_auto_gen.go`, mirrored
from `tools/database/dbc/enums.go` and holding only the values the tables use, so the two cannot drift.
`Misc` stays a plain int: what it selects depends on the aura - a modified spell property for
`A_ADD_PCT_MODIFIER`, a stat for `A_MOD_TOTAL_STAT_PERCENTAGE`, a school mask for `A_MOD_DAMAGE_DONE` -
so there is no single enum to name it with.

Name the effect by aura rather than reading `Direct` whenever a spell has more than one. Which effect
lands in `Direct` is the generator's choice, not a promise, so a caller that depends on it breaks
silently the day the ordering changes.

`Effect` panics when nothing matches, and also when **two** effects match: 186 ranked spells carry a
duplicate aura/misc pair, and returning the first is how a caller ends up reading the wrong half of a
talent. Index into `Effects` where the pair cannot tell them apart.

**The value is in the client's units.** A percentage is an integer here - Improved Righteous Fury's
threat bonus reads `16`, not `0.16` - so the `/100` stays at the call site. It is deliberately not
folded into the generator the way the rage `/10` is: whether a value is a percentage depends on the
aura, so a blanket rule would be wrong for some rows and invisible when it was.

## Talents

A talent is read by the points spent in it, not registered at a rank it has, so the ladder has its own
three readers. All of them answer the identity at rank 0 - an untaken talent - where `ByRank` would
panic:

```go
genRanks.Moonfury.FractionAt(rank)        // 0.10 at 5/5 - the client's 10, over 100
genRanks.NaturesReach.ValueAt(rank)       // 20 at 2/2  - the client's number as it stands
genRanks.LivingSpirit.MultiplierAt(rank)  // 1.15 at 5/5 - 1 + the fraction
```

That replaces the `<literal> * float64(x.Talents.Y)` idiom, and with it the `if rank > 0` guard the
caller would otherwise need.

**`MultiplierAt` takes its sign from the data.** Improved Righteous Fury states its damage reduction as
-2 / -4 / -6, so rank 3 gives 0.94 and nobody writes the minus. Where the sim's parameter runs the other
way - `AddReducedCritTakenPercent` wants a positive amount for a reduction the client states negative -
negate at the call site so the disagreement is visible.

**Ladders are not always the per-point literal times the rank.** Most are: of 125 percent talents, 122
scale linearly, so `0.02 * rank` was already right and the table only adds provenance. The ones that do
not are the reason to read it - Improved Righteous Fury is 16 / 33 / 50, not 16 / 32 / 48, and shaman
Elemental Weapons is 7 / 14 / 20, not 7 / 14 / 21.

### Picking the effect

A talent with one effect per rank needs nothing further. One with several does, and `ValueAt` panics
rather than guess:

```go
genRanks.ImprovedRighteousFury.
    Effect(shared.A_ADD_PCT_MODIFIER, shared.SPELLMOD_ALL_EFFECTS).MultiplierAt(rank)   // 1.50 threat
genRanks.ImprovedRighteousFury.
    Effect(shared.A_ADD_FLAT_MODIFIER, shared.SPELLMOD_EFFECT2).MultiplierAt(rank)      // 0.94 taken
```

**Do not pick the effect by which one matches the number.** Survival of the Fittest states +1/2/3% to
all stats and -1/-2/-3% crit taken; both ladders fit, and an automatic pass attached the stat effect to
the crit-taken call site. What the call site does decides it, and the mod's `Kind` usually says:
Improved Moonfire's two mods are `SpellMod_DamageDone_Flat` and `SpellMod_BonusCrit_Percent`, so one
takes `SPELLMOD_DAMAGE` and the other `SPELLMOD_CRITICAL_CHANCE`.

Where a talent modifies damage and its DoT with the same ladder, the sim has one mod against the
client's two. Either aura reads the same number; `SPELLMOD_DAMAGE` is the convention here.

### The Misc value

`Misc` says what an effect applies to, and what it means depends on the aura: a modified spell property
under `A_ADD_PCT_MODIFIER` and `A_ADD_FLAT_MODIFIER`, a stat under `A_MOD_TOTAL_STAT_PERCENTAGE`, a
school mask under `A_MOD_DAMAGE_DONE`. There is no single enum for it, so it stays an int.

For the two modifier auras the `SPELLMOD_*` constants name it. Those are hand-written in
`sim/common/shared/spell_rank_talents.go`, not mirrored from the client, which ships no name list for
them - each carries the talents it was read off, and the four resting on one or two talents each are
marked as thinner. Audit the comment before trusting the name.

## Worked examples

### Direct damage

```go
var exorcismRanks = genRanks.Exorcism.BySpellID(27138)

func (paladin *Paladin) registerExorcism() {
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: exorcismRanks.SpellID},
		Rank:             exorcismRanks.Rank,
		ManaCost:         core.ManaCostOptions{FlatCost: exorcismRanks.Cost},
		BonusCoefficient: exorcismRanks.Direct.BonusCoefficient(),
		MaxRange:         exorcismRanks.MaxRange,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: exorcismRanks.GCD, CastTime: exorcismRanks.CastTime},
			CD:          core.Cooldown{Timer: paladin.getExorcismTimer(), Duration: exorcismRanks.Cooldown},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, exorcismRanks.Direct.Damage(sim), spell.OutcomeMagicHitAndCrit)
		},
	})
}
```

### A damage-over-time effect

The tick and its schedule both come from the table, so `NumberOfTicks` and `TickLength` stop being
hand-written:

```go
var swpRanks = genRanks.ShadowWordPain.BySpellID(25368)

func (priest *Priest) registerShadowWordPain() {
	tick := swpRanks.Periodic.(shared.SpellRankPeriodic)

	priest.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: swpRanks.SpellID},
		ManaCost: core.ManaCostOptions{FlatCost: swpRanks.Cost},

		Dot: core.DotConfig{
			Aura:             core.Aura{Label: "ShadowWordPain-" + swpRanks.GetRankLabel()},
			NumberOfTicks:    tick.NumberOfTicks,
			TickLength:       tick.TickLength,
			BonusCoefficient: tick.Coef,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, tick.Tick)
			},
		},
	})
}
```

### A heal, and a mana restore

```go
holyLight := genRanks.HolyLight.BySpellID(27136)
low, high := holyLight.Heal.Range()          // both ends in one call

layOnHands := genRanks.LayOnHands.BySpellID(27154)
mana := shared.SpellRankMin(layOnHands.Energize)   // 900; rank 1 has no Energize at all, and reads 0
```

### Registering several ranks

Downranking registers more than one, and the spec chooses which:

```go
// Starfire's ladder runs 1-8; the sim registers only these two.
genRanks.Starfire.Ranks(6, 8).RegisterAll(druid.registerStarfireSpell)

func (druid *Druid) registerStarfireSpell(rank shared.SpellRank) {
	druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ActionID: core.ActionID{SpellID: rank.SpellID},
		Rank:     rank.Rank,
		ManaCost: core.ManaCostOptions{FlatCost: rank.Cost},
		Cast:     core.CastConfig{DefaultCast: core.Cast{GCD: rank.GCD, CastTime: rank.CastTime}},
		// ...
	})
}
```

Each rank is its own registered spell with its own ActionID, which is what lets an APL name a downrank.

## Attack power

Attack power scaling is **not** in the client data - one effect in 38357 carries a nonzero
`BonusCoefficientFromAP` - so melee coefficients live in server script and have to be supplied by hand:

One coefficient for the whole ladder:

```go
// Rupture's ranks are all periodic, so the coefficient goes on the tick.
var ruptureRanks = shared.WithSpellRankPeriodicAPCoef(genRanks.Rupture, 0.18)
```

Or one per rank, the way the spell power coefficient already varies because each row carries its own:

```go
var ruptureRanks = shared.WithSpellRankPeriodicAPCoefs(genRanks.Rupture, map[int32]float64{
	1: 0.04, 2: 0.06, 3: 0.08, 4: 0.10, 5: 0.12, 6: 0.15, 7: 0.18,
})
```

**Every rank in the table has to be named.** Leave one out and it panics rather than scaling that rank
off nothing, and naming a rank the ladder does not have panics too - so a ladder that gains a rank in a
later client build fails loudly instead of quietly mis-scaling.

|                                        |                                |
| -------------------------------------- | ------------------------------ |
| `WithSpellRankAPCoef(t, c)`            | one coefficient, on `Direct`   |
| `WithSpellRankPeriodicAPCoef(t, c)`    | one coefficient, on `Periodic` |
| `WithSpellRankAPCoefs(t, map)`         | per rank, on `Direct`          |
| `WithSpellRankPeriodicAPCoefs(t, map)` | per rank, on `Periodic`        |

All four return a copy, so the generated table keeps what the database said. All four panic if the role
is nil on any rank - check the generated table first, `genRanks.Mangle` is the _learn-spell_ entry
(`Effect = 36`) and carries no value at all - and if the table already carries a coefficient, because a
value that appears upstream should be noticed, not silently shadowed.

## Regenerating

```
go run ./tools/database/gen_spellranks
```

Reads `tools/database/wowsims.db` and rewrites every `sim/<class>/spell_ranks_auto_gen.go`. It is its
own binary rather than a mode of `gen_db` on purpose: `gen_db` imports the sim, and the sim reads these
tables, so a stale generated file would stop the generator that fixes it from compiling.

Nothing lists which spells to generate. The generator walks `dbc.Classes`, takes each class's own skill
lines and emits a table for every spell in them whose subtext reads `Rank N`. A new family appears on
its own; if one you expect is missing, look at the `// Not generated:` comment at the head of the class
file, which names every family that could not be resolved and why.

`go test ./tools/database/ -run GeneratedRankTables` re-derives every value from the database and
compares it to the committed tables, so a hand-edited or stale generated file fails the build. It skips
when `wowsims.db` is absent.

## Traps

**The data contains ranks the game never grants.** Fireball 38692 and Frostbolt 38697 are rank 14
entries at level 70 whose `SkillLineAbility` rows and spell attributes are byte-for-byte
indistinguishable from the real rank 13s, so the generator cannot filter them. Reaching for
`HighestRank()` on those families silently casts a spell that does not exist - it cost 1.2% DPS when it
happened during the mage port. Use `BySpellID`.

**`HighestRank()` is not "the rank my spec casts".** It is the largest rank number present, which is
also not the last element: Flamestrike is declared rank 7 then rank 6.

**A rank can carry nothing in a role.** Lay on Hands rank 1 restores no mana where ranks 2-4 do, so
`rank.Energize` is nil there. The `SpellRank*` helpers read nil as zero; a direct field access does not.

**Never delete a generated file before regenerating it.** The class package stops compiling, and
`gen_db` - which imports the sim - then cannot build either. Regenerate over the top, or
`git checkout HEAD -- <path>` to get back.

**A melee ability states its bonus as weapon damage, not school damage.** Sinister Strike's +98 is
`Effect = 121` (normalised weapon damage); the generator reads 17, 58 and 121 alongside school damage,
so those land in `Direct`. `Effect = 31` (weapon percent damage) is deliberately excluded - it is a
multiplier on the swing, not an amount a rank can carry.

**Rage costs are divided by ten on the way in.** The client stores rage on a 0-1000 bar, so Heroic
Strike's cost reads 150 where the player sees 15. `Cost` is always in the units the sim uses; mana,
energy and focus need no conversion, and only rage does. An `Energize` effect that restores rage would
still be in tenths - nothing generated today does.

**A new client table needs a settings line.** `SpellCastTimes` was missing from
`generator-settings.json` and cast times read zero until it was added and `make db` re-run. Adding a
table is one line; the extractor needs no code.
