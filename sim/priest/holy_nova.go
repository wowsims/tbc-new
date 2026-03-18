package priest

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var HolyNovaRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 15237, Cost: 185, MinDamage: 29, MaxDamage: 34, Coefficient: 0.161},
	{Rank: 2, SpellID: 15430, Cost: 290, MinDamage: 52, MaxDamage: 61, Coefficient: 0.161},
	{Rank: 3, SpellID: 15431, Cost: 400, MinDamage: 79, MaxDamage: 92, Coefficient: 0.161},
	{Rank: 4, SpellID: 27799, Cost: 520, MinDamage: 110, MaxDamage: 127, Coefficient: 0.161},
	{Rank: 5, SpellID: 27800, Cost: 635, MinDamage: 146, MaxDamage: 168, Coefficient: 0.161},
	{Rank: 6, SpellID: 27801, Cost: 750, MinDamage: 188, MaxDamage: 217, Coefficient: 0.161},
	{Rank: 7, SpellID: 25331, Cost: 875, MinDamage: 244, MaxDamage: 283, Coefficient: 0.161},
}

func (priest *Priest) registerHolyNovaSpell(rankConfig shared.SpellRankConfig) {
	spell := priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellHolyNova,
		Rank:           rankConfig.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           priest.DefaultSpellCritMultiplier(),
		BonusCoefficient:         rankConfig.Coefficient,
		ThreatMultiplier:         0,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := priest.CalcAndRollDamageRange(sim, rankConfig.MinDamage, rankConfig.MaxDamage)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)

			baseHeal := priest.CalcAndRollDamageRange(sim, rankConfig.MinDamage, rankConfig.MaxDamage)
			spell.CalcAndDealHealing(sim, spell.Unit, baseHeal, spell.OutcomeHealing)
		},
	})

	priest.HolyNova = append(priest.HolyNova, spell)
}
