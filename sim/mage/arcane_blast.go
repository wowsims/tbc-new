package mage

import (
	"github.com/wowsims/tbc/sim/core"
)

var arcaneBlastRank = genRanks.ArcaneBlast.BySpellID(30451)

func (mage *Mage) registerArcaneBlastSpell() {

	//https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=30451
	arcaneBlastCoefficient := 0.71399998665

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 30451},
		SpellSchool:    core.SpellSchoolArcane,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: MageSpellArcaneBlast,
		ManaCost: core.ManaCostOptions{
			FlatCost: arcaneBlastRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      arcaneBlastRank.GCD,
				CastTime: arcaneBlastRank.CastTime,
			},
		},

		DamageMultiplier: 1,
		BonusCoefficient: arcaneBlastCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := arcaneBlastRank.Direct.Damage(sim)
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			if result.Landed() {
				mage.ArcaneChargesAura.Activate(sim)
				mage.ArcaneChargesAura.AddStack(sim)
			}
		},
	})
}
