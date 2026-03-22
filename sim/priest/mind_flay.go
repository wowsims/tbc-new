package priest

import (
	"fmt"
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
	"time"
)

var MindFlayRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 15407, Cost: 45, DotTickDamage: 25},
	{Rank: 2, SpellID: 17311, Cost: 70, DotTickDamage: 42},
	{Rank: 3, SpellID: 17312, Cost: 100, DotTickDamage: 62},
	{Rank: 4, SpellID: 17313, Cost: 135, DotTickDamage: 87},
	{Rank: 5, SpellID: 17314, Cost: 165, DotTickDamage: 110},
	{Rank: 6, SpellID: 18807, Cost: 205, DotTickDamage: 142},
	{Rank: 7, SpellID: 25387, Cost: 230, DotTickDamage: 176},
}

// mindFlayTickCoefficient is the SP coefficient applied per tick.
// Total channel coefficient ~0.57 split across 3 ticks.
const mindFlayTickCoefficient = 0.1905

func (priest *Priest) registerMindFlaySpell(rankConfig shared.SpellRankConfig) {
	spell := priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagChanneled | core.SpellFlagAPL,
		ClassSpellMask: PriestSpellMindFlay,
		Rank:           rankConfig.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           priest.DefaultSpellCritMultiplier(),
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("MindFlay-%d", rankConfig.Rank),
			},
			NumberOfTicks:        3,
			TickLength:           time.Second,
			AffectedByCastSpeed:  true,
			HasteReducesDuration: true,
			BonusCoefficient:     mindFlayTickCoefficient,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, rankConfig.DotTickDamage)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
				spell.DealOutcome(sim, result)
			}
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			if useSnapshot {
				dot := spell.Dot(target)
				return dot.CalcSnapshotDamage(sim, target, spell.OutcomeExpectedMagicHit)
			}
			return spell.CalcPeriodicDamage(sim, target, rankConfig.DotTickDamage, spell.OutcomeExpectedMagicHit)
		},
	})

	priest.MindFlay = append(priest.MindFlay, spell)
}
