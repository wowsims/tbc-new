package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const conflagrateCoeff = 0.429

func (warlock *Warlock) registerConflagrate() {
	warlock.Conflagrate = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 30912},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellConflagrate,

		ManaCost: core.ManaCostOptions{FlatCost: 305},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Duration: time.Second * 10,
			},
		},
		DamageMultiplier: 1.0,
		CritMultiplier:   warlock.DefaultCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: conflagrateCoeff,
		RechargeTime:     time.Second * 10,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			//tie this to landed/hit
			target.GetAura("Immolate (DoT)").Deactivate(sim)

		},
	})
}
