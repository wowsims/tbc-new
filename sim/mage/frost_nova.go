package mage

import (
	"github.com/wowsims/tbc/sim/core"
)

var frostNovaRank = genRanks.FrostNova.BySpellID(27088)

func (mage *Mage) registerFrostNovaSpell() {

	frostNovaCoefficient := 0.18799999356 // Per https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=exact%253A122 Field "EffetBonusCoefficient"

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27088},
		SpellSchool:    core.SpellSchoolFrost,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL | core.SpellFlagBinary,
		ClassSpellMask: MageSpellFrostNova,

		ManaCost: core.ManaCostOptions{
			FlatCost: frostNovaRank.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: frostNovaRank.GCD,
			},
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: frostNovaRank.Cooldown,
			},
		},

		DamageMultiplier: 1,
		BonusCoefficient: frostNovaCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			baseDamage := frostNovaRank.Direct.Damage(sim)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}
