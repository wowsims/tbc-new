package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

var ItemSetCryptstalkerArmor = core.NewItemSet(core.ItemSet{
	Name: "Cryptstalker Armor",
	ID:   530,
	Bonuses: map[int32]core.ApplySetBonus{
		// (2) Set: Increases the duration of your Rapid Fire by 4 secs.
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_BuffDuration_Flat,
				ClassMask: HunterSpellRapidFire,
				TimeValue: time.Second * 4,
			}).ExposeToAPL(28755)
		},
		// (4) Set: While your pet is active, increases Attack Power by 50 for both you and your pet.
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			hunter := agent.(HunterAgent).GetHunter()
			if hunter.Pet == nil {
				return
			}

			apBuff := stats.Stats{
				stats.AttackPower:       50,
				stats.RangedAttackPower: 50,
			}
			ownerAura := hunter.RegisterAura(core.Aura{
				Label:      "Stalker's Ally",
				ActionID:   core.ActionID{SpellID: 28757},
				Duration:   core.NeverExpires,
				BuildPhase: setBonusAura.BuildPhase,
			}).AttachStatsBuff(
				apBuff,
			)

			petAura := hunter.Pet.RegisterAura(core.Aura{
				Label:      "Stalker's Ally",
				ActionID:   core.ActionID{SpellID: 28758},
				Duration:   core.NeverExpires,
				BuildPhase: setBonusAura.BuildPhase,
			}).AttachStatsBuff(
				apBuff,
			)

			if setBonusAura.BuildPhase == core.CharacterBuildPhaseGear {
				core.MakePermanent(ownerAura)
				core.MakePermanent(petAura)
			} else {
				setBonusAura.AttachDependentAura(ownerAura).AttachDependentAura(petAura)
			}

			setBonusAura.ExposeToAPL(28756)
		},
		// (6) Set: Your ranged critical hits cause an Adrenaline Rush, granting you 50 mana.
		6: func(agent core.Agent, setBonusAura *core.Aura) {
			hunter := agent.(HunterAgent).GetHunter()
			manaMetrics := hunter.NewManaMetrics(core.ActionID{SpellID: 28753})

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:            "Adrenaline Rush",
				MetricsActionID: core.ActionID{SpellID: 28752},
				Callback:        core.CallbackOnSpellHitDealt,
				Outcome:         core.OutcomeCrit,
				ProcMask:        core.ProcMaskRanged,

				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					hunter.AddMana(sim, 50, manaMetrics)
				},
			}).ExposeToAPL(28752)
		},
		// (8) Set: Reduces the mana cost of your Multi-Shot and Aimed Shot by 20.
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_PowerCost_Flat,
				ClassMask: HunterSpellMultiShot | HunterSpellAimedShot,
				IntValue:  -20,
			}).ExposeToAPL(28751)
		},
	},
})

var ItemSetBeastLordArmor = core.NewItemSet(core.ItemSet{
	Name: "Beast Lord Armor",
	ID:   650,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			hunter := agent.(HunterAgent).GetHunter()

			exploitedWeakness := hunter.RegisterAura(core.Aura{
				Label:    "Exploited Weakness",
				ActionID: core.ActionID{SpellID: 37482},
				Duration: time.Second * 15,
			}).AttachStatBuff(stats.ArmorPenetration, 600)

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:           "Improved Kill Command",
				Callback:       core.CallbackOnCastComplete,
				ClassSpellMask: HunterSpellKillCommand,

				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					exploitedWeakness.Activate(sim)
				},
			}).ExposeToAPL(37483)
		},
	},
})

var ItemSetDemonStalkerArmor = core.NewItemSet(core.ItemSet{
	Name: "Demon Stalker Armor",
	ID:   651,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_PowerCost_Pct,
				ClassMask:  HunterSpellMultiShot,
				FloatValue: -0.1,
			}).ExposeToAPL(37485)
		},
	},
})

var ItemSetRiftStalkerArmor = core.NewItemSet(core.ItemSet{
	Name: "Rift Stalker Armor",
	ID:   652,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			hunter := agent.(HunterAgent).GetHunter()
			if hunter.Pet == nil {
				return
			}

			metrics := hunter.NewHealthMetrics(core.ActionID{SpellID: 37382})

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:               "Pet Healing",
				Callback:           core.CallbackOnSpellHitDealt,
				Outcome:            core.OutcomeLanded,
				RequireDamageDealt: true,

				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					hunter.Pet.GainHealth(sim, result.Damage*0.15, metrics)
				},
			}).ExposeToAPL(37381)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_BonusCrit_Percent,
				ClassMask:  HunterSpellSteadyShot,
				FloatValue: 5,
			}).ExposeToAPL(37505)
		},
	},
})

var ItemSetGronnstalkersArmor = core.NewItemSet(core.ItemSet{
	Name: "Gronnstalker's Armor",
	ID:   669,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			hunter := agent.(HunterAgent).GetHunter()
			hunter.GronnStalker2PcAura = setBonusAura.ExposeToAPL(38390)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Pct,
				ClassMask:  HunterSpellSteadyShot,
				FloatValue: 0.1,
			}).ExposeToAPL(38392)
		},
	},
})

