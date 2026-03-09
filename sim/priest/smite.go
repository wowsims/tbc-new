package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var SmiteRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 585, Cost: 13, MinDamage: 13, MaxDamage: 16, Coefficient: 0.123, CastTimeSeconds: 1.5},
	{Rank: 2, SpellID: 591, Cost: 30, MinDamage: 28, MaxDamage: 32, Coefficient: 0.271, CastTimeSeconds: 2.0},
	{Rank: 3, SpellID: 598, Cost: 50, MinDamage: 48, MaxDamage: 53, Coefficient: 0.554},
	{Rank: 4, SpellID: 984, Cost: 80, MinDamage: 84, MaxDamage: 94, Coefficient: 0.714},
	{Rank: 5, SpellID: 1004, Cost: 115, MinDamage: 136, MaxDamage: 152, Coefficient: 0.714},
	{Rank: 6, SpellID: 6060, Cost: 160, MinDamage: 207, MaxDamage: 229, Coefficient: 0.714},
	{Rank: 7, SpellID: 10933, Cost: 210, MinDamage: 299, MaxDamage: 331, Coefficient: 0.714},
	{Rank: 8, SpellID: 10934, Cost: 265, MinDamage: 401, MaxDamage: 445, Coefficient: 0.714},
	{Rank: 9, SpellID: 25363, Cost: 380, MinDamage: 539, MaxDamage: 603, Coefficient: 0.714},
	{Rank: 10, SpellID: 25364, Cost: 410, MinDamage: 640, MaxDamage: 716, Coefficient: 0.714},
}

func (priest *Priest) registerSmiteSpell(rankConfig shared.SpellRankConfig) {
	// Divine Fury (Holy Tier 3): -0.1s cast time per rank, max 5 ranks = -0.5s
	castTimeReduction := time.Duration(priest.Talents.DivineFury) * 100 * time.Millisecond

	// Default cast time for Smite is 2.5 seconds (ranks 3+)
	baseCastTime := core.DurationFromSeconds(2.5)
	if rankConfig.CastTimeSeconds > 0 {
		baseCastTime = core.DurationFromSeconds(rankConfig.CastTimeSeconds)
	}

	spell := priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellSmite,
		Rank:           rankConfig.Rank,
		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: baseCastTime - castTimeReduction,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           priest.DefaultSpellCritMultiplier(),
		BonusCoefficient:         rankConfig.Coefficient,
		ThreatMultiplier:         1 - []float64{0, .07, .14, .20}[priest.Talents.SilentResolve],

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := priest.CalcAndRollDamageRange(sim, rankConfig.MinDamage, rankConfig.MaxDamage)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.DealDamage(sim, result)
		},
	})

	priest.Smite = append(priest.Smite, spell)
}
