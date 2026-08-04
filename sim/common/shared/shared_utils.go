package shared

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type ProcStatBonusEffect struct {
	Name               string
	ItemID             int32
	EnchantID          int32
	MaxStacks          int32
	Callback           core.AuraCallback
	ProcMask           core.ProcMask
	Outcome            core.HitOutcome
	RequireDamageDealt bool
	ClassSpellsOnly    bool

	// Any other custom proc conditions not covered by the above fields.
	CustomProcCondition core.CustomStatBuffProcCondition
}

type DamageEffect struct {
	SpellID          int32
	School           core.SpellSchool
	MinDmg           float64
	MaxDmg           float64
	BonusCoefficient float64
	IsMelee          bool
	ProcMask         core.ProcMask
	Outcome          OutcomeType
	Flags            core.SpellFlag
}

type ExtraSpellInfo struct {
	Spell   *core.Spell
	Trigger func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult)
}

type ItemVariant struct {
	ItemID   int32
	ItemName string
}

type CustomProcHandler func(sim *core.Simulation, procAura *core.StatBuffAura)

func NewProcStatBonusEffectWithDamageProc(config ProcStatBonusEffect, damage DamageEffect) {
	procMask := core.ProcMaskEmpty
	if damage.ProcMask != core.ProcMaskUnknown {
		procMask = damage.ProcMask
	}

	factory_StatBonusEffect(config, func(agent core.Agent) ExtraSpellInfo {
		character := agent.GetCharacter()

		procSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:                 core.ActionID{SpellID: damage.SpellID},
			SpellSchool:              damage.School,
			ProcMask:                 procMask,
			Flags:                    core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
			DamageMultiplier:         1,
			CritMultiplier:           character.DefaultMeleeCritMultiplier(),
			DamageMultiplierAdditive: 1,
			ThreatMultiplier:         1,
			BonusCoefficient:         damage.BonusCoefficient,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, sim.Roll(damage.MinDmg, damage.MaxDmg), GetOutcome(spell, damage.Outcome))
			},
		})

		return ExtraSpellInfo{
			Spell: procSpell,
			Trigger: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				procSpell.Cast(sim, result.Target)
			},
		}
	})
}

