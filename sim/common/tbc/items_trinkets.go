package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// Figurine - Dawnstone Crab
	core.NewItemEffect(24125, func(agent core.Agent) {
		character := agent.GetCharacter()
		actionId := core.ActionID{SpellID: 31039, ItemID: 24125}

		aura := character.NewTemporaryStatsAura(
			"Dawnstone Crab",
			core.ActionID{SpellID: 31039},
			stats.Stats{stats.DodgeRating: 125},
			time.Second*20,
		)

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    actionId,
			SpellSchool: core.SpellSchoolNature,
			ProcMask:    core.ProcMaskEmpty,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				aura.Activate(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell:    spell,
			Type:     core.CooldownTypeSurvival,
			BuffAura: aura,
			ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
				return character.CurrentHealthPercent() < 0.4
			},
		})
	})

	// Figurine - Nightseye Panther
	// Use: Increases attack power by 320 for 12 sec. (3 Min Cooldown)
	core.NewItemEffect(24128, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 12
		aura := character.NewTemporaryStatsAura(
			"Nightseye Panther",
			core.ActionID{SpellID: 31047},
			stats.Stats{stats.AttackPower: 320, stats.RangedAttackPower: 320},
			duration,
		)

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: 24128},
			SpellSchool: core.SpellSchoolPhysical,
			ProcMask:    core.ProcMaskEmpty,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 3,
				},
				SharedCD: core.Cooldown{
					Timer:    character.GetOffensiveTrinketCD(),
					Duration: duration,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				aura.Activate(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell:    spell,
			Type:     core.CooldownTypeDPS,
			BuffAura: aura,
		})
	})

	// Jom Gabbar
	// Use: Increases attack power by 65 and an additional 65 every 2 sec. Lasts 20 sec. (2 Min Cooldown)
	core.NewItemEffect(23570, func(agent core.Agent) {
		character := agent.GetCharacter()
		actionID := core.ActionID{SpellID: 29602}
		duration := time.Second * 20
		bonusPerStack := stats.Stats{
			stats.AttackPower:       65,
			stats.RangedAttackPower: 65,
		}

		jomGabbarAura := character.GetOrRegisterAura(core.Aura{
			Label:     "Jom Gabbar",
			ActionID:  actionID,
			Duration:  duration,
			MaxStacks: 10,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				core.StartPeriodicAction(sim, core.PeriodicActionOptions{
					Period:          time.Second * 2,
					NumTicks:        10,
					Priority:        core.ActionPriorityAuto,
					TickImmediately: true,
					OnAction: func(sim *core.Simulation) {
						aura.AddStack(sim)
					},
				})
			},
			OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
				character.AddStatsDynamic(sim, bonusPerStack.Multiply(float64(newStacks-oldStacks)))
			},
		})

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID: actionID,
			Flags:    core.SpellFlagNoOnCastComplete,
			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
				SharedCD: core.Cooldown{
					Timer:    character.GetOffensiveTrinketCD(),
					Duration: duration,
				},
			},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				jomGabbarAura.Activate(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeDPS,
		})
	})

	// Figurine of the Colossus
	core.NewItemEffect(27529, func(agent core.Agent) {
		character := agent.GetCharacter()
		actionId := core.ActionID{ItemID: 27529}

		healSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 33090},
			ProcMask:    core.ProcMaskEmpty,
			SpellSchool: core.SpellSchoolHoly,
			Flags:       core.SpellFlagIgnoreTargetModifiers | core.SpellFlagIgnoreAttackerModifiers,

			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					NonEmpty: true,
				},
			},

			DamageMultiplier: 1,
			CritMultiplier:   character.DefaultHealingCritMultiplier(),
			ThreatMultiplier: 0.5,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealHealing(sim, target, 120, spell.OutcomeHealing)
			},
		})

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:            "Vigilance of the Colossus",
			MetricsActionID: actionId,
			Duration:        time.Second * 20,
			Outcome:         core.OutcomeBlock,
			Callback:        core.CallbackOnSpellHitTaken,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				healSpell.Cast(sim, &character.Unit)
			},
		})

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    actionId,
			SpellSchool: core.SpellSchoolHoly,
			ProcMask:    core.ProcMaskEmpty,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				procAura.Activate(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeSurvival,
			BuffAura: &core.StatBuffAura{
				Aura:            procAura,
				BuffedStatTypes: []stats.Stat{stats.Health},
			},
			ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
				return character.CurrentHealthPercent() < 0.4
			},
		})
	})

	// Argussian Compass
	core.NewItemEffect(27770, func(agent core.Agent) {
		character := agent.GetCharacter()

		damageAbsorptionAura := character.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
			Aura: core.Aura{
				Label:    "Argussian Compass",
				ActionID: core.ActionID{SpellID: 39228},
				Duration: time.Second * 20,
			},
			MaxAbsorbPerHit: 68,
			ShieldStrengthCalculator: func(_ *core.Unit) float64 {
				return 1150
			},
		})

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: 27770},
			SpellSchool: core.SpellSchoolHoly,
			ProcMask:    core.ProcMaskEmpty,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				damageAbsorptionAura.Activate(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeSurvival,
			BuffAura: &core.StatBuffAura{
				Aura:            damageAbsorptionAura.Aura,
				BuffedStatTypes: []stats.Stat{stats.Health},
			},
			ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
				return character.CurrentHealthPercent() < 0.4
			},
		})
	})

	// Hourglass of the Unraveller
	core.NewItemEffect(28034, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 10
		value := 300.0

		aura := character.NewTemporaryStatsAura(
			"Rage of the Unraveller",
			core.ActionID{SpellID: 33649},
			stats.Stats{stats.AttackPower: value, stats.RangedAttackPower: value},
			duration,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:              "Hourglass of the Unraveller",
			ActionID:          core.ActionID{ItemID: 28034},
			SpellFlagsExclude: core.SpellFlagSuppressEquipProcs,
			ProcChance:        0.1,
			ICD:               time.Second * 50,
			ProcMask:          core.ProcMaskMeleeOrRanged,
			Outcome:           core.OutcomeCrit,
			Callback:          core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})

		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(28034)
		character.AddStatProcBuff(28034, aura, false, eligibleSlots)
		character.ItemSwap.RegisterProc(28034, procAura)
	})

	// Romulo's Poison Vial
	core.NewItemEffect(28579, func(agent core.Agent) {
		character := agent.GetCharacter()

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 34587},
			SpellSchool: core.SpellSchoolNature,

			ProcMask: core.ProcMaskSpellProc | core.ProcMaskSpellDamageProc,
			Flags:    core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := sim.Roll(222, 332)
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHit)
			},
		})

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Romulo's Poison Vial",
			ActionID:           core.ActionID{ItemID: 28579},
			SpellFlagsExclude:  core.SpellFlagSuppressEquipProcs,
			DPM:                character.NewLegacyPPMManager(1, core.ProcMaskMeleeOrRanged),
			RequireDamageDealt: true,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				spell.Cast(sim, result.Target)
			},
		})

		character.ItemSwap.RegisterProc(28579, procAura)
	})

	// The Lightning Capacitor
	core.NewItemEffect(28785, func(agent core.Agent) {
		character := agent.GetCharacter()

		lightningBolt := character.RegisterSpell(core.SpellConfig{
			ActionID:     core.ActionID{SpellID: 37661},
			SpellSchool:  core.SpellSchoolNature,
			ProcMask:     core.ProcMaskEmpty,
			Flags:        core.SpellFlagPassiveSpell | core.SpellFlagIgnoreAttackerModifiers,
			MissileSpeed: 20,

			DamageMultiplier: 1,
			CritMultiplier:   character.DefaultSpellCritMultiplier(),

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.WaitTravelTime(sim, func(s *core.Simulation) {
					baseDamage := sim.Roll(694, 806)
					// https://www.wowhead.com/tbc/item=28785/the-lightning-capacitor#comments
					// It can crit, may need some testing
					spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
				})

			},
		})

		icd := core.Cooldown{
			Timer:    character.NewTimer(),
			Duration: time.Millisecond * 2500,
		}

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
					icd.Use(sim)
				}
			},
		})

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "The Lightning Capacitor",
			ActionID: core.ActionID{ItemID: 28785},
			ProcMask: core.ProcMaskSpellOrSpellProc,
			Outcome:  core.OutcomeCrit,
			Callback: core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !icd.IsReady(sim) {
					return
				}

				lightningCapacitorAura.Activate(sim)
				lightningCapacitorAura.AddStack(sim)
			},
		})

		character.ItemSwap.RegisterProc(28785, procAura)
	})

	// Eye of Magtheridon
	core.NewItemEffect(28789, func(agent core.Agent) {
		character := agent.GetCharacter()

		value := 170.0
		aura := character.NewTemporaryStatsAura(
			"Recurring Power",
			core.ActionID{SpellID: 34747},
			stats.Stats{stats.SpellDamage: value, stats.HealingPower: value},
			time.Second*10,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:              "Eye of Magtheridon",
			ActionID:          core.ActionID{ItemID: 28789},
			SpellFlagsExclude: core.SpellFlagSuppressEquipProcs,
			ClassSpellsOnly:   true,
			Outcome:           core.OutcomeMiss,
			Callback:          core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})

		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(28789)
		character.AddStatProcBuff(34747, aura, false, eligibleSlots)
		character.ItemSwap.RegisterProc(28789, procAura)
	})

	// Spyglass of the Hidden Fleet
	core.NewItemEffect(30620, func(agent core.Agent) {
		character := agent.GetCharacter()

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: 30620},
			SpellSchool: core.SpellSchoolNature,

			ProcMask: core.ProcMaskSpellProc | core.ProcMaskSpellDamageProc,
			Flags:    core.SpellFlagNoOnCastComplete,

			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					NonEmpty: true,
				},
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
				IgnoreHaste: true,
			},

			DamageMultiplier: 1,
			CritMultiplier:   character.DefaultMeleeCritMultiplier(),
			ThreatMultiplier: 1,

			Hot: core.DotConfig{
				Aura: core.Aura{
					Label: "Regeneration",
				},

				NumberOfTicks: 4,
				TickLength:    time.Second * 3,

				OnSnapshot: func(_ *core.Simulation, target *core.Unit, dot *core.Dot) {
					dot.SnapshotHeal(target, 325)
				},

				OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
					dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.Hot(&character.Unit).Apply(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeSurvival,
		})
	})

	// Prism of Inner Calm
	core.NewItemEffect(30621, func(agent core.Agent) {
		character := agent.GetCharacter()

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Prism of Inner Calm",
			ActionID:           core.ActionID{ItemID: 30621},
			Outcome:            core.OutcomeCrit,
			Callback:           core.CallbackOnSpellHitDealt,
			RequireDamageDealt: true,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				threatReduction := 150.0
				if spell.ProcMask.Matches(core.ProcMaskSpellDamageProc) {
					threatReduction = 1000
				}
				spell.FlatThreatBonus -= threatReduction
				result.Threat = spell.ThreatFromDamage(sim, result.Outcome, result.Damage, spell.Unit.AttackTables[result.Target.UnitIndex])
				spell.FlatThreatBonus += threatReduction
			},
		})

		character.ItemSwap.RegisterProc(30621, procAura)
	})

	// Sextant of Unstable Currents
	core.NewItemEffect(30626, func(agent core.Agent) {
		character := agent.GetCharacter()

		value := 190.0
		aura := character.NewTemporaryStatsAura(
			"Unstable Currents",
			core.ActionID{SpellID: 38348},
			stats.Stats{stats.SpellDamage: value, stats.HealingPower: value},
			time.Second*15,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Sextant of Unstable Currents",
			ActionID:           core.ActionID{ItemID: 30626},
			ProcChance:         0.2,
			ICD:                time.Second * 45,
			Outcome:            core.OutcomeCrit,
			Callback:           core.CallbackOnSpellHitDealt,
			RequireDamageDealt: true,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})

		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(30626)
		character.AddStatProcBuff(38348, aura, false, eligibleSlots)
		character.ItemSwap.RegisterProc(30626, procAura)
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

		meleeProcAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:              "Darkmoon Card: Crusade (Melee)",
			ActionID:          core.ActionID{SpellID: 39438},
			SpellFlagsExclude: core.SpellFlagSuppressEquipProcs,
			ProcMask:          core.ProcMaskMeleeOrRanged,
			Outcome:           core.OutcomeLanded,
			Callback:          core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				meleeAura.Activate(sim)
				meleeAura.AddStack(sim)
			},
		})

		casterProcAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:            "Darkmoon Card: Crusade (Caster)",
			ActionID:        core.ActionID{SpellID: 39440},
			ProcMask:        core.ProcMaskSpellOrSpellProc,
			ClassSpellsOnly: true,
			Outcome:         core.OutcomeLanded,
			Callback:        core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				casterAura.Activate(sim)
				casterAura.AddStack(sim)
			},
		})

		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(31856)
		character.AddStatProcBuff(39438, meleeAura, false, eligibleSlots)
		character.AddStatProcBuff(39441, casterAura, false, eligibleSlots)
		character.ItemSwap.RegisterProc(31856, meleeProcAura)
		character.ItemSwap.RegisterProc(31856, casterProcAura)
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
			Name:              "Darkmoon Card: Wrath",
			ActionID:          core.ActionID{ItemID: 31857},
			SpellFlagsExclude: core.SpellFlagSuppressEquipProcs,
			ProcMask:          core.ProcMaskDirect,
			Outcome:           core.OutcomeLanded,
			Callback:          core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if result.Outcome.Matches(core.OutcomeCrit) {
					aura.Deactivate(sim)
				} else {
					aura.Activate(sim)
					aura.AddStack(sim)
				}
			},
		})

		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(31857)
		character.AddStatProcBuff(39442, aura, false, eligibleSlots)
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
			Name:               "Darkmoon Card: Vengeance",
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

	// Blackened Naaru Sliver
	core.NewItemEffect(34427, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := core.MakeStackingAura(character, core.StackingStatAura{
			Aura: core.Aura{
				Label:     "Combat Insight",
				ActionID:  core.ActionID{SpellID: 45041},
				Duration:  time.Second * 20,
				MaxStacks: 10,
			},
			BonusPerStack: stats.Stats{stats.AttackPower: 44, stats.RangedAttackPower: 44},
		})

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:            "Battle Trance",
			MetricsActionID: core.ActionID{SpellID: 45040},
			Duration:        time.Second * 20,
			ProcMask:        core.ProcMaskMeleeOrRanged,
			Outcome:         core.OutcomeLanded,
			Callback:        core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if aura.IsActive() {
					aura.AddStack(sim)
				}
			},
		})

		triggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:              "Blackened Naaru Sliver",
			ActionID:          core.ActionID{ItemID: 34427},
			SpellFlagsExclude: core.SpellFlagSuppressEquipProcs,
			ICD:               time.Second * 45,
			ProcChance:        0.1,
			ProcMask:          core.ProcMaskMeleeOrRanged,
			Outcome:           core.OutcomeLanded,
			Callback:          core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
				procAura.Activate(sim)
			},
		})

		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(34427)
		character.AddStatProcBuff(45041, aura, false, eligibleSlots)
		character.ItemSwap.RegisterProc(34427, triggerAura)
	})

	// Commendation of Kael'thas
	core.NewItemEffect(34473, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := character.NewTemporaryStatsAura(
			"Evasive Maneuvers",
			core.ActionID{SpellID: 45058},
			stats.Stats{stats.DodgeRating: 152},
			time.Second*10,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Evasive Maneuvers",
			ActionID:           core.ActionID{ItemID: 34473},
			ProcMask:           core.ProcMaskMelee,
			ICD:                time.Second * 30,
			RequireDamageDealt: true,
			TriggerImmediately: true,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitTaken,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				if character.CurrentHealthPercent() < 0.35 {
					aura.Activate(sim)
				}
			},
		})

		character.ItemSwap.RegisterProc(34473, procAura)
	})

	// Figurine - Empyrean Tortoise
	core.NewItemEffect(35693, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := character.NewTemporaryStatsAura(
			"Empyrean Tortoise",
			core.ActionID{SpellID: 46780},
			stats.Stats{stats.DodgeRating: 165},
			time.Second*20,
		)

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: 35693},
			SpellSchool: core.SpellSchoolNature,
			ProcMask:    core.ProcMaskEmpty,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				aura.Activate(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell:    spell,
			Type:     core.CooldownTypeSurvival,
			BuffAura: aura,
			ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
				return character.CurrentHealthPercent() < 0.4
			},
		})
	})

	// Pendant of the Violet Eye
	core.NewItemEffect(28727, func(agent core.Agent) {
		character := agent.GetCharacter()

		stackingAura := core.MakeStackingAura(character, core.StackingStatAura{
			Aura: core.Aura{
				Label:     "Enlightenment",
				ActionID:  core.ActionID{SpellID: 35095},
				Duration:  core.NeverExpires,
				MaxStacks: 20,
			},
			BonusPerStack: stats.Stats{stats.MP5: 21},
		})

		procAura := character.RegisterAura(core.Aura{
			Label:    "Enlightenment Trigger",
			ActionID: core.ActionID{SpellID: 29601},
			Duration: time.Second * 20,

			OnExpire: func(_ *core.Aura, sim *core.Simulation) {
				stackingAura.Deactivate(sim)
			},
		}).AttachProcTrigger(core.ProcTrigger{
			Callback: core.CallbackOnCastComplete,
			ProcMask: core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,

			ExtraCondition: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
				return spell.CurCast.Cost > 0
			},

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				stackingAura.Activate(sim)
				stackingAura.AddStack(sim)
			},
		})

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: 28727},
			SpellSchool: core.SpellSchoolPhysical,
			ProcMask:    core.ProcMaskEmpty,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				spell.RelatedSelfBuff.Activate(sim)
			},

			RelatedSelfBuff: procAura,
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeMana,
		})

		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(28727)
		character.AddStatProcBuff(35095, stackingAura, false, eligibleSlots)
		character.ItemSwap.RegisterProc(28727, procAura)
	})

	// Memento of Tyrande
	core.NewItemEffect(32496, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := character.NewTemporaryStatsAura(
			"Wisdom",
			core.ActionID{SpellID: 37656},
			stats.Stats{stats.MP5: 76},
			time.Second*15,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:            "Memento of Tyrande",
			Callback:        core.CallbackOnCastComplete,
			ProcMask:        core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
			MetricsActionID: core.ActionID{SpellID: 37655},
			ClassSpellsOnly: true,
			ProcChance:      0.1,
			ICD:             time.Second * 50,

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})

		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(32496)
		character.AddStatProcBuff(37656, aura, false, eligibleSlots)
		character.ItemSwap.RegisterProc(32496, procAura)
	})

	// Shifting Naaru Sliver
	// Use: Conjures a Power Circle lasting for 15 sec.  While standing in this circle, the caster gains up to 320 spell damage and healing.
	core.NewItemEffect(34429, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 15
		aura := character.NewTemporaryStatsAura("Power Circle", core.ActionID{SpellID: 45042}, stats.Stats{stats.SpellDamage: 320, stats.HealingPower: 320}, duration)

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: 34429},
			SpellSchool: core.SpellSchoolPhysical,
			ProcMask:    core.ProcMaskEmpty,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
				SharedCD: core.Cooldown{
					Timer:    character.GetOffensiveTrinketCD(),
					Duration: duration,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				aura.Activate(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell:    spell,
			Type:     core.CooldownTypeDPS,
			BuffAura: aura,
		})
	})
}
