package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

// Hammer of Wrath
// https://www.wowhead.com/tbc/spell=27180
//
// Hurls a hammer that strikes an enemy for Holy damage.
// Only usable on enemies that have 20% or less health.
func (paladin *Paladin) registerHammerOfWrath() {
	var ranks = []struct {
		level        int32
		spellID      int32
		manaCost     int32
		minValue     float64
		maxValue     float64
		coeff        float64
		scaleLevel   int32
		scalingCoeff float64
	}{
		{},
		{level: 44, spellID: 24275, manaCost: 235, minValue: 304, maxValue: 336, coeff: 0.429, scaleLevel: 49, scalingCoeff: 2.40},
		{level: 52, spellID: 24274, manaCost: 290, minValue: 399, maxValue: 441, coeff: 0.429, scaleLevel: 57, scalingCoeff: 2.70},
		{level: 60, spellID: 24239, manaCost: 340, minValue: 504, maxValue: 556, coeff: 0.429, scaleLevel: 65, scalingCoeff: 3.10},
		{level: 68, spellID: 27180, manaCost: 440, minValue: 665, maxValue: 735, coeff: 0.429, scaleLevel: 73, scalingCoeff: 3.50},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 6,
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		minDamage := ranks[rank].minValue + ranks[rank].scalingCoeff*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)
		maxDamage := ranks[rank].maxValue + ranks[rank].scalingCoeff*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)

		hammerOfWrath := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskRangedSpecial,
			Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHammerOfWrath,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: 30,

			ManaCost: core.ManaCostOptions{
				FlatCost: ranks[rank].manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      time.Millisecond * 500,
					CastTime: time.Millisecond * 500,
				},
				CD: cd,
			},

			BonusCoefficient: ranks[rank].coeff,
			CritMultiplier:   paladin.DefaultMeleeCritMultiplier(),

			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return sim.IsExecutePhase20()
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				result := spell.CalcDamage(sim, target, sim.Roll(minDamage, maxDamage), spell.OutcomeMeleeSpecialHitAndCrit)
				spell.DealDamage(sim, result)
			},
		})

		paladin.HammerOfWraths = append(paladin.HammerOfWraths, hammerOfWrath)
	}
}
