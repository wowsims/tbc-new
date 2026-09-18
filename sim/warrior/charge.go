package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

var chargeRank = genRanks.Charge.BySpellID(11578)

func (war *Warrior) registerCharge() {
	actionID := core.ActionID{SpellID: chargeRank.SpellID}
	metrics := war.NewRageMetrics(actionID)

	chargeMinRange := chargeRank.MinRange

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
		MaxRange:       chargeRank.MaxRange,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: chargeRank.Cooldown,
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
