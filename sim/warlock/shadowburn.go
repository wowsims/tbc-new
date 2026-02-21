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
		ClassSpellMask: WarlockSpellSearingPain,

		ManaCost: core.ManaCostOptions{FlatCost: 515},
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
			dmgRoll := warlock.CalcAndRollDamageRange(sim, 597, 665)
			spell.CalcDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)

		},
	})
}
