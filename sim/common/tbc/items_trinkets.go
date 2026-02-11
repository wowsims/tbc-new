package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// Quagmirran's Eye
	core.NewItemEffect(27683, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 6
		spellHasteRating := 320.0

		quagmirransEyeAura := character.NewTemporaryStatsAura(
			"Spell Haste Trinket",
			core.ActionID{SpellID: 33297},
			stats.Stats{stats.SpellHasteRating: spellHasteRating},
			duration,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Quagmirran's Eye",
			ActionID:           core.ActionID{ItemID: 27683},
			ProcChance:         .1,
			ICD:                time.Second * 45,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			RequireDamageDealt: true,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				quagmirransEyeAura.Activate(sim)
			},
		})

		character.ItemSwap.RegisterProc(27683, procAura)
	})

	// Hourglass of the Unraveller
	core.NewItemEffect(28034, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 6
		value := 300.0

		aura := character.NewTemporaryStatsAura(
			"Rage of the Unraveller",
			core.ActionID{SpellID: 33649},
			stats.Stats{stats.MeleeHasteRating: value},
			duration,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:       "Hourglass of the Unraveller",
			ActionID:   core.ActionID{ItemID: 28034},
			ProcChance: 0.1,
			ICD:        time.Second * 50,
			ProcMask:   core.ProcMaskMeleeOrRanged,
			Outcome:    core.OutcomeCrit,
			Callback:   core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})

		character.ItemSwap.RegisterProc(28034, procAura)
	})

	// The Lightning Capacitor
	core.NewItemEffect(28785, func(agent core.Agent) {
		character := agent.GetCharacter()

		lightningBolt := character.RegisterSpell(core.SpellConfig{
			ActionID:     core.ActionID{SpellID: 42372},
			SpellSchool:  core.SpellSchoolNature,
			ProcMask:     core.ProcMaskEmpty,
			Flags:        core.SpellFlagPassiveSpell | core.SpellFlagIgnoreAttackerModifiers,
			MissileSpeed: 28, // this is a guess atm

			DamageMultiplier: 1,
			CritMultiplier:   character.DefaultSpellCritMultiplier(),

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.WaitTravelTime(sim, func(s *core.Simulation) {
					baseDamage := sim.Roll(694, 806)
					//https://www.wowhead.com/tbc/item=28785/the-lightning-capacitor#comments
					//It can crit, may need some testing
					spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicCrit)
				})

			},
		})

		lightningCapacitorAura := character.RegisterAura(core.Aura{
			Label:     "Electrical Charge",
			ActionID:  core.ActionID{SpellID: 37658},
			Duration:  core.NeverExpires,
			MaxStacks: 3,
			OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
				if newStacks >= 3 {
					aura.SetStacks(sim, newStacks%3)
					aura.Deactivate(sim)
					lightningBolt.Proc(sim, character.CurrentTarget)
				}
			},
		})

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "The Lightning Capacitor",
			ActionID: core.ActionID{ItemID: 28785},
			ProcMask: core.ProcMaskSpellOrSpellProc,
			ICD:      time.Millisecond * 2500,
			Outcome:  core.OutcomeCrit,
			Callback: core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				lightningCapacitorAura.Activate(sim)
				lightningCapacitorAura.AddStack(sim)
			},
		})

		character.ItemSwap.RegisterProc(28785, procAura)
	})

	// Dragonspine Trophy
	core.NewItemEffect(28830, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 6
		value := 325.0

		aura := character.NewTemporaryStatsAura(
			"Haste",
			core.ActionID{SpellID: 34775},
			stats.Stats{stats.MeleeHasteRating: value},
			duration,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Dragonspine Trophy",
			ActionID: core.ActionID{ItemID: 28830},
			DPM:      character.NewLegacyPPMManager(1, core.ProcMaskMeleeOrRanged),
			ICD:      time.Second * 20,
			ProcMask: core.ProcMaskMeleeOrRanged,
			Outcome:  core.OutcomeLanded,
			Callback: core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})

		character.ItemSwap.RegisterProc(28830, procAura)
	})

	// Darkmoon Card: Crusade
	core.NewItemEffect(31856, func(agent core.Agent) {
		character := agent.GetCharacter()

		meleeAura := core.MakeStackingAura(character, core.StackingStatAura{
			Aura: core.Aura{
				Label:     "Aura of the Crusader (Melee)",
				ActionID:  core.ActionID{SpellID: 39438},
				Duration:  time.Second * 10,
				MaxStacks: 20,
			},
			BonusPerStack: stats.Stats{stats.AttackPower: 6, stats.RangedAttackPower: 6},
		})

		casterAura := core.MakeStackingAura(character, core.StackingStatAura{
			Aura: core.Aura{
				Label:     "Aura of the Crusader (Caster)",
				ActionID:  core.ActionID{SpellID: 39441},
				Duration:  time.Second * 10,
				MaxStacks: 10,
			},
			BonusPerStack: stats.Stats{stats.SpellDamage: 8},
		})

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Darkmoon Card: Crusade",
			ActionID: core.ActionID{ItemID: 31856},
			ProcMask: core.ProcMaskDirect | core.ProcMaskProc,
			Outcome:  core.OutcomeLanded,
			Callback: core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura := core.Ternary(spell.ProcMask.Matches(core.ProcMaskSpellDamageProc), casterAura, meleeAura)
				aura.Activate(sim)
				aura.AddStack(sim)
			},
		})

		character.ItemSwap.RegisterProc(28830, procAura)
	})

	// Darkmoon Card: Wrath
	core.NewItemEffect(31857, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := core.MakeStackingAura(character, core.StackingStatAura{
			Aura: core.Aura{
				Label:     "Aura of Wrath",
				ActionID:  core.ActionID{SpellID: 39442},
				Duration:  time.Second * 10,
				MaxStacks: 20,
			},
			BonusPerStack: stats.Stats{stats.MeleeCritRating: 17, stats.SpellCritRating: 17},
		})

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Darkmoon Card: Wrath",
			ActionID: core.ActionID{ItemID: 31857},
			ProcMask: core.ProcMaskDirect,
			Outcome:  core.OutcomeLanded,
			Callback: core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if result.Outcome.Matches(core.OutcomeCrit) {
					aura.Deactivate(sim)
				} else {
					aura.Activate(sim)
					aura.AddStack(sim)
				}
			},
		})

		character.ItemSwap.RegisterProc(31857, procAura)
	})

	// Darkmoon Card: Vengeance
	core.NewItemEffect(31858, func(agent core.Agent) {
		character := agent.GetCharacter()

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 39445},
			SpellSchool: core.SpellSchoolHoly,

			ProcMask: core.ProcMaskEmpty,
			Flags:    core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete | core.SpellFlagIgnoreResists,

			DamageMultiplier: 1,
			CritMultiplier:   character.DefaultMeleeCritMultiplier(),
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := sim.Roll(95, 115)
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialCritOnly)
			},
		})

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Darkmoon Card: Wrath",
			ActionID:           core.ActionID{ItemID: 31858},
			ProcMask:           core.ProcMaskDirect,
			ProcChance:         0.1,
			RequireDamageDealt: true,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitTaken,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				spell.Cast(sim, result.Target)
			},
		})

		character.ItemSwap.RegisterProc(31858, procAura)
	})

}
