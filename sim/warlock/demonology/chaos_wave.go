package demonology

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/warlock"
)

const chaosWaveScale = 1
const chaosWaveCoeff = 1.167

func (demonology *DemonologyWarlock) registerChaosWave() {
	demonology.ChaosWave = demonology.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 124916},
		SpellSchool:    core.SpellSchoolChaos,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAoE | core.SpellFlagAPL,
		ClassSpellMask: warlock.WarlockSpellChaosWave,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		Charges:      2,
		RechargeTime: time.Second * 15,

		DamageMultiplier: 1,
		CritMultiplier:   demonology.DefaultCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: chaosWaveCoeff,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return demonology.IsInMeta() && demonology.CanSpendDemonicFury(80)
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// keep stacks in sync as they're shared
			demonology.HandOfGuldan.ConsumeCharge(sim)
			demonology.SpendDemonicFury(sim, 80, spell.ActionID)
			pa := sim.GetConsumedPendingActionFromPool()
			pa.NextActionAt = sim.CurrentTime + time.Millisecond*1300 // Fixed delay of 1.3 seconds
			pa.Priority = core.ActionPriorityAuto

			pa.OnAction = func(sim *core.Simulation) {
				spell.CalcAndDealAoeDamage(sim, demonology.CalcScalingSpellDmg(chaosWaveScale), spell.OutcomeMagicHitAndCrit)
			}

			sim.AddPendingAction(pa)
		},
	})
}
