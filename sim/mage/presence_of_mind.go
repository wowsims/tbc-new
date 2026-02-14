package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerPresenceOfMindSpell() {
	if !mage.Talents.PresenceOfMind {
		return
	}

	presenceOfMindMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll ^ (MageSpellInstantCast | MageSpellBlizzard | MageSpellEvocation),
		FloatValue: -1,
		Kind:       core.SpellMod_CastTime_Pct,
	})

	var pomSpell *core.Spell
	mage.PresenceOfMindAura = mage.RegisterAura(core.Aura{
		Label:    "Presence of Mind",
		ActionID: core.ActionID{SpellID: 12043},
		Duration: time.Hour,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			presenceOfMindMod.Activate()
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			presenceOfMindMod.Deactivate()
			pomSpell.CD.Use(sim)
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !spell.Matches(MageSpellsAll ^ (MageSpellInstantCast | MageSpellEvocation)) {
				return
			}
			if spell.DefaultCast.CastTime == 0 {
				return
			}
			aura.Deactivate(sim)
		},
	})

	pomSpell = mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 12043},
		Flags:          core.SpellFlagNoOnCastComplete,
		ClassSpellMask: MageSpellPresenceOfMind,
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
			mage.PresenceOfMindAura.Activate(sim)
		},
		RelatedSelfBuff: mage.PresenceOfMindAura,
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: pomSpell,
		Type:  core.CooldownTypeDPS,
	})
}
