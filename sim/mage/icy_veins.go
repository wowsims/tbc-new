package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerIcyVeinsSpell() {
	if !mage.Talents.IcyVeins {
		return
	}

	icyVeinsMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll,
		FloatValue: -.2,
		Kind:       core.SpellMod_CastTime_Pct,
	})

	mage.IcyVeins = mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 12472},
		Flags:          core.SpellFlagNoOnCastComplete,
		ClassSpellMask: MageSpellIcyVeins,
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
			mage.IcyVeinsAura.Activate(sim)
		},
	})

	mage.IcyVeinsAura = mage.RegisterAura(core.Aura{
		Label:    "Icy Veins",
		ActionID: core.ActionID{SpellID: 12472},
		Duration: time.Second * 20,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			icyVeinsMod.Activate()
			mage.IcyVeins.CD.Use(sim)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Deactivate(sim)
			icyVeinsMod.Deactivate()
		},
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: mage.IcyVeins,
		Type:  core.CooldownTypeDPS,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			if icyVeinsMod.IsActive {
				return false
			}

			return true
		},
	})

}
