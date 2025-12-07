package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const agonyScale = 0.0255
const agonyCoeff = 0.0255

func (warlock *Warlock) registerCurseOfAgony() {
	warlock.CurseOfAgony = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 980},
		Flags:          core.SpellFlagAPL,
		ProcMask:       core.ProcMaskSpellDamage,
		SpellSchool:    core.SpellSchoolShadow,
		ClassSpellMask: WarlockSpellCurseOfAgony,

		ThreatMultiplier: 1,
		DamageMultiplier: 1,
		BonusCoefficient: agonyCoeff,
		CritMultiplier:   1,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 1,
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Agony",
				Tag:   "Affliction",
			},

			TickLength:          2 * time.Second,
			NumberOfTicks:       12,
			AffectedByCastSpeed: false,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, 1356)
			},

			BonusCoefficient: agonyCoeff,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			dot := spell.Dot(target)

			// Always compare fully stacked agony damage
			if useSnapshot {
				result := dot.CalcSnapshotDamage(sim, target, dot.OutcomeTick)
				result.Damage *= 10
				result.Damage /= dot.TickPeriod().Seconds()
				return result
			} else {
				result := spell.CalcPeriodicDamage(sim, target, warlock.CalcScalingSpellDmg(agonyScale), spell.OutcomeExpectedMagicCrit)
				result.Damage *= 10
				result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
				return result
			}
		},
	})
}
