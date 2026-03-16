package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (hunter *Hunter) registerArcaneShotSpell() {
	hunter.ArcaneShot = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27019},
		SpellSchool:    core.SpellSchoolArcane,
		ClassSpellMask: HunterSpellArcaneShot,
		ProcMask:       core.ProcMaskRangedSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		MissileSpeed: 40,
		MinRange:     core.MaxMeleeRange,
		MaxRange:     HunterBaseMaxRange,

		ManaCost: core.ManaCostOptions{
			FlatCost: 230,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 6,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   hunter.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.RangedAttackPower(target)*0.15 +
				hunter.talonOfAlarBonus() +
				273

			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
