package core

import (
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// Registers all consume-related effects to the Agent.
func applyConsumeEffects(agent Agent, partyBuffs *proto.PartyBuffs) {
	character := agent.GetCharacter()
	consumables := character.Consumables
	if consumables == nil {
		return
	}

	if consumables.FlaskId != 0 {
		flask := ConsumablesByID[consumables.FlaskId]
		character.AddStats(flask.Stats)
	}

	if consumables.BattleElixirId != 0 {
		// Elixir of Demonslaying
		if consumables.BattleElixirId == 9224 {
			character.Env.RegisterPostFinalizeEffect(func() {
				for _, at := range character.AttackTables {
					at.MobTypeBonusStats[proto.MobType_MobTypeDemon] = at.MobTypeBonusStats[proto.MobType_MobTypeDemon].Add(stats.Stats{
						stats.AttackPower:       265,
						stats.RangedAttackPower: 265,
					})
				}
			})
		} else {
			elixir := ConsumablesByID[consumables.BattleElixirId]
			character.AddStats(elixir.Stats)
		}
	}

	if consumables.GuardianElixirId != 0 {
		// Gift of Arthas
		if consumables.GuardianElixirId == 9088 {
			character.AddStat(stats.ShadowResistance, 10)
			auras := character.NewEnemyAuraArray(func(target *Unit) *Aura {
				return GiftOfArthasAura(target)
			})
			procSpell := character.RegisterSpell(SpellConfig{
				ActionID:    ActionID{SpellID: 11374},
				SpellSchool: SpellSchoolNature,
				ProcMask:    ProcMaskEmpty,

				FlatThreatBonus: 90,

				ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
					spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
					auras.Get(target).Activate(sim)
				},
			})

			character.MakeProcTriggerAura(ProcTrigger{
				Name:       "Gift of Arthas - Trigger",
				ICD:        time.Second * 3,
				ProcChance: 0.3,
				Outcome:    OutcomeLanded,
				Callback:   CallbackOnSpellHitTaken,
				Handler: func(sim *Simulation, spell *Spell, _ *SpellResult) {
					procSpell.Cast(sim, spell.Unit)
				},
			})
		} else {
			elixir := ConsumablesByID[consumables.GuardianElixirId]
			character.AddStats(elixir.Stats)
		}
	}
	if consumables.FoodId != 0 {
		food := ConsumablesByID[consumables.FoodId]
		character.AddStats(food.Stats)
	}

	// Static Imbues
	if consumables.MhImbueId != 0 && partyBuffs.WindfuryTotem == proto.TristateEffect_TristateEffectMissing {
		registerStaticImbue(agent, consumables.MhImbueId, true)
	}
	if consumables.OhImbueId != 0 {
		registerStaticImbue(agent, consumables.OhImbueId, false)
	}

	// Scrolls
	if consumables.ScrollAgi {
		character.AddStat(stats.Agility, 20)
	}
	if consumables.ScrollStr {
		character.AddStat(stats.Strength, 20)
	}
	if consumables.ScrollInt {
		character.AddStat(stats.Intellect, 20)
	}
	if consumables.ScrollSpi {
		character.AddStat(stats.Spirit, 20)
	}
	if consumables.ScrollArm {
		character.AddStat(stats.Armor, 300)
	}

	registerPotionCD(agent, consumables)
	registerConjuredCD(agent, consumables)
	registerExplosivesCD(agent, consumables)
	registerDrumsCD(agent, consumables)
}

var PotionAuraTag = "Potion"

func registerPotionCD(agent Agent, consumes *proto.ConsumesSpec) {
	character := agent.GetCharacter()
	potion := consumes.PotId

	if potion != 0 {
		potMCD := makePotionActivationSpell(potion, character)
		potMCD.Spell.Flags |= SpellFlagCombatPotion
		character.AddMajorCooldown(potMCD)
	}
}

var AlchStoneItemIDs = []int32{136197, 80508, 96252, 96253, 96254, 44322, 44323, 44324}

