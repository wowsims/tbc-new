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
		character.AddStat(stats.Spirit, 30)
	}
	if consumables.ScrollArm {
		character.AddStat(stats.Armor, 300)
	}

	// Pet Consumes
	for _, pet := range character.Pets {
		if pet.isGuardian {
			continue
		}

		if consumables.PetScrollAgi {
			pet.AddStat(stats.Agility, 20)
		}
		if consumables.PetScrollStr {
			pet.AddStat(stats.Strength, 20)
		}
		if consumables.PetFoodId != 0 {
			petFood := ConsumablesByID[consumables.PetFoodId]
			pet.AddStats(petFood.Stats)
		}
	}

	drumsBombsSharedTimer := character.NewTimer()

	registerPotionCD(agent, consumables)
	registerConjuredCD(agent, consumables)
	registerExplosivesCD(agent, consumables, drumsBombsSharedTimer)
	registerDrumsCD(agent, consumables, drumsBombsSharedTimer)
}

var PotionAuraTag = "Potion"

func registerPotionCD(agent Agent, consumes *proto.ConsumesSpec) {
	character := agent.GetCharacter()
	defaultPotion := consumes.PotId

	for _, potionId := range consumes.Potions {
		potion := ConsumablesByID[potionId]
		if potion.Type == proto.ConsumableType_ConsumableTypePotion {
			potMCD := makePotionActivationSpell(potion.Id, character)
			if defaultPotion == potion.Id {
				potMCD.Spell.Flags |= SpellFlagCombatPotion
				character.AddMajorCooldown(potMCD)
			}
		}
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
		mcd.Spell.Flags |= SpellFlagEncounterOnly | SpellFlagPotion | SpellFlagAPL
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
				totalRegen := character.ManaRegenPerSecondWhileCasting() * 5
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

	for _, conjuredId := range consumes.ConjuredItems {
		var conjuredMCD MajorCooldown
		switch conjuredId {
		case 22788:
			conjuredMCD = makeConjuredActivationSpell(conjuredId, character)

			flameCapProc := character.RegisterSpell(SpellConfig{
				ActionID:    conjuredMCD.Spell.ActionID,
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
				ActionID:   conjuredMCD.Spell.ActionID,
				Duration:   time.Minute * 1,
				ProcChance: 0.185,
				ProcMask:   ProcMaskMeleeOrRanged,
				Outcome:    OutcomeLanded,
				Callback:   CallbackOnSpellHitDealt,
				Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
					flameCapProc.Cast(sim, result.Target)
				},
			})

			flameCapAura := character.NewTemporaryStatsAura("Flame Cap", conjuredMCD.Spell.ActionID, stats.Stats{stats.FireDamage: 80}, time.Minute)
			flameCapAura.AttachDependentAura(procTrigger)

			oldApplyEffects := conjuredMCD.Spell.ApplyEffects
			conjuredMCD.Spell.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
				oldApplyEffects(sim, target, spell)
				flameCapAura.Activate(sim)
			}
			conjuredMCD.Spell.RelatedSelfBuff = flameCapAura.Aura
		default:
			conjuredMCD = makeConjuredActivationSpell(conjuredId, character)
		}

		if consumes.NightmareSeed {
			conjuredMCD = makeConjuredActivationSpell(22797, character)
		}

		if conjuredMCD.Spell != nil {
			oldShouldActivate := conjuredMCD.ShouldActivate
			conjuredMCD.ShouldActivate = func(sim *Simulation, character *Character) bool {
				return oldShouldActivate(sim, character) && consumes.ConjuredId == conjuredId
			}
			character.AddMajorCooldown(conjuredMCD)
		}
	}

}

func makeConjuredActivationSpell(conjuredId int32, character *Character) MajorCooldown {
	conjured := ConsumablesByID[conjuredId]
	categoryCooldownDuration := TernaryDuration(conjured.CategoryCooldownDuration > 0, conjured.CategoryCooldownDuration, time.Minute*2)
	mcd := makeConjuredActivationSpellInternal(conjured, character)

	if mcd.Spell != nil {
		mcd.Spell.Flags |= SpellFlagConjured | SpellFlagAPL
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

func makeConjuredActivationSpellInternal(conjured Consumable, character *Character) MajorCooldown {
	cooldownDuration := TernaryDuration(conjured.CooldownDuration > 0, conjured.CooldownDuration, time.Minute*2)

	conjuredCast := CastConfig{
		CD: Cooldown{
			Timer:    character.NewTimer(),
			Duration: cooldownDuration,
		},
		SharedCD: Cooldown{
			Timer:    character.GetConjuredCD(),
			Duration: cooldownDuration,
		},
	}

	actionID := ActionID{ItemID: conjured.Id}
	var aura *StatBuffAura
	mcd := MajorCooldown{
		Spell: character.GetOrRegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast:     conjuredCast,
		}),
	}
	if conjured.BuffDuration > 0 {
		// Add stat buff aura if applicable
		aura = character.NewTemporaryStatsAura(conjured.Name, actionID, conjured.Stats, conjured.BuffDuration)
		mcd.Spell.RelatedSelfBuff = aura.Aura
		mcd.Type = aura.InferCDType()
	}
	var gains []resourceGainConfig
	resourceMetrics := make(map[proto.ResourceType]*ResourceMetrics)

	for _, effectID := range conjured.EffectIds {
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
			gain := config.min + TernaryFloat64(config.spread > 1, sim.RandomFloat(conjured.Name)*config.spread, config.spread)
			switch config.resType {
			case proto.ResourceType_ResourceTypeHealth:
				gain *= character.PseudoStats.HealingTakenMultiplier
			case proto.ResourceType_ResourceTypeEnergy:
				// Thistle Tea 100 - 2 * max(0, CharacterLevel - 40) energy gain
				if conjured.Id == 7676 {
					gain -= 2 * max(0, CharacterLevel-40)
				}
			}
			character.ExecuteResourceGain(sim, config.resType, gain, resourceMetrics[config.resType])
		}
	}

	mcd.ShouldActivate = func(sim *Simulation, character *Character) bool {
		shouldActivate := true
		for _, config := range gains {
			switch config.resType {
			case proto.ResourceType_ResourceTypeMana:
				totalRegen := character.ManaRegenPerSecondWhileCasting() * 5
				manaGain := config.min + config.spread
				shouldActivate = character.MaxMana()-(character.CurrentMana()+totalRegen) >= manaGain
			case proto.ResourceType_ResourceTypeEnergy:
				if conjured.Id == 7676 {
					gain := (config.min + config.spread) - 2*max(0, CharacterLevel-40)
					shouldActivate = character.MaximumEnergy()-(character.CurrentEnergy()) >= gain
				}
			}
		}
		return shouldActivate
	}

	return mcd

}

