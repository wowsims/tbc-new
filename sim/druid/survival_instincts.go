package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (druid *Druid) registerSurvivalInstinctsCD() {
	actionID := core.ActionID{SpellID: 61336}

	druid.SurvivalInstinctsAura = druid.RegisterAura(core.Aura{
		Label:    "Survival Instincts",
		ActionID: actionID,
		Duration: time.Second * 12,
		OnGain: func(aura *core.Aura, _ *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier *= 0.5
		},

		OnExpire: func(aura *core.Aura, _ *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier /= 0.5
		},
	})

	druid.SurvivalInstincts = druid.RegisterSpell(Cat|Bear, core.SpellConfig{
		ActionID: actionID,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.SurvivalInstinctsAura.Activate(sim)
		},
		RelatedSelfBuff: druid.SurvivalInstinctsAura,
	})

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.SurvivalInstincts.Spell,
		Type:  core.CooldownTypeSurvival,
	})
}
