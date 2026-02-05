package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

var PVPSet = core.NewItemSet(core.ItemSet{
	Name: "Gladiator's Vestments",
	ID:   577,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			agent.GetCharacter().AddStat(stats.ResilienceRating, 35)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			rogue := agent.(RogueAgent).GetRogue()
			rogue.HasPvpEnergy = true
		},
	},
})

var Tier4 = core.NewItemSet(core.ItemSet{
	Name: "Netherblade",
	ID:   621,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			agent.(RogueAgent).GetRogue().SliceAndDiceBonusDuration += 3
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			rogue := agent.(RogueAgent).GetRogue()
			pointMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 37168})
			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:           "Netherblade Combo Point",
				ActionID:       core.ActionID{SpellID: 37168},
				ProcChance:     0.15,
				ClassSpellMask: RogueSpellFinisher,
				Callback:       core.CallbackOnApplyEffects,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					rogue.AddComboPoints(sim, 1, pointMetrics)
				},
			})

		},
	},
})

var Tier5 = core.NewItemSet(core.ItemSet{
	Name: "Deathmantle",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			agent.(RogueAgent).GetRogue().DeathmantleBonus = 40
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			rogue := agent.(RogueAgent).GetRogue()
			mod := rogue.GetOrRegisterAura(core.Aura{
				Label:    "Coup de Grace",
				Duration: time.Second * 15,
				ActionID: core.ActionID{SpellID: 37171},
			}).AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_PowerCost_Pct,
				ClassMask:  RogueSpellFinisher,
				FloatValue: -2,
			})
			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name: "Deathmantle Proc Trigger",
				DPM:  rogue.NewLegacyPPMManager(1.0, core.ProcMaskMelee),
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					mod.Activate(sim)
				},
			}).ExposeToAPL(37170)
		},
	},
})

var Tier6 = core.NewItemSet(core.ItemSet{
	Name: "Slayer's Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			rogue := agent.(RogueAgent).GetRogue()
			rogue.SliceAndDiceBonusFlat += 0.05
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Flat,
				ClassMask:  RogueSpellBackstab | RogueSpellSinisterStrike | RogueSpellMutilate | RogueSpellHemorrhage,
				FloatValue: 0.06,
			})
		},
	},
})
