package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (war *Warrior) registerCharge() {
	actionID := core.ActionID{SpellID: 100}
	metrics := war.NewRageMetrics(actionID)

	chargeMinRange := 8.0

	aura := war.RegisterAura(core.Aura{
		Label:    "Charge",
		ActionID: actionID,
		Duration: 15 * time.Second,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			war.MultiplyMovementSpeed(sim, 3.0)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			war.MultiplyMovementSpeed(sim, 1.0/3.0)
		},
	})

	war.RegisterMovementCallback(func(sim *core.Simulation, position float64, kind core.MovementUpdateType) {
		if kind == core.MovementEnd && aura.IsActive() {
			aura.Deactivate(sim)
		}
	})

	war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolPhysical,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskCharge,
		MinRange:       chargeMinRange,
		MaxRange:       25,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: 15 * time.Second,
			},
			IgnoreHaste: true,
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return sim.CurrentTime < 0 && war.StanceMatches(BattleStance)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
			war.AddRage(sim, war.ChargeRageGain, metrics)
			war.MoveTo(chargeMinRange-3.5, sim) // movement aura is discretized in 1 yard intervals, so need to overshoot to guarantee melee range
		},
	})
}
