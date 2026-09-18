package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var MindBlastRankMap = shared.RankTable{
	{Rank: 1, SpellID: 8092, Cost: 50, Direct: &shared.Amount{Min: 42, Max: 46, Coef: 0.268}},
	{Rank: 2, SpellID: 8102, Cost: 80, Direct: &shared.Amount{Min: 76, Max: 83, Coef: 0.364}},
	{Rank: 3, SpellID: 8103, Cost: 110, Direct: &shared.Amount{Min: 117, Max: 126, Coef: 0.42857}},
	{Rank: 4, SpellID: 8104, Cost: 150, Direct: &shared.Amount{Min: 174, Max: 184, Coef: 0.42857}},
	{Rank: 5, SpellID: 8105, Cost: 185, Direct: &shared.Amount{Min: 225, Max: 239, Coef: 0.42857}},
	{Rank: 6, SpellID: 8106, Cost: 225, Direct: &shared.Amount{Min: 288, Max: 307, Coef: 0.42857}},
	{Rank: 7, SpellID: 10945, Cost: 265, Direct: &shared.Amount{Min: 356, Max: 377, Coef: 0.42857}},
	{Rank: 8, SpellID: 10946, Cost: 310, Direct: &shared.Amount{Min: 437, Max: 461, Coef: 0.42857}},
	{Rank: 9, SpellID: 10947, Cost: 350, Direct: &shared.Amount{Min: 516, Max: 544, Coef: 0.42857}},
	{Rank: 10, SpellID: 25372, Cost: 380, Direct: &shared.Amount{Min: 571, Max: 602, Coef: 0.42857}},
	{Rank: 11, SpellID: 25375, Cost: 450, Direct: &shared.Amount{Min: 711, Max: 752, Coef: 0.42857}},
}

func (priest *Priest) registerMindBlastSpell(rank shared.RankRow, cdTimer *core.Timer) {

	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellMindBlast,
		Rank:           rank.Rank,
		ManaCost: core.ManaCostOptions{
			FlatCost: rank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: 1500 * time.Millisecond,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: 8 * time.Second,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		BonusCoefficient:         rank.Direct.Coef,
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := priest.CalcAndRollDamageRange(sim, rank.Direct.Min, rank.Direct.Max)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.DealDamage(sim, result)
		},
	})
}
