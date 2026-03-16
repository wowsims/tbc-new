package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
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

		character.ItemSwap.RegisterProc(32330, aura)
	})

	// Ivory Idol of the Moongoddess
	core.NewItemEffect(27518, func(agent core.Agent) {
		// Increases the damage of your Starfire spell by up to 55.
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Starfire Damage",
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask:  DruidSpellStarfire,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 55.0,
		}))

		character.ItemSwap.RegisterProc(32330, aura)
	})

	core.NewItemEffect(32486, func(agent core.Agent) {
		// Ashtongue Talisman of Equilibrium
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
