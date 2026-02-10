package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerArcanePowerSpell() {
	if !mage.Talents.ArcanePower {
		return
	}

	arcanePowerCostMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll,
		FloatValue: .3,
		Kind:       core.SpellMod_PowerCost_Pct,
	})

	arcanePowerDmgMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll,
		FloatValue: .3,
		Kind:       core.SpellMod_DamageDone_Pct,
	})

	var arcanePowerSpell *core.Spell
	mage.ArcanePowerAura = mage.RegisterAura(core.Aura{
		Label:    "Arcane Power",
		ActionID: core.ActionID{SpellID: 12042},
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			arcanePowerCostMod.Activate()
			arcanePowerDmgMod.Activate()
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			arcanePowerCostMod.Deactivate()
			arcanePowerDmgMod.Deactivate()
			arcanePowerSpell.CD.Use(sim)
		},
	})

	arcanePowerSpell = mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 12042},
		Flags:          core.SpellFlagNoOnCastComplete,
		ClassSpellMask: MageSpellArcanePower,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: time.Second * 180,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			mage.ArcanePowerAura.Activate(sim)
		},
		RelatedSelfBuff: mage.ArcanePowerAura,
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: arcanePowerSpell,
		Type:  core.CooldownTypeDPS,
	})
}
