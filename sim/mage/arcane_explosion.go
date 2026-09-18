package mage

import (
	"github.com/wowsims/tbc/sim/core"
)

var arcaneExplosionRank = genRanks.ArcaneExplosion.BySpellID(27082)

func (mage *Mage) registerArcaneExplosionSpell() {
	arcaneExplosionCoefficient := 0.21400000155

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27082},
		SpellSchool:    core.SpellSchoolArcane,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: MageSpellArcaneExplosion,

		ManaCost: core.ManaCostOptions{
			FlatCost: arcaneExplosionRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: arcaneExplosionRank.GCD,
			},
		},

		DamageMultiplier: 1,
		BonusCoefficient: arcaneExplosionCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := arcaneExplosionRank.Direct.Damage(sim)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}
