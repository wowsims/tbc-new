package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// Communal Totem of Lightning
	core.NewItemEffect(186071, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Lightning Damage",
		}).AttachSpellMod(core.SpellModConfig{
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 10.0,
			ClassMask:  SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload,
		}))

		character.ItemSwap.RegisterProc(186071, aura)
	})

	// Skycall Totem
	core.NewItemEffect(33506, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := character.NewTemporaryStatsAura("Energized", core.ActionID{SpellID: 43751}, stats.Stats{stats.SpellHasteRating: 100}, time.Second*10)
		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:           "Skycall Totem",
			ActionID:       core.ActionID{ItemID: 43750},
			Callback:       core.CallbackOnSpellHitDealt,
			Outcome:        core.OutcomeLanded,
			ClassSpellMask: SpellMaskLightningBolt | SpellMaskLightningBoltOverload,
			ProcChance:     0.15,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})

		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(33506)
		character.AddStatProcBuff(39441, aura, false, eligibleSlots)
		character.ItemSwap.RegisterProc(33506, procAura)
	})

	// Totem of Ancestral Guidance
	core.NewItemEffect(32330, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Lightning Damage",
		}).AttachSpellMod(core.SpellModConfig{
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 85.0,
			ClassMask:  SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload,
		}))

		character.ItemSwap.RegisterProc(32330, aura)
	})

	// Totem of Impact (A)
	core.NewItemEffect(27984, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Shock Damage",
		}).AttachSpellMod(core.SpellModConfig{
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 46.0,
			ClassMask:  SpellMaskShock,
		}))

		character.ItemSwap.RegisterProc(27984, aura)
	})

	// Totem of Impact (H)
	core.NewItemEffect(27947, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Shock Damage",
		}).AttachSpellMod(core.SpellModConfig{
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 46.0,
			ClassMask:  SpellMaskShock,
		}))

		character.ItemSwap.RegisterProc(27947, aura)
	})

	// Totem of Lightning
	core.NewItemEffect(28066, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Reduced Lightning Cost",
		}).AttachSpellMod(core.SpellModConfig{
			Kind:      core.SpellMod_PowerCost_Flat,
			IntValue:  -15,
			ClassMask: SpellMaskLightningBolt,
		}))

		character.ItemSwap.RegisterProc(28066, aura)
	})

	// Totem of Rage
	// core.NewItemEffect(22395, func(agent core.Agent) {
	// 	character := agent.GetCharacter()

	// 	aura := core.MakePermanent(character.RegisterAura(core.Aura{
	// 		Label: "Increased Shock Damage",
	// 	}).AttachSpellMod(core.SpellModConfig{
	// 		Kind:       core.SpellMod_BaseDamage_Flat,
	// 		FloatValue: 30.0,
	// 		ClassMask:  SpellMaskShock,
	// 	}))

	// 	character.ItemSwap.RegisterProc(22395, aura)
	// })

	// Totem of the Storm
	core.NewItemEffect(23199, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Lightning Damage",
		}).AttachSpellMod(core.SpellModConfig{
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 33.0,
			ClassMask:  SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload,
		}))

		character.ItemSwap.RegisterProc(23199, aura)
	})

	// Totem of the Void
	core.NewItemEffect(28248, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Lightning Damage",
		}).AttachSpellMod(core.SpellModConfig{
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 55.0,
			ClassMask:  SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload,
		}))

		character.ItemSwap.RegisterProc(28248, aura)
	})

	core.NewItemEffect(27815, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()

		bonusAP := 80.0
		aura := core.MakePermanent(shaman.RegisterAura(core.Aura{
			Label: "Increased Windfury Weapon AP Bonus",
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				shaman.WindfuryAPBonus += bonusAP
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				shaman.WindfuryAPBonus -= bonusAP
			},
		}))

		shaman.ItemSwap.RegisterProc(27815, aura)
	})

	// Ashtongue Talisman of Vision
	core.NewItemEffect(32491, func(agent core.Agent) {
		// Lesser Healing Wave has a 10% chance to grant 170 mana,
		// Lightning Bolt has a 15% chance to grant up to 170 mana,
		// and Stormstrike has a 50% chance to grant up to 275 attack power for 10 sec.
		// Note: Lesser Healing Wave is not implemented
		shaman := agent.(ShamanAgent).GetShaman()
		ashtongueAuraSS := shaman.NewTemporaryStatsAura("Power Surge", core.ActionID{SpellID: 40466}, stats.Stats{stats.AttackPower: 275}, time.Second*10)
		manaMetrics := shaman.NewManaMetrics(core.ActionID{SpellID: 40465})

		aura := core.MakePermanent(shaman.RegisterAura(core.Aura{
			Label:    "Ashtongue Talisman of Vision - Trigger",
			ActionID: core.ActionID{ItemID: 32491},
		}).AttachProcTrigger(core.ProcTrigger{
			ClassSpellMask: SpellMaskStormstrikeCast | SpellMaskLightningBolt,
			Callback:       core.CallbackOnSpellHitDealt,
			Outcome:        core.OutcomeLanded,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell.Matches(SpellMaskStormstrikeCast) && sim.Proc(0.5, "Ashtongue Talisman of Vision - Stormstrike") {
					ashtongueAuraSS.Activate(sim)
				} else if spell.Matches(SpellMaskLightningBolt) && sim.Proc(0.15, "Ashtongue Talisman of Vision - Lightning Bolt") {
					shaman.AddMana(sim, 170, manaMetrics)
				}

			},
		}))

		shaman.ItemSwap.RegisterProc(32491, aura)
	})
}