func (character *Character) HasAlchStone() bool {
	alchStoneEquipped := false
	for _, itemID := range AlchStoneItemIDs {
		alchStoneEquipped = alchStoneEquipped || character.HasTrinketEquipped(itemID)
	}
	return character.HasProfession(proto.Profession_Alchemy) && alchStoneEquipped
}

func makePotionActivationSpell(potionId int32, character *Character) MajorCooldown {
	potion := ConsumablesByID[potionId]
	categoryCooldownDuration := TernaryDuration(potion.CategoryCooldownDuration > 0, potion.CategoryCooldownDuration, time.Minute*2)
	mcd := makePotionActivationSpellInternal(potion, character)

	if mcd.Spell != nil {
		// Mark as 'Encounter Only' so that users are forced to select the generic Potion
		// placeholder action instead of specific potion spells, in APL prepull. This
		// prevents a mismatch between Consumes and Rotation settings.
		mcd.Spell.Flags |= SpellFlagEncounterOnly | SpellFlagPotion
		oldApplyEffects := mcd.Spell.ApplyEffects
		mcd.Spell.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
			oldApplyEffects(sim, target, spell)
			if sim.CurrentTime < 0 {
				spell.SharedCD.Set(sim.CurrentTime + categoryCooldownDuration)
				character.UpdateMajorCooldowns()
			}
		}
	}
	return mcd

}

type resourceGainConfig struct {
	resType proto.ResourceType
	min     float64
	spread  float64
}

func makePotionActivationSpellInternal(potion Consumable, character *Character) MajorCooldown {
	stoneMul := TernaryFloat64(character.HasAlchStone(), 1.4, 1.0)
	cooldownDuration := TernaryDuration(potion.CooldownDuration > 0, potion.CooldownDuration, time.Minute*2)

	potionCast := CastConfig{
		CD: Cooldown{
			Timer:    character.NewTimer(),
			Duration: cooldownDuration,
		},
		SharedCD: Cooldown{
			Timer:    character.GetPotionCD(),
			Duration: cooldownDuration,
		},
	}

	actionID := ActionID{ItemID: potion.Id}
	var aura *StatBuffAura
	mcd := MajorCooldown{
		Spell: character.GetOrRegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast:     potionCast,
		}),
	}
	if potion.BuffDuration > 0 {
		// Add stat buff aura if applicable
		aura = character.NewTemporaryStatsAura(potion.Name, actionID, potion.Stats, potion.BuffDuration)
		mcd.Spell.RelatedSelfBuff = aura.Aura
		mcd.Type = aura.InferCDType()
	}
	var gains []resourceGainConfig
	resourceMetrics := make(map[proto.ResourceType]*ResourceMetrics)

	for _, effectID := range potion.EffectIds {
		e := SpellEffectsById[effectID]
		resourceType := e.GetResourceType()
		if e.Type == proto.EffectType_EffectTypeResourceGain && resourceType != 0 {
			if resourceType == proto.ResourceType_ResourceTypeMana && mcd.Type != CooldownTypeSurvival {
				mcd.Type = CooldownTypeMana
			} else if resourceType == proto.ResourceType_ResourceTypeHealth {
				mcd.Type = CooldownTypeSurvival
			} else {
				mcd.Type = CooldownTypeDPS
			}
			gains = append(gains, resourceGainConfig{
				resType: resourceType,
				min:     e.MinEffectSize,
				spread:  e.EffectSpread,
			})
			if _, exists := resourceMetrics[resourceType]; !exists {
				resourceMetrics[resourceType] = character.Metrics.NewResourceMetrics(actionID, resourceType)
			}
			// Preload resource types that are found on this item
			if resourceMetrics[resourceType] == nil {
				resourceMetrics[resourceType] = character.Metrics.NewResourceMetrics(actionID, resourceType)
			}
		}
	}

	mcd.Spell.ApplyEffects = func(sim *Simulation, _ *Unit, _ *Spell) {
		if aura != nil {
			aura.Activate(sim)
		}
		for _, config := range gains {
			gain := config.min + sim.RandomFloat(potion.Name)*config.spread
			gain *= stoneMul
			if config.resType == proto.ResourceType_ResourceTypeHealth {
				gain *= character.PseudoStats.HealingTakenMultiplier
			}
			character.ExecuteResourceGain(sim, config.resType, gain, resourceMetrics[config.resType])
		}
	}

	mcd.ShouldActivate = func(sim *Simulation, character *Character) bool {
		shouldActivate := true
		for _, config := range gains {
			switch config.resType {
			case proto.ResourceType_ResourceTypeMana:
				totalRegen := character.ManaRegenPerSecondWhileCombat() * 5
				manaGain := config.min + config.spread
				manaGain *= stoneMul
				shouldActivate = character.MaxMana()-(character.CurrentMana()+totalRegen) >= manaGain
			}
		}
		return shouldActivate
	}

	return mcd

}

