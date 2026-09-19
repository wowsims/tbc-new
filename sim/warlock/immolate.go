package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var immolateRank = genRanks.Immolate.BySpellID(27215)
var immolateTick = immolateRank.Periodic.(shared.SpellRankPeriodic)
var immolateCoeff = immolateRank.Direct.BonusCoefficient()
var immolateDotCoeff = immolateTick.Coef

func (warlock *Warlock) registerImmolate() {
	actionID := core.ActionID{SpellID: 27215}
	tickCount := immolateTick.NumberOfTicks
	warlock.ImmolateTickBaseDamage = immolateTick.Tick

	warlock.Immolate = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellImmolate,

		ManaCost: core.ManaCostOptions{
			FlatCost: immolateRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      immolateRank.GCD,
				CastTime: immolateRank.CastTime,
			},
		},

		DamageMultiplier: 1,
		DefenseType:      core.DefenseTypeMagic,
		ThreatMultiplier: 1,
		BonusCoefficient: immolateCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, immolateRank.Direct.Damage(sim), spell.OutcomeMagicHitAndCrit)
			if result.Landed() {
				spell.RelatedDotSpell.Dot(target).Apply(sim)
			}

			spell.DealDamage(sim, result)
		},
	})

	warlock.Immolate.RelatedDotSpell = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(1),
		SpellSchool:    core.SpellSchoolFire,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: WarlockSpellImmolateDot,
		Flags:          core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Immolate (DoT)",
			},
			NumberOfTicks:    tickCount,
			TickLength:       immolateTick.TickLength,
			BonusCoefficient: immolateDotCoeff,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, warlock.ImmolateTickBaseDamage)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			dot := spell.Dot(target)
			if useSnapshot {
				result := dot.CalcSnapshotDamage(sim, target, dot.OutcomeTick)
				result.Damage /= dot.TickPeriod().Seconds()
				return result
			} else {
				result := spell.CalcPeriodicDamage(sim, target, immolateTick.Tick*float64(immolateTick.NumberOfTicks), spell.OutcomeExpectedMagicHit)
				result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
				return result
			}
		},
	})
}
