package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

// Dungeon Set 3 - DPS - Plate
var ItemSetDoomplateBattlegear = core.NewItemSet(core.ItemSet{
	Name: "Doomplate Battlegear",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.MeleeHitRating, 35).
				ExposeToAPL(37610)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your attacks have a chance to grant you 160 attack power for 15 sec.
			aura := agent.GetCharacter().NewTemporaryStatsAura(
				"Heroic Resolution",
				core.ActionID{SpellID: 37612},
				stats.Stats{stats.AttackPower: 160, stats.RangedAttackPower: 160},
				time.Second*15,
			)

			setBonusAura.
				AttachProcTrigger(core.ProcTrigger{
					Name:       "Doomplate Battlegear - 4PC",
					ProcChance: 0.02,
					Callback:   core.CallbackOnSpellHitDealt,
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						aura.Activate(sim)
					},
				}).
				ExposeToAPL(37611)
		},
	},
})

// Dungeon Set 3 - DPS - Leather
var ItemSetWastewalkerArmor = core.NewItemSet(core.ItemSet{
	Name: "Wastewalker Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases your hit rating by 35.
			setBonusAura.
				AttachStatBuff(stats.MeleeHitRating, 35).
				ExposeToAPL(37610)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your attacks have a chance to grant you 160 attack power for 15 sec.
			aura := agent.GetCharacter().NewTemporaryStatsAura(
				"Heroic Resolution",
				core.ActionID{SpellID: 37612},
				stats.Stats{stats.AttackPower: 160, stats.RangedAttackPower: 160},
				time.Second*15,
			)

			setBonusAura.
				AttachProcTrigger(core.ProcTrigger{
					Name:       "Wastewalker Armor - 4PC",
					ProcChance: 0.02,
					Callback:   core.CallbackOnSpellHitDealt,
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						aura.Activate(sim)
					},
				}).
				ExposeToAPL(37611)
		},
	},
})

// Blacksmithing - Plate
var ItemSetBurningRage = core.NewItemSet(core.ItemSet{
	Name: "Burning Rage",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.MeleeHitRating, 20).
				ExposeToAPL(41678)
		},
	},
})

// Dungeon Set 3 - Cloth
var ItemSetManaEtchedRegalia = core.NewItemSet(core.ItemSet{
	ID:   658,
	Name: "Mana-Etched Regalia",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases your spell hit rating by 35
			// Spell Hit Rating - 37607
			setBonusAura.AttachStatBuff(stats.SpellHitRating, 35)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your harmful spells have a chance to grant you up to 110 spell damage and healing for 15 sec
			// Spell Power Bonus - 37619
			clothie := agent.GetCharacter()

			bonusPower := clothie.NewTemporaryStatsAura("Spell Power Bonus", core.ActionID{SpellID: 37619}, stats.Stats{stats.SpellDamage: 110, stats.HealingPower: 110}, time.Second*15)

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:       "Mana Etched Regalia 4pc",
				ProcChance: 0.02,
				ProcMask:   core.ProcMaskSpellDamage,
				Callback:   core.CallbackOnSpellHitDealt,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					bonusPower.Activate(sim)
				},
			})
		},
	},
})

func init() {}
