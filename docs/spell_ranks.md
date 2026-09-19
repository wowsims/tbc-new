# Spell Ranks

Every ranked spell in the game has a table generated from the client database, checked in at
`sim/<class>/spell_ranks_auto_gen.go`. A spell reads its numbers from that table instead of carrying
hand-transcribed literals.

- [Using a rank](#using-a-rank)
- [The value shapes](#the-value-shapes)
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
shared.SpellRankPeriodic {Tick, TickLength, NumberOfTicks, Coef, APCoef}  // a tick and its schedule
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
tickDamage := rank.Periodic.Damage(sim)   // a tick is already the answer, so nothing is rolled
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
| `Cost`, `CostPct`             | in the units the sim uses - see the rage trap below                                                                                        |
| `CastTime`, `GCD`, `Cooldown` | zero for a channel, whose duration carries it                                                                                              |
| `MinRange`, `MaxRange`        | `core` gates the cast on both; zero means ungated. `MinRange` is the dead zone on a charge, and is nonzero on only 212 spells in the build |
| `MissileSpeed`                | yards per second, which `core` turns into the delay before the damage lands. Zero is an instant hit                                        |

`MissileSpeed` is the one to be careful with: giving a spell a speed it did not have delays its damage
and moves goldens, so check the sim is not already modelling it elsewhere. Arcane Missiles is the case
to know - the channel carries no speed because the missile spell does, and that one is not a ranked row.

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

All four return a copy and panic if the role is nil on any rank, so check the generated table first -
`genRanks.Mangle`, for instance, is the _learn-spell_ entry (`Effect = 36`) and carries no value at
all.

Both return a copy, so the generated table keeps what the database said, and both panic if that table
already carries a coefficient - a value that appears upstream should be noticed, not silently shadowed.

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

`go test ./tools/database/ -run SpellRankCalibration` re-derives every value from the database and
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
