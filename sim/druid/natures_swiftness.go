package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (druid *Druid) registerNaturesSwiftness() {
	if druid.Spec == proto.Spec_SpecGuardianDruid {
		return
	}

	actionID := core.ActionID{SpellID: 132158}
	cdTimer := druid.NewTimer()
	cd := time.Minute * 1

	nsAura := druid.RegisterAura(core.Aura{
		Label:    "Nature's Swiftness",
		ActionID: actionID,
		Duration: core.NeverExpires,

		OnReset: func(_ *core.Aura, _ *core.Simulation) {
			druid.HealingTouch.FormMask = Humanoid | Moonkin
		},

		OnGain: func(_ *core.Aura, _ *core.Simulation) {
			druid.HealingTouch.FormMask |= Cat
		},

		OnExpire: func(_ *core.Aura, _ *core.Simulation) {
			druid.HealingTouch.FormMask ^= Cat
		},

		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !spell.Matches(DruidSpellHealingTouch) {
				return
			}
			aura.Deactivate(sim)
			cdTimer.Set(sim.CurrentTime + cd)
			druid.UpdateMajorCooldowns()
		},
	}).AttachSpellMod(core.SpellModConfig{
		ClassMask:  DruidHealingNonInstantSpells,
		Kind:       core.SpellMod_CastTime_Pct,
		FloatValue: -1,
	}).AttachSpellMod(core.SpellModConfig{
		ClassMask:  DruidHealingNonInstantSpells,
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -2,
	}).AttachSpellMod(core.SpellModConfig{
		ClassMask:  DruidHealingNonInstantSpells,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.5,
	})

	druid.NaturesSwiftness = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID:        actionID,
		Flags:           core.SpellFlagNoOnCastComplete,
		RelatedSelfBuff: nsAura,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: cd,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},
	})

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.NaturesSwiftness.Spell,
		Type:  core.CooldownTypeDPS,
	})
}
