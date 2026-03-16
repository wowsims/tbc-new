package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (warlock *Warlock) registerDeathCoil() {

	warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27223},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellDeathCoil,
		MissileSpeed:   24,
		MaxRange:       30,

		ManaCost: core.ManaCostOptions{FlatCost: 600},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: time.Minute * 2,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   warlock.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 0.214,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 526, spell.OutcomeMagicHit)
		},
	})
}
