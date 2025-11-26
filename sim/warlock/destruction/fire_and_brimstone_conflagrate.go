package destruction

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/warlock"
)

func (destruction *DestructionWarlock) registerFireAndBrimstoneConflagrate() {
	destruction.FABConflagrate = destruction.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 108685},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAoE | core.SpellFlagAPL,
		ClassSpellMask: warlock.WarlockSpellFaBConflagrate,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		Charges:      2,
		RechargeTime: time.Second * 12,

		DamageMultiplier: 1,
		CritMultiplier:   destruction.DefaultCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: conflagrateCoeff,
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return destruction.BurningEmbers.CanSpend(10)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if !destruction.FABAura.IsActive() {
				destruction.FABAura.Activate(sim)
			}

			// reduce damage for this spell based on mastery
			reduction := destruction.getFABReduction()
			spell.DamageMultiplier *= reduction

			// keep charges in sync
			destruction.Conflagrate.ConsumeCharge(sim)
			for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
				result := spell.CalcAndDealDamage(
					sim,
					aoeTarget,
					destruction.CalcAndRollDamageRange(sim, conflagrateScale, conflagrateVariance),
					spell.OutcomeMagicHitAndCrit)

				var emberGain int32 = 1

				// ember lottery
				if sim.Proc(0.15, "Ember Lottery") {
					emberGain *= 2
				}

				if result.DidCrit() {
					emberGain += 1
				}

				destruction.BurningEmbers.Gain(sim, float64(emberGain), spell.ActionID)
			}
			spell.DamageMultiplier /= reduction
			destruction.BurningEmbers.Spend(sim, 10, spell.ActionID)
		},
	})
}
