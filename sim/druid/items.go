package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// Idol of Brutality (23198): +50 flat to Maul, +10 flat to Swipe.
	core.NewItemEffect(23198, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		aura := core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Idol of Brutality",
			OnGain: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolMaulBonus += 50
				druid.IdolSwipeBonus += 10
			},
			OnExpire: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolMaulBonus -= 50
				druid.IdolSwipeBonus -= 10
			},
		}))
		druid.ItemSwap.RegisterProc(23198, aura)
	})

	// Idol of Terror (33509): Mangle has 85% chance to grant 65 agility for 10 sec (10s ICD).
	core.NewItemEffect(33509, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		procAura := druid.NewTemporaryStatsAura("Idol of Terror Proc", core.ActionID{SpellID: 43737}, stats.Stats{stats.Agility: 65}, time.Second*10)

		aura := core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Idol of Terror",
		}).AttachProcTrigger(core.ProcTrigger{
			ClassSpellMask: DruidSpellMangle,
			Callback:       core.CallbackOnSpellHitDealt,
			ProcChance:     0.85,
			ICD:            time.Second * 10,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				procAura.Activate(sim)
			},
		}))
		druid.ItemSwap.RegisterProc(33509, aura)
	})

	// Idol of the White Stag (32257): Mangle also increases attack power by 94 for 20 sec.
	core.NewItemEffect(32257, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		procAura := druid.NewTemporaryStatsAura("Idol of the White Stag Proc", core.ActionID{SpellID: 41037}, stats.Stats{stats.AttackPower: 94}, time.Second*20)

		aura := core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Idol of the White Stag",
		}).AttachProcTrigger(core.ProcTrigger{
			ClassSpellMask: DruidSpellMangle,
			Callback:       core.CallbackOnSpellHitDealt,
			ProcChance:     1.0,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				procAura.Activate(sim)
			},
		}))
		druid.ItemSwap.RegisterProc(32257, aura)
	})

	// Idol of the Moon
	core.NewItemEffect(23197, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Improved Moonfire",
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask:  DruidSpellMoonfire,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 33.0,
		}))

		character.ItemSwap.RegisterProc(32330, aura)
	})

	// Ivory Idol of the Moongoddess
	core.NewItemEffect(27518, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Starfire Damage",
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask:  DruidSpellStarfire,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 55.0,
		}))

		character.ItemSwap.RegisterProc(27518, aura)
	})

	// Living Root of the Wildheart
	core.NewItemEffect(30664, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()

		buffAuras := map[DruidForm]*core.StatBuffAura{}
		if druid.Talents.MoonkinForm {
			buffAuras[Moonkin] = druid.NewTemporaryStatsAura("Living Root Moonkin Proc", core.ActionID{SpellID: 37343}, stats.Stats{stats.SpellDamage: 209}, time.Second*15)
		}
		buffAuras[Humanoid] = druid.NewTemporaryStatsAura("Living Root Humanoid Proc", core.ActionID{SpellID: 37344}, stats.Stats{stats.SpellDamage: 175}, time.Second*15)
		buffAuras[Bear] = druid.NewTemporaryStatsAura("Living Root Bear Proc", core.ActionID{SpellID: 37340}, stats.Stats{stats.Armor: 4070}, time.Second*15)
		buffAuras[Cat] = druid.NewTemporaryStatsAura("Living Root Cat Proc", core.ActionID{SpellID: 37341}, stats.Stats{stats.Strength: 64}, time.Second*15)

		aura := core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Living Root of the Wildheart",
		}).AttachProcTrigger(core.ProcTrigger{
			Callback:   core.CallbackOnSpellHitDealt,
			ProcMask:   core.ProcMaskDirect,
			ProcChance: 0.03,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				buffAuras[druid.form].Activate(sim)
			},
		}))
		druid.ItemSwap.RegisterProc(30664, aura)
	})

	// Idol of the Avenger
	core.NewItemEffect(31025, func(agent core.Agent) {
		// Increases the damage dealt by Wrath by 25.
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Wrath Damage",
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask:  DruidSpellWrath,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 25.0,
		}))

		character.ItemSwap.RegisterProc(31025, aura)
	})

	// Staff of Natural Fury
	core.NewItemEffect(31334, func(agent core.Agent) {
		// Reduces the base Mana cost of your shapeshifting spells by 200.
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Staff of Natural Fury",
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask: DruidSpellCatForm | DruidSpellBearForm,
			Kind:      core.SpellMod_PowerCost_Flat,
			IntValue:  -200,
		}))
		character.ItemSwap.RegisterProc(31334, aura)
	})

	// Idol of the Beast (25667): Increases the damage dealt by Ferocious Bite by 14 per combo point.
	core.NewItemEffect(25667, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		aura := core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Idol of the Beast",
			OnGain: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolFerociousBiteBonus += 14
			},
			OnExpire: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolFerociousBiteBonus -= 14
			},
		}))
		druid.ItemSwap.RegisterProc(25667, aura)
	})

	// Idol of the Wild (28064): Increases the damage dealt by Mangle by 24 (Cat) or 52 (Bear).
	core.NewItemEffect(28064, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		aura := core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Idol of the Wild",
			OnGain: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolMangleCatBonus += 24
				druid.IdolMangleBearBonus += 52
			},
			OnExpire: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolMangleCatBonus -= 24
				druid.IdolMangleBearBonus -= 52
			},
		}))
		druid.ItemSwap.RegisterProc(28064, aura)
	})

	// Idol of Feral Shadows (28372): Increases the damage dealt by Rip by 7 per combo point per tick.
	core.NewItemEffect(28372, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		aura := core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Idol of Feral Shadows",
			OnGain: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolRipBonus += 7
			},
			OnExpire: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolRipBonus -= 7
			},
		}))
		druid.ItemSwap.RegisterProc(28372, aura)
	})

	// Everbloom Idol (29390): Increases the damage dealt by Shred by 88.
	core.NewItemEffect(29390, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		aura := core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Everbloom Idol",
			OnGain: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolShredBonus += 88
			},
			OnExpire: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolShredBonus -= 88
			},
		}))
		druid.ItemSwap.RegisterProc(29390, aura)
	})

	// Idol of the Raven Goddess
	core.NewItemEffect(32387, func(agent core.Agent) {
		// Implemented naively in druid.go
	})

	// Idol of Ursoc (27744): Increases the damage dealt by Lacerate by 8 per tick per stack.
	core.NewItemEffect(27744, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		aura := core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Idol of Ursoc",
			OnGain: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolLacerateBonus += 8
			},
			OnExpire: func(_ *core.Aura, _ *core.Simulation) {
				druid.IdolLacerateBonus -= 8
			},
		}))
		druid.ItemSwap.RegisterProc(27744, aura)
	})

	// Ashtongue Talisman of Equilibrium
	core.NewItemEffect(32486, func(agent core.Agent) {
		// Mangle has a 40% chance to grant 140 Strength for 8 sec,
		// Starfire has a 25% chance to grant up to 150 spell damage for 8 sec, and
		// Rejuvenation has a 25% chance to grant up to 210 healing for 8 sec.
		druid := agent.(DruidAgent).GetDruid()
		ashtongueAuraMangle := druid.NewTemporaryStatsAura("Ashtongue Talisman of Equilibrium (Mangle)", core.ActionID{SpellID: 32486}, stats.Stats{stats.Strength: 140}, time.Second*8)
		ashtongueAuraStarfire := druid.NewTemporaryStatsAura("Ashtongue Talisman of Equilibrium (Starfire)", core.ActionID{SpellID: 32486}, stats.Stats{stats.SpellDamage: 150}, time.Second*8)

		aura := core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Ashtongue Talisman of Equilibrium",
		}).AttachProcTrigger(core.ProcTrigger{
			ClassSpellMask: DruidSpellMangle,
			Callback:       core.CallbackOnSpellHitDealt,
			ProcChance:     0.4,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				ashtongueAuraMangle.Activate(sim)
			},
		}).AttachProcTrigger(core.ProcTrigger{
			ClassSpellMask: DruidSpellStarfire,
			Callback:       core.CallbackOnSpellHitDealt,
			ProcChance:     0.25,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				ashtongueAuraStarfire.Activate(sim)
			},
		}))

		druid.ItemSwap.RegisterProc(32486, aura)
		druid.ItemSwap.RegisterActive(32486)
	})
}
