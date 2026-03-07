package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

func (paladin *Paladin) getAvengersShieldTimer() *core.Timer {
	if paladin.avengersShieldTimer == nil {
		paladin.avengersShieldTimer = paladin.NewTimer()
	}
	return paladin.avengersShieldTimer
}

var AvengersShieldRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 31935, Cost: 500, MinDamage: 270, MaxDamage: 330, Coefficient: 0.193},
	{Rank: 2, SpellID: 32699, Cost: 615, MinDamage: 370, MaxDamage: 452, Coefficient: 0.193},
	{Rank: 3, SpellID: 32700, Cost: 780, MinDamage: 494, MaxDamage: 602, Coefficient: 0.193},
}

// Avenger's Shield (Talent)
// https://www.wowhead.com/tbc/spell=31935
//
// Hurls a holy shield at the enemy, dealing Holy damage, dazing them and
// then jumping to additional nearby enemies. Affects 3 total targets.
func (paladin *Paladin) registerAvengersShield(rankConfig shared.SpellRankConfig) {
	paladin.AvengersShields = append(paladin.AvengersShields, paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskAvengersShield,
		Rank:           rankConfig.Rank,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultSpellCritMultiplier(),

		MaxRange:     30,
		MissileSpeed: 35,

		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      time.Second,
				CastTime: time.Second,
			},
			CD: core.Cooldown{
				Timer:    paladin.getAvengersShieldTimer(),
				Duration: time.Second * 30,
			},
		},

		BonusCoefficient: rankConfig.Coefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := sim.Roll(rankConfig.MinDamage, rankConfig.MaxDamage)
			spell.CalcAndDealCleaveDamage(sim, target, 3, damage, spell.OutcomeMagicHitAndCrit)
		},
	}))
}
