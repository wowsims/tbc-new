package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const drainLifeCoeff = 0.143

func (warlock *Warlock) registerDrainLife() {
	healthMetric := warlock.NewHealthMetrics(core.ActionID{SpellID: 689})
	resultSlice := make(core.SpellResultSlice, 1)

	warlock.DrainLife = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 689},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagChanneled | core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellDrainLife,

		ManaCost: core.ManaCostOptions{FlatCost: 425},
		Cast:     core.CastConfig{DefaultCast: core.Cast{GCD: core.GCDDefault}},

		DamageMultiplierAdditive: 1,
		CritMultiplier:           warlock.DefaultCritMultiplier(),
		ThreatMultiplier:         1,
		BonusCoefficient:         drainLifeCoeff,

		Dot: core.DotConfig{
			Aura:             core.Aura{Label: "Drain Life"},
			NumberOfTicks:    5,
			TickLength:       1 * time.Second,
			BonusCoefficient: drainLifeCoeff,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, _ bool) {
				dot.Snapshot(target, warlock.CalcScalingSpellDmg(drainLifeCoeff))
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				resultSlice[0] = dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeSnapshotCrit)

				warlock.GainHealth(sim, warlock.MaxHealth()*0.02, healthMetric)

				// if callback != nil {
				// 	callback(resultSlice, dot.Spell, sim)
				// }
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
				spell.DealOutcome(sim, result)
			}
		},
	})
}
