package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var SmiteRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 585, Cost: 20, MinDamage: 15, MaxDamage: 20, Coefficient: 0.123, CastTimeSeconds: 1.5},
	{Rank: 2, SpellID: 591, Cost: 30, MinDamage: 28, MaxDamage: 34, Coefficient: 0.271, CastTimeSeconds: 2.0},
	{Rank: 3, SpellID: 598, Cost: 60, MinDamage: 58, MaxDamage: 67, Coefficient: 0.554},
	{Rank: 4, SpellID: 984, Cost: 95, MinDamage: 97, MaxDamage: 112, Coefficient: 0.714},
	{Rank: 5, SpellID: 1004, Cost: 140, MinDamage: 158, MaxDamage: 178, Coefficient: 0.714},
	{Rank: 6, SpellID: 6060, Cost: 185, MinDamage: 222, MaxDamage: 250, Coefficient: 0.714},
	{Rank: 7, SpellID: 10933, Cost: 230, MinDamage: 298, MaxDamage: 335, Coefficient: 0.714},
	{Rank: 8, SpellID: 10934, Cost: 280, MinDamage: 384, MaxDamage: 429, Coefficient: 0.714},
	{Rank: 9, SpellID: 25363, Cost: 300, MinDamage: 422, MaxDamage: 470, Coefficient: 0.714},
	{Rank: 10, SpellID: 25364, Cost: 385, MinDamage: 549, MaxDamage: 616, Coefficient: 0.714},
}

func (priest *Priest) registerSmiteSpell(rankConfig shared.SpellRankConfig) {

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
				CastTime: core.TernaryDuration(rankConfig.CastTimeSeconds > 0, core.DurationFromSeconds(rankConfig.CastTimeSeconds), 2500*time.Millisecond),
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   priest.DefaultSpellCritMultiplier(),
		BonusCoefficient: rankConfig.Coefficient,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := priest.CalcAndRollDamageRange(sim, rankConfig.MinDamage, rankConfig.MaxDamage)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})

	priest.Smite = append(priest.Smite, spell)
}
