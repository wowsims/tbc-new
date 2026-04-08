package druid

import (
	"github.com/wowsims/tbc/sim/core"
)

func (druid *Druid) registerSwipeBearSpell() {
	druid.Swipe = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26997},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellSwipe,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost:   20,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		CritMultiplier:   druid.FeralCritMultiplier(),
		ThreatMultiplier: 1.75,
		MaxRange:         core.MaxMeleeRange,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range druid.Env.Encounter.AllTargetUnits {
				baseDamage := 108 + druid.IdolSwipeBonus + 0.07*spell.MeleeAttackPower(aoeTarget)
				spell.CalcAndDealDamage(sim, aoeTarget, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			}
		},
	})
}
