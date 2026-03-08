package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var VampiricTouchRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 34914, Cost: 325, DotTickDamage: 90, Coefficient: 0.2},
	{Rank: 2, SpellID: 34916, Cost: 400, DotTickDamage: 120, Coefficient: 0.2},
	{Rank: 3, SpellID: 34917, Cost: 425, DotTickDamage: 130, Coefficient: 0.2},
}

func (priest *Priest) registerVampiricTouchSpell(rankConfig shared.SpellRankConfig) {
	manaMetrics := priest.NewManaMetrics(core.ActionID{SpellID: rankConfig.SpellID}.WithTag(1))

	vtManaAura := priest.GetOrRegisterAura(core.Aura{
		Label:    "VampiricTouch-ManaReturn",
		Duration: core.NeverExpires,
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.SpellSchool.Matches(core.SpellSchoolShadow) || !result.Landed() || result.Damage == 0 {
				return
			}
			priest.AddMana(sim, result.Damage*0.05, manaMetrics)
		},
		OnPeriodicDamageDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.SpellSchool.Matches(core.SpellSchoolShadow) || result.Damage == 0 || spell.ClassSpellMask == PriestSpellVampiricTouch {
				return
			}
			priest.AddMana(sim, result.Damage*0.05, manaMetrics)
		},
	})

	spell := priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellVampiricTouch,
		Rank:           rankConfig.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: 1500 * time.Millisecond,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("VampiricTouch-%d", rankConfig.Rank),
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					vtManaAura.Activate(sim)
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					vtManaAura.Deactivate(sim)
				},
			},
			NumberOfTicks:       5,
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

	priest.VampiricTouch = append(priest.VampiricTouch, spell)
}
