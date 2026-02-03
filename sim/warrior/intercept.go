package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (war *Warrior) registerIntercept() {
	actionID := core.ActionID{SpellID: 20252}
	chargeMinRange := 8.0

	var spell *core.Spell
	var interceptTarget *core.Unit

	aura := war.RegisterAura(core.Aura{
		Label:    "Intercept",
		ActionID: actionID,
		Duration: 15 * time.Second,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			war.MultiplyMovementSpeed(sim, 3.0)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			war.MultiplyMovementSpeed(sim, 1.0/3.0)
			spell.CalcAndDealDamage(sim, interceptTarget, 105, spell.OutcomeAlwaysHit)
		},
	})

	war.RegisterMovementCallback(func(sim *core.Simulation, position float64, kind core.MovementUpdateType) {
		if kind == core.MovementEnd && aura.IsActive() {
			aura.Deactivate(sim)
		}
	})

	spell = war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolPhysical,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskIntercept,
		MinRange:       chargeMinRange,
		MaxRange:       25,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: 30 * time.Second,
			},
			IgnoreHaste: true,
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(BerserkerStance)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			target = interceptTarget
			aura.Activate(sim)
			war.MoveTo(chargeMinRange-1, sim) // movement aura is discretized in 1 yard intervals, so need to overshoot to guarantee melee range
		},
	})
}
