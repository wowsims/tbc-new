package priest

import (
	"fmt"
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
	"time"
)

var MindFlayRankMap = shared.RankTable{
	{Rank: 1, SpellID: 15407, Cost: 45, Periodic: &shared.Periodic{Tick: 25}},
	{Rank: 2, SpellID: 17311, Cost: 70, Periodic: &shared.Periodic{Tick: 42}},
	{Rank: 3, SpellID: 17312, Cost: 100, Periodic: &shared.Periodic{Tick: 62}},
	{Rank: 4, SpellID: 17313, Cost: 135, Periodic: &shared.Periodic{Tick: 87}},
	{Rank: 5, SpellID: 17314, Cost: 165, Periodic: &shared.Periodic{Tick: 110}},
	{Rank: 6, SpellID: 18807, Cost: 205, Periodic: &shared.Periodic{Tick: 142}},
	{Rank: 7, SpellID: 25387, Cost: 230, Periodic: &shared.Periodic{Tick: 176}},
}

// mindFlayTickCoefficient is the SP coefficient applied per tick.
// Total channel coefficient ~0.57 split across 3 ticks.
const mindFlayTickCoefficient = 0.1905

func (priest *Priest) registerMindFlaySpell(rank shared.RankRow) {
	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagChanneled | core.SpellFlagAPL,
		ClassSpellMask: PriestSpellMindFlay,
		Rank:           rank.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rank.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("MindFlay-%d", rank.Rank),
			},
			NumberOfTicks:        3,
			TickLength:           time.Second,
			AffectedByCastSpeed:  true,
			HasteReducesDuration: true,
			BonusCoefficient:     mindFlayTickCoefficient,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, rank.Periodic.Tick)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			if useSnapshot {
				dot := spell.Dot(target)
				return dot.CalcSnapshotDamage(sim, target, spell.OutcomeExpectedMagicHit)
			}
			return spell.CalcPeriodicDamage(sim, target, rank.Periodic.Tick, spell.OutcomeExpectedMagicHit)
		},
	})
}
