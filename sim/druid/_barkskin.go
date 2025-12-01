package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (druid *Druid) registerBarkskinCD() {
	actionId := core.ActionID{SpellID: 22812}

	druid.BarkskinAura = druid.RegisterAura(core.Aura{
		Label:    "Barkskin",
		ActionID: actionId,
		Duration: time.Second * 12,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			druid.PseudoStats.DamageTakenMultiplier *= 0.8
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.PseudoStats.DamageTakenMultiplier /= 0.8
		},
	})

	druid.Barkskin = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID: actionId,
		Flags:    core.SpellFlagAPL | core.SpellFlagReadinessTrinket,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: core.TernaryDuration(druid.Spec == proto.Spec_SpecGuardianDruid, time.Second*30, time.Second*60),
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.BarkskinAura.Activate(sim)

			if sim.CurrentTime > 0 {
				druid.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime)
			}
		},
		RelatedSelfBuff: druid.BarkskinAura,
	})

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.Barkskin.Spell,
		Type:  core.CooldownTypeSurvival,
	})
}
