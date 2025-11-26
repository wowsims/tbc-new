package feral

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/druid"
)

// T16 Feral
var ItemSetBattlegearOfTheShatteredVale = core.NewItemSet(core.ItemSet{
	ID:                      1197,
	DisabledInChallengeMode: true,
	Name:                    "Battlegear of the Shattered Vale",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Omen of Clarity increases damage of Shred, Mangle, Swipe, and Ravage by 50% for 6 sec.
			cat := agent.(*FeralDruid)
			cat.registerFeralFury(setBonusAura)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// After using Tiger's Fury, your next finishing move will restore 3 combo points on your current target after being used.
			cat := agent.(*FeralDruid)
			cat.registerFeralRage()

			setBonusAura.OnCastComplete = func(_ *core.Aura, sim *core.Simulation, spell *core.Spell) {
				if spell.Matches(druid.DruidSpellTigersFury) {
					cat.FeralRageAura.Activate(sim)
				}
			}
		},
	},
})

func (cat *FeralDruid) registerFeralFury(setBonusTracker *core.Aura) {
	cat.FeralFuryBonus = setBonusTracker
	meleeAbilityMask := druid.DruidSpellMangleCat | druid.DruidSpellShred | druid.DruidSpellRavage | druid.DruidSpellSwipeCat

	feralFuryMod := cat.AddDynamicMod(core.SpellModConfig{
		ClassMask:  meleeAbilityMask,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.5,
	})

	cat.FeralFuryAura = cat.RegisterAura(core.Aura{
		Label:    "Feral Fury 2PT16",
		ActionID: core.ActionID{SpellID: 144865},
		Duration: time.Second * 6,

		OnGain: func(_ *core.Aura, _ *core.Simulation) {
			feralFuryMod.Activate()
		},

		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Matches(meleeAbilityMask) {
				aura.Deactivate(sim)
			}
		},

		OnExpire: func(_ *core.Aura, _ *core.Simulation) {
			feralFuryMod.Deactivate()
		},
	})
}

func (cat *FeralDruid) registerFeralRage() {
	actionID := core.ActionID{SpellID: 146874}
	cpMetrics := cat.NewComboPointMetrics(actionID)

	var resultLanded bool

	cat.FeralRageAura = cat.RegisterAura(core.Aura{
		Label:    "Feral Rage 4PT16",
		ActionID: actionID,
		Duration: time.Second * 12,

		OnSpellHitDealt: func(_ *core.Aura, _ *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.Matches(druid.DruidSpellFinisher) && result.Landed() {
				resultLanded = true
			}
		},

		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !spell.Matches(druid.DruidSpellFinisher) {
				return
			}

			if spell.Matches(druid.DruidSpellSavageRoar) || resultLanded {
				aura.Unit.AddComboPoints(sim, 3, cpMetrics)
				resultLanded = false
				aura.Deactivate(sim)
			}
		},
	})
}

func init() {
}
