package shared

import (
	"fmt"
	"math"
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

	// What adds a stack while a stacking trinket's window is open. Derived from the container
	// spell's own proc flags, which are not the ones that open the window.
	// For example: Blackened Naaru Sliver opens on a melee hit
	// and then stacks on every attack for the next 20s.
	StackCallback core.AuraCallback
	StackProcMask core.ProcMask
	StackOutcome  core.HitOutcome

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
			// Set only for the stacking trinkets, where the trigger opens a window aura that
			// accumulates a separate stat aura. The handler activates the window rather than the
			// stat aura, so a re-proc restarts the window instead of refreshing a duration the
			// game does not refresh when a stack lands.
			var windowAura *core.Aura
			if stackingAura := effect.StackingAura; stackingAura != nil {
				procAura, windowAura = character.NewTemporaryStatBuffWithStacks(core.TemporaryStatBuffWithStacksConfig{
					AuraLabel:            config.Name + " Proc",
					ActionID:             procAction,
					Duration:             time.Millisecond * time.Duration(effect.EffectDurationMs),
					MaxStacks:            stackingAura.MaxCumulativeStacks,
					BonusPerStack:        stats.FromProtoMap(stackingAura.ScalingOptions[int32(0)].Stats),
					StackingAuraActionID: core.ActionID{SpellID: stackingAura.BuffId},
					StackingAuraLabel:    config.Name + " Stacks",
					TimePerStack:         time.Millisecond * time.Duration(effect.GetStackPeriodMs()),
					TickImmediately:      true,
					StacksFromEvent:      effect.GetStackProc() != nil,
				})
			} else if effect.MaxCumulativeStacks > 0 {
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
					if windowAura != nil {
						windowAura.Activate(sim)
					} else {
						procAura.Activate(sim)
						if effect.MaxCumulativeStacks > 0 {
							procAura.AddStack(sim)
						}
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

			// Event-driven stacks come from their own trigger: the container's proc flags decide
			// what counts, and it only does anything while the window is open. A timer-driven
			// stacking aura fills itself and needs none of this.
			if stackProc := effect.GetStackProc(); stackProc != nil && windowAura != nil && config.StackCallback != core.CallbackEmpty {
				// Attached to the window rather than registered as its own aura: it is then only
				// live while the window is open, needs no active check, and cannot outlive the
				// item the way a permanent trigger would across an item swap.
				stackingAura := procAura
				windowAura.AttachProcTriggerCallback(&character.Unit, core.ProcTrigger{
					Name:       config.Name + " Stack Trigger",
					Callback:   config.StackCallback,
					ProcMask:   config.StackProcMask,
					Outcome:    config.StackOutcome,
					ProcChance: stackProc.GetProcChance(),
					DPM:        stackTriggerDPM(character, stackProc, config.StackProcMask),
					ICD:        time.Millisecond * time.Duration(stackProc.IcdMs),
					Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
						if !stackingAura.IsActive() {
							return
						}
						stackingAura.AddStack(sim)
					},
				})
			}

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

			// A database-resolved stacking trinket keeps the window and the stacks in separate
			// auras, so the stacks, the per-stack stats and the stat aura's identity all come
			// from the nested aura rather than from the effect itself. The window is then always
			// what bounds it, whatever the config asked for.
			statAuraID := auraID
			maxStacks := itemEffect.MaxCumulativeStacks
			perStack := itemEffect.ScalingOptions[int32(0)].Stats
			windowBounded := config.TrinketLimitsDuration
			if stackingAura := itemEffect.StackingAura; stackingAura != nil {
				statAuraID = core.ActionID{SpellID: stackingAura.BuffId}
				maxStacks = stackingAura.MaxCumulativeStacks
				perStack = stackingAura.ScalingOptions[int32(0)].Stats
				windowBounded = true
			}

			duration := core.TernaryDuration(windowBounded, core.NeverExpires, auraDuration)
			statAura := core.MakeStackingAura(character, core.StackingStatAura{
				Aura: core.Aura{
					Label:     config.Name + " Proc",
					ActionID:  statAuraID,
					Duration:  duration,
					MaxStacks: maxStacks,
				},
				BonusPerStack: stats.FromProtoMap(perStack),
			})

			// If trinket limits duration create a separate proc aura
			var procAura *core.Aura = statAura.Aura
			if windowBounded {
				procAura = character.RegisterAura(core.Aura{
					Label:    fmt.Sprintf("%s Limit Aura %s", config.Name, itemEffect.BuffName),
					ActionID: auraID,
					Duration: auraDuration,
					OnExpire: func(_ *core.Aura, sim *core.Simulation) {
						statAura.Deactivate(sim)
					},
				})
			}

			var stackDPM *core.DynamicProcManager
			if stackProc := itemEffect.GetStackProc(); stackProc != nil && stackProc.GetPpm() > 0 {
				stackDPM = character.NewLegacyPPMManager(stackProc.GetPpm(), config.ProcMask)
			}

			procAura.AttachProcTriggerCallback(&character.Unit, core.ProcTrigger{
				Name:               config.Name,
				Callback:           config.Callback,
				ProcMask:           config.ProcMask,
				SpellFlags:         config.SpellFlags,
				Outcome:            config.Outcome,
				RequireDamageDealt: config.RequireDamageDealt,
				ProcChance:         core.TernaryFloat64(stackDPM == nil, config.ProcChance, 0),
				DPM:                stackDPM,
				Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					if !statAura.IsActive() {
						return
					}

					if itemEffect.StacksDecay {
						statAura.RemoveStack(sim)
					} else {
						statAura.AddStack(sim)
					}
				},
			})

			// Only share a cooldown when the effect says it belongs to a category,
			// the same rule NewSimpleStatActive applies.
			var sharedCD core.Cooldown
			if onUse := itemEffect.GetOnUse(); onUse != nil && onUse.CategoryId > 0 {
				sharedCDDuration := time.Millisecond * time.Duration(onUse.CategoryCooldownMs)
				if sharedCDDuration <= 0 {
					sharedCDDuration = time.Millisecond * time.Duration(itemEffect.EffectDurationMs)
				}
				sharedCD = core.Cooldown{
					Timer:    character.GetOrInitSpellCategoryTimer(onUse.CategoryId),
					Duration: sharedCDDuration,
				}
			}

			spell := character.RegisterSpell(core.SpellConfig{
				ActionID: core.ActionID{ItemID: config.ID},
				Flags:    core.SpellFlagNoOnCastComplete,

				Cast: core.CastConfig{
					CD: core.Cooldown{
						Timer:    character.NewTimer(),
						Duration: config.CD,
					},
					SharedCD: sharedCD,
				},

				ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
					statAura.Activate(sim)
					if procAura != statAura.Aura {
						procAura.Activate(sim)
					}
					if itemEffect.StacksDecay {
						statAura.SetStacks(sim, maxStacks)
					}
				},

				RelatedSelfBuff: statAura.Aura,
			})

			character.AddMajorCooldown(core.MajorCooldown{
				Spell:    spell,
				Type:     core.CooldownTypeDPS,
				BuffAura: statAura,
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

		// Per-character copy. config is captured once at registration and this body runs for
		// every character the effect applies to, so filling the trigger in place would hand
		// the second character the first one's DPM - a proc manager bound to another unit,
		// carrying its proc timing.
		triggerConfig := config.Trigger

		if core.ActionID.IsEmptyAction(triggerConfig.ActionID) {
			triggerConfig.ActionID = triggerActionID
		}

		if config.TriggerDPM != nil {
			triggerConfig.DPM = config.TriggerDPM(character)
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

		triggerConfig.TriggerImmediately = true
		triggerConfig.Handler = func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			// Land the extra damage on whatever was hit, not on the primary target.
			target := character.CurrentTarget
			if result != nil && result.Target != nil {
				target = result.Target
			}
			damageSpell.Cast(sim, target)
		}
		triggerAura := character.MakeProcTriggerAura(triggerConfig)

		if isEnchant {
			character.ItemSwap.RegisterEnchantProc(effectID, triggerAura)
		} else {
			character.ItemSwap.RegisterProc(effectID, triggerAura)
		}
	})
}

