package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

func (paladin *Paladin) getHammerOfWrathTimer() *core.Timer {
	if paladin.hammerOfWrathTimer == nil {
		paladin.hammerOfWrathTimer = paladin.NewTimer()
	}
	return paladin.hammerOfWrathTimer
}

var HammerOfWrathRankMap = shared.RankTable{
	{Rank: 1, SpellID: 24275, Cost: 235, Direct: &shared.Amount{Min: 316, Max: 348, Coef: 0.429}},
	{Rank: 2, SpellID: 24274, Cost: 290, Direct: &shared.Amount{Min: 412, Max: 455, Coef: 0.429}},
	{Rank: 3, SpellID: 24239, Cost: 340, Direct: &shared.Amount{Min: 519, Max: 572, Coef: 0.429}},
	{Rank: 4, SpellID: 27180, Cost: 440, Direct: &shared.Amount{Min: 672, Max: 742, Coef: 0.429}},
}

// Hammer of Wrath
// https://www.wowhead.com/tbc/spell=27180
//
// Hurls a hammer that strikes an enemy for Holy damage.
// Only usable on enemies that have 20% or less health.
func (paladin *Paladin) registerHammerOfWrath(rankConfig shared.RankRow) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := rankConfig.Direct.Min
	maxDamage := rankConfig.Direct.Max
	coefficient := rankConfig.Direct.Coef

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeRanged,
		ProcMask:       core.ProcMaskRangedSpecial,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHammerOfWrath,
		Rank:           rankConfig.Rank,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange:     30,
		MissileSpeed: 35,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCDMin:   time.Millisecond * 500,
				GCD:      time.Millisecond * 500,
				CastTime: time.Millisecond * 500,
			},
			CD: core.Cooldown{
				Timer:    paladin.getHammerOfWrathTimer(),
				Duration: time.Second * 6,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				castTime := paladin.ApplyCastSpeedForSpell(cast.CastTime, spell)
				paladin.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+castTime)
			},
		},

		BonusCoefficient: coefficient,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return sim.IsExecutePhase20()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, sim.Roll(minDamage, maxDamage), spell.OutcomeRangedHitAndCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
