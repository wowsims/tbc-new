package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const frostboltCoefficient = 0.81400001049 // Per https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=exact%253A38697 Field: "BonusCoefficient"

func (mage *Mage) frostBoltConfig(config core.SpellConfig) core.SpellConfig {
	return core.SpellConfig{
		ActionID:       config.ActionID,
		SpellSchool:    core.SpellSchoolFrost,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          config.Flags,
		ClassSpellMask: MageSpellFrostbolt,
		MissileSpeed:   28,

		ManaCost: config.ManaCost,
		Cast:     config.Cast,

		DamageMultiplier: config.DamageMultiplier,
		CritMultiplier:   mage.DefaultSpellCritMultiplier(),
		BonusCoefficient: frostboltCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: config.ApplyEffects,
	}
}

func (mage *Mage) registerFrostboltSpell() {
	actionID := core.ActionID{SpellID: 116}

	mage.RegisterSpell(mage.frostBoltConfig(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL | core.SpellFlagBinary,

		ManaCost: core.ManaCostOptions{
			FlatCost: 345,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Second * 3,
			},
		},

		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := mage.CalcAndRollDamageRange(sim, 629, 680)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}))
}
