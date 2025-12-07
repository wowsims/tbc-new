package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

var shadowBurnCoeff = 0.429

func (warlock *Warlock) registerShadowBurn() {

	warlock.Shadowburn = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 17877},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellShadowBurn,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: time.Second * 15,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   warlock.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: shadowBurnCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, warlock.CalcScalingSpellDmg(shadowBurnCoeff), spell.OutcomeMagicHitAndCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