func factory_StatBonusEffect(config ProcStatBonusEffect, extraSpell func(agent core.Agent) ExtraSpellInfo) {
	isEnchant := config.EnchantID != 0

	// Ignore empty dummy implementations
	if config.Callback == core.CallbackEmpty {
		return
	}

	// Soft fail to allow for overrides for bad effects
	if isEnchant {
		if core.HasEnchantEffect(config.EnchantID) {
			return
		}
	} else {
		if core.HasItemEffect(config.ItemID) {
			return
		}
	}

	var effectFn func(id int32, effect core.ApplyEffect)
	var effectID int32
	var triggerActionID core.ActionID
	if isEnchant {
		effectID = config.EnchantID
		effectFn = core.NewEnchantEffect
		triggerActionID = core.ActionID{SpellID: effectID}
	} else {
		effectID = config.ItemID
		effectFn = core.NewItemEffect
		triggerActionID = core.ActionID{ItemID: effectID}
	}

	effectFn(effectID, func(agent core.Agent) {
		character := agent.GetCharacter()
		var eligibleSlots []proto.ItemSlot
		procEffects := make(map[int32]*proto.ItemEffect)
		if isEnchant {
			eligibleSlots = character.ItemSwap.EligibleSlotsForEffect(effectID)
			ench := core.GetEnchantByEffectID(effectID)
			for _, effect := range ench.EnchantEffects {
				if effect.GetProc() != nil {
					procEffects[effect.BuffId] = effect
				}
			}
		} else {
			eligibleSlots = character.ItemSwap.EligibleSlotsForItem(effectID)

			item := core.GetItemByID(effectID)
			if item.ItemEffects != nil {
				for _, effect := range item.ItemEffects {
					if effect.GetProc() != nil {
						procEffects[effect.BuffId] = effect
					}
				}
			}
		}

		if len(procEffects) == 0 {
			err, _ := fmt.Printf("Error getting proc effects for item/enchant %v", effectID)
			panic(err)
		}

		for _, effect := range procEffects {
			proc := effect.GetProc()
			procAction := core.ActionID{SpellID: effect.BuffId}
			var procAura *core.StatBuffAura
			if effect.MaxCumulativeStacks > 0 {
				procAura = core.MakeStackingAura(character, core.StackingStatAura{
					Aura: core.Aura{
						Label:     config.Name + " Proc",
						ActionID:  procAction,
						Duration:  time.Millisecond * time.Duration(effect.EffectDurationMs),
						MaxStacks: effect.MaxCumulativeStacks,
					},
					BonusPerStack: stats.FromProtoMap(effect.ScalingOptions[int32(0)].Stats),
				})
			} else {
				procAura = character.NewTemporaryStatsAura(
					config.Name+" Proc",
					procAction,
					stats.FromProtoMap(effect.ScalingOptions[int32(0)].Stats),
					time.Millisecond*time.Duration(effect.EffectDurationMs),
				)
			}

			var dpm *core.DynamicProcManager
			if proc.GetPpm() > 0 {
				if config.ProcMask == core.ProcMaskUnknown {
					if isEnchant {
						dpm = character.NewDynamicLegacyProcForEnchant(effectID, proc.GetPpm(), 0)
					} else {
						dpm = character.NewDynamicLegacyProcForWeapon(effectID, proc.GetPpm(), 0)
					}
				} else {
					dpm = character.NewLegacyPPMManager(proc.GetPpm(), config.ProcMask)
				}
			}

			procAura.CustomProcCondition = config.CustomProcCondition
			var customHandler CustomProcHandler
			if config.CustomProcCondition != nil {
				customHandler = func(sim *core.Simulation, procAura *core.StatBuffAura) {
					if procAura.CanProc(sim) {
						procAura.Activate(sim)
					} else {
						// reset ICD condition was not fulfilled
						if procAura.Icd != nil && procAura.Icd.Duration != 0 {
							procAura.Icd.Reset()
						}
					}
				}
			}
			var procSpell ExtraSpellInfo
			if extraSpell != nil {
				procSpell = extraSpell(agent)
			}

			handler := func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if customHandler != nil {
					customHandler(sim, procAura)
				} else {
					procAura.Activate(sim)
					if effect.MaxCumulativeStacks > 0 {
						procAura.AddStack(sim)
					}
					if procSpell.Spell != nil {
						procSpell.Trigger(sim, spell, result)
					}
				}
			}

			triggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
				ActionID:           triggerActionID,
				Name:               config.Name,
				Callback:           config.Callback,
				ProcMask:           config.ProcMask,
				Outcome:            config.Outcome,
				RequireDamageDealt: config.RequireDamageDealt,
				ClassSpellsOnly:    config.ClassSpellsOnly,
				ProcChance:         proc.GetProcChance(),
				DPM:                dpm,
				ICD:                time.Millisecond * time.Duration(proc.IcdMs),
				Handler:            handler,
			})

			if proc.IcdMs != 0 {
				procAura.Icd = triggerAura.Icd
			}
			if isEnchant {
				character.ItemSwap.RegisterEnchantProcWithSlots(effectID, triggerAura, eligibleSlots)
			} else {
				character.ItemSwap.RegisterProcWithSlots(effectID, triggerAura, eligibleSlots)
			}

			character.AddStatProcBuff(effectID, procAura, isEnchant, eligibleSlots)

		}
	})
}

func NewProcStatBonusEffectWithVariants(config ProcStatBonusEffect, variants []ItemVariant) {
	var maxItemID int32

	for _, variant := range variants {
		maxItemID = max(maxItemID, variant.ItemID)
	}

	for _, variant := range variants {
		config.Name = variant.ItemName
		config.ItemID = variant.ItemID
		core.AddEffectsToTest = (config.ItemID == maxItemID)
		NewProcStatBonusEffect(config)
	}

	core.AddEffectsToTest = true
}

func NewProcStatBonusEffect(config ProcStatBonusEffect) {
	factory_StatBonusEffect(config, nil)
}

