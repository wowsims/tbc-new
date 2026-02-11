package core

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func applyRaceEffects(agent Agent) {
	character := agent.GetCharacter()

	switch character.Race {
	case proto.Race_RaceBloodElf:
		character.stats[stats.ArcaneResistance] += 5
		character.stats[stats.FireResistance] += 5
		character.stats[stats.FrostResistance] += 5
		character.stats[stats.NatureResistance] += 5
		character.stats[stats.ShadowResistance] += 5

		var actionID ActionID

		var resourceMetrics *ResourceMetrics = nil
		if resourceMetrics == nil {
			if character.HasEnergyBar() {
				actionID = ActionID{SpellID: 25046}
				resourceMetrics = character.NewEnergyMetrics(actionID)
			} else if character.HasManaBar() {
				actionID = ActionID{SpellID: 28730}
				resourceMetrics = character.NewManaMetrics(actionID)
			}
		}

		spell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, spell *Spell) {
				if spell.Unit.HasEnergyBar() {
					spell.Unit.AddEnergy(sim, 10, resourceMetrics)
				} else if spell.Unit.HasManaBar() {
					spell.Unit.AddMana(sim, 10, resourceMetrics)
				}
			},
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell:    spell,
			Type:     CooldownTypeDPS,
			Priority: CooldownPriorityLow,
			ShouldActivate: func(sim *Simulation, character *Character) bool {
				if spell.Unit.HasEnergyBar() {
					return character.CurrentEnergy() <= character.maxEnergy-10
				}
				return true
			},
		})
	case proto.Race_RaceDraenei:
		character.stats[stats.ShadowResistance] += 10

		switch character.Class {
		case proto.Class_ClassHunter, proto.Class_ClassPaladin, proto.Class_ClassWarrior:
			MakePermanent(DraneiRacialAura(character, false))
		case proto.Class_ClassMage, proto.Class_ClassPriest, proto.Class_ClassShaman:
			MakePermanent(DraneiRacialAura(character, true))
		}

		character.RegisterSpell(SpellConfig{
			ActionID:    ActionID{SpellID: 28880},
			Flags:       SpellFlagAPL | SpellFlagHelpful | SpellFlagIgnoreModifiers,
			ProcMask:    ProcMaskSpellHealing,
			SpellSchool: SpellSchoolHoly,

			MaxRange: 40,

			Cast: CastConfig{
				DefaultCast: Cast{
					CastTime: time.Millisecond * 1500,
				},
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Second * 15,
				},
			},

			DamageMultiplier: 1.0,
			CritMultiplier:   character.DefaultSpellCritMultiplier(),
			ThreatMultiplier: 1.0,

			Hot: DotConfig{
				Aura: Aura{
					Label: "Gift of the Naaru" + character.Label,
				},
				NumberOfTicks:       5,
				TickLength:          time.Second * 3,
				AffectedByCastSpeed: false,
				OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
					healValue := float64((35.0 + 15*CharacterLevel) / dot.ExpectedTickCount())
					dot.Spell.CalcAndDealPeriodicHealing(sim, target, healValue, dot.OutcomeTick)
				},
			},

			ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
				spell.Hot(target).Activate(sim)
			},
		})
	case proto.Race_RaceDwarf:
		character.stats[stats.FrostResistance] += 10

		hasGunEquipped := func() bool {
			ranged := character.Ranged()
			return ranged != nil && (ranged.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeGun)
		}

		aura := character.GetOrRegisterAura(Aura{
			Label:      "Gun Specialization",
			ActionID:   ActionID{SpellID: 20595},
			Duration:   NeverExpires,
			BuildPhase: Ternary(hasGunEquipped(), CharacterBuildPhaseBase, CharacterBuildPhaseNone),
		}).AttachStatBuff(stats.RangedCritPercent, 1)

		character.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}, func(sim *Simulation, slot proto.ItemSlot) {
			if hasGunEquipped() {
				aura.Activate(sim)
			} else {
				aura.Deactivate(sim)
			}
		})

		actionID := ActionID{SpellID: 20594}

		stoneFormAura := character.NewTemporaryStatsAuraWrapped("Stoneform", actionID, stats.Stats{}, time.Second*8, func(aura *Aura) {
			aura.ApplyOnGain(func(aura *Aura, sim *Simulation) {
				character.PseudoStats.ArmorMultiplier *= 1.1
			})
			aura.ApplyOnExpire(func(aura *Aura, sim *Simulation) {
				character.PseudoStats.ArmorMultiplier /= 1.1
			})
		})

		spell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				DefaultCast: Cast{
					GCD: GCDDefault,
				},
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 3,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				stoneFormAura.Activate(sim)
			},

			RelatedSelfBuff: stoneFormAura.Aura,
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell: spell,
			Type:  CooldownTypeDPS,
		})
	case proto.Race_RaceGnome:
		character.stats[stats.ArcaneResistance] += 10
		character.MultiplyStat(stats.Intellect, 1.05)
	case proto.Race_RaceHuman:
		character.MultiplyStat(stats.Spirit, 1.10)
		applyWeaponSpecialization(character, "Mace Specialization ", 20864, false, proto.WeaponType_WeaponTypeMace)
		applyWeaponSpecialization(character, "Sword Specialization ", 20597, false, proto.WeaponType_WeaponTypeSword)
	case proto.Race_RaceNightElf:
		character.stats[stats.NatureResistance] += 10
		character.PseudoStats.BaseDodgeChance += 0.01

	case proto.Race_RaceOrc:
		// Command (Pet damage +5%)
		for _, pet := range character.Pets {
			MakePermanent(pet.GetOrRegisterAura(Aura{
				Label:    "Command",
				ActionID: ActionID{SpellID: TernaryInt32(character.Class == proto.Class_ClassWarlock, 20575, 20576)},
				Duration: NeverExpires,
			})).AttachMultiplicativePseudoStatBuff(&pet.PseudoStats.DamageDealtMultiplier, 1.05)
		}

		// Blood Fury
		actionID := ActionID{SpellID: 33697}
		apFormula := float64(character.Level)*4 + 2
		spFormula := float64(character.Level)*2 + 3
		apBonus := 0.0
		spBonus := 0.0

		switch character.Class {
		case proto.Class_ClassWarrior,
			proto.Class_ClassRogue,
			proto.Class_ClassHunter:
			apBonus = apFormula
		case proto.Class_ClassShaman:
			spBonus = spFormula
			apBonus = apFormula
		case proto.Class_ClassWarlock:
			spBonus = spFormula
		}

		buffStats := stats.Stats{
			stats.AttackPower:       apBonus,
			stats.RangedAttackPower: apBonus,
			stats.SpellDamage:       spBonus,
		}

		RegisterTemporaryStatsOnUseCD(character,
			"Blood Fury",
			buffStats,
			time.Second*15,
			SpellConfig{
				ActionID: actionID,
				Cast: CastConfig{
					CD: Cooldown{
						Timer:    character.NewTimer(),
						Duration: time.Minute * 2,
					},
				},
			})

		applyWeaponSpecialization(character, "Axe Specialization", 20574, false, proto.WeaponType_WeaponTypeAxe)
	case proto.Race_RaceTauren:
		character.stats[stats.NatureResistance] += 10
		character.MultiplyStat(stats.Health, 1.05)
	case proto.Race_RaceTroll:
		hasBowEquipped := func() bool {
			ranged := character.Ranged()
			return ranged != nil && (ranged.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeBow)
		}

		bowAura := character.GetOrRegisterAura(Aura{
			Label:      "Bow Specialization",
			ActionID:   ActionID{SpellID: 26290},
			Duration:   NeverExpires,
			BuildPhase: Ternary(hasBowEquipped(), CharacterBuildPhaseBase, CharacterBuildPhaseNone),
		}).AttachStatBuff(stats.RangedCritPercent, 1)

		hasThwrowingEquipped := func() bool {
			ranged := character.Ranged()
			return ranged != nil && (ranged.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeThrown)
		}

		throwingAura := character.GetOrRegisterAura(Aura{
			Label:      "Throwing Specialization",
			ActionID:   ActionID{SpellID: 20558},
			Duration:   NeverExpires,
			BuildPhase: Ternary(hasThwrowingEquipped(), CharacterBuildPhaseBase, CharacterBuildPhaseNone),
		}).AttachStatBuff(stats.RangedCritPercent, 1)

		character.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}, func(sim *Simulation, slot proto.ItemSlot) {
			if hasBowEquipped() {
				bowAura.Activate(sim)
			} else {
				bowAura.Deactivate(sim)
			}
			if hasThwrowingEquipped() {
				throwingAura.Activate(sim)
			} else {
				throwingAura.Deactivate(sim)
			}
		})

		// Beast Slaying (+5% damage to beasts)
		if character.CurrentTarget.MobType == proto.MobType_MobTypeBeast {
			MakePermanent(character.GetOrRegisterAura(Aura{
				Label:    "Beast Slaying",
				ActionID: ActionID{SpellID: 20557},
				Duration: NeverExpires,
			})).AttachMultiplicativePseudoStatBuff(&character.PseudoStats.DamageDealtMultiplier, 1.05)
		}

		// Berserking
		baseSpellConfig := SpellConfig{
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 3,
				},
			},
		}

		baseAuraConfig := Aura{
			ActionID: baseSpellConfig.ActionID,
			Duration: time.Second * 10,
		}

		createBerserkingSpell := func(labelSuffix string, tag int32, percentage float64) {
			var resourceMetrics *ResourceMetrics = nil
			if resourceMetrics == nil {
				if character.HasEnergyBar() {
					baseSpellConfig.ActionID = ActionID{SpellID: 26297}.WithTag(tag)
					resourceMetrics = character.NewEnergyMetrics(baseSpellConfig.ActionID)
					baseSpellConfig.EnergyCost = EnergyCostOptions{
						Cost: 10,
					}
				} else if character.HasRageBar() {
					baseSpellConfig.ActionID = ActionID{SpellID: 26296}.WithTag(tag)
					resourceMetrics = character.NewManaMetrics(baseSpellConfig.ActionID)
					baseSpellConfig.RageCost = RageCostOptions{
						Cost: 5,
					}
				} else if character.HasManaBar() {
					baseSpellConfig.ActionID = ActionID{SpellID: 20554}.WithTag(tag)
					resourceMetrics = character.NewManaMetrics(baseSpellConfig.ActionID)
					baseSpellConfig.ManaCost = ManaCostOptions{
						FlatCost: int32(character.BaseMana * 0.06),
					}
				}
			}

			auraConfig := baseAuraConfig
			auraConfig.Label = fmt.Sprintf("Berserking (%s)", labelSuffix)

			berserkingAura := character.RegisterAura(auraConfig)
			berserkingAura.
				ApplyOnGain(func(aura *Aura, sim *Simulation) {
					character.MultiplyCastSpeed(sim, percentage)
					character.MultiplyAttackSpeed(sim, percentage)
				}).
				ApplyOnExpire(func(aura *Aura, sim *Simulation) {
					character.MultiplyAttackSpeed(sim, 1/percentage)
					character.MultiplyCastSpeed(sim, 1/percentage)
				})

			berserkingSpellConfig := baseSpellConfig
			berserkingSpellConfig.RelatedSelfBuff = berserkingAura
			berserkingSpellConfig.ApplyEffects = func(sim *Simulation, _ *Unit, _ *Spell) {
				berserkingAura.Activate(sim)
			}
			berserkingSpell := character.RegisterSpell(berserkingSpellConfig)

			character.AddMajorCooldown(MajorCooldown{
				Spell: berserkingSpell,
				Type:  CooldownTypeDPS,
			})

		}

		createBerserkingSpell("10%", 1, 1.1)
		createBerserkingSpell("30%", 2, 1.3)

	case proto.Race_RaceUndead:
		character.stats[stats.ShadowResistance] += 10
	}
}

