package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerScorchSpell() {

	scorchCoefficient := 0.42899999022 // Per https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=exact%253A2948 Field: "BonusCoefficient"
	procChance := []float64{0, 0.33, 0.66, 1}[mage.Talents.ImprovedScorch]

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 2948},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: MageSpellScorch,

		ManaCost: core.ManaCostOptions{
			FlatCost: 180,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 1500,
			},
		},

		DamageMultiplierAdditive: 1,
		CritMultiplier:           mage.DefaultSpellCritMultiplier(),
		BonusCoefficient:         scorchCoefficient,
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := mage.CalcAndRollDamageRange(sim, 304, 361)
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			if result.Landed() && mage.Talents.ImprovedScorch > 0 {
				if sim.Proc(procChance, "Improved Scorch") {
					aura := mage.ImprovedScorchAuras.Get(target)
					aura.Activate(sim)
					aura.AddStack(sim)
				}
			}
		},

		RelatedAuraArrays: mage.ImprovedScorchAuras.ToMap(),
	})
}