// Takes in the SpellResult for the triggering spell, and returns the total damage
// of a *fresh* Ignite triggered by that spell. Roll-over damage
// calculations for existing Ignites are handled internally.
type IgniteDamageCalculator func(result *core.SpellResult) float64

type IgniteConfig struct {
	ActionID           core.ActionID
	ClassSpellMask     int64
	SpellSchool        core.SpellSchool
	DisableCastMetrics bool
	DotAuraLabel       string
	DotAuraTag         string
	ProcTrigger        core.ProcTrigger // Ignores the Handler field and creates a custom one, but uses all others.
	DamageCalculator   IgniteDamageCalculator
	IncludeAuraDelay   bool // "munching" and "free roll-over" interactions
	NumberOfTicks      int32
	TickLength         time.Duration
	ParentAura         *core.Aura
}

func RegisterIgniteEffect(unit *core.Unit, config IgniteConfig) *core.Spell {
	spellFlags := core.SpellFlagIgnoreModifiers | core.SpellFlagNoSpellMods | core.SpellFlagNoOnCastComplete

	if config.DisableCastMetrics {
		spellFlags |= core.SpellFlagPassiveSpell
	}

	if config.SpellSchool == 0 {
		config.SpellSchool = core.SpellSchoolFire
	}

	if config.NumberOfTicks == 0 {
		config.NumberOfTicks = 2
	}

	if config.TickLength == 0 {
		config.TickLength = time.Second * 2
	}

	igniteSpell := unit.RegisterSpell(core.SpellConfig{
		ActionID:         config.ActionID,
		SpellSchool:      config.SpellSchool,
		ProcMask:         core.ProcMaskSpellProc,
		ClassSpellMask:   config.ClassSpellMask,
		Flags:            spellFlags,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:     config.DotAuraLabel,
				Tag:       config.DotAuraTag,
				MaxStacks: math.MaxInt32,
			},

			NumberOfTicks:       config.NumberOfTicks,
			TickLength:          config.TickLength,
			AffectedByCastSpeed: false,

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Spell.CalcAndDealPeriodicDamage(sim, target, dot.SnapshotBaseDamage, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Apply(sim)
		},
	})

	refreshIgnite := func(sim *core.Simulation, target *core.Unit, damagePerTick float64) {
		// Cata Ignite
		// 1st ignite application = 4s, split into 2 ticks (2s, 0s)
		// Ignite refreshes: Duration = 4s + MODULO(remaining duration, 2), max 6s. Split damage over 3 ticks at 4s, 2s, 0s.
		dot := igniteSpell.Dot(target)
		dot.SnapshotBaseDamage = damagePerTick
		igniteSpell.Cast(sim, target)
		dot.Aura.SetStacks(sim, int32(dot.SnapshotBaseDamage))
	}

	var scheduledRefresh *core.PendingAction
	procTrigger := config.ProcTrigger
	procTrigger.TriggerImmediately = true
	procTrigger.Handler = func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		target := result.Target
		dot := igniteSpell.Dot(target)
		outstandingDamage := dot.OutstandingDmg()
		newDamage := config.DamageCalculator(result)
		totalDamage := outstandingDamage + newDamage
		newTickCount := dot.BaseTickCount + core.TernaryInt32(dot.IsActive(), 1, 0)
		damagePerTick := totalDamage / float64(newTickCount)

		if config.IncludeAuraDelay {
			// Rough 2-bucket model for the aura update delay distribution based
			// on PTR measurements. Most updates occur on either the same or very
			// next spell batch after the proc, and can therefore be modeled by a
			// 0-10 ms random draw. But a reasonable minority fraction take ~10x
			// longer than this to fire. The origin of these longer delays is
			// likely not actually random in reality, but can be treated that way
			// in practice since the player cannot play around them.
			var delaySeconds float64

			if sim.Proc(0.75, "Aura Delay") {
				delaySeconds = 0.010 * sim.RandomFloat("Aura Delay")
			} else {
				delaySeconds = 0.090 + 0.020*sim.RandomFloat("Aura Delay")
			}

			applyDotAt := sim.CurrentTime + core.DurationFromSeconds(delaySeconds)

			// Cancel any prior aura updates already in the queue
			if (scheduledRefresh != nil) && (scheduledRefresh.NextActionAt > sim.CurrentTime) {
				scheduledRefresh.Cancel(sim)

				if sim.Log != nil {
					unit.Log(sim, "Previous %s proc was munched due to server aura delay", config.DotAuraLabel)
				}
			}

			// Schedule a delayed refresh of the DoT with cached damagePerTick value (allowing for "free roll-overs")
			if sim.Log != nil {
				unit.Log(sim, "Schedule travel (%0.1f ms) for %s", delaySeconds*1000, config.DotAuraLabel)

				if dot.IsActive() && (dot.NextTickAt() < applyDotAt) {
					unit.Log(sim, "%s rolled with %0.3f damage both ticking and rolled into next", config.DotAuraLabel, outstandingDamage)
				}
			}

			scheduledRefresh = core.NewDelayedAction(core.DelayedActionOptions{
				DoAt:     applyDotAt,
				Priority: core.ActionPriorityDOT,

				OnAction: func(_ *core.Simulation) {
					refreshIgnite(sim, target, damagePerTick)
				},
			})

			sim.AddPendingAction(scheduledRefresh)
		} else {
			refreshIgnite(sim, target, damagePerTick)
		}
	}

	if config.ParentAura != nil {
		config.ParentAura.AttachProcTrigger(procTrigger)
	} else {
		unit.MakeProcTriggerAura(procTrigger)
	}

	return igniteSpell
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

// A stack rate given as PPM needs a proc manager rather than a flat chance; a chance-based rate
// needs none.
func stackTriggerDPM(character *core.Character, stackProc *proto.ProcEffect, mask core.ProcMask) *core.DynamicProcManager {
	if stackProc.GetPpm() <= 0 {
		return nil
	}
	return character.NewLegacyPPMManager(stackProc.GetPpm(), mask)
}
