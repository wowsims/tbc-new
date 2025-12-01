package demonology

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/warlock"
)

const voidRayScale = 0.525
const voidRayVariance = 0.1
const voidRayCoeff = 0.234

func (demonology *DemonologyWarlock) registerVoidRay() {
	demonology.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 115422},
		SpellSchool:    core.SpellSchoolShadowFlame,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAoE | core.SpellFlagAPL,
		ClassSpellMask: warlock.WarlockSpellVoidray,
		MissileSpeed:   38,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		DamageMultiplier: 1.0,
		CritMultiplier:   demonology.DefaultCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: voidRayCoeff,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return demonology.IsInMeta() && demonology.CanSpendDemonicFury(80)
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			demonology.SpendDemonicFury(sim, 80, spell.ActionID)
			spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, func(sim *core.Simulation, _ *core.Spell) float64 {
				return demonology.CalcAndRollDamageRange(sim, voidRayScale, voidRayVariance)
			})
		},
	})
}
