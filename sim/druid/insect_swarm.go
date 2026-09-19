package druid

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var insectSwarmRank = genRanks.InsectSwarm.BySpellID(27013)
var insectSwarmTick = insectSwarmRank.Periodic.(shared.SpellRankPeriodic)

func (druid *Druid) registerInsectSwarmSpell() {
	druid.InsectSwarm = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: insectSwarmRank.SpellID},
		SpellSchool:    core.SpellSchoolNature,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: DruidSpellInsectSwarm,
		Flags:          core.SpellFlagAPL | core.SpellFlagBinary,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		MaxRange:         insectSwarmRank.MaxRange,

		ManaCost: core.ManaCostOptions{
			FlatCost: insectSwarmRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: insectSwarmRank.GCD,
			},
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Insect Swarm",
			},

			NumberOfTicks:       insectSwarmTick.NumberOfTicks,
			TickLength:          insectSwarmTick.TickLength,
			AffectedByCastSpeed: false,
			BonusCoefficient:    insectSwarmTick.Coef,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, insectSwarmTick.Tick)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)

			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}

			spell.DealOutcome(sim, result)
		},
	})
}
