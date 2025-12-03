package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const soulfireCoeff = 1.15
const soulfireVariance = 0.2

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
				CastTime: time.Millisecond*6000 - (time.Millisecond * 400 * time.Duration(warlock.Talents.Bane)),
			},
			CD: core.Cooldown{
				Duration: 1 * time.Minute,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   warlock.DefaultCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: soulfireCoeff,
	})

}