func NewSimpleStatActive(itemID int32) {

	// Soft fail to allow for overrides for bad effects
	if core.HasItemEffect(itemID) {
		return
	}

	core.NewItemEffect(itemID, func(agent core.Agent) {
		item := core.GetItemByID(itemID)
		if item == nil {
			panic(fmt.Sprintf("No item with ID: %d", itemID))
		}

		itemEffects := item.ItemEffects
		if len(itemEffects) == 0 {
			panic(fmt.Sprintf("No effects data for item with ID: %d", itemID))
		}

		hasEffect := false
		for idx, itemEffect := range itemEffects {
			onUseData := itemEffect.GetOnUse()

			if onUseData == nil {
				if !hasEffect && idx == len(itemEffects)-1 {
					panic(fmt.Sprintf("No active effects found for item with ID: %d!", itemID))
				}
				continue
			}

			hasEffect = true
			spellConfig := core.SpellConfig{
				ActionID: core.ActionID{ItemID: itemID},
			}

			character := agent.GetCharacter()
			spellConfig.Cast.CD = core.Cooldown{
				Timer:    character.NewTimer(),
				Duration: time.Duration(onUseData.CooldownMs) * time.Millisecond,
			}
			// if SpellCategoryID is 0 we seemingly do not share cd with anything
			// Say Darkmoon Card: Earthquake and Ruthless Gladiator's Emblem of Cruelty even though tooltip shows as such
			if onUseData.CategoryId > 0 {
				sharedCDDuration := time.Duration(onUseData.CategoryCooldownMs) * time.Millisecond
				if sharedCDDuration == 0 {
					sharedCDDuration = time.Millisecond * time.Duration(itemEffect.EffectDurationMs)
				}

				sharedCDTimer := character.GetOrInitSpellCategoryTimer(onUseData.CategoryId)
				spellConfig.Cast.SharedCD = core.Cooldown{
					Timer:    sharedCDTimer,
					Duration: sharedCDDuration,
				}
			}

			core.RegisterTemporaryStatsOnUseCD(character, itemEffect.BuffName, stats.FromProtoMap(itemEffect.ScalingOptions[int32(0)].Stats), time.Millisecond*time.Duration(itemEffect.EffectDurationMs), spellConfig)
		}
	})
}

type StackingStatBonusCD struct {
	Name               string
	ID                 int32
	Duration           time.Duration
	CD                 time.Duration
	Callback           core.AuraCallback
	ProcMask           core.ProcMask
	SpellFlags         core.SpellFlag
	Outcome            core.HitOutcome
	RequireDamageDealt bool
	ProcChance         float64
	IsDefensive        bool

	// The stacks will only be granted as long as the trinket is active
	TrinketLimitsDuration bool
}

// Creates a new stacking stats bonus aura based on the configuration. If Bonus is not given, the ItemEffect of the item will be used
// to determine the correct values.
func NewStackingStatBonusCD(config StackingStatBonusCD) {
	core.NewItemEffect(config.ID, func(agent core.Agent) {
		character := agent.GetCharacter()

		item := core.GetItemByID(config.ID)
		if item == nil {
			panic(fmt.Sprintf("No item with ID: %d", config.ID))
		}

		itemEffects := item.ItemEffects
		if len(itemEffects) == 0 {
			panic(fmt.Sprintf("No effects data for item with ID: %d", config.ID))
		}

		for _, itemEffect := range itemEffects {
			auraID := core.ActionID{SpellID: itemEffect.BuffId}
			auraDuration := time.Millisecond * time.Duration(itemEffect.EffectDurationMs)
			if auraID.IsEmptyAction() {
				auraID = core.ActionID{ItemID: config.ID}
			}

			duration := core.TernaryDuration(config.TrinketLimitsDuration, core.NeverExpires, auraDuration)
			statAura := core.MakeStackingAura(character, core.StackingStatAura{
				Aura: core.Aura{
					Label:     config.Name + " Proc",
					ActionID:  auraID,
					Duration:  duration,
					MaxStacks: itemEffect.MaxCumulativeStacks,
				},
				BonusPerStack: stats.FromProtoMap(itemEffect.ScalingOptions[int32(0)].Stats),
			})

			// If trinket limits duration create a separate proc aura
			var procAura *core.Aura = statAura.Aura
			if config.TrinketLimitsDuration {
				procAura = character.RegisterAura(core.Aura{
					Label:    fmt.Sprintf("%s Limit Aura %s", config.Name, itemEffect.BuffName),
					ActionID: auraID,
					Duration: auraDuration,
					OnExpire: func(_ *core.Aura, sim *core.Simulation) {
						statAura.Deactivate(sim)
					},
				})
			}

			procAura.AttachProcTriggerCallback(&character.Unit, core.ProcTrigger{
				Name:               config.Name,
				Callback:           config.Callback,
				ProcMask:           config.ProcMask,
				SpellFlags:         config.SpellFlags,
				Outcome:            config.Outcome,
				RequireDamageDealt: config.RequireDamageDealt,
				ProcChance:         config.ProcChance,
				Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					statAura.AddStack(sim)
				},
			})

			var sharedTimer *core.Timer
			if config.IsDefensive {
				sharedTimer = character.GetDefensiveTrinketCD()
			} else {
				sharedTimer = character.GetOffensiveTrinketCD()
			}

			spell := character.RegisterSpell(core.SpellConfig{
				ActionID: core.ActionID{ItemID: config.ID},
				Flags:    core.SpellFlagNoOnCastComplete,

				Cast: core.CastConfig{
					CD: core.Cooldown{
						Timer:    character.NewTimer(),
						Duration: config.CD,
					},
					SharedCD: core.Cooldown{
						Timer:    sharedTimer,
						Duration: config.Duration,
					},
				},

				ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
					statAura.Activate(sim)
				},

				RelatedSelfBuff: statAura.Aura,
			})

			character.AddMajorCooldown(core.MajorCooldown{
				Spell: spell,
				Type:  core.CooldownTypeDPS,
			})
		}
	})
}

