package mage

import (
	"github.com/wowsims/tbc/sim/core"
)

var fireBlastRank = genRanks.FireBlast.BySpellID(27079)

func (mage *Mage) registerFireBlastSpell() {

	mage.FireBlast = mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27079},
		SpellSchool:    core.SpellSchoolFire,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: MageSpellFireBlast,

		ManaCost: core.ManaCostOptions{
			FlatCost: fireBlastRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: fireBlastRank.GCD,
			},
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: fireBlastRank.Cooldown,
			},
		},

		DamageMultiplier: 1,
		BonusCoefficient: fireBlastRank.Direct.BonusCoefficient(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := fireBlastRank.Direct.Damage(sim)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}
