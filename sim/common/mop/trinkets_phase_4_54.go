package mop

import (
	"fmt"
	"time"

	"github.com/wowsims/mop/sim/common/shared"
	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
)

type readinessTrinketConfig struct {
	itemVersionMap   shared.ItemVersionMap
	baseTrinketLabel string
	buffAuraLabel    string
	buffAuraID       int32
	buffedStat       stats.Stat
	buffDuration     time.Duration
	icd              time.Duration
	cdrAuraIDs       map[proto.Spec]int32
}

type multistrikeTrinketConfig struct {
	itemVersionMap   shared.ItemVersionMap
	baseTrinketLabel string
	buffAuraLabel    string
	buffAuraID       int32
	buffedStat       stats.Stat
	buffDuration     time.Duration
	icd              time.Duration
	rppm             float64
}

func init() {
	newReadinessTrinket := func(config *readinessTrinketConfig) {
		config.itemVersionMap.RegisterAll(func(version shared.ItemVersion, itemID int32, versionLabel string) {
			core.NewItemEffect(itemID, func(agent core.Agent, state proto.ItemLevelState) {
				character := agent.GetCharacter()

				auraID, exists := config.cdrAuraIDs[character.Spec]
				var cdrAura *core.Aura
				if exists {
					cdr := 1.0 / (1.0 + core.GetItemEffectScaling(itemID, 0.00989999995, state)/100)
					cdrAura = core.MakePermanent(character.RegisterAura(core.Aura{
						Label:    fmt.Sprintf("Readiness (%s)", versionLabel),
						ActionID: core.ActionID{SpellID: auraID},
					}).AttachSpellMod(core.SpellModConfig{
						Kind:       core.SpellMod_Cooldown_Multiplier,
						SpellFlag:  core.SpellFlagReadinessTrinket,
						FloatValue: cdr,
					}))
				}

				stats := stats.Stats{}
				stats[config.buffedStat] = core.GetItemEffectScaling(itemID, 2.97300004959, state)

				aura := character.NewTemporaryStatsAura(
					fmt.Sprintf("%s (%s)", config.buffAuraLabel, versionLabel),
					core.ActionID{SpellID: config.buffAuraID},
					stats,
					config.buffDuration,
				)

				triggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
					Name:       fmt.Sprintf("%s (%s) - Trigger", config.baseTrinketLabel, versionLabel),
					ProcChance: 0.15,
					ICD:        config.icd,
					ProcMask:   core.ProcMaskDirect | core.ProcMaskProc,
					Outcome:    core.OutcomeLanded,
					Callback:   core.CallbackOnSpellHitDealt,
					Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
						aura.Activate(sim)
					},
				})

				aura.Icd = triggerAura.Icd
				eligibleSlots := character.ItemSwap.EligibleSlotsForItem(itemID)
				character.AddStatProcBuff(itemID, aura, false, eligibleSlots)
				character.ItemSwap.RegisterProcWithSlots(itemID, triggerAura, eligibleSlots)
				if cdrAura != nil {
					character.ItemSwap.RegisterProcWithSlots(itemID, cdrAura, eligibleSlots)
				}
			})
		})
	}

	// Assurance of Consequence
	// Increases the cooldown recovery rate of six of your major abilities by 47%.
	// Effective for Agility-based damage roles only.
	//
	// Your attacks have a chance to grant you 14039 Agility for 20 sec.
	// (15% chance, 115 sec cooldown) (Proc chance: 15%, 1.917m cooldown)
	newReadinessTrinket(&readinessTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             104974,
			shared.ItemVersionNormal:          102292,
			shared.ItemVersionHeroic:          104476,
			shared.ItemVersionWarforged:       105223,
			shared.ItemVersionHeroicWarforged: 105472,
			shared.ItemVersionFlexible:        104725,
		},
		baseTrinketLabel: "Assurance of Consequence",
		buffAuraLabel:    "Dextrous",
		buffAuraID:       146308,
		buffedStat:       stats.Agility,
		buffDuration:     time.Second * 20,
		icd:              time.Second * 115,
		cdrAuraIDs: map[proto.Spec]int32{
			// Druid
			// Missing: Bear Hug, Ironbark, Nature's Swiftness
			proto.Spec_SpecFeralDruid:       145961,
			proto.Spec_SpecGuardianDruid:    145962,
			proto.Spec_SpecRestorationDruid: 145963,
			// Hunter
			proto.Spec_SpecBeastMasteryHunter: 145964,
			proto.Spec_SpecMarksmanshipHunter: 145965,
			proto.Spec_SpecSurvivalHunter:     145966,
			// Rogue
			// Missing: Cloak of Shadows, Evasion, JuJu Escape
			proto.Spec_SpecAssassinationRogue: 145983,
			proto.Spec_SpecCombatRogue:        145984,
			proto.Spec_SpecSubtletyRogue:      145985,
			// Priest - NOTE: Priests seem to have a Aura for this
			// Missing: Divine Hymn, Guardian Spirit, Hymn of Hope, Inner Focus, Pain Suppression, Power Word: Barrier, Void Shift
			// proto.Spec_SpecDisciplinePriest: 145981,
			// proto.Spec_SpecHolyPriest:       145982,
			// Shaman
			// Missing: Mana Tide Totem, Spirit Link Totem
			proto.Spec_SpecEnhancementShaman: 145986,
			proto.Spec_SpecRestorationShaman: 145988,
			// Monk
			// Missing: Zen Meditation, Life Cocoon, Revival, Thunder Focus Tea, Flying Serpent Kick
			proto.Spec_SpecBrewmasterMonk: 145967,
			proto.Spec_SpecMistweaverMonk: 145968,
			proto.Spec_SpecWindwalkerMonk: 145969,
		},
	})

	// Evil Eye of Galakras
	// Increases the cooldown recovery rate of six of your major abilities by 1%. Effective for Strength-based
	// damage roles only.
	//
	// Your attacks have a chance to grant you 11761 Strength for 10 sec.
	// (15% chance, 55 sec cooldown) (Proc chance: 15%, 55s cooldown)
	newReadinessTrinket(&readinessTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             104993,
			shared.ItemVersionNormal:          102298,
			shared.ItemVersionHeroic:          104495,
			shared.ItemVersionWarforged:       105242,
			shared.ItemVersionHeroicWarforged: 105491,
			shared.ItemVersionFlexible:        104744,
		},
		baseTrinketLabel: "Evil Eye of Galakras",
		buffAuraLabel:    "Outrage",
		buffAuraID:       146245,
		buffedStat:       stats.Strength,
		buffDuration:     time.Second * 10,
		icd:              time.Second * 55,
		cdrAuraIDs: map[proto.Spec]int32{
			// Death Knight
			proto.Spec_SpecBloodDeathKnight:  145958,
			proto.Spec_SpecFrostDeathKnight:  145959,
			proto.Spec_SpecUnholyDeathKnight: 145960,
			// Paladin
			// Missing: Divine Plea, Hand Of Protection, Divine Shield, Hand Of Purity
			proto.Spec_SpecHolyPaladin:        145978,
			proto.Spec_SpecProtectionPaladin:  145976,
			proto.Spec_SpecRetributionPaladin: 145975,
			// Warrior
			// Missing: Die by the Sword, Mocking Banner
			proto.Spec_SpecArmsWarrior:       145990,
			proto.Spec_SpecFuryWarrior:       145991,
			proto.Spec_SpecProtectionWarrior: 145992,
		},
	})

	getMultistrikeSpell := func(character *core.Character, spellID int32, spellSchool core.SpellSchool) *core.Spell {
		return character.GetOrRegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: spellID},
			SpellSchool: spellSchool,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagIgnoreArmor | core.SpellFlagIgnoreModifiers | core.SpellFlagPassiveSpell | core.SpellFlagNoSpellMods,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
		})
	}

	getMultistrikeSpells := func(character *core.Character) (*core.Spell, *core.Spell) {
		var physicalSpellID int32
		if character.Class == proto.Class_ClassHunter {
			physicalSpellID = 146069
		} else {
			physicalSpellID = 146061
		}

		physicalSpell := getMultistrikeSpell(character, physicalSpellID, core.SpellSchoolPhysical)
		magicSpell := physicalSpell

		switch character.Class {
		case proto.Class_ClassDruid:
			magicSpell = getMultistrikeSpell(character, 146064, core.SpellSchoolArcane)
		case proto.Class_ClassMage:
			var magicSpellID int32
			var school core.SpellSchool
			if character.Spec == proto.Spec_SpecArcaneMage {
				magicSpellID = 146070
				school = core.SpellSchoolArcane
			} else {
				magicSpellID = 146067
				school = core.SpellSchoolFrostfire
			}
			magicSpell = getMultistrikeSpell(character, magicSpellID, school)
		case proto.Class_ClassMonk:
			magicSpell = getMultistrikeSpell(character, 146075, core.SpellSchoolNature)
		case proto.Class_ClassPriest:
			var magicSpellID int32
			var school core.SpellSchool
			if character.Spec == proto.Spec_SpecShadowPriest {
				magicSpellID = 146065
				school = core.SpellSchoolShadow
			} else {
				magicSpellID = 146063
				school = core.SpellSchoolHoly
			}
			magicSpell = getMultistrikeSpell(character, magicSpellID, school)
		case proto.Class_ClassShaman:
			magicSpell = getMultistrikeSpell(character, 146071, core.SpellSchoolNature)
		case proto.Class_ClassWarlock:
			magicSpell = getMultistrikeSpell(character, 146065, core.SpellSchoolShadow)
		}

		return physicalSpell, magicSpell
	}

	blackoutKickTickID := core.ActionID{SpellID: 100784}.WithTag(2)
	newMultistrikeTrinket := func(config *multistrikeTrinketConfig) {
		config.itemVersionMap.RegisterAll(func(version shared.ItemVersion, itemID int32, versionLabel string) {
			core.NewItemEffect(itemID, func(agent core.Agent, state proto.ItemLevelState) {
				character := agent.GetCharacter()

				var baseDamage float64
				applyEffects := func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeAlwaysHit)
				}

				physicalSpell, magicSpell := getMultistrikeSpells(character)

				multistrikeTriggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
					Name:               fmt.Sprintf("%s (%s) - Multistrike Trigger", config.baseTrinketLabel, versionLabel),
					ProcChance:         core.GetItemEffectScaling(itemID, 0.03539999947, state) / 1000,
					Outcome:            core.OutcomeLanded,
					Callback:           core.CallbackOnSpellHitDealt | core.CallbackOnPeriodicDamageDealt,
					RequireDamageDealt: true,

					ExtraCondition: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
						return spell.ProcMask != core.ProcMaskEmpty
					},

					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						baseDamage = result.Damage / 3.0

						// Special case for Windwalker Blackout Kick DoTs which does physical damage but procs the nature damage spell
						if spell.SpellSchool.Matches(core.SpellSchoolPhysical) && !spell.ActionID.SameAction(blackoutKickTickID) {
							physicalSpell.ApplyEffects = applyEffects
							physicalSpell.Cast(sim, result.Target)
						} else {
							magicSpell.ApplyEffects = applyEffects
							magicSpell.Cast(sim, result.Target)
						}
					},
				})

				stats := stats.Stats{}
				stats[config.buffedStat] = core.GetItemEffectScaling(itemID, 2.97300004959, state)

				statBuffAura := character.NewTemporaryStatsAura(
					fmt.Sprintf("%s (%s)", config.buffAuraLabel, versionLabel),
					core.ActionID{SpellID: config.buffAuraID},
					stats,
					config.buffDuration,
				)

				statBuffTriggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
					Name:     fmt.Sprintf("%s (%s) - Stat Trigger", config.baseTrinketLabel, versionLabel),
					ICD:      config.icd,
					Outcome:  core.OutcomeLanded,
					Callback: core.CallbackOnSpellHitDealt,

					DPM: character.NewRPPMProcManager(itemID, false, false, core.ProcMaskDirect|core.ProcMaskProc, core.RPPMConfig{
						PPM: config.rppm,
					}),

					Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
						statBuffAura.Activate(sim)
					},
				})

				statBuffAura.Icd = statBuffTriggerAura.Icd

				eligibleSlots := character.ItemSwap.EligibleSlotsForItem(itemID)
				character.AddStatProcBuff(itemID, statBuffAura, false, eligibleSlots)
				character.ItemSwap.RegisterProcWithSlots(itemID, statBuffTriggerAura, eligibleSlots)
				character.ItemSwap.RegisterProcWithSlots(itemID, multistrikeTriggerAura, eligibleSlots)
			})
		})
	}

	// Haromm's Talisman
	// Your attacks have a 16.7% chance to trigger Multistrike, which deals instant additional damage to your target equal to 33% of the original damage dealt.
	//
	// Your attacks have a chance to grant you 14039 Agility for 10 sec.
	// (Approximately 0.92 procs per minute)
	newMultistrikeTrinket(&multistrikeTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             105029,
			shared.ItemVersionNormal:          102301,
			shared.ItemVersionHeroic:          104531,
			shared.ItemVersionWarforged:       105278,
			shared.ItemVersionHeroicWarforged: 105527,
			shared.ItemVersionFlexible:        104780,
		},
		baseTrinketLabel: "Haromm's Talisman",
		buffAuraLabel:    "Vicious",
		buffAuraID:       148903,
		buffedStat:       stats.Agility,
		buffDuration:     time.Second * 10,
		icd:              time.Second * 10,
		rppm:             0.92000001669,
	})

	// Kardris' Toxic Totem
	// Your attacks have a 16.7% chance to trigger Multistrike, which deals instant additional damage to your target equal to 33% of the original damage dealt.
	//
	// Your attacks have a chance to grant 14039 Intellect for 10 sec.
	// (Approximately 0.92 procs per minute)
	newMultistrikeTrinket(&multistrikeTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             105042,
			shared.ItemVersionNormal:          102300,
			shared.ItemVersionHeroic:          104544,
			shared.ItemVersionWarforged:       105291,
			shared.ItemVersionHeroicWarforged: 105540,
			shared.ItemVersionFlexible:        104793,
		},
		baseTrinketLabel: "Kardris' Toxic Totem",
		buffAuraLabel:    "Toxic Power",
		buffAuraID:       148906,
		buffedStat:       stats.Intellect,
		buffDuration:     time.Second * 10,
		icd:              time.Second * 10,
		rppm:             0.92000001669,
	})

	// Purified Bindings of Immerseus
	// Your attacks have a chance to grant 606 Intellect for 20 sec.
	// (Proc chance: 15%, 1.917m cooldown)
	// Amplifies your Critical Strike damage and healing, Haste, Mastery, and Spirit by 1%.
	shared.ItemVersionMap{
		shared.ItemVersionLFR:             104924,
		shared.ItemVersionNormal:          102293,
		shared.ItemVersionHeroic:          104426,
		shared.ItemVersionWarforged:       105173,
		shared.ItemVersionHeroicWarforged: 105422,
		shared.ItemVersionFlexible:        104675,
	}.RegisterAll(func(version shared.ItemVersion, itemID int32, versionLabel string) {
		label := "Purified Bindings of Immerseus"

		core.NewItemEffect(itemID, func(agent core.Agent, state proto.ItemLevelState) {
			character := agent.GetCharacter()
			statValue := core.GetItemEffectScaling(itemID, 2.97300004959, state)

			critDamageValue := 1 + core.GetItemEffectScaling(itemID, 0.00088499999, state)/100
			hasteValue := 1 + core.GetItemEffectScaling(itemID, 0.00176999997, state)/100
			masteryValue := 1 + core.GetItemEffectScaling(itemID, 0.00176999997, state)/100
			spiritValue := 1 + core.GetItemEffectScaling(itemID, 0.00176999997, state)/100

			statAura := core.MakePermanent(character.RegisterAura(core.Aura{
				Label:      fmt.Sprintf("Amplification (%s)", versionLabel),
				ActionID:   core.ActionID{SpellID: 146051},
				BuildPhase: core.CharacterBuildPhaseGear,
			})).
				AttachStatDependency(character.NewDynamicMultiplyStat(stats.HasteRating, hasteValue)).
				AttachStatDependency(character.NewDynamicMultiplyStat(stats.MasteryRating, masteryValue)).
				AttachStatDependency(character.NewDynamicMultiplyStat(stats.Spirit, spiritValue)).
				AttachMultiplicativePseudoStatBuff(&character.PseudoStats.CritDamageMultiplier, critDamageValue)

			aura := character.NewTemporaryStatsAura(
				fmt.Sprintf("Expanded Mind (%s)", versionLabel),
				core.ActionID{SpellID: 146046},
				stats.Stats{stats.Intellect: statValue},
				time.Second*20,
			)

			triggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
				Name:       fmt.Sprintf("%s (%s)", label, versionLabel),
				ICD:        time.Second * 115,
				ProcChance: 0.15,
				Outcome:    core.OutcomeLanded,
				Callback:   core.CallbackOnSpellHitDealt,
				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					aura.Activate(sim)
				},
			})

			eligibleSlots := character.ItemSwap.EligibleSlotsForItem(itemID)
			character.AddStatProcBuff(itemID, aura, false, eligibleSlots)
			character.ItemSwap.RegisterProcWithSlots(itemID, triggerAura, eligibleSlots)
			character.ItemSwap.RegisterProcWithSlots(itemID, statAura, eligibleSlots)
		})
	})
}
