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

var Dungeon3 = core.NewItemSet(core.ItemSet{
	Name: "Assassination Armor",
	ID:   620,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Cheap Shot and Kidney Shot attacks grant you 160 haste rating for 6 sec.
			// NYI
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Eviscerate and Envenom abilities cost 10 less energy.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_PowerCost_Flat,
				ClassMask: RogueSpellEviscerate | RogueSpellEnvenom,
				IntValue:  -10,
			})
		},
	},
})

var Tier4 = core.NewItemSet(core.ItemSet{
	Name: "Netherblade",
	ID:   621,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
					agent.(RogueAgent).GetRogue().SliceAndDiceBonusDuration += time.Second * 3
				}).
				ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
					agent.(RogueAgent).GetRogue().SliceAndDiceBonusDuration -= time.Second * 3
				})
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
				OnApplyEffects: func(aura *core.Aura, sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					if spell.Matches(RogueSpellFinisher) {
						aura.Deactivate(sim)
					}
				},
			}).AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_PowerCost_Pct,
				ClassMask:  RogueSpellFinisher,
				FloatValue: -2,
			})
			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:     "Deathmantle Proc Trigger",
				ProcMask: core.ProcMaskMelee,
				Outcome:  core.OutcomeLanded,
				Callback: core.CallbackOnSpellHitDealt,
				DPM:      rogue.NewLegacyPPMManager(1.0, core.ProcMaskMelee),
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					mod.Activate(sim)
				},
			})
		},
	},
})

var Tier6 = core.NewItemSet(core.ItemSet{
	Name: "Slayer's Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
					agent.(RogueAgent).GetRogue().SliceAndDiceBonusFlat += 0.05
				}).
				ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
					agent.(RogueAgent).GetRogue().SliceAndDiceBonusFlat -= 0.05
				})
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

func init() {
	// Warp-Spring Coil
	core.NewItemEffect(30450, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := character.NewTemporaryStatsAura(
			"Warp-Spring Coil",
			core.ActionID{SpellID: 37174},
			stats.Stats{stats.ArmorPenetration: 1000},
			time.Second*15,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Perceived Weakness",
			ActionID:           core.ActionID{ItemID: 30450},
			ProcMask:           core.ProcMaskMeleeSpecial,
			ICD:                time.Second * 30,
			RequireDamageDealt: true,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			ProcChance:         0.25,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})

		character.ItemSwap.RegisterProc(30450, procAura)
	})
}
