package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var HammerOfWrathRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 24275, Cost: 235, MinDamage: 316, MaxDamage: 348, Coefficient: 0.429},
	{Rank: 2, SpellID: 24274, Cost: 290, MinDamage: 412, MaxDamage: 455, Coefficient: 0.429},
	{Rank: 3, SpellID: 24239, Cost: 340, MinDamage: 519, MaxDamage: 572, Coefficient: 0.429},
	{Rank: 4, SpellID: 27180, Cost: 440, MinDamage: 672, MaxDamage: 742, Coefficient: 0.429},
}

// Hammer of Wrath
// https://www.wowhead.com/tbc/spell=27180
//
// Hurls a hammer that strikes an enemy for Holy damage.
// Only usable on enemies that have 20% or less health.
func (paladin *Paladin) registerHammerOfWrath(rankConfig shared.SpellRankConfig) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := rankConfig.MinDamage
	maxDamage := rankConfig.MaxDamage
	coefficient := rankConfig.Coefficient

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 6,
	}

	hammerOfWrath := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskRangedSpecial,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHammerOfWrath,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: 30,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      time.Millisecond * 500,
				CastTime: time.Millisecond * 500,
			},
			CD: cd,
		},

		BonusCoefficient: coefficient,
		CritMultiplier:   paladin.DefaultMeleeCritMultiplier(),

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return sim.IsExecutePhase20()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, sim.Roll(minDamage, maxDamage), spell.OutcomeMeleeSpecialHitAndCrit)
			spell.DealDamage(sim, result)
		},
	})

	paladin.HammerOfWraths = append(paladin.HammerOfWraths, hammerOfWrath)
}
