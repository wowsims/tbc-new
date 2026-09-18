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

var HammerOfWrathRankMap = genRanks.HammerOfWrath

// Hammer of Wrath
// https://www.wowhead.com/tbc/spell=27180
//
// Hurls a hammer that strikes an enemy for Holy damage.
// Only usable on enemies that have 20% or less health.
func (paladin *Paladin) registerHammerOfWrath(rankConfig shared.SpellRank) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := shared.SpellRankMin(rankConfig.Direct)
	maxDamage := shared.SpellRankMax(rankConfig.Direct)
	coefficient := rankConfig.Direct.BonusCoefficient()

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

		MaxRange:     rankConfig.MaxRange,
		MissileSpeed: rankConfig.MissileSpeed,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCDMin:   time.Millisecond * 500,
				GCD:      rankConfig.GCD,
				CastTime: rankConfig.CastTime,
			},
			CD: core.Cooldown{
				Timer:    paladin.getHammerOfWrathTimer(),
				Duration: rankConfig.Cooldown,
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
