package mage

import (
	"github.com/wowsims/tbc/sim/core"
)

const blastWaveCoefficient = 0.1930000037

var blastWaveRank = genRanks.BlastWave.BySpellID(33933)

func (mage *Mage) registerBlastWaveSpell() {
	if !mage.Talents.BlastWave {
		return
	}

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 33933},
		Flags:          core.SpellFlagAPL,
		SpellSchool:    core.SpellSchoolFire,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: MageSpellBlastWave,

		BonusCoefficient: blastWaveRank.Direct.BonusCoefficient(),
		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ManaCost: core.ManaCostOptions{
			FlatCost: blastWaveRank.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: blastWaveRank.GCD,
			},
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: blastWaveRank.Cooldown,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := blastWaveRank.Direct.Damage(sim)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
			//The above returns a result slice if you want to implement the daze on the targets hit
		},
	})
}
