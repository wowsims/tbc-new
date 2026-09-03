package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

// Enrage (Dire Bear Form): generates 2 Rage/sec for 10 sec (20 total),
// reduces base armor by 16%. Intensity talent grants instant rage on activation.
func (druid *Druid) registerEnrageSpell() {
	actionID := core.ActionID{SpellID: 5229}
	rageMetrics := druid.NewRageMetrics(actionID)

	const armorReduction = 0.16

	druid.EnrageAura = druid.RegisterAura(core.Aura{
		Label:    "Enrage",
		ActionID: actionID,
		Duration: 10 * time.Second,
	}).AttachMultiplicativePseudoStatBuff(&druid.PseudoStats.ArmorMultiplier, 1-armorReduction)

	// Deactivate Enrage when leaving Bear Form.
	druid.BearFormAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		if !druid.Env.MeasuringStats {
			druid.EnrageAura.Deactivate(sim)
		}
	})

	druid.Enrage = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:       actionID,
		ClassSpellMask: DruidSpellEnrage,
		Flags:          core.SpellFlagAPL,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: 1 * time.Minute,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			// Intensity talent: instantly generate 4/7/10 rage on cast.
			if druid.IntensityEnrageRageBonus > 0 {
				druid.AddRage(sim, druid.IntensityEnrageRageBonus, rageMetrics)
			}
			druid.EnrageAura.Activate(sim)
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   time.Second,
				NumTicks: 10,
				Priority: core.ActionPriorityRegen,
				OnAction: func(sim *core.Simulation) {
					druid.AddRage(sim, 2, rageMetrics)
				},
			})
		},
	})
}
