package arcane

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/mage"
)

func (arcane *ArcaneMage) registerArcanePowerCD() {
	arcane.ArcanePowerDamageMod = arcane.AddDynamicMod(core.SpellModConfig{
		FloatValue: 0.20,
		Kind:       core.SpellMod_DamageDone_Pct,
	})

	arcanePowerCostMod := arcane.AddDynamicMod(core.SpellModConfig{
		ClassMask:  mage.MageSpellsAllDamaging,
		FloatValue: 0.1,
		Kind:       core.SpellMod_PowerCost_Pct,
	})

	actionID := core.ActionID{SpellID: 12042}
	arcane.arcanePowerAura = arcane.RegisterAura(core.Aura{
		Label:    "Arcane Power",
		ActionID: actionID,
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			arcane.ArcanePowerDamageMod.Activate()
			arcanePowerCostMod.Activate()
		},
		OnExpire: func(_ *core.Aura, sim *core.Simulation) {
			arcane.ArcanePowerDamageMod.Deactivate()
			arcanePowerCostMod.Deactivate()
		},
	})

	arcane.arcanePower = arcane.RegisterSpell(core.SpellConfig{
		ActionID:        actionID,
		ClassSpellMask:  mage.MageSpellArcanePower,
		Flags:           core.SpellFlagNoOnCastComplete,
		RelatedSelfBuff: arcane.arcanePowerAura,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    arcane.NewTimer(),
				Duration: time.Second * 90,
			},
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			arcane.arcanePowerAura.Activate(sim)
		},
	})

	arcane.AddMajorCooldown(core.MajorCooldown{
		Spell: arcane.arcanePower,
		Type:  core.CooldownTypeDPS,
	})
}
