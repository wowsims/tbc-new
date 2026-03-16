package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const (
	InsectSwarmBonusCoeff    = 0.12700000405
	InsectSwarmNumberOfTicks = 6
	InsectSwarmTickLength    = time.Second * 2
	InsectSwarmTotalDamage   = 792
)

func (druid *Druid) registerInsectSwarmSpell() {
	druid.InsectSwarm = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27013},
		SpellSchool:    core.SpellSchoolArcane | core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: DruidSpellInsectSwarm,
		Flags:          core.SpellFlagAPL,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ManaCost: core.ManaCostOptions{
			FlatCost: 175,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Insect Swarm",
			},

			NumberOfTicks:       InsectSwarmNumberOfTicks,
			TickLength:          InsectSwarmTickLength,
			AffectedByCastSpeed: false,
			BonusCoefficient:    InsectSwarmBonusCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, InsectSwarmTotalDamage)
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
