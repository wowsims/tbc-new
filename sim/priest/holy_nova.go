package priest

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var HolyNovaRankMap = shared.RankTable{
	{Rank: 1, SpellID: 15237, Cost: 185, Direct: &shared.Amount{Min: 29, Max: 34, Coef: 0.161}},
	{Rank: 2, SpellID: 15430, Cost: 290, Direct: &shared.Amount{Min: 52, Max: 61, Coef: 0.161}},
	{Rank: 3, SpellID: 15431, Cost: 400, Direct: &shared.Amount{Min: 79, Max: 92, Coef: 0.161}},
	{Rank: 4, SpellID: 27799, Cost: 520, Direct: &shared.Amount{Min: 110, Max: 127, Coef: 0.161}},
	{Rank: 5, SpellID: 27800, Cost: 635, Direct: &shared.Amount{Min: 146, Max: 168, Coef: 0.161}},
	{Rank: 6, SpellID: 27801, Cost: 750, Direct: &shared.Amount{Min: 188, Max: 217, Coef: 0.161}},
	{Rank: 7, SpellID: 25331, Cost: 875, Direct: &shared.Amount{Min: 244, Max: 283, Coef: 0.161}},
}

func (priest *Priest) registerHolyNovaSpell(rank shared.RankRow) {
	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.SpellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellHolyNova,
		Rank:           rank.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rank.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		BonusCoefficient:         rank.Direct.Coef,
		ThreatMultiplier:         0,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := priest.CalcAndRollDamageRange(sim, rank.Direct.Min, rank.Direct.Max)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)

			baseHeal := priest.CalcAndRollDamageRange(sim, rank.Direct.Min, rank.Direct.Max)
			spell.CalcAndDealHealing(sim, spell.Unit, baseHeal, spell.OutcomeHealing)
		},
	})
}