var ConjuredAuraTag = "Conjured"

func registerConjuredCD(agent Agent, consumes *proto.ConsumesSpec) {
	character := agent.GetCharacter()

	//Todo: Implement dynamic handling like pots etc.
	switch consumes.ConjuredId {
	case 5512:
		actionID := ActionID{ItemID: 5512}
		healthMetrics := character.NewHealthMetrics(actionID)

		spell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				SharedCD: Cooldown{
					Timer:    character.GetConjuredCD(),
					Duration: time.Minute * 2,
				},

				// Enforce only one HS per fight
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 60,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				character.GainHealth(sim, 0.45*character.baseStats[stats.Health], healthMetrics)
			},
		})
		character.AddMajorCooldown(MajorCooldown{
			Spell: spell,
			Type:  CooldownTypeSurvival,
		})
	case 7676:
		actionID := ActionID{ItemID: 7676}
		energyMetrics := character.NewEnergyMetrics(actionID)

		spell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				SharedCD: Cooldown{
					Timer:    character.GetConjuredCD(),
					Duration: time.Minute * 2,
				},

				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 5,
				},
			},
			ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
				character.AddEnergy(sim, 40, energyMetrics)
			},
		})
		character.AddMajorCooldown(MajorCooldown{
			Spell: spell,
			Type:  CooldownTypeDPS,
		})
	case 22788:

		actionID := ActionID{ItemID: 22788}

		flameCapProc := character.RegisterSpell(SpellConfig{
			ActionID:    actionID,
			SpellSchool: SpellSchoolFire,
			ProcMask:    ProcMaskEmpty,

			DamageMultiplier: 1,
			CritMultiplier:   character.DefaultSpellCritMultiplier(),
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
				spell.CalcAndDealDamage(sim, target, 40, spell.OutcomeMagicHitAndCrit)
			},
		})

		procTrigger := character.MakeProcTriggerAura(ProcTrigger{
			Name:       "Flame Cap - Proc",
			ActionID:   actionID,
			Duration:   time.Minute * 1,
			ProcChance: 0.185,
			ProcMask:   ProcMaskMeleeOrRanged,
			Outcome:    OutcomeLanded,
			Callback:   CallbackOnSpellHitDealt,
			Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
				flameCapProc.Cast(sim, result.Target)
			},
		})

		flameCapAura := character.NewTemporaryStatsAura("Flame Cap", actionID, stats.Stats{stats.FireDamage: 80}, time.Minute)
		flameCapAura.AttachDependentAura(procTrigger)

		spell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.GetConjuredCD(),
					Duration: time.Minute * 3,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				flameCapAura.Activate(sim)
			},
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell: spell,
			Type:  CooldownTypeDPS,
		})
	case 20520:
		actionID := ActionID{ItemID: 20520}
		manaMetrics := character.NewManaMetrics(actionID)
		// damageTakenManaMetrics := character.NewManaMetrics(ActionID{SpellID: 33776})
		spell := character.RegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.GetConjuredCD(),
					Duration: time.Minute * 15,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				// Restores 900 to 1500 mana. (2 Min Cooldown)
				manaGain := sim.RollWithLabel(900, 1500, "dark rune")
				character.AddMana(sim, manaGain, manaMetrics)

				// if character.Class == proto.Class_ClassPaladin {
				// 	// Paladins gain extra mana from self-inflicted damage
				// 	// TO-DO: It is possible for damage to be resisted or to crit
				// 	// This would affect mana returns for Paladins
				// 	manaFromDamage := manaGain * 2.0 / 3.0 * 0.1
				// 	character.AddMana(sim, manaFromDamage, damageTakenManaMetrics, false)
				// }
			},
		})
		character.AddMajorCooldown(MajorCooldown{
			Spell: spell,
			Type:  CooldownTypeMana,
			ShouldActivate: func(sim *Simulation, character *Character) bool {
				// Only pop if we have less than the max mana provided by the potion minus 1mp5 tick.
				totalRegen := character.ManaRegenPerSecondWhileCombat() * 5
				return character.MaxMana()-(character.CurrentMana()+totalRegen) >= 1500
			},
		})
	}

	if consumes.NightmareSeed {
		aura := character.NewTemporaryStatsAura(
			"Nightmare Seed",
			ActionID{SpellID: 28726},
			stats.Stats{stats.Health: 2000},
			time.Second*30,
		)

		spell := character.RegisterSpell(SpellConfig{
			ActionID: ActionID{ItemID: 22797},
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.GetConjuredCD(),
					Duration: time.Minute * 3,
				},
			},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				aura.Activate(sim)
			},
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell: spell,
			Type:  CooldownTypeSurvival,
		})
	}
}

