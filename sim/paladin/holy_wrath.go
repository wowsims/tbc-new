package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (paladin *Paladin) getHolyWrathTimer() *core.Timer {
	if paladin.holyWrathTimer == nil {
		paladin.holyWrathTimer = paladin.NewTimer()
	}
	return paladin.holyWrathTimer
}

var HolyWrathRankMap = shared.RankTable{
	{Rank: 1, SpellID: 2812, Cost: 550, Direct: &shared.Amount{Min: 368, Max: 435, Coef: 0.286}},
	{Rank: 2, SpellID: 10318, Cost: 685, Direct: &shared.Amount{Min: 497, Max: 584, Coef: 0.286}},
	{Rank: 3, SpellID: 27139, Cost: 825, Direct: &shared.Amount{Min: 637, Max: 748, Coef: 0.286}},
}

// Holy Wrath
// https://www.wowhead.com/tbc/spell=2812/holy-wrath
//
// Sends bolts of holy power in all directions, causing Holy damage
// to all Undead and Demon targets within 20 yds.
// 2 sec cast, 1 min cooldown.
func (paladin *Paladin) registerHolyWrath(rankConfig shared.RankRow) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := rankConfig.Direct.Min
	maxDamage := rankConfig.Direct.Max
	coefficient := rankConfig.Direct.Coef

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHolyWrath,
		Rank:           rankConfig.Rank,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange:     20,
		MissileSpeed: 20,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Second * 2,
			},
			CD: core.Cooldown{
				Timer:    paladin.getHolyWrathTimer(),
				Duration: time.Minute,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				castTime := paladin.ApplyCastSpeedForSpell(cast.CastTime, spell)
				paladin.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+castTime)
			},
		},

		BonusCoefficient: coefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			results := []*core.SpellResult{}
			for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
				if aoeTarget.MobType == proto.MobType_MobTypeUndead || aoeTarget.MobType == proto.MobType_MobTypeDemon {
					damage := sim.Roll(minDamage, maxDamage)
					results = append(results, spell.CalcDamage(sim, aoeTarget, damage, spell.OutcomeMagicHitAndCrit))
				}
			}

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				for _, result := range results {
					spell.DealDamage(sim, result)
				}
			})
		},
	})
}
