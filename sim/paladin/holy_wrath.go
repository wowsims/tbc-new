package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

var HolyWrathRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 2812, Cost: 550, MinDamage: 368, MaxDamage: 435, Coefficient: 0.286},
	{Rank: 2, SpellID: 10318, Cost: 685, MinDamage: 497, MaxDamage: 584, Coefficient: 0.286},
	{Rank: 3, SpellID: 27139, Cost: 825, MinDamage: 637, MaxDamage: 748, Coefficient: 0.286},
}

// Holy Wrath
// https://www.wowhead.com/tbc/spell=2812/holy-wrath
//
// Sends bolts of holy power in all directions, causing Holy damage
// to all Undead and Demon targets within 20 yds.
// 2 sec cast, 1 min cooldown.
func (paladin *Paladin) registerHolyWrath(rankConfig shared.SpellRankConfig) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := rankConfig.MinDamage
	maxDamage := rankConfig.MaxDamage
	coefficient := rankConfig.Coefficient

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Minute,
	}

	paladin.HolyWraths = append(paladin.HolyWraths, paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHolyWrath,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultSpellCritMultiplier(),

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Second * 2,
			},
			CD: cd,
		},

		BonusCoefficient: coefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := sim.Roll(minDamage, maxDamage)
			for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
				if aoeTarget.MobType == proto.MobType_MobTypeUndead || aoeTarget.MobType == proto.MobType_MobTypeDemon {
					spell.CalcAndDealDamage(sim, aoeTarget, damage, spell.OutcomeMagicHitAndCrit)
				}
			}
		},
	}))
}
