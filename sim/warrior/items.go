package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

// Dungeon Set 3 - Tank
var ItemSetBoldArmor = core.NewItemSet(core.ItemSet{
	Name: "Bold Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask: SpellMaskShouts,
					Kind:      core.SpellMod_PowerCost_Flat,
					IntValue:  -2,
				}).
				ExposeToAPL(37512)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			warrior := agent.(WarriorAgent).GetWarrior()

			setBonusAura.
				ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
					warrior.ChargeRageGain += 5
				}).
				ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
					warrior.ChargeRageGain -= 5
				}).
				ExposeToAPL(37513)
		},
	},
})

// T4 - DPS
var ItemSetWarbringerBattlegear = core.NewItemSet(core.ItemSet{
	Name: "Warbringer Battlegear",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask: SpellMaskWhirlwind,
					Kind:      core.SpellMod_PowerCost_Flat,
					IntValue:  -5,
				}).
				ExposeToAPL(37518)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			warrior := agent.(WarriorAgent).GetWarrior()
			actionID := core.ActionID{SpellID: 37521}
			rageMetrics := warrior.NewRageMetrics(actionID)

			setBonusAura.
				AttachProcTrigger(core.ProcTrigger{
					Name:               "Warbringer Battlegear - 4PC",
					TriggerImmediately: true,
					Callback:           core.CallbackOnSpellHitDealt,
					Outcome:            core.OutcomeParry | core.OutcomeDodge,
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						warrior.AddRage(sim, 2, rageMetrics)
					},
				}).
				ExposeToAPL(actionID.SpellID)
		},
	},
})

// T4 - Tank
var ItemSetWarbringerArmor = core.NewItemSet(core.ItemSet{
	Name: "Warbringer Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			warrior := agent.(WarriorAgent).GetWarrior()
			actionID := core.ActionID{SpellID: 37515}

			shield := warrior.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
				Aura: core.Aura{
					Label:    "Blade Turning" + warrior.Label,
					ActionID: actionID,
					Duration: time.Second * 15,
				},
				ShieldStrengthCalculator: func(_ *core.Unit) float64 {
					return 200
				},
			})

			setBonusAura.
				AttachProcTrigger(core.ProcTrigger{
					Name:               "Warbringer Armor - 2PC",
					TriggerImmediately: true,
					Callback:           core.CallbackOnSpellHitTaken,
					Outcome:            core.OutcomeParry,
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						shield.Activate(sim)
					},
				}).
				ExposeToAPL(actionID.SpellID)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			warrior := agent.(WarriorAgent).GetWarrior()

			revengeMod := warrior.AddDynamicMod(core.SpellModConfig{
				ClassMask:  SpellMaskDirectDamageSpells,
				Kind:       core.SpellMod_DamageDone_Flat,
				FloatValue: 0.1,
			})

			setBonusAura.
				AttachProcTrigger(core.ProcTrigger{
					Name:               "Warbringer Armor 4PC - Trigger",
					TriggerImmediately: true,
					ClassSpellMask:     SpellMaskRevenge,
					Callback:           core.CallbackOnSpellHitDealt,
					Outcome:            core.OutcomeLanded,
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						revengeMod.Activate()
					},
				}).
				AttachProcTrigger(core.ProcTrigger{
					Name:               "Warbringer Armor 4PC - Deactivate",
					TriggerImmediately: true,
					ClassSpellMask:     SpellMaskDirectDamageSpells,
					Callback:           core.CallbackOnSpellHitDealt,
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						revengeMod.Deactivate()
					},
				}).
				ExposeToAPL(38407)
		},
	},
})

// T5 - DPS
var ItemSetDestroyerBattlegear = core.NewItemSet(core.ItemSet{
	Name: "Destroyer Battlegear",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask: SpellMaskWhirlwind,
					Kind:      core.SpellMod_PowerCost_Flat,
					IntValue:  -5,
				}).
				ExposeToAPL(37518)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			warrior := agent.(WarriorAgent).GetWarrior()
			actionID := core.ActionID{SpellID: 37521}
			rageMetrics := warrior.NewRageMetrics(actionID)

			setBonusAura.
				AttachProcTrigger(core.ProcTrigger{
					Name:               "Destroyer Battlegear - 4PC",
					TriggerImmediately: true,
					Callback:           core.CallbackOnSpellHitDealt,
					Outcome:            core.OutcomeParry | core.OutcomeDodge,
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						warrior.AddRage(sim, 2, rageMetrics)
					},
				}).
				ExposeToAPL(actionID.SpellID)
		},
	},
})

// T5 - Tank
var ItemSetDestroyerArmor = core.NewItemSet(core.ItemSet{
	Name: "Destroyer Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			warrior := agent.(WarriorAgent).GetWarrior()
			actionID := core.ActionID{SpellID: 37529}

			aura := warrior.NewTemporaryStatsAura(
				"Overpower",
				actionID,
				stats.Stats{stats.AttackPower: 100},
				time.Second*5,
			)

			setBonusAura.
				AttachProcTrigger(core.ProcTrigger{
					Name:               "Destroyer Armor - 2PC",
					ClassSpellMask:     SpellMaskOverpower,
					TriggerImmediately: true,
					Callback:           core.CallbackOnSpellHitDealt,
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						aura.Activate(sim)
					},
				}).
				ExposeToAPL(actionID.SpellID)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask: SpellMaskMortalStrike | SpellMaskBloodthirst,
					Kind:      core.SpellMod_PowerCost_Flat,
					IntValue:  -5,
				}).
				ExposeToAPL(37535)
		},
	},
})

// T6 - DPS
var ItemSetOnslaughtBattlegear = core.NewItemSet(core.ItemSet{
	Name: "Onslaught Battlegear",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask: SpellMaskExecute,
					Kind:      core.SpellMod_PowerCost_Flat,
					IntValue:  -3,
				}).
				ExposeToAPL(38398)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask:  SpellMaskMortalStrike | SpellMaskBloodthirst,
					Kind:       core.SpellMod_DamageDone_Flat,
					FloatValue: 0.05,
				}).
				ExposeToAPL(38399)
		},
	},
})

// T6 - Tank
var ItemSetOnslaughtArmor = core.NewItemSet(core.ItemSet{
	Name: "Onslaught Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.ExposeToAPL(38408)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask:  SpellMaskShieldSlam,
					Kind:       core.SpellMod_DamageDone_Flat,
					FloatValue: 0.1,
				}).
				ExposeToAPL(38407)
		},
	},
})

func init() {}
