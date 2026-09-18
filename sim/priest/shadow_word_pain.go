package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var ShadowWordPainRankMap = shared.RankTable{
	{Rank: 1, SpellID: 589, Cost: 25, Periodic: &shared.Periodic{Tick: 5, Coef: 0.0732}},
	{Rank: 2, SpellID: 594, Cost: 50, Periodic: &shared.Periodic{Tick: 11, Coef: 0.114}},
	{Rank: 3, SpellID: 970, Cost: 95, Periodic: &shared.Periodic{Tick: 22, Coef: 0.169}},
	{Rank: 4, SpellID: 992, Cost: 155, Periodic: &shared.Periodic{Tick: 39, Coef: 0.183}},
	{Rank: 5, SpellID: 2767, Cost: 230, Periodic: &shared.Periodic{Tick: 61, Coef: 0.183}},
	{Rank: 6, SpellID: 10892, Cost: 305, Periodic: &shared.Periodic{Tick: 85, Coef: 0.183}},
	{Rank: 7, SpellID: 10893, Cost: 385, Periodic: &shared.Periodic{Tick: 112, Coef: 0.183}},
	{Rank: 8, SpellID: 10894, Cost: 470, Periodic: &shared.Periodic{Tick: 142, Coef: 0.183}},
	{Rank: 9, SpellID: 25367, Cost: 510, Periodic: &shared.Periodic{Tick: 167, Coef: 0.183}},
	{Rank: 10, SpellID: 25368, Cost: 575, Periodic: &shared.Periodic{Tick: 206, Coef: 0.183}},
}

func (priest *Priest) registerShadowWordPainSpell(rank shared.RankRow) {
	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellShadowWordPain,
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
				Label: fmt.Sprintf("ShadowWordPain-%d", rank.Rank),
			},
			NumberOfTicks:       6,
			TickLength:          3 * time.Second,
			AffectedByCastSpeed: false, // DoT ticks not haste-affected in TBC
			BonusCoefficient:    rank.Periodic.Coef,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, rank.Periodic.Tick)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
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
