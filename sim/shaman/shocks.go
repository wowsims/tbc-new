package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var earthShockRank = genRanks.EarthShock.BySpellID(25454)
var flameShockRank = genRanks.FlameShock.BySpellID(25457)
var frostShockRank = genRanks.FrostShock.BySpellID(25464)

// Shared logic for all shocks.
func (shaman *Shaman) newShockSpellConfig(spellID int32, spellSchool core.SpellSchool, baseFlatCost int32, shockTimer *core.Timer, bonusCoefficient float64, gcd time.Duration, cooldown time.Duration, maxRange float64) core.SpellConfig {
	actionID := core.ActionID{SpellID: spellID}

	return core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: spellSchool,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShamanSpell | SpellFlagShock | core.SpellFlagAPL | SpellFlagInstant,
		MaxRange:    maxRange,

		ManaCost: core.ManaCostOptions{
			FlatCost: baseFlatCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: gcd,
			},
			CD: core.Cooldown{
				Timer:    shockTimer,
				Duration: cooldown,
			},
		},

		DamageMultiplier: 1,
		BonusCoefficient: bonusCoefficient,
		ThreatMultiplier: 1,
	}
}

func (shaman *Shaman) registerEarthShockSpell(shockTimer *core.Timer) {
	config := shaman.newShockSpellConfig(earthShockRank.SpellID, core.SpellSchoolNature, earthShockRank.Cost, shockTimer, earthShockRank.Direct.BonusCoefficient(), earthShockRank.GCD, earthShockRank.Cooldown, earthShockRank.MaxRange)
	config.ClassSpellMask = SpellMaskEarthShock
	config.Flags |= core.SpellFlagBinary
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := earthShockRank.Direct.Damage(sim)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	}

	shaman.EarthShock = shaman.RegisterSpell(config)
}

func (shaman *Shaman) registerFlameShockSpell(shockTimer *core.Timer) {
	tick := flameShockRank.Periodic.(shared.SpellRankPeriodic)

	config := shaman.newShockSpellConfig(flameShockRank.SpellID, core.SpellSchoolFire, flameShockRank.Cost, shockTimer, flameShockRank.Direct.BonusCoefficient(), flameShockRank.GCD, flameShockRank.Cooldown, flameShockRank.MaxRange)
	config.ClassSpellMask = SpellMaskFlameShockDirect
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := flameShockRank.Direct.Damage(sim)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			spell.RelatedDotSpell.Cast(sim, target)
		}
		spell.DealDamage(sim, result)
	}
	shaman.FlameShock = shaman.RegisterSpell(config)

	shaman.FlameShock.RelatedDotSpell = shaman.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: flameShockRank.SpellID, Tag: 1},
		SpellSchool:      core.SpellSchoolFire,
		DefenseType:      core.DefenseTypeMagic,
		ProcMask:         core.ProcMaskSpellDamage,
		Flags:            config.Flags & ^core.SpellFlagAPL | core.SpellFlagPassiveSpell,
		ClassSpellMask:   SpellMaskFlameShockDot,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Flame Shock",
			},
			NumberOfTicks:    tick.Ticks,
			TickLength:       tick.Period,
			BonusCoefficient: tick.Coef,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, tick.Tick)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Apply(sim)
		},
		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			dot := spell.Dot(target)
			if useSnapshot {
				result := dot.CalcSnapshotDamage(sim, target, dot.OutcomeTick)
				result.Damage /= dot.TickPeriod().Seconds()
				return result
			} else {
				result := spell.CalcPeriodicDamage(sim, target, tick.Tick, spell.OutcomeExpectedMagicCrit)
				result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
				return result
			}
		},
	})
}

func (shaman *Shaman) registerFrostShockSpell(shockTimer *core.Timer) {
	config := shaman.newShockSpellConfig(frostShockRank.SpellID, core.SpellSchoolFrost, frostShockRank.Cost, shockTimer, frostShockRank.Direct.BonusCoefficient(), frostShockRank.GCD, frostShockRank.Cooldown, frostShockRank.MaxRange)
	config.ClassSpellMask = SpellMaskFrostShock
	config.Flags |= core.SpellFlagBinary
	config.ThreatMultiplier *= 2
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := frostShockRank.Direct.Damage(sim)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	}

	shaman.FrostShock = shaman.RegisterSpell(config)
}

func (shaman *Shaman) registerShocks() {
	shockTimer := shaman.NewTimer()
	shaman.registerEarthShockSpell(shockTimer)
	shaman.registerFlameShockSpell(shockTimer)
	shaman.registerFrostShockSpell(shockTimer)
}
