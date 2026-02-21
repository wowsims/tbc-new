package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

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
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := rankConfig.MinDamage
	maxDamage := rankConfig.MaxDamage
	coefficient := rankConfig.Coefficient

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 30,
	}

	paladin.AvengersShields = append(paladin.AvengersShields, paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskAvengersShield,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultSpellCritMultiplier(),

		MaxRange:     30,
		MissileSpeed: 35,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      time.Second,
				CastTime: time.Second,
			},
			CD: cd,
		},

		BonusCoefficient: coefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := sim.Roll(minDamage, maxDamage)
			spell.CalcAndDealCleaveDamage(sim, target, 3, damage, spell.OutcomeMagicHitAndCrit)
		},
	}))
}
