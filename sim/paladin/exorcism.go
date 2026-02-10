package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

// Exorcism
// https://www.wowhead.com/tbc/spell=10314
//
// Causes X to Y Holy damage to an Undead or Demon target.
func (paladin *Paladin) registerExorcism() {
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
		{level: 20, spellID: 879, 	manaCost: 70,  minValue: 84,  maxValue: 96,  coeff: 0.429, scaleLevel: 25, scalingCoeff: 1.20},
		{level: 28, spellID: 5614, 	manaCost: 115, minValue: 152, maxValue: 172, coeff: 0.429, scaleLevel: 33, scalingCoeff: 1.60},
		{level: 36, spellID: 5615, 	manaCost: 155, minValue: 217, maxValue: 245, coeff: 0.429, scaleLevel: 41, scalingCoeff: 2.00},
		{level: 44, spellID: 10312, manaCost: 200, minValue: 304, maxValue: 342, coeff: 0.429, scaleLevel: 49, scalingCoeff: 2.40},
		{level: 52, spellID: 10313, manaCost: 240, minValue: 393, maxValue: 439, coeff: 0.429, scaleLevel: 57, scalingCoeff: 2.80},
		{level: 60, spellID: 10314, manaCost: 295, minValue: 505, maxValue: 563, coeff: 0.429, scaleLevel: 65, scalingCoeff: 3.20},
		{level: 68, spellID: 27138, manaCost: 340, minValue: 619, maxValue: 691, coeff: 0.429, scaleLevel: 73, scalingCoeff: 3.50},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 15,
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		minDamage := ranks[rank].minValue + ranks[rank].scalingCoeff*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)
		maxDamage := ranks[rank].maxValue + ranks[rank].scalingCoeff*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)

		exorcism := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskSpellDamage,
			Flags:          core.SpellFlagAPL,
			ClassSpellMask: SpellMaskExorcism,

			ManaCost: core.ManaCostOptions{
				FlatCost: ranks[rank].manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
				},
				CD: cd,
			},

			MaxRange: 30,
			BonusCoefficient: ranks[rank].coeff,

			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return target.MobType == proto.MobType_MobTypeUndead || target.MobType == proto.MobType_MobTypeDemon
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				result := spell.CalcDamage(sim, target, sim.Roll(minDamage, maxDamage), spell.OutcomeMagicHitAndCrit)
				spell.DealDamage(sim, result)
			},
		})

		paladin.Exorcisms = append(paladin.Exorcisms, exorcism)
	}
}
