package mage

import (
	"github.com/wowsims/tbc/sim/core"
)

const frostboltCoefficient = 0.81400001049 // Per https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=exact%253A38697 Field: "BonusCoefficient"

func (mage *Mage) frostBoltConfig(config core.SpellConfig) core.SpellConfig {
	return core.SpellConfig{
		ActionID:       config.ActionID,
		SpellSchool:    core.SpellSchoolFrost,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          config.Flags,
		ClassSpellMask: MageSpellFrostbolt,
		MissileSpeed:   28,

		ManaCost: config.ManaCost,
		Cast:     config.Cast,

		DamageMultiplier: config.DamageMultiplier,
		BonusCoefficient: frostboltRank.Direct.BonusCoefficient(),
		ThreatMultiplier: 1,

		ApplyEffects: config.ApplyEffects,
	}
}

var frostboltRank = genRanks.Frostbolt.BySpellID(27072)

func (mage *Mage) registerFrostboltSpell() {
	actionID := core.ActionID{SpellID: 27072}

	mage.RegisterSpell(mage.frostBoltConfig(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL | core.SpellFlagBinary,

		ManaCost: core.ManaCostOptions{
			FlatCost: frostboltRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      frostboltRank.GCD,
				CastTime: frostboltRank.CastTime,
			},
		},

		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := frostboltRank.Direct.Damage(sim)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}))
}
