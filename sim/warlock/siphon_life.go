package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const siphonLifeCoeff = 0.1

func (warlock *Warlock) registerSiphonLifeSpell() {
	actionID := core.ActionID{SpellID: 18265}
	baseCost := 410.0

	healthMetrics := warlock.NewHealthMetrics(actionID)

	warlock.SiphonLife = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolShadow,
		ClassSpellMask: WarlockSpellSiphonLife,
		Flags:          core.SpellFlagAPL,
		BaseCost:       baseCost,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				Cost: baseCost,
				GCD:  core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "SiphonLife",
				Tag:   "Affliction",
			},
			NumberOfTicks:       10,
			TickLength:          3 * time.Second,
			AffectedByCastSpeed: false,
			BonusCoefficient:    0.1,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, 63)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				var result = dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)

				healthToRegain := result.Damage * (1 * warlock.PseudoStats.BonusHealingTaken)
				warlock.GainHealth(sim, healthToRegain, healthMetrics)
				dot.Spell.ApplyAOEThreat(healthToRegain * 0.5)
			},
		},
	})
}