var SuperSapperActionID = ActionID{ItemID: 23827}
var GoblinSapperActionID = ActionID{ItemID: 10646}
var FelIronBombActionID = ActionID{ItemID: 23736}
var AdamantiteGrenadeActionID = ActionID{ItemID: 23737}
var GnomishFlameTurretActionID = ActionID{ItemID: 23841}

func registerExplosivesCD(agent Agent, consumes *proto.ConsumesSpec) {
	character := agent.GetCharacter()
	if !character.HasProfession(proto.Profession_Engineering) {
		return
	}
	if !consumes.GoblinSapper && !consumes.SuperSapper && consumes.ExplosiveId == 0 {
		return
	}
	sharedTimer := character.NewTimer()

	if consumes.SuperSapper {
		character.AddMajorCooldown(MajorCooldown{
			Spell:    character.newSuperSapperSpell(sharedTimer),
			Type:     CooldownTypeDPS | CooldownTypeExplosive,
			Priority: CooldownPriorityLow + 30,
		})
	}
	if consumes.GoblinSapper {
		character.AddMajorCooldown(MajorCooldown{
			Spell:    character.newGoblinSapperSpell(sharedTimer),
			Type:     CooldownTypeDPS | CooldownTypeExplosive,
			Priority: CooldownPriorityLow + 20,
		})
	}
	if consumes.ExplosiveId > 0 {
		var filler *Spell
		switch consumes.ExplosiveId {
		case 30217:
			filler = character.newAdamantiteGrenadeSpell(sharedTimer)
		case 30216:
			filler = character.newFelIronBombSpell(sharedTimer)
		case 30526:
			// Summon Gnomish Turret? Just treat it like a DoT? TBD
		}

		character.AddMajorCooldown(MajorCooldown{
			Spell:    filler,
			Type:     CooldownTypeDPS | CooldownTypeExplosive,
			Priority: CooldownPriorityLow + 10,
		})
	}
}

