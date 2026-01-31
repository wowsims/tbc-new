package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const blastWaveCoefficient = 0.1930000037

func (mage *Mage) registerBlastWaveSpell() {
	if !mage.Talents.BlastWave {
		return
	}

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 11113},
		Flags:          core.SpellFlagAPL,
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: MageSpellBlastWave,

		BonusCoefficient: blastWaveCoefficient,
		DamageMultiplier: 1,
		CritMultiplier:   mage.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,

		ManaCost: core.ManaCostOptions{
			FlatCost: 620,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: time.Second * 30,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := mage.CalcAndRollDamageRange(sim, 616, 724)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
			//The above returns a result slice if you want to implement the daze on the targets hit
		},
	})
}
