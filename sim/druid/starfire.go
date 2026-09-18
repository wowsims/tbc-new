package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var StarfireRankMap = shared.RankTable{
	{Rank: 6, SpellID: 9876, Cost: 315, Direct: &shared.Amount{Min: 463, Max: 543, Coef: 1}},
	{Rank: 8, SpellID: 26986, Cost: 370, Direct: &shared.Amount{Min: 550, Max: 647, Coef: 1}},
}

func (druid *Druid) registerStarfireSpell(rankConfig shared.RankRow) {
	spell := druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolArcane,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: DruidSpellStarfire,
		Flags:          core.SpellFlagAPL,
		Rank:           rankConfig.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 3500,
			},
		},

		BonusCoefficient: rankConfig.Direct.Coef,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := druid.CalcAndRollDamageRange(sim, rankConfig.Direct.Min, rankConfig.Direct.Max)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})

	druid.Starfire = append(druid.Starfire, spell)
}
