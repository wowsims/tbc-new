package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const shadowBoltCoeff = 1.38

func (warlock *Warlock) registerShadowBolt() {

	warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 686},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellShadowBolt,
		MissileSpeed:   20,

		ManaCost: core.ManaCostOptions{BaseCostPercent: 5.5},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: 2500 * time.Millisecond,
			},
		},

		DamageMultiplierAdditive: 1,
		CritMultiplier:           warlock.DefaultSpellCritMultiplier(),
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, 1000, spell.OutcomeMagicHitAndCrit)
			existingAura := target.GetAurasWithTag("ImprovedShadowBolt")

			if len(existingAura) == 0 || existingAura[0].Duration != core.NeverExpires {
				if result.Landed() && result.Outcome.Matches(core.OutcomeCrit) {
					if !warlock.ImpShadowboltAura.IsActive() {

						warlock.ImpShadowboltAura.Activate(sim)
					}
					warlock.ImpShadowboltAura.SetStacks(sim, 4)
				}
			}

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