func NewStackingStatBonusEffectWithVariants(config ProcStatBonusEffect, variants []ItemVariant) {
	var maxItemID int32

	for _, variant := range variants {
		maxItemID = max(maxItemID, variant.ItemID)
	}

	for _, variant := range variants {
		config.Name = variant.ItemName
		config.ItemID = variant.ItemID
		core.AddEffectsToTest = (config.ItemID == maxItemID)
		factory_StatBonusEffect(config, nil)
	}

	core.AddEffectsToTest = true
}

// func NewStackingStatBonusEffect(config StackingStatBonusEffect) {
// 	// Ignore empty dummy implementations
// 	if config.Callback == core.CallbackEmpty {
// 		return
// 	}

// 	if core.HasItemEffect(config.ItemID) {
// 		return
// 	}

// 	core.NewItemEffect(config.ItemID, func(agent core.Agent) {
// 		character := agent.GetCharacter()
// 		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(config.ItemID)
// 		item := core.GetItemByID(config.ItemID)

// 		for _, itemEffect := range item.ItemEffects {

// 			var procEffect *proto.ItemEffect
// 			if itemEffect != nil {
// 				if itemEffect.GetProc() != nil {
// 					procEffect = itemEffect
// 				}
// 			}

// 			if procEffect == nil {
// 				err, _ := fmt.Printf("Error getting proc effect for item/enchant %v", config.ItemID)
// 				panic(err)
// 			}

// 			proc := procEffect.GetProc()
// 			procAction := core.ActionID{SpellID: procEffect.BuffId}
// 			procAura := core.MakeStackingAura(character, core.StackingStatAura{
// 				Aura: core.Aura{
// 					Label:     config.Name + " Proc",
// 					ActionID:  procAction,
// 					Duration:  time.Millisecond * time.Duration(procEffect.EffectDurationMs),
// 					MaxStacks: config.MaxStacks,
// 				},
// 				BonusPerStack: stats.FromProtoMap(procEffect.ScalingOptions[int32(0)].Stats),
// 			})

// 			var dpm *core.DynamicProcManager
// 			if proc.GetPpm() > 0 {
// 				if config.ProcMask == core.ProcMaskUnknown {
// 					dpm = character.NewDynamicLegacyProcForEnchant(config.ItemID, proc.GetPpm(), 0)
// 				} else {
// 					dpm = character.NewLegacyPPMManager(proc.GetPpm(), config.ProcMask)
// 				}
// 			}

// 			triggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
// 				ActionID:           core.ActionID{ItemID: config.ItemID},
// 				Name:               config.Name,
// 				Callback:           config.Callback,
// 				ProcMask:           config.ProcMask,
// 				SpellFlags:         config.SpellFlags,
// 				Outcome:            config.Outcome,
// 				RequireDamageDealt: config.RequireDamageDealt,
// 				ProcChance:         proc.GetProcChance(),
// 				DPM:                dpm,
// 				ICD:                time.Millisecond * time.Duration(proc.IcdMs),
// 				Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
// 					procAura.Activate(sim)
// 					procAura.AddStack(sim)
// 				},
// 			})

// 			procAura.Icd = triggerAura.Icd
// 			character.AddStatProcBuff(config.ItemID, procAura, false, eligibleSlots)
// 			character.ItemSwap.RegisterProcWithSlots(config.ItemID, triggerAura, eligibleSlots)

// 		}
// 	})
// }

type OutcomeType uint64

