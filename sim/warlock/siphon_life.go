package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const siphonLifeCoeff = 0.1

func (warlock *Warlock) registerSiphonLifeSpell() {
	actionID := core.ActionID{SpellID: 6353}
	baseCost := 410.0
	resultSlice := make(core.SpellResultSlice, 1)
	healthMetrics := warlock.NewHealthMetrics(actionID)

	warlock.SiphonLife = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolShadow,
		ClassSpellMask: WarlockSpellSiphonLife,
		BaseCost:       baseCost,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				Cost: baseCost,
				GCD:  core.GCDDefault,
			},
		},
		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:    "Corruption",
				Tag:      "Affliction",
				ActionID: core.ActionID{SpellID: 6353},
			},
			NumberOfTicks:       6,
			TickLength:          3 * time.Second,
			AffectedByCastSpeed: false,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, warlock.CalcScalingSpellDmg(siphonLifeCoeff))
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				resultSlice[0] = dot.CalcSnapshotDamage(sim, target, dot.OutcomeSnapshotCrit)

				// if onTickCallback != nil {
				// 	onTickCallback(resultSlice, dot.Spell, sim)
				// }

				dot.Spell.DealPeriodicDamage(sim, resultSlice[0])

				healthToRegain := resultSlice[0].Damage * (1 * warlock.PseudoStats.BonusHealingTaken)
				warlock.GainHealth(sim, healthToRegain, healthMetrics)
				dot.Spell.ApplyAOEThreat(healthToRegain * 0.5)
			},
		},
	})
}
