package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const conflagrateCoeff = 0.429

func (warlock *Warlock) registerConflagrate() {

	if !warlock.Talents.Conflagrate {
		return
	}

	warlock.Conflagrate = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 17962},
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
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return target.HasActiveAura("Immolate (DoT)-1")
		},

		DamageMultiplier: 1.0,
		CritMultiplier:   warlock.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: conflagrateCoeff,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			//tie this to landed/hit
			dmgRoll := warlock.CalcAndRollDamageRange(sim, 579, 721)
			result := spell.CalcAndDealDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)

			if result.Outcome == core.OutcomeLanded || result.Outcome == core.OutcomePartial {
				target.GetAura("Immolate (DoT)").Deactivate(sim)
			}
		},
	})
}
