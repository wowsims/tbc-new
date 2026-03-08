package hunter

import (
	"github.com/wowsims/tbc/sim/core"
)

func (hunter *Hunter) registerScorpidStingSpell() {
	auraArray := hunter.NewEnemyAuraArray(func(unit *core.Unit) *core.Aura {
		aura := core.ScorpidStingAura(unit)
		aura.Tag = "Sting"
		return aura
	})

	hunter.ScorpidSting = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 3043},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskProc,
		ClassSpellMask: HunterSpellScorpidSting,
		Flags:          core.SpellFlagAPL,

		MissileSpeed: 40,
		MinRange:     core.MaxMeleeRange,
		MaxRange:     HunterBaseMaxRange,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 9,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		CritMultiplier:   0,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeRangedHit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				if result.Landed() {
					aura := auraArray.Get(target)
					activeSting := target.GetActiveAuraWithTag("Sting")
					if activeSting != nil && activeSting != aura {
						activeSting.Deactivate(sim)
					}
					aura.Activate(sim)
				}
				spell.DealOutcome(sim, result)
			})
		},
	})
}
