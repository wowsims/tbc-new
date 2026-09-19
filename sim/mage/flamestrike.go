package mage

import (
	"fmt"
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var FlameStrikeRankMap = genRanks.Flamestrike.Ranks(7, 6)

func (mage *Mage) registerFlamestrike(rankConfig shared.SpellRank) {
	tick := rankConfig.Periodic.(shared.SpellRankPeriodic)

	flameStrikeCoefficient := 0.23600000143 // Per https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=exact%253A2120 Field: "BonusCoefficient"
	flameStrikeDotCoefficient := 0.02999999933

	spell := mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolFire,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: MageSpellFlamestrike,
		Rank:           rankConfig.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      rankConfig.GCD,
				CastTime: rankConfig.CastTime,
			},
		},

		DamageMultiplier: 1,
		BonusCoefficient: flameStrikeCoefficient,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: fmt.Sprintf("Flamestrike DoT %s", rankConfig.GetRankLabel()),
			},
			NumberOfTicks:    tick.NumberOfTicks,
			TickLength:       tick.TickLength,
			BonusCoefficient: flameStrikeDotCoefficient,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, tick.Tick)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeTick)
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			baseDamage := rankConfig.Direct.Damage(sim)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.AOEDot().Apply(sim)
		},
	})

	mage.Flamestrike = append(mage.Flamestrike, spell)
}
