package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

// Holy Wrath
// https://www.wowhead.com/tbc/spell=2812/holy-wrath
//
// Sends bolts of holy power in all directions, causing Holy damage
// to all Undead and Demon targets within 20 yds.
func (paladin *Paladin) registerHolyWrath() {
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
		{level: 50, spellID: 2812, manaCost: 550, minValue: 362, maxValue: 428, coeff: 0.286, scaleLevel: 54, scalingCoeff: 1.60},
		{level: 60, spellID: 10318, manaCost: 700, minValue: 490, maxValue: 576, coeff: 0.286, scaleLevel: 64, scalingCoeff: 1.90},
		{level: 69, spellID: 27139, manaCost: 805, minValue: 635, maxValue: 745, coeff: 0.286, scaleLevel: 73, scalingCoeff: 2.20},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Minute,
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		minDamage := ranks[rank].minValue + ranks[rank].scalingCoeff*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)
		maxDamage := ranks[rank].maxValue + ranks[rank].scalingCoeff*float64(min(paladin.Level, ranks[rank].scaleLevel)-ranks[rank].level)

		paladin.HolyWraths = append(paladin.HolyWraths, paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskSpellDamage,
			Flags:          core.SpellFlagAPL,
			ClassSpellMask: SpellMaskHolyWrath,

			ManaCost: core.ManaCostOptions{
				FlatCost: ranks[rank].manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD:      core.GCDDefault,
					CastTime: time.Second * 2,
				},
				CD: cd,
			},

			BonusCoefficient: ranks[rank].coeff,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := sim.Roll(minDamage, maxDamage)
				for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
					if aoeTarget.MobType == proto.MobType_MobTypeUndead || aoeTarget.MobType == proto.MobType_MobTypeDemon {
						spell.CalcAndDealDamage(sim, aoeTarget, baseDamage, spell.OutcomeMagicHitAndCrit)
					}
				}
			},
		}))
	}
}