var SuperSapperActionID = ActionID{ItemID: 23827}
var GoblinSapperActionID = ActionID{ItemID: 10646}
var EzThroDynamiteTwoActionID = ActionID{ItemID: 18588}
var CrystalChargeActionID = ActionID{ItemID: 11566}
var FelIronBombActionID = ActionID{ItemID: 23736}
var AdamantiteGrenadeActionID = ActionID{ItemID: 23737}
var GnomishFlameTurretActionID = ActionID{ItemID: 23841}

func registerExplosivesCD(agent Agent, consumes *proto.ConsumesSpec, sharedTimer *Timer) {
	character := agent.GetCharacter()
	if !character.HasProfession(proto.Profession_Engineering) {
		return
	}
	if !consumes.GoblinSapper && !consumes.SuperSapper && consumes.ExplosiveId == 0 {
		return
	}

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
		case 18588:
			filler = character.newEzThroDynamiteTwoSpell(sharedTimer)
		case 15239:
			filler = character.newCrystalChargeSpell(sharedTimer)
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
func (character *Character) newBasicExplosiveSpellConfig(sharedTimer *Timer, actionID ActionID, school SpellSchool, minDamage float64, maxDamage float64, speed float64, castTime time.Duration, cooldown Cooldown) SpellConfig {
	dealSelfDamage := actionID.SameAction(SuperSapperActionID) || actionID.SameAction(GoblinSapperActionID)

	return SpellConfig{
		ActionID:     actionID,
		SpellSchool:  school,
		ProcMask:     ProcMaskEmpty,
		MissileSpeed: speed,

		Cast: CastConfig{
			DefaultCast: Cast{
				CastTime: castTime,
			},
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
			spell.CalcAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
			if speed > 0 {
				spell.WaitTravelTime(sim, func(sim *Simulation) {
					spell.DealBatchedAoeDamage(sim)
				})
			} else {
				spell.DealBatchedAoeDamage(sim)
			}

			if dealSelfDamage {
				baseDamage := sim.Roll(minDamage, maxDamage)
				spell.CalcAndDealDamage(sim, &character.Unit, baseDamage, spell.OutcomeMagicHitAndCrit)
			}
		},
	}
}
func (character *Character) newSuperSapperSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, SuperSapperActionID, SpellSchoolFire, 900, 1500, 0, 0, Cooldown{Timer: character.NewTimer(), Duration: time.Minute * 5}))
}
func (character *Character) newGoblinSapperSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, GoblinSapperActionID, SpellSchoolFire, 450, 750, 0, 0, Cooldown{Timer: character.NewTimer(), Duration: time.Minute * 5}))
}
func (character *Character) newAdamantiteGrenadeSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, AdamantiteGrenadeActionID, SpellSchoolFire, 450, 750, 14, time.Second, Cooldown{}))
}
func (character *Character) newFelIronBombSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, FelIronBombActionID, SpellSchoolFire, 330, 770, 14, time.Second, Cooldown{}))
}
func (character *Character) newCrystalChargeSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, CrystalChargeActionID, SpellSchoolFire, 383, 517, 0, 0, Cooldown{}))
}
func (character *Character) newEzThroDynamiteTwoSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, EzThroDynamiteTwoActionID, SpellSchoolFire, 213, 287, 14, time.Second, Cooldown{}))
}

func registerDrumsCD(agent Agent, consumables *proto.ConsumesSpec, sharedTimer *Timer) {
	if consumables.DrumsId > 0 && int(consumables.DrumsId) < len(proto.Drums_value) {
		character := agent.GetCharacter()
		config := drumsSpellConfig(character, consumables.DrumsId, false)
		config.Cast = CastConfig{
			DefaultCast: Cast{
				CastTime: time.Second,
				GCD:      GCDDefault,
			},
			CD: Cooldown{
				Timer:    character.NewTimer(),
				Duration: time.Minute * 2,
			},
			SharedCD: Cooldown{
				Timer:    sharedTimer,
				Duration: time.Minute * 2,
			},
		}
		spell := character.RegisterSpell(config)

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

		if imbueId == 34340 && character.AutoAttacks.Ranged() != nil {
			character.AutoAttacks.Ranged().BaseDamageMin += 12
			character.AutoAttacks.Ranged().BaseDamageMax += 12
		}
	case 28891: // Consecrated Sharpening Stone
		character.Env.RegisterPostFinalizeEffect(func() {
			for _, at := range character.AttackTables {
				at.MobTypeBonusStats[proto.MobType_MobTypeUndead] = at.MobTypeBonusStats[proto.MobType_MobTypeUndead].Add(stats.Stats{
					stats.AttackPower:       100,
					stats.RangedAttackPower: 100,
				})
			}
		})
	}
}