const (
	OutcomeDefault                  = 0
	OutcomeMeleeCanCrit OutcomeType = iota
	OutcomeMeleeNoCrit
	OutcomeMeleeNoBlockDodgeParry
	OutcomeMeleeNoBlockDodgeParryCrit
	OutcomeSpellCanCrit
	OutcomeSpellNoCrit
	OutcomeSpellNoMissCanCrit
	OutcomeRangedCanCrit
	OutcomeAlwaysHit
)

type ProcDamageEffect struct {
	ItemID     int32
	SpellID    int32
	EnchantID  int32
	Trigger    core.ProcTrigger
	TriggerDPM func(*core.Character) *core.DynamicProcManager
	School     core.SpellSchool
	MinDmg     float64
	MaxDmg     float64
	IsMelee    bool
	Flags      core.SpellFlag
	Outcome    OutcomeType
}

func GetOutcome(spell *core.Spell, outcome OutcomeType) core.OutcomeApplier {
	switch outcome {
	case OutcomeMeleeCanCrit:
		return spell.OutcomeMeleeSpecialHitAndCrit
	case OutcomeMeleeNoCrit:
		return spell.OutcomeMeleeSpecialHit
	case OutcomeMeleeNoBlockDodgeParry:
		return spell.OutcomeMeleeSpecialNoBlockDodgeParry
	case OutcomeMeleeNoBlockDodgeParryCrit:
		return spell.OutcomeMeleeSpecialNoBlockDodgeParryNoCrit
	case OutcomeSpellCanCrit:
		return spell.OutcomeMagicHitAndCrit
	case OutcomeSpellNoMissCanCrit:
		return spell.OutcomeMagicCrit
	case OutcomeSpellNoCrit:
		return spell.OutcomeMagicHit
	case OutcomeRangedCanCrit:
		return spell.OutcomeRangedHitAndCrit
	case OutcomeAlwaysHit:
		return spell.OutcomeAlwaysHit
	default:
		return spell.OutcomeMagicHitAndCrit
	}
}

func NewProcDamageEffect(config ProcDamageEffect) {
	isEnchant := config.EnchantID != 0

	var effectFn func(id int32, effect core.ApplyEffect)
	var effectID int32
	var triggerActionID core.ActionID

	if isEnchant {
		effectID = config.EnchantID
		effectFn = core.NewEnchantEffect
		triggerActionID = core.ActionID{SpellID: config.SpellID}
	} else {
		effectID = config.ItemID
		effectFn = core.NewItemEffect
		triggerActionID = core.ActionID{ItemID: config.ItemID}
	}

	effectFn(effectID, func(agent core.Agent) {
		character := agent.GetCharacter()

		minDmg := config.MinDmg
		maxDmg := config.MaxDmg

		critMultiplier := core.TernaryFloat64(config.IsMelee, character.DefaultMeleeCritMultiplier(), character.DefaultSpellCritMultiplier())

		if core.ActionID.IsEmptyAction(config.Trigger.ActionID) {
			config.Trigger.ActionID = triggerActionID
		}

		if config.TriggerDPM != nil {
			config.Trigger.DPM = config.TriggerDPM(character)
		}

		damageSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: config.SpellID},
			SpellSchool: config.School,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       config.Flags,

			DamageMultiplier: 1,
			CritMultiplier:   critMultiplier,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, sim.Roll(minDmg, maxDmg), GetOutcome(spell, config.Outcome))
			},
		})

		triggerConfig := config.Trigger
		triggerConfig.TriggerImmediately = true
		triggerConfig.Handler = func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			damageSpell.Cast(sim, character.CurrentTarget)
		}
		triggerAura := character.MakeProcTriggerAura(triggerConfig)

		if isEnchant {
			character.ItemSwap.RegisterEnchantProc(effectID, triggerAura)
		} else {
			character.ItemSwap.RegisterProc(effectID, triggerAura)
		}
	})
}

type SpellRankConfig struct {
	Rank             int32
	SpellID          int32
	Cost             int32
	MinDamage        float64
	MaxDamage        float64
	DotTickDamage    float64
	Coefficient      float64
	ThreatMultiplier float64
	FlatThreatBonus  float64
	CastTimeSeconds  float64 // Optional: specify only if overriding default cast time
}

type SpellRankMap []SpellRankConfig
type SpellRankFactory func(config SpellRankConfig)

func (spell SpellRankConfig) GetRankLabel() string {
	return fmt.Sprintf("Rank %d", spell.Rank)
}

func (ranks SpellRankMap) RegisterAll(factory SpellRankFactory) {
	for _, rankConfig := range ranks {
		factory(rankConfig)
	}
}
