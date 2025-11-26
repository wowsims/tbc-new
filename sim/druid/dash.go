package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (druid *Druid) registerDashCD() {
	actionID := core.ActionID{SpellID: 1850}

	druid.DashAura = druid.RegisterAura(core.Aura{
		Label:    "Dash",
		ActionID: actionID,
		Duration: time.Second * 15,
	})

	exclusiveSpeedEffect := druid.DashAura.NewActiveMovementSpeedEffect(0.7)

	druid.Dash = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID:        actionID,
		RelatedSelfBuff: druid.DashAura,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return !exclusiveSpeedEffect.Category.AnyActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			if !druid.InForm(Cat) {
				druid.CatFormAura.Activate(sim)
			}

			druid.DashAura.Activate(sim)
		},
	})

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.Dash.Spell,
		Type:  core.CooldownTypeDPS,

		ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
			return (character.DistanceFromTarget > core.MaxMeleeRange) && (character.GetAura("Nitro Boosts") == nil)
		},
	})
}