func init() {
	// Thori'dal, the Star's Fury
	core.NewItemEffect(ThoridalTheStarsFuryItemID, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()

		isEquipped := hunter.HasItemEquipped(ThoridalTheStarsFuryItemID, []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged})
		buildPhase := core.Ternary(isEquipped, core.CharacterBuildPhaseGear, core.CharacterBuildPhaseNone)

		hasteAura := hunter.RegisterAura(core.Aura{
			Label:      "Legendary Bow Haste",
			ActionID:   core.ActionID{SpellID: 44972},
			Duration:   core.NeverExpires,
			BuildPhase: buildPhase,

			// Tried to do this with ExclusiveEffects but damn that was wonky and didn't work right...
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				if hunter.quiverBonusAura != nil {
					hunter.quiverBonusAura.Deactivate(sim)
				}
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				if hunter.quiverBonusAura != nil && sim.CurrentTime > 0 {
					hunter.quiverBonusAura.Activate(sim)
				}
			},
		}).AttachMultiplicativePseudoStatBuff(
			&hunter.PseudoStats.RangedSpeedMultiplier,
			quiverHasteMultipliers[proto.HunterOptions_Speed15],
		)

		ammoAura := hunter.RegisterAura(core.Aura{
			Label:    "Requires No Ammo",
			ActionID: core.ActionID{SpellID: 46699},
			Duration: core.NeverExpires,

			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				hunter.AmmoDamageBonus = 0
			},
		})

		if isEquipped {
			core.MakePermanent(hasteAura)
			core.MakePermanent(ammoAura)
		}

		hunter.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}, func(sim *core.Simulation, _ proto.ItemSlot) {
			if ranged := hunter.AutoAttacks.Ranged(); ranged != nil &&
				!hunter.HasItemEquipped(ThoridalTheStarsFuryItemID, []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}) {
				hunter.AmmoDamageBonus = hunter.AmmoDPS * ranged.SwingSpeed
				ranged.BaseDamageMin += hunter.AmmoDamageBonus
				ranged.BaseDamageMax += hunter.AmmoDamageBonus
			}
		})

		hunter.ItemSwap.RegisterProc(ThoridalTheStarsFuryItemID, hasteAura)
		hunter.ItemSwap.RegisterProc(ThoridalTheStarsFuryItemID, ammoAura)
	})

	// Beast-tamer's Shoulders
	core.NewItemEffect(30892, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()

		hunter.Pet.PseudoStats.DamageDealtMultiplier *= 1.03
		hunter.Pet.AddStat(stats.PhysicalCritPercent, 3)
	})

	// Black Bow of the Betrayer
	const BlackBowOfTheBetrayerItemID = 32336
	core.NewItemEffect(BlackBowOfTheBetrayerItemID, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()

		manaMetrics := hunter.NewManaMetrics(core.ActionID{SpellID: 29471})

		procAura := hunter.MakeProcTriggerAura(core.ProcTrigger{
			Name:              "Black Bow of the Betrayer",
			MetricsActionID:   core.ActionID{ItemID: 46939},
			SpellFlagsExclude: core.SpellFlagSuppressWeaponProcs,
			Callback:          core.CallbackOnSpellHitDealt,
			Outcome:           core.OutcomeLanded,
			ProcMask:          core.ProcMaskRanged,

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				hunter.AddMana(sim, 8, manaMetrics)
			},
		})

		hunter.ItemSwap.RegisterProc(BlackBowOfTheBetrayerItemID, procAura)
	})

	// Ashtongue Talisman of Swiftness
	const AshtongueTalismanOfSwiftnessItemID = 32487
	core.NewItemEffect(AshtongueTalismanOfSwiftnessItemID, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()
		eligibleSlots := hunter.ItemSwap.EligibleSlotsForItem(AshtongueTalismanOfSwiftnessItemID)

		statsAura := hunter.NewTemporaryStatsAura(
			"Deadly Aim",
			core.ActionID{SpellID: 40487},
			stats.Stats{
				stats.AttackPower:       275,
				stats.RangedAttackPower: 275,
			},
			time.Second*8,
		)

		procAura := hunter.MakeProcTriggerAura(core.ProcTrigger{
			Name:            "Ashtongue Talisman of Swiftness",
			MetricsActionID: core.ActionID{SpellID: 40485},
			Callback:        core.CallbackOnSpellHitDealt,
			ClassSpellMask:  HunterSpellSteadyShot,
			Outcome:         core.OutcomeLanded,
			ProcChance:      0.15,

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				statsAura.Activate(sim)
			},
		})

		hunter.AddStatProcBuff(AshtongueTalismanOfSwiftnessItemID, statsAura, false, eligibleSlots)
		hunter.ItemSwap.RegisterProcWithSlots(AshtongueTalismanOfSwiftnessItemID, procAura, eligibleSlots)
	})

	// Talon of Al'ar
	const TalonOfAlarItemID = 30448
	core.NewItemEffect(TalonOfAlarItemID, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()

		hunter.TalonOfAlarAura = hunter.RegisterAura(core.Aura{
			Label:    "Shot Power",
			ActionID: core.ActionID{SpellID: 37508},
			Duration: time.Second*6 + 1,
		})

		procAura := hunter.MakeProcTriggerAura(core.ProcTrigger{
			Name:            "Improved Shots",
			MetricsActionID: core.ActionID{SpellID: 37507},
			Callback:        core.CallbackOnSpellHitDealt,
			ClassSpellMask:  HunterSpellArcaneShot,
			Outcome:         core.OutcomeLanded,

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				hunter.TalonOfAlarAura.Activate(sim)
			},
		})

		hunter.ItemSwap.RegisterProc(TalonOfAlarItemID, procAura)
	})
}

func (hunter *Hunter) talonOfAlarBonus() float64 {
	if hunter.TalonOfAlarAura.IsActive() {
		return 40
	}
	return 0
}

func (hunter *Hunter) addPvpGloves() {
	hunter.RegisterPvPGloveMod(
		[]int32{23279, 22862, 16463, 16571, 35475, 35377, 28806, 28614, 28335, 31961, 33665, 34991},
		core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Pct,
			ClassMask:  HunterSpellMultiShot,
			FloatValue: 0.05,
		})
}
