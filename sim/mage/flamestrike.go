package mage

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var FlameStrikeRankMap = shared.SpellRankMap{
	{Rank: 7, SpellID: 27086, Cost: 1175, MinDamage: 480, MaxDamage: 585, DotTickDamage: 106, ThreatMultiplier: 1},
	{Rank: 6, SpellID: 10216, Cost: 990, MinDamage: 383, MaxDamage: 468, DotTickDamage: 85, ThreatMultiplier: 1},
}

func (mage *Mage) registerFlamestrike(rankConfig shared.SpellRankConfig) {
	flameStrikeCoefficient := 0.23600000143 // Per https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=exact%253A2120 Field: "BonusCoefficient"
	flameStrikeDotCoefficient := 0.02999999933

	spell := mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: MageSpellFlamestrike,
		Rank:           rankConfig.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Second * 3,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   mage.DefaultSpellCritMultiplier(),
		BonusCoefficient: flameStrikeCoefficient,
		ThreatMultiplier: rankConfig.ThreatMultiplier,

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: fmt.Sprintf("Flamestrike DoT %s", rankConfig.GetRankLabel()),
			},
			NumberOfTicks:    4,
			TickLength:       time.Second * 2,
			BonusCoefficient: flameStrikeDotCoefficient,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, rankConfig.DotTickDamage)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeTick)
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			baseDamage := mage.CalcAndRollDamageRange(sim, rankConfig.MinDamage, rankConfig.MaxDamage)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.AOEDot().Apply(sim)
		},
	})

	mage.Flamestrike = append(mage.Flamestrike, spell)
}
