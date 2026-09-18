package druid

import (
	"github.com/wowsims/tbc/sim/core"
)

var swipeRank = genRanks.Swipe.BySpellID(26997)

func (druid *Druid) registerSwipeBearSpell() {
	druid.Swipe = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: swipeRank.SpellID},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellSwipe,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost:   swipeRank.Cost,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: swipeRank.GCD,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		MaxRange:         core.MaxMeleeRange,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			numHits := min(3, len(druid.Env.Encounter.AllTargetUnits))
			for i := 0; i < numHits; i++ {
				aoeTarget := druid.Env.Encounter.AllTargetUnits[i]
				baseDamage := swipeRank.Direct.Damage(sim) + druid.IdolSwipeBonus + 0.07*spell.MeleeAttackPower(aoeTarget)
				spell.CalcAndDealDamage(sim, aoeTarget, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			}
		},
	})
}