// Creates a spell object for the common explosive case.
func (character *Character) newBasicExplosiveSpellConfig(sharedTimer *Timer, actionID ActionID, school SpellSchool, minDamage float64, maxDamage float64, cooldown Cooldown) SpellConfig {
	dealSelfDamage := actionID.SameAction(SuperSapperActionID) || actionID.SameAction(GoblinSapperActionID)

	return SpellConfig{
		ActionID:    actionID,
		SpellSchool: school,
		ProcMask:    ProcMaskEmpty,

		Cast: CastConfig{
			CD: cooldown,
			SharedCD: Cooldown{
				Timer:    sharedTimer,
				Duration: time.Minute,
			},
		},

		// Explosives always have 1% resist chance, so just give them hit cap.
		BonusHitPercent:  100,
		DamageMultiplier: 1,
		CritMultiplier:   2,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			baseDamage := sim.Roll(minDamage, maxDamage) * sim.Encounter.AOECapMultiplier()
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)

			if dealSelfDamage {
				baseDamage := sim.Roll(minDamage, maxDamage)
				spell.CalcAndDealDamage(sim, &character.Unit, baseDamage, spell.OutcomeMagicHitAndCrit)
			}
		},
	}
}
func (character *Character) newSuperSapperSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, SuperSapperActionID, SpellSchoolFire, 900, 1500, Cooldown{Timer: character.NewTimer(), Duration: time.Minute * 5}))
}
func (character *Character) newGoblinSapperSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, GoblinSapperActionID, SpellSchoolFire, 450, 750, Cooldown{Timer: character.NewTimer(), Duration: time.Minute * 5}))
}
func (character *Character) newAdamantiteGrenadeSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, AdamantiteGrenadeActionID, SpellSchoolFire, 450, 750, Cooldown{}))
}
func (character *Character) newFelIronBombSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, AdamantiteGrenadeActionID, SpellSchoolFire, 330, 770, Cooldown{}))
}

func registerDrumsCD(agent Agent, consumables *proto.ConsumesSpec) {
	if consumables.DrumsId > 0 {
		character := agent.GetCharacter()
		actionID := ActionID{SpellID: consumables.DrumsId}
		var drumLabel string
		var drumStats stats.Stats
		var duration time.Duration
		switch consumables.DrumsId {
		case 351355:
			drumLabel = "Drums of Battle"
			drumStats = stats.Stats{stats.MeleeHasteRating: 80, stats.SpellHasteRating: 80}
			duration = time.Second * 30
		case 351360:
			drumLabel = "Drums of War"
			drumStats = stats.Stats{stats.AttackPower: 60, stats.RangedAttackPower: 60, stats.SpellDamage: 30}
			duration = time.Second * 30
		case 351358:
			drumLabel = "Drums of Restoration"
			drumStats = stats.Stats{stats.MP5: 200}
			duration = time.Second * 15
		}
		aura := character.NewTemporaryStatsAura(drumLabel, actionID, drumStats, duration)

		spell := character.GetOrRegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			ProcMask: ProcMaskEmpty,

			Cast: CastConfig{
				CD: Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},

			ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
				aura.Activate(sim)
			},

			RelatedSelfBuff: aura.Aura,
		})

		character.AddMajorCooldown(MajorCooldown{
			Spell:    spell,
			Type:     CooldownTypeDPS,
			Priority: CooldownPriorityDrums,
		})
	}
}

func registerStaticImbue(agent Agent, imbueId int32, isMH bool) {
	character := agent.GetCharacter()
	switch imbueId {
	case 25123: // Mana Oil
		character.AddStat(stats.HealingPower, 25)
		character.AddStat(stats.MP5, 12)
	case 25122: // Briliant Wizard Oil
		character.AddStat(stats.SpellDamage, 36)
		character.AddStat(stats.SpellCritRating, 14)
	case 28017: // Superior Wizard Oil
		character.AddStat(stats.SpellDamage, 42)
	case 29453, 34340: // Addy Stone
		character.AddStat(stats.MeleeCritRating, 14)
		if isMH {
			character.AutoAttacks.MH().BaseDamageMax += 12
			character.AutoAttacks.MH().BaseDamageMin += 12
		} else {
			character.AutoAttacks.OH().BaseDamageMax += 12
			character.AutoAttacks.OH().BaseDamageMin += 12
		}
	}
}
