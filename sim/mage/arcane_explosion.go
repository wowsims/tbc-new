package mage

import (
	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerArcaneExplosionSpell() {
	arcaneExplosionCoefficient := 0.21400000155

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 1449},
		SpellSchool:    core.SpellSchoolArcane,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: MageSpellArcaneExplosion,

		ManaCost: core.ManaCostOptions{
			FlatCost: 545,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   mage.DefaultSpellCritMultiplier(),
		BonusCoefficient: arcaneExplosionCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := mage.CalcAndRollDamageRange(sim, 377, 407)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}