func applyWeaponSpecialization(character *Character, label string, spellID int32, oneHand bool, weaponTypes ...proto.WeaponType) {
	mask := Ternary(oneHand, character.GetDynamicProcMaskForTypesAndHand(false, weaponTypes...), character.GetDynamicProcMaskForTypes(weaponTypes...))
	expertiseBonus := 5 * ExpertisePerQuarterPercentReduction

	expSpellMod := character.AddDynamicMod(SpellModConfig{
		Kind: SpellMod_Custom,
		ApplyCustom: func(mod *SpellMod, spell *Spell) {
			if spell.ProcMask.Matches(ProcMaskMeleeOH) && !spell.ProcMask.Matches(ProcMaskMeleeMH) {
				spell.BonusExpertiseRating += mod.GetFloatValue()
			}
		},
		RemoveCustom: func(mod *SpellMod, spell *Spell) {
			if spell.ProcMask.Matches(ProcMaskMeleeOH) && !spell.ProcMask.Matches(ProcMaskMeleeMH) {
				spell.BonusExpertiseRating -= mod.GetFloatValue()
			}
		},
		FloatValue: expertiseBonus,
	})

	expStatAura := character.RegisterAura(Aura{
		Label:    fmt.Sprintf("ExpertiseStatAura (%s)", label),
		Duration: NeverExpires,
	}).AttachStatBuff(stats.ExpertiseRating, expertiseBonus)

	aura := character.RegisterAura(Aura{
		Label:      label,
		ActionID:   ActionID{SpellID: spellID},
		BuildPhase: Ternary(mask.Matches(ProcMaskMeleeMH), CharacterBuildPhaseBase, CharacterBuildPhaseNone),
		Duration:   NeverExpires,

		OnReset: func(aura *Aura, sim *Simulation) {
			if *mask != ProcMaskUnknown {
				aura.Activate(sim)
			}
		},

		OnGain: func(aura *Aura, sim *Simulation) {
			// Always add if main-hand matches
			if mask.Matches(ProcMaskMeleeMH) {
				expStatAura.Activate(sim)
				if *mask == ProcMaskMeleeMH {
					// Remove from off-hand attacks if only main-hand matches
					expSpellMod.UpdateFloatValue(-expertiseBonus)
					expSpellMod.Activate()
				}
			} else if mask.Matches(ProcMaskMeleeOH) {
				// Only add specifically to off-hand attacks
				expSpellMod.UpdateFloatValue(expertiseBonus)
				expSpellMod.Activate()
			}
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			expStatAura.Deactivate(sim)
			expSpellMod.Deactivate()
		},
	})

	character.RegisterItemSwapCallback(AllWeaponSlots(), func(sim *Simulation, slot proto.ItemSlot) {
		aura.Deactivate(sim)
		if mask.Matches(ProcMaskMelee) {
			aura.Activate(sim)
		}
	})
}
