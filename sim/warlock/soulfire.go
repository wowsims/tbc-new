package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const soulfireCoeff = 1.15

func (warlock *Warlock) registerSoulfire() {
	warlock.Soulfire = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 6353},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellSoulFire,
		MissileSpeed:   21,

		ManaCost: core.ManaCostOptions{FlatCost: 250},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond*6000 - time.Duration(400*warlock.Talents.Bane),
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: 1 * time.Minute,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   warlock.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: soulfireCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			dmgRoll := warlock.CalcAndRollDamageRange(sim, 1003, 1257)
			result := spell.CalcDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})

}
