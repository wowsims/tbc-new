package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var ShadowWordPainRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 589, Cost: 25, DotTickDamage: 5, Coefficient: 0.0732},
	{Rank: 2, SpellID: 594, Cost: 50, DotTickDamage: 11, Coefficient: 0.114},
	{Rank: 3, SpellID: 970, Cost: 95, DotTickDamage: 22, Coefficient: 0.169},
	{Rank: 4, SpellID: 992, Cost: 155, DotTickDamage: 39, Coefficient: 0.183},
	{Rank: 5, SpellID: 2767, Cost: 230, DotTickDamage: 61, Coefficient: 0.183},
	{Rank: 6, SpellID: 10892, Cost: 305, DotTickDamage: 85, Coefficient: 0.183},
	{Rank: 7, SpellID: 10893, Cost: 385, DotTickDamage: 112, Coefficient: 0.183},
	{Rank: 8, SpellID: 10894, Cost: 470, DotTickDamage: 142, Coefficient: 0.183},
	{Rank: 9, SpellID: 25367, Cost: 510, DotTickDamage: 167, Coefficient: 0.183},
	{Rank: 10, SpellID: 25368, Cost: 575, DotTickDamage: 206, Coefficient: 0.183},
}

func (priest *Priest) registerShadowWordPainSpell(rankConfig shared.SpellRankConfig) {
	spell := priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellShadowWordPain,
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
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("ShadowWordPain-%d", rankConfig.Rank),
			},
			NumberOfTicks:       6,
			TickLength:          3 * time.Second,
			AffectedByCastSpeed: false, // DoT ticks not haste-affected in TBC
			BonusCoefficient:    rankConfig.Coefficient,

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
			}
			spell.DealOutcome(sim, result)
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			if useSnapshot {
				dot := spell.Dot(target)
				return dot.CalcSnapshotDamage(sim, target, spell.OutcomeExpectedMagicHit)
			}
			return spell.CalcPeriodicDamage(sim, target, rankConfig.DotTickDamage, spell.OutcomeExpectedMagicHit)
		},
	})

	priest.ShadowWordPain = append(priest.ShadowWordPain, spell)
}
