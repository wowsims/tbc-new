package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (mage *Mage) registerFireBlastSpell() {

	if mage.Spec == proto.Spec_SpecFireMage {
		return
	}

	fireBlastVariance := 0.17000000179    // Per https://wago.tools/db2/SpellEffect?build=5.5.0.61217&filter%5BSpellID%5D=exact%253A2136 Field: "Variance"
	fireBlastScaling := 0.78899997473     // Per https://wago.tools/db2/SpellEffect?build=5.5.0.61217&filter%5BSpellID%5D=exact%253A2136 Field: "Coefficient"
	fireBlastCoefficient := 1.01199996471 // Per https://wago.tools/db2/SpellEffect?build=5.5.0.61217&filter%5BSpellID%5D=exact%253A2136 Field: "BonusCoefficient"

	mage.FireBlast = mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 2136},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL | core.SpellFlagAoE,
		ClassSpellMask: MageSpellFireBlast,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 2,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: time.Second * 8,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   mage.DefaultCritMultiplier(),
		BonusCoefficient: fireBlastCoefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := mage.CalcAndRollDamageRange(sim, fireBlastScaling, fireBlastVariance)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}
