package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var corruptionRank = genRanks.Corruption.BySpellID(27216)
var corruptionTick = corruptionRank.Periodic.(shared.SpellRankPeriodic)
var corruptionCoeff = corruptionTick.Coef

func (warlock *Warlock) registerCorruption() *core.Spell {
	tickCount := corruptionTick.Ticks
	warlock.CorruptionTickBaseDamage = corruptionTick.Tick

	warlock.Corruption = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27216},
		SpellSchool:    core.SpellSchoolShadow,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellCorruption,

		DamageMultiplier: 1,
		ManaCost:         core.ManaCostOptions{FlatCost: corruptionRank.Cost},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      corruptionRank.GCD,
				CastTime: corruptionRank.CastTime,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)

			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},
		BonusCoefficient: corruptionCoeff,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Corruption",
				Tag:   "Affliction",
			},
			NumberOfTicks:    tickCount,
			TickLength:       corruptionTick.Period,
			BonusCoefficient: corruptionCoeff,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, warlock.CorruptionTickBaseDamage)
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
				result := spell.CalcPeriodicDamage(sim, target, corruptionTick.Tick*float64(corruptionTick.Ticks), spell.OutcomeExpectedMagicHit)
				result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
				return result
			}
		},
	})

	return warlock.Corruption
}
