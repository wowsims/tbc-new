package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerArcaneMissilesSpell() {
	// Values found at https://wago.tools/db2/SpellEffect?build=5.5.0.60802&filter%5BSpellID%5D=exact%253A7268
	arcaneMissilesCoefficient := 0.28600001335
	actionID := core.ActionID{SpellID: 7268}

	arcaneMissilesTickSpell := mage.GetOrRegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(1),
		SpellSchool:    core.SpellSchoolArcane,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: MageSpellArcaneMissilesTick,
		MissileSpeed:   20,

		DamageMultiplier: 1,
		CritMultiplier:   mage.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: arcaneMissilesCoefficient,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, 286, spell.OutcomeTickMagicHitAndCrit)
			spell.SpellMetrics[result.Target.UnitIndex].Casts--
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})

	mage.RegisterSpell(core.SpellConfig{
		ActionID:         actionID, // Real SpellID: 5143
		SpellSchool:      core.SpellSchoolArcane,
		ProcMask:         core.ProcMaskSpellDamage,
		Flags:            core.SpellFlagChanneled | core.SpellFlagAPL,
		ClassSpellMask:   MageSpellArcaneMissilesCast,
		DamageMultiplier: 0,

		ManaCost: core.ManaCostOptions{
			FlatCost: 785,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "ArcaneMissiles",
			},
			NumberOfTicks:        5,
			TickLength:           time.Second,
			HasteReducesDuration: true,
			AffectedByCastSpeed:  true,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				arcaneMissilesTickSpell.Cast(sim, target)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
			//Casts is out here and not in Landed() to refer to the number of attempted casts
			arcaneMissilesTickSpell.SpellMetrics[target.UnitIndex].Casts++
			if result.Landed() {
				spell.Dot(target).Apply(sim)
				arcaneMissilesTickSpell.SpellMetrics[target.UnitIndex].Hits++
			}
		},
	})
}
