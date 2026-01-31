package mage

import "github.com/wowsims/tbc/sim/core"

func (mage *Mage) registerSlowSpell() {
	if !mage.Talents.Slow {
		return
	}

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 31589},
		ClassSpellMask: MageSpellSlow,
		SpellSchool:    core.SpellSchoolArcane,
		Flags:          core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 20,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				aura := mage.SlowAuras.Get(target)
				aura.Activate(sim)
			}
		},
	})
}
