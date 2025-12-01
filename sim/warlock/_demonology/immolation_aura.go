package demonology

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/warlock"
)

const immolationAuraScale = 0.17499999702
const immolationAuraCoeff = 0.17499999702

func (demonology *DemonologyWarlock) registerImmolationAura() {
	var baseDamage = demonology.CalcScalingSpellDmg(immolationAuraScale)

	immolationAura := demonology.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 104025},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAoE | core.SpellFlagAPL | core.SpellFlagNoMetrics,
		ClassSpellMask: warlock.WarlockSpellImmolationAura,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
		},

		DamageMultiplierAdditive: 1,
		CritMultiplier:           demonology.DefaultCritMultiplier(),
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Immolation Aura (DoT)",
			},

			TickLength:           time.Second,
			NumberOfTicks:        8,
			AffectedByCastSpeed:  true,
			HasteReducesDuration: true,
			BonusCoefficient:     immolationAuraCoeff,
			IsAOE:                true,

			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				if !demonology.CanSpendDemonicFury(25) {
					dot.Deactivate(sim)
					return
				}

				demonology.SpendDemonicFury(sim, 25, dot.Spell.ActionID)
				dot.Spell.CalcAndDealPeriodicAoeDamage(sim, baseDamage, dot.OutcomeTick)
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return demonology.IsInMeta() && demonology.CanSpendDemonicFury(25)
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		},
	})

	demonology.Metamorphosis.RelatedSelfBuff.ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
		immolationAura.AOEDot().Deactivate(sim)
	})
}
