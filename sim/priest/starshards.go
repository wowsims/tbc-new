package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

// Starshards - Night Elf Racial
// Arcane school DoT, 0 mana cost, 30s cooldown, 15s duration
var StarshardsRankMap = shared.RankTable{
	{Rank: 1, SpellID: 10797, Cost: 0, Periodic: &shared.Periodic{Tick: 12, Coef: 0.167}},
	{Rank: 2, SpellID: 19296, Cost: 0, Periodic: &shared.Periodic{Tick: 23, Coef: 0.167}},
	{Rank: 3, SpellID: 19299, Cost: 0, Periodic: &shared.Periodic{Tick: 40, Coef: 0.167}},
	{Rank: 4, SpellID: 19302, Cost: 0, Periodic: &shared.Periodic{Tick: 58, Coef: 0.167}},
	{Rank: 5, SpellID: 19303, Cost: 0, Periodic: &shared.Periodic{Tick: 79, Coef: 0.167}},
	{Rank: 6, SpellID: 19304, Cost: 0, Periodic: &shared.Periodic{Tick: 105, Coef: 0.167}},
	{Rank: 7, SpellID: 19305, Cost: 0, Periodic: &shared.Periodic{Tick: 130, Coef: 0.167}},
	{Rank: 8, SpellID: 25446, Cost: 0, Periodic: &shared.Periodic{Tick: 157, Coef: 0.167}},
}

func (priest *Priest) registerStarshardsSpell(rank shared.RankRow, cdTimer *core.Timer) {

	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.SpellID},
		SpellSchool:    core.SpellSchoolArcane,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellStarshards,
		Rank:           rank.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: 0,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: 30 * time.Second,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("Starshards-%d", rank.Rank),
			},
			NumberOfTicks:       5,
			TickLength:          3 * time.Second,
			AffectedByCastSpeed: false,
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
