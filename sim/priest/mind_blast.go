package priest

import (
	//"fmt"
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var MindBlastRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 8092, Cost: 50, MinDamage: 42, MaxDamage: 46, Coefficient: 0.268},
	{Rank: 2, SpellID: 8102, Cost: 80, MinDamage: 76, MaxDamage: 83, Coefficient: 0.364},
	{Rank: 3, SpellID: 8103, Cost: 110, MinDamage: 117, MaxDamage: 126, Coefficient: 0.42857},
	{Rank: 4, SpellID: 8104, Cost: 150, MinDamage: 174, MaxDamage: 184, Coefficient: 0.42857},
	{Rank: 5, SpellID: 8105, Cost: 185, MinDamage: 225, MaxDamage: 239, Coefficient: 0.42857},
	{Rank: 6, SpellID: 8106, Cost: 225, MinDamage: 288, MaxDamage: 307, Coefficient: 0.42857},
	{Rank: 7, SpellID: 8107, Cost: 265, MinDamage: 356, MaxDamage: 377, Coefficient: 0.42857},
	{Rank: 8, SpellID: 10946, Cost: 310, MinDamage: 437, MaxDamage: 461, Coefficient: 0.42857},
	{Rank: 9, SpellID: 10947, Cost: 350, MinDamage: 516, MaxDamage: 544, Coefficient: 0.42857},
	{Rank: 10, SpellID: 25372, Cost: 380, MinDamage: 571, MaxDamage: 602, Coefficient: 0.42857},
	{Rank: 11, SpellID: 25375, Cost: 450, MinDamage: 711, MaxDamage: 752, Coefficient: 0.42857},
}

func (priest *Priest) registerMindBlastSpell(rankConfig shared.SpellRankConfig) {
	// Improved Mind Blast (Shadow Tier 2): -0.5s CD per rank, max 5 ranks = -2.5s
	cdReduction := time.Duration(priest.Talents.ImprovedMindBlast) * 500 * time.Millisecond
	cdTimer := priest.NewTimer()

	spell := priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellMindBlast,
		Rank:           rankConfig.Rank,
		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: 1500 * time.Millisecond,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: 8*time.Second - cdReduction,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           priest.DefaultSpellCritMultiplier(),
		BonusCoefficient:         rankConfig.Coefficient,
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := priest.CalcAndRollDamageRange(sim, rankConfig.MinDamage, rankConfig.MaxDamage)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			if result.Landed() {
				// TODO: apply Shadow Weaving debuff (1 stack per hit, max 5)
			}

			spell.DealDamage(sim, result)
		},
	})

	priest.MindBlast = append(priest.MindBlast, spell)
}
