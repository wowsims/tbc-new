package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var SmiteRankMap = shared.RankTable{
	{Rank: 1, SpellID: 585, Cost: 20, Direct: &shared.Amount{Min: 15, Max: 20, Coef: 0.123}},
	{Rank: 2, SpellID: 591, Cost: 30, Direct: &shared.Amount{Min: 28, Max: 34, Coef: 0.271}},
	{Rank: 3, SpellID: 598, Cost: 60, Direct: &shared.Amount{Min: 58, Max: 67, Coef: 0.554}},
	{Rank: 4, SpellID: 984, Cost: 95, Direct: &shared.Amount{Min: 97, Max: 112, Coef: 0.714}},
	{Rank: 5, SpellID: 1004, Cost: 140, Direct: &shared.Amount{Min: 158, Max: 178, Coef: 0.714}},
	{Rank: 6, SpellID: 6060, Cost: 185, Direct: &shared.Amount{Min: 222, Max: 250, Coef: 0.714}},
	{Rank: 7, SpellID: 10933, Cost: 230, Direct: &shared.Amount{Min: 298, Max: 335, Coef: 0.714}},
	{Rank: 8, SpellID: 10934, Cost: 280, Direct: &shared.Amount{Min: 384, Max: 429, Coef: 0.714}},
	{Rank: 9, SpellID: 25363, Cost: 300, Direct: &shared.Amount{Min: 422, Max: 470, Coef: 0.714}},
	{Rank: 10, SpellID: 25364, Cost: 385, Direct: &shared.Amount{Min: 549, Max: 616, Coef: 0.714}},
}

var smiteCastTimes = map[int32]time.Duration{
	1: 1500 * time.Millisecond,
	2: 2000 * time.Millisecond,
}

func (priest *Priest) registerSmiteSpell(rank shared.RankRow) {
	castTime := 2500 * time.Millisecond
	if ct, ok := smiteCastTimes[rank.Rank]; ok {
		castTime = ct
	}

	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.SpellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellSmite,
		Rank:           rank.Rank,
		ManaCost: core.ManaCostOptions{
			FlatCost: rank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: castTime,
			},
		},

		DamageMultiplier: 1,
		BonusCoefficient: rank.Direct.Coef,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := priest.CalcAndRollDamageRange(sim, rank.Direct.Min, rank.Direct.Max)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}
