package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

// Devouring Plague - Undead Racial
// Shadow school DoT, 3 min cooldown, 24s duration

var DevouringPlagueRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 2944, Cost: 215, DotTickDamage: 19, Coefficient: 0.1},
	{Rank: 2, SpellID: 19276, Cost: 350, DotTickDamage: 34, Coefficient: 0.1},
	{Rank: 3, SpellID: 19277, Cost: 495, DotTickDamage: 50, Coefficient: 0.1},
	{Rank: 4, SpellID: 19278, Cost: 645, DotTickDamage: 68, Coefficient: 0.1},
	{Rank: 5, SpellID: 19279, Cost: 810, DotTickDamage: 89, Coefficient: 0.1},
	{Rank: 6, SpellID: 19280, Cost: 985, DotTickDamage: 113, Coefficient: 0.1},
	{Rank: 7, SpellID: 25467, Cost: 1145, DotTickDamage: 152, Coefficient: 0.1},
}

func (priest *Priest) registerDevouringPlagueSpell(rankConfig shared.SpellRankConfig, cdTimer *core.Timer) {
	healthMetrics := priest.NewHealthMetrics(core.ActionID{SpellID: rankConfig.SpellID}.WithTag(1))

	spell := priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellDevouringPlague,
		Rank:           rankConfig.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: 180 * time.Second,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("DevouringPlague-%d", rankConfig.Rank),
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					aura.AttachProcTrigger(core.ProcTrigger{
						Name:               "DevouringPlague-Heal",
						Callback:           core.CallbackOnPeriodicDamageTaken,
						ClassSpellMask:     PriestSpellDevouringPlague,
						RequireDamageDealt: true,
						Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
							priest.GainHealth(sim, result.Damage, healthMetrics)
						},
					})
				},
			},
			NumberOfTicks:       8,
			TickLength:          3 * time.Second,
			AffectedByCastSpeed: false,
			BonusCoefficient:    rankConfig.Coefficient,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, rankConfig.DotTickDamage)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
				spell.Dot(target).TickOnce(sim)
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

	priest.DevouringPlague = append(priest.DevouringPlague, spell)
}
