package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerArcaneBlastSpell() {

	//https://wago.tools/db2/SpellEffect?build=5.5.0.60802&filter%5BSpellID%5D=30451
	arcaneBlastCoefficient := 0.71399998665

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 30451},
		SpellSchool:    core.SpellSchoolArcane,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: MageSpellArcaneBlast,
		ManaCost: core.ManaCostOptions{
			FlatCost: 195,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 2500,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   mage.DefaultSpellCritMultiplier(),
		BonusCoefficient: arcaneBlastCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := mage.CalcAndRollDamageRange(sim, 668, 772)
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			if result.Landed() {
				mage.ArcaneChargesAura.Activate(sim)
				mage.ArcaneChargesAura.AddStack(sim)
			}
		},
	})
}
