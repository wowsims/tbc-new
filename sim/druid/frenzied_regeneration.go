package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (druid *Druid) registerFrenziedRegenerationSpell() {
	actionID := core.ActionID{SpellID: 22842}
	healthMetrics := druid.NewHealthMetrics(actionID)
	rageMetrics := druid.NewRageMetrics(actionID)

	druid.FrenziedRegenerationAura = druid.RegisterAura(core.Aura{
		Label:    "Frenzied Regeneration",
		ActionID: actionID,
		Duration: time.Second * 10,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			druid.PseudoStats.HealingTakenMultiplier *= 1.3
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.PseudoStats.HealingTakenMultiplier /= 1.3
		},
	})

	druid.FrenziedRegeneration = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskEmpty,
		ClassSpellMask: DruidSpellFrenziedRegeneration,
		Flags:          core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Second * 180,
			},
			IgnoreHaste: true,
		},

		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.FrenziedRegenerationAura.Activate(sim)

			// Each second, converts up to 10 rage into 10 HP per rage.
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   time.Second * 1,
				NumTicks: 10,
				Priority: core.ActionPriorityDOT,
				OnAction: func(sim *core.Simulation) {
					if !druid.FrenziedRegenerationAura.IsActive() {
						return
					}
					rage := min(druid.CurrentRage(), 10)
					if rage > 0 {
						druid.SpendRage(sim, rage, rageMetrics)
						druid.GainHealth(sim, rage*10, healthMetrics)
					}
				},
			})
		},
	})
}
