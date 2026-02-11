package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerCombustionSpell() {
	if !mage.Talents.Combustion {
		return
	}

	actionID := core.ActionID{SpellID: 11129}
	cd := core.Cooldown{
		Timer:    mage.NewTimer(),
		Duration: time.Minute * 3,
	}

	numCrits := 0
	critPerStack := 0.0

	critMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellFire,
		FloatValue: critPerStack,
		Kind:       core.SpellMod_BonusCrit_Percent,
	})

	combustAura := mage.RegisterAura(core.Aura{
		Label:     "Combustion",
		ActionID:  actionID,
		Duration:  core.NeverExpires,
		MaxStacks: 20,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			numCrits = 0
			critMod.Activate()
		},

		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			critMod.Deactivate()
			cd.Use(sim)
			mage.UpdateMajorCooldowns()
		},

		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
			critPerStack = float64(newStacks) * 10
			critMod.UpdateFloatValue(critPerStack)
		},

		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.SpellSchool.Matches(core.SpellSchoolFire) {
				return
			}

			if spell.SameAction(mage.Ignite.ActionID) {
				return
			}

			if !result.Landed() {
				return
			}

			if numCrits >= 3 {
				aura.Deactivate(sim)
				return
			}

			aura.AddStack(sim)

			if result.DidCrit() {
				numCrits++
				if numCrits == 3 {
					aura.Deactivate(sim)
				}
			}
		},
	})

	combustSpell := mage.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: cd,
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !combustAura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			combustAura.Activate(sim)
		},

		RelatedSelfBuff: combustAura,
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: combustSpell,
		Type:  core.CooldownTypeDPS,
	})
}
