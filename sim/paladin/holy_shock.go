package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var HolyShockRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 20473, Cost: 335, MinDamage: 277, MaxDamage: 299, Coefficient: 0.429},
	{Rank: 2, SpellID: 20929, Cost: 410, MinDamage: 379, MaxDamage: 409, Coefficient: 0.429},
	{Rank: 3, SpellID: 20930, Cost: 485, MinDamage: 496, MaxDamage: 628, Coefficient: 0.429},
	{Rank: 4, SpellID: 27174, Cost: 575, MinDamage: 614, MaxDamage: 664, Coefficient: 0.429},
	{Rank: 5, SpellID: 33072, Cost: 650, MinDamage: 721, MaxDamage: 779, Coefficient: 0.429},
}

// Holy Shock
// https://www.wowhead.com/tbc/spell=20473
//
// Blasts the target with Holy energy, causing X to Y Holy damage to an enemy,
// or X*1.267 to Y*1.267 healing to an ally.
func (paladin *Paladin) registerHolyShock(rankConfig shared.SpellRankConfig) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := rankConfig.MinDamage
	maxDamage := rankConfig.MaxDamage
	coefficient := rankConfig.Coefficient

	// Holy Shock heals for 1.267x the damage component of the spell.
	healingCoeff := 1.267

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 15,
	}

	holyShock := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHolyShock,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultSpellCritMultiplier(),

		MaxRange: 20,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: cd,
		},

		BonusCoefficient: coefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if target.IsOpponent(target) {
				damage := sim.Roll(minDamage, maxDamage)
				spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
			} else {
				// Temporarily configure the spell as a healing spell
				spell.Flags |= core.SpellFlagHelpful
				originalProcMask := spell.ProcMask
				spell.ProcMask = core.ProcMaskSpellHealing

				// TODO: Use healing power instead of holy power for healing calculations
				healing := sim.Roll(minDamage, maxDamage) * healingCoeff
				spell.CalcAndDealHealing(sim, target, healing, spell.OutcomeHealingCrit)

				// Reset the spell to its original configuration
				spell.Flags &= ^core.SpellFlagHelpful
				spell.ProcMask = originalProcMask
			}
		},
	})

	paladin.HolyShocks = append(paladin.HolyShocks, holyShock)
}
