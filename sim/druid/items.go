package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
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

	// Idol of the Raven Goddess
	core.NewItemEffect(32387, func(agent core.Agent) {
		// Implemented naively in druid.go
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
