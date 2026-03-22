package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// Despair
	core.NewItemEffect(28573, func(agent core.Agent) {
		character := agent.GetCharacter()

		spell := character.GetOrRegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 34580},
			ProcMask:    core.ProcMaskEmpty,
			SpellSchool: core.SpellSchoolPhysical,
			Flags:       core.SpellFlagPassiveSpell | core.SpellFlagIgnoreResists,

			DamageMultiplier: 1,
			CritMultiplier:   character.DefaultMeleeCritMultiplier(),
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, 600, spell.OutcomeMeleeSpecialNoBlockDodgeParry)
			},
		})

		getDpm := func() *core.DynamicProcManager {
			return character.NewStaticLegacyPPMManager(
				1,
				*character.GetDynamicProcMaskForWeaponEffect(28573),
			)
		}

		dpm := getDpm()

		procTrigger := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Despair",
			DPM:                dpm,
			TriggerImmediately: true,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				spell.Cast(sim, result.Target)
			},
		})

		character.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *core.Simulation, slot proto.ItemSlot) {
			dpm = getDpm()
		})

		character.ItemSwap.RegisterProc(28573, procTrigger)
	})

	// Rod of the Sun King
	core.NewItemEffect(29996, func(agent core.Agent) {
		character := agent.GetCharacter()
		actionID := core.ActionID{SpellID: 36070}
		var resourceMetrics *core.ResourceMetrics = nil
		if character.HasEnergyBar() {
			resourceMetrics = character.NewEnergyMetrics(actionID)
		} else if character.HasRageBar() {
			resourceMetrics = character.NewRageMetrics(actionID)
		} else {
			return
		}

		spell := character.GetOrRegisterSpell(core.SpellConfig{
			ActionID: actionID,
			ProcMask: core.ProcMaskEmpty,
			Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagNoMetrics,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				if character.HasEnergyBar() {
					character.AddEnergy(sim, 10, resourceMetrics)
				} else if character.HasRageBar() {
					character.AddRage(sim, 5, resourceMetrics)
				}
			},
		})

		resourceGainDpm := func() *core.DynamicProcManager {
			return character.NewStaticLegacyPPMManager(
				1,
				*character.GetDynamicProcMaskForWeaponEffect(29996),
			)
		}

		dpm := resourceGainDpm()

		procTrigger := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Power of the Sun King",
			DPM:                dpm,
			TriggerImmediately: true,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				spell.Cast(sim, result.Target)
			},
		})

		character.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
			dpm = resourceGainDpm()
		})

		character.ItemSwap.RegisterProc(29996, procTrigger)
	})

	// World Breaker
	core.NewItemEffect(30090, func(agent core.Agent) {
		character := agent.GetCharacter()
		var aura *core.Aura
		aura = character.RegisterAura(core.Aura{
			Label:     "World Breaker",
			ActionID:  core.ActionID{SpellID: 36111},
			Duration:  time.Second * 4,
			MaxStacks: 2,
		}).
			AttachStatBuff(stats.MeleeCritRating, 900).
			AttachProcTrigger(core.ProcTrigger{
				Name:               "World Breaker - Consume",
				ProcMask:           core.ProcMaskMelee,
				Callback:           core.CallbackOnSpellHitDealt,
				TriggerImmediately: true,
				Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
					if aura.IsActive() {
						aura.RemoveStack(sim)
					}
				},
			})

		getDpm := func() *core.DynamicProcManager {
			return character.NewStaticLegacyPPMManager(
				1,
				*character.GetDynamicProcMaskForWeaponEffect(30090),
			)
		}

		dpm := getDpm()

		procTrigger := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "World Breaker - Trigger",
			DPM:      dpm,
			Outcome:  core.OutcomeLanded,
			Callback: core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
				aura.AddStack(sim)
			},
		})

		character.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *core.Simulation, slot proto.ItemSlot) {
			dpm = getDpm()
		})

		character.ItemSwap.RegisterProc(30090, procTrigger)
	})

	// Blinkstrike
	core.NewItemEffect(31332, func(agent core.Agent) {
		character := agent.GetCharacter()
		var blinkStrikeSpell *core.Spell

		extraAttackDPM := func() *core.DynamicProcManager {
			return character.NewStaticLegacyPPMManager(
				1,
				*character.GetDynamicProcMaskForWeaponEffect(31332),
			)
		}

		dpm := extraAttackDPM()

		procTrigger := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Blinkstrike",
			DPM:                dpm,
			TriggerImmediately: true,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				character.AutoAttacks.MaybeReplaceMHSwing(sim, blinkStrikeSpell).Cast(sim, result.Target)
			},
		})

		procTrigger.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
			config := *character.AutoAttacks.MHConfig()
			config.ActionID = config.ActionID.WithTag(31332)
			config.Flags |= core.SpellFlagPassiveSpell
			blinkStrikeSpell = character.GetOrRegisterSpell(config)
		})

		character.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
			dpm = extraAttackDPM()
		})

		character.ItemSwap.RegisterProc(31332, procTrigger)
	})

	// Warglaives of Azzinoth
	core.NewItemSet(core.ItemSet{
		Name: "The Twin Blades of Azzinoth",
		Bonuses: map[int32]core.ApplySetBonus{
			2: func(agent core.Agent, setBonusAura *core.Aura) {
				character := agent.GetCharacter()

				if character.Class != proto.Class_ClassRogue && character.Class != proto.Class_ClassWarrior {
					return
				}

				aura := character.NewTemporaryStatsAura(
					"The Twin Blades of Azzinoth",
					core.ActionID{SpellID: 41435},
					stats.Stats{stats.MeleeHasteRating: 450},
					time.Second*10,
				)

				hasteDPM := func() *core.DynamicProcManager {
					return character.NewStaticLegacyPPMManager(
						1,
						character.GetProcMaskForTypes(proto.WeaponType_WeaponTypeSword),
					)
				}

				dpm := hasteDPM()

				setBonusAura.
					AttachProcTrigger(core.ProcTrigger{
						Name:     "The Twin Blades of Azzinoth - Trigger",
						DPM:      dpm,
						ICD:      time.Second * 45,
						Outcome:  core.OutcomeLanded,
						Callback: core.CallbackOnSpellHitDealt,
						Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
							aura.Activate(sim)
						},
					}).
					ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
						for _, at := range character.AttackTables {
							at.MobTypeBonusStats[proto.MobType_MobTypeDemon] = at.MobTypeBonusStats[proto.MobType_MobTypeDemon].Add(stats.Stats{
								stats.AttackPower:       200,
								stats.RangedAttackPower: 200,
							})
						}
					}).
					ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
						for _, at := range character.AttackTables {
							at.MobTypeBonusStats[proto.MobType_MobTypeDemon] = at.MobTypeBonusStats[proto.MobType_MobTypeDemon].Subtract(stats.Stats{
								stats.AttackPower:       200,
								stats.RangedAttackPower: 200,
							})
						}
					}).
					ExposeToAPL(41434)

				character.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
					dpm = hasteDPM()
				})
			},
		},
	})

}
