package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

///////////////////////////////////////////////////////////////////////////
//							Dungeon
///////////////////////////////////////////////////////////////////////////

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

///////////////////////////////////////////////////////////////////////////
//							Blacksmithing
///////////////////////////////////////////////////////////////////////////

// Blacksmithing - Plate
var ItemSetBurningRage = core.NewItemSet(core.ItemSet{
	Name:               "Burning Rage",
	RequiredProfession: proto.Profession_Blacksmithing,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.MeleeHitRating, 20).
				ExposeToAPL(41678)
		},
	},
})

// Blacksmithing - Plate
var ItemSetKhoriumWard = core.NewItemSet(core.ItemSet{
	ID:                 565,
	Name:               "Khorium Ward",
	RequiredProfession: proto.Profession_Blacksmithing,
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatsBuff(stats.Stats{stats.HealingPower: 55, stats.SpellDamage: 19}).
				ExposeToAPL(41783)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//							Leatherworking
///////////////////////////////////////////////////////////////////////////

// Leatherworking - Dragonscale
var ItemSetNetherstrikeArmor = core.NewItemSet(core.ItemSet{
	ID:   617,
	Name: "Netherstrike Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatsBuff(stats.Stats{stats.SpellDamage: 23, stats.HealingPower: 23}).
				ExposeToAPL(41828)
		},
	},
})

var ItemSetNetherscaleArmor = core.NewItemSet(core.ItemSet{
	ID:   616,
	Name: "Netherscale Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.MeleeHitRating, 20).
				ExposeToAPL(39982)
		},
	},
})

// Leatherworking - Tribal
var ItemSetWindhawkArmor = core.NewItemSet(core.ItemSet{
	ID:   618,
	Name: "Windhawk Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.MP5, 8).
				ExposeToAPL(41591)
		},
	},
})

// Leatherworking - Elemental
var ItemSetPrimalIntent = core.NewItemSet(core.ItemSet{
	ID:   619,
	Name: "Primal Intent",
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.AttackPower, 40).
				ExposeToAPL(41832)
		},
	},
})

var ItemSetFelscaleArmor = core.NewItemSet(core.ItemSet{
	ID:   611,
	Name: "Felscale Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.MeleeCritRating, 15).
				ExposeToAPL(41748)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.Stamina, 20).
				ExposeToAPL(41747)
		},
	},
})

var ItemSetThickDraenicArmor = core.NewItemSet(core.ItemSet{
	ID:   613,
	Name: "Thick Draenic Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.MeleeHitRating, 15).
				ExposeToAPL(41878)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.MeleeCritRating, 15).
				ExposeToAPL(41877)
		},
	},
})

// Leatherworking - Fel Skin
var ItemSetFelSkin = core.NewItemSet(core.ItemSet{
	ID:                 573,
	Name:               "Fel Skin",
	RequiredProfession: proto.Profession_Leatherworking,
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.DodgeRating, 20).
				ExposeToAPL(41746)
		},
	},
})

// Leatherworking - Strength of the Clefthoof
var ItemSetStrengthOfTheClefthoof = core.NewItemSet(core.ItemSet{
	ID:                 574,
	Name:               "Strength of the Clefthoof",
	RequiredProfession: proto.Profession_Leatherworking,
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.Strength, 20).
				ExposeToAPL(41846)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//							Tailoring
///////////////////////////////////////////////////////////////////////////

// Tailoring - Battlecast
var ItemSetBattlecastGarb = core.NewItemSet(core.ItemSet{
	ID:                 572,
	Name:               "Battlecast Garb",
	RequiredProfession: proto.Profession_Tailoring,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases the chance spell pushback and spell interrupt will be resisted by 5%.
			// Not modeled in the sim.
			setBonusAura.ExposeToAPL(41660)
		},
	},
})

// Tailoring - Spellstrike
var ItemSetSpellstrikeInfusion = core.NewItemSet(core.ItemSet{
	ID:                 559,
	Name:               "Spellstrike Infusion",
	RequiredProfession: proto.Profession_Tailoring,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Gives a chance when your harmful spells land to increase the damage of your spells and effects by 92 for 10 sec.
			// Spell Damage Bonus - 32108
			character := agent.GetCharacter()
			bonusPower := character.NewTemporaryStatsAura("Lesser Spell Blasting", core.ActionID{SpellID: 32108}, stats.Stats{stats.SpellDamage: 92}, time.Second*10)

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:            "Spellstrike Infusion 2pc",
				ProcChance:      0.05,
				ProcMask:        core.ProcMaskSpellOrSpellProc,
				Callback:        core.CallbackOnSpellHitDealt,
				ClassSpellsOnly: true,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					bonusPower.Activate(sim)
				},
			})
		},
	},
})

// Tailoring - Spellfire
var ItemSetWrathOfSpellfire = core.NewItemSet(core.ItemSet{
	ID:   552,
	Name: "Wrath of Spellfire",
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases spell damage by up to 7% of your total Intellect.
			character := agent.GetCharacter()
			statDep := character.Unit.NewDynamicStatDependency(stats.Intellect, stats.SpellDamage, 0.07)
			setBonusAura.AttachStatDependency(statDep)
		},
	},
})

// Tailoring - Whitemend
var ItemSetWhitemendWisdom = core.NewItemSet(core.ItemSet{
	ID:                 571,
	Name:               "Whitemend Wisdom",
	RequiredProfession: proto.Profession_Tailoring,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases healing by up to 10% of your total Intellect.
			character := agent.GetCharacter()
			statDep := character.Unit.NewDynamicStatDependency(stats.Intellect, stats.HealingPower, 0.10)
			setBonusAura.AttachStatDependency(statDep)
		},
	},
})

// Tailoring - Shadow's Embrace
var ItemSetShadowsEmbrace = core.NewItemSet(core.ItemSet{
	ID:   618,
	Name: "Shadow's Embrace",
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Frost and Shadow damage spells heal you for 2% of the damage they deal.
			// TODO
		},
	},
})

// Tailoring - Soulcloth Embrace
var ItemSetSoulclothEmbrace = core.NewItemSet(core.ItemSet{
	ID:   557,
	Name: "Soulcloth Embrace",
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.SpellHitRating, 16)
		},
	},
})

// Tailoring - Primal Mooncloth
var ItemSetPrimalMooncloth = core.NewItemSet(core.ItemSet{
	ID:   554,
	Name: "Primal Mooncloth",
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			// Allow 5% of your Mana regeneration to continue while casting.
			character := agent.GetCharacter()

			setBonusAura.
				ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
					character.PseudoStats.SpiritRegenRateCasting += 0.05
					character.UpdateManaRegenRates()
				}).
				ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
					character.PseudoStats.SpiritRegenRateCasting -= 0.05
					character.UpdateManaRegenRates()
				})
		},
	},
})

// Tailoring - Netherweave
var ItemSetImbuedNetherweave = core.NewItemSet(core.ItemSet{
	ID:   556,
	Name: "Imbued Netherweave",
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.SpellCritRating, 28)
		},
	},
})

// Tailoring - Netherweave
var ItemSetNetherweaveVestments = core.NewItemSet(core.ItemSet{
	ID:   555,
	Name: "Netherweave Vestments",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatsBuff(stats.Stats{stats.SpellDamage: 23, stats.HealingPower: 23})
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.SpellCritRating, 14)
		},
	},
})

// Tailoring - Arcanoweave
var ItemSetArcanoweaveVestments = core.NewItemSet(core.ItemSet{
	ID:   558,
	Name: "Arcanoweave Vestments",
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.SpellHitRating, 16)
		},
	},
})

func init() {}
