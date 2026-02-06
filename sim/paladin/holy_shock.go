package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)


// Holy Shock
// https://www.wowhead.com/tbc/spell=20473
// 
// Blasts the target with Holy energy, causing X to Y Holy damage to an enemy, or X*1.267 to Y*1.267 healing to an ally.
func (paladin *Paladin) registerHolyShock() {
	var ranks = []struct {
		level        int32
		spellID      int32
		manaCost     int32
		minValue     float64
		maxValue     float64
		coeff        float64
	}{
		{},
		{level: 40, spellID: 20473, manaCost: 335, minValue: 277, maxValue: 299, coeff: 0.429},
		{level: 48, spellID: 20929, manaCost: 410, minValue: 379, maxValue: 409, coeff: 0.429},
		{level: 56, spellID: 20930, manaCost: 485, minValue: 496, maxValue: 628, coeff: 0.429},
		{level: 64, spellID: 27174, manaCost: 575, minValue: 614, maxValue: 664, coeff: 0.429},
		{level: 70, spellID: 33072, manaCost: 650, minValue: 721, maxValue: 779, coeff: 0.429},
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		// Holy Shock heals for 1.267x the damage component of the spell.
		healingCoeff := 1.267

		holyShock := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskSpellDamage,
			Flags:          core.SpellFlagAPL,
			ClassSpellMask: SpellMaskHolyShock,

			MaxRange: 20,

			ManaCost: core.ManaCostOptions{
				FlatCost: ranks[rank].manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: core.Cooldown{
					Timer:    paladin.NewTimer(),
					Duration: time.Second * 15,
				},
			},

			BonusCoefficient: ranks[rank].coeff,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				if target.IsOpponent(target) {
					damage := sim.Roll(ranks[rank].minValue, ranks[rank].maxValue)
					spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
				} else {
					// Temporarily configure the spell as a healing spell
					spell.Flags |= core.SpellFlagHelpful
					originalProcMask := spell.ProcMask
					spell.ProcMask = core.ProcMaskSpellHealing

					// TODO: Use healing power instead of holy power for healing calculations
					healing := sim.Roll(ranks[rank].minValue, ranks[rank].maxValue) * healingCoeff
					spell.CalcAndDealHealing(sim, target, healing, spell.OutcomeHealingCrit)

					// Reset the spell to its original configuration
					spell.Flags &= ^core.SpellFlagHelpful
					spell.ProcMask = originalProcMask
				}
			},
		})

		paladin.HolyShocks = append(paladin.HolyShocks, holyShock)
	}
}
