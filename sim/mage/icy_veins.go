package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerIcyVeinsSpell() {
	if !mage.Talents.IcyVeins {
		return
	}

	mage.IcyVeinsAura = mage.RegisterAura(core.Aura{
		Label:    "Icy Veins",
		ActionID: core.ActionID{SpellID: 12472},
		Duration: time.Second * 20,
	}).AttachMultiplyCastSpeed(1.2)

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
		RelatedSelfBuff: mage.IcyVeinsAura,
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: mage.IcyVeins,
		Type:  core.CooldownTypeDPS,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			if mage.IcyVeinsAura.IsActive() {
				return false
			}

			return true
		},
	})

}
