package arcane

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/mage"
)

func (arcane *ArcaneMage) registerArcaneBlastSpell() {

	//https://wago.tools/db2/SpellEffect?build=5.5.0.60802&filter%5BSpellID%5D=30451
	arcaneBlastVariance := 0.15000000596
	arcaneBlastCoefficient := 0.77700001001
	arcaneBlastScaling := 0.77700001001

	arcane.RegisterSpell(core.SpellConfig{

		ActionID:       core.ActionID{SpellID: 30451},
		SpellSchool:    core.SpellSchoolArcane,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: mage.MageSpellArcaneBlast,
		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 1.666667,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 2000,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   arcane.DefaultCritMultiplier(),
		BonusCoefficient: arcaneBlastCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := arcane.CalcAndRollDamageRange(sim, arcaneBlastScaling, arcaneBlastVariance)
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			if result.Landed() {
				arcane.ArcaneChargesAura.Activate(sim)
				arcane.ArcaneChargesAura.AddStack(sim)
			}
		},
	})
}
