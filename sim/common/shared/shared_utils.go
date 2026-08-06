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
	// Set when the client bars the damage spell from critting, which the sim has no way to know:
	// it is a spell attribute, so only the database generator can see it.
	CannotCrit bool
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

// Whether a proc's damage is dealt as a melee hit. The school decides it: anything non-physical
// rolls against the spell tables and takes the spell crit multiplier. IsMelee stays honoured on top
// of that for a caller that means physical damage without saying so through the school.
func isMeleeDamage(school core.SpellSchool, isMelee bool) bool {
	return isMelee || school.Matches(core.SpellSchoolPhysical)
}

func damageCritMultiplier(character *core.Character, school core.SpellSchool, isMelee bool) float64 {
	return core.TernaryFloat64(isMeleeDamage(school, isMelee), character.DefaultMeleeCritMultiplier(), character.DefaultSpellCritMultiplier())
}

// The outcome a proc's damage rolls when the caller states none. Physical damage goes through the
// melee table, everything else through the spell one, and a spell the client bars from critting
// takes the no-crit variant of whichever it rolls against.
func damageOutcome(school core.SpellSchool, isMelee bool, cannotCrit bool, outcome OutcomeType) OutcomeType {
	if outcome != OutcomeDefault {
		return outcome
	}

	if isMeleeDamage(school, isMelee) {
		if cannotCrit {
			return OutcomeMeleeNoCrit
		}

		return OutcomeMeleeCanCrit
	}

	if cannotCrit {
		return OutcomeSpellNoCrit
	}

	return OutcomeSpellCanCrit
}

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
			CritMultiplier:           damageCritMultiplier(character, damage.School, damage.IsMelee),
			DamageMultiplierAdditive: 1,
			ThreatMultiplier:         1,
			BonusCoefficient:         damage.BonusCoefficient,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, sim.Roll(damage.MinDmg, damage.MaxDmg), GetOutcome(spell, damageOutcome(damage.School, damage.IsMelee, damage.CannotCrit, damage.Outcome)))
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
	// Ignore empty dummy implementations
	if config.Callback == core.CallbackEmpty {
		return
	}

	source := config.effectSource()

	// Soft fail to allow for overrides for bad effects
	if source.isAlreadyImplemented() {
		return
	}

	triggerActionID := source.actionID()

	source.registerEffect(func(agent core.Agent) {
		character := agent.GetCharacter()
		eligibleSlots := source.eligibleSlots(character)

		procEffects := source.procEffects()
		if len(procEffects) == 0 {
			panic(fmt.Sprintf("Error getting proc effects for item/enchant %v", source.id))
		}

		for _, effect := range procEffects {
			proc := effect.GetProc()

			// windowAura is set only for the stacking trinkets, where the trigger opens a window
			// that accumulates a separate stat aura. The handler then activates the window rather
			// than the stat aura, so a re-proc restarts the window instead of refreshing a duration
			// the game does not refresh when a stack lands.
			procAura, windowAura := buildProcAura(character, config, effect)

			dpm := procDPM(character, config, source, proc)

			procAura.CustomProcCondition = config.CustomProcCondition

			var procSpell ExtraSpellInfo
			if extraSpell != nil {
				procSpell = extraSpell(agent)
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
				Handler:            procHandler(config, effect, procAura, windowAura, procSpell),
			})

			attachStackTrigger(character, config, effect, procAura, windowAura)

			// Carried on the stacking path too. Nothing in the stacking machinery reads this -
			// CanProc consults only IsSwapped and CustomProcCondition - so it gates nothing. What
			// it feeds is GetMatchingItemProcAuras, which drops any aura whose Icd is nil, and with
			// it every ICD-aware APL value. Hand-written trinkets all set it, so leaving it nil
			// would hide the generated ones from those APLs.
			if proc.IcdMs != 0 {
				procAura.Icd = triggerAura.Icd
			}

			source.registerProc(character, triggerAura, eligibleSlots)
			source.registerWeaponEnchantBuff(character, procAura)
			character.AddStatProcBuff(source.id, procAura, source.isEnchant, eligibleSlots)
		}
	})
}

// What the proc does when it fires. When a custom condition refuses, the ICD is rolled back so the
// next opportunity still counts instead of the effect being locked out by a proc that never
// happened.
func procHandler(config ProcStatBonusEffect, effect *proto.ItemEffect, procAura *core.StatBuffAura, windowAura *core.Aura, procSpell ExtraSpellInfo) func(*core.Simulation, *core.Spell, *core.SpellResult) {
	return func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		// A custom condition gates the body rather than replacing it. Written as its own branch it
		// skipped the window, the stack accumulation and the extra spell, so any override set on an
		// item whose database entry resolves a stacking aura would open nothing.
		if config.CustomProcCondition != nil && !procAura.CanProc(sim) {
			if procAura.Icd != nil && procAura.Icd.Duration != 0 {
				procAura.Icd.Reset()
			}

			return
		}

		// Activating the window and not the stat aura is what makes a re-proc restart the window
		// instead of refreshing stacks the game would not refresh.
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

// The three shapes a proc buff comes in. Only the first returns a second aura: there the trigger
// opens a window and the stat aura inside it accumulates, so the caller has two things to wire.
func buildProcAura(character *core.Character, config ProcStatBonusEffect, effect *proto.ItemEffect) (*core.StatBuffAura, *core.Aura) {
	label := config.Name + " Proc"
	action := core.ActionID{SpellID: effect.BuffId}
	duration := time.Millisecond * time.Duration(effect.EffectDurationMs)

	if stackingAura := effect.StackingAura; stackingAura != nil {
		return character.NewTemporaryStatBuffWithStacks(core.TemporaryStatBuffWithStacksConfig{
			AuraLabel:            label,
			ActionID:             action,
			Duration:             duration,
			MaxStacks:            stackingAura.MaxCumulativeStacks,
			BonusPerStack:        stats.FromProtoMap(stackingAura.GetScalingOptions()[int32(0)].GetStats()),
			StackingAuraActionID: core.ActionID{SpellID: stackingAura.BuffId},
			StackingAuraLabel:    config.Name + " Stacks",
			TimePerStack:         time.Millisecond * time.Duration(effect.GetStackPeriodMs()),
			TickImmediately:      true,
			StacksFromEvent:      effect.GetStackProc() != nil,
		})
	}

	if effect.MaxCumulativeStacks > 0 {
		return core.MakeStackingAura(character, core.StackingStatAura{
			Aura: core.Aura{
				Label:     label,
				ActionID:  action,
				Duration:  duration,
				MaxStacks: effect.MaxCumulativeStacks,
			},
			BonusPerStack: stats.FromProtoMap(effect.GetScalingOptions()[int32(0)].GetStats()),
		}), nil
	}

	return character.NewTemporaryStatsAura(label, action, stats.FromProtoMap(effect.GetScalingOptions()[int32(0)].GetStats()), duration), nil
}

func procDPM(character *core.Character, config ProcStatBonusEffect, source effectSource, proc *proto.ProcEffect) *core.DynamicProcManager {
	if proc.GetPpm() <= 0 {
		return nil
	}

	if config.ProcMask != core.ProcMaskUnknown {
		return character.NewLegacyPPMManager(proc.GetPpm(), config.ProcMask)
	}

	// With no mask of its own the rate has to be read off whatever the effect sits on.
	if source.isEnchant {
		return character.NewDynamicLegacyProcForEnchant(source.id, proc.GetPpm(), 0)
	}

	return character.NewDynamicLegacyProcForWeapon(source.id, proc.GetPpm(), 0)
}

// Event-driven stacks come from their own trigger: the container's proc flags decide what counts,
// and it only does anything while the window is open. A timer-driven stacking aura fills itself and
// needs none of this, so this is a no-op for everything else.
func attachStackTrigger(character *core.Character, config ProcStatBonusEffect, effect *proto.ItemEffect, statAura *core.StatBuffAura, windowAura *core.Aura) {
	stackProc := effect.GetStackProc()
	if stackProc == nil || windowAura == nil || config.StackCallback == core.CallbackEmpty {
		return
	}

	// Attached to the window rather than registered as its own aura: it is then only live while
	// the window is open, needs no active check, and cannot outlive the item the way a permanent
	// trigger would across an item swap.
	windowAura.AttachProcTriggerCallback(&character.Unit, core.ProcTrigger{
		Name:       config.Name + " Stack Trigger",
		Callback:   config.StackCallback,
		ProcMask:   config.StackProcMask,
		Outcome:    config.StackOutcome,
		ProcChance: stackProc.GetProcChance(),
		DPM:        stackTriggerDPM(character, stackProc, config.StackProcMask),
		ICD:        time.Millisecond * time.Duration(stackProc.IcdMs),
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			if !statAura.IsActive() {
				return
			}
			statAura.AddStack(sim)
		},
	})
}

// Registers the same effect once per item that carries it. Only the highest ID is added to the
// test suite, so that a dozen re-issues of one trinket do not each get their own fixture entry.
func forEachVariant(config ProcStatBonusEffect, variants []ItemVariant, register func(config ProcStatBonusEffect)) {
	var maxItemID int32
	for _, variant := range variants {
		maxItemID = max(maxItemID, variant.ItemID)
	}

	for _, variant := range variants {
		config.Name = variant.ItemName
		config.ItemID = variant.ItemID
		core.AddEffectsToTest = (config.ItemID == maxItemID)
		register(config)
	}

	core.AddEffectsToTest = true
}

func NewProcStatBonusEffectWithVariants(config ProcStatBonusEffect, variants []ItemVariant) {
	forEachVariant(config, variants, NewProcStatBonusEffect)
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
		character := agent.GetCharacter()

		onUseEffects := core.FilterSlice(itemEffectsFor(itemID), func(effect *proto.ItemEffect) bool {
			return effect.GetOnUse() != nil
		})
		if len(onUseEffects) == 0 {
			panic(fmt.Sprintf("No active effects found for item with ID: %d!", itemID))
		}

		for _, itemEffect := range onUseEffects {
			spellConfig := core.SpellConfig{
				ActionID: core.ActionID{ItemID: itemID},
			}
			spellConfig.Cast.CD = core.Cooldown{
				Timer:    character.NewTimer(),
				Duration: time.Duration(itemEffect.GetOnUse().CooldownMs) * time.Millisecond,
			}
			spellConfig.Cast.SharedCD = sharedCooldown(character, itemEffect)

			core.RegisterTemporaryStatsOnUseCD(character, itemEffect.BuffName, stats.FromProtoMap(itemEffect.GetScalingOptions()[int32(0)].GetStats()), time.Millisecond*time.Duration(itemEffect.EffectDurationMs), spellConfig)
		}
	})
}

type StackingStatBonusCD struct {
	Name               string
	ID                 int32
	CD                 time.Duration
	Callback           core.AuraCallback
	ProcMask           core.ProcMask
	SpellFlags         core.SpellFlag
	Outcome            core.HitOutcome
	RequireDamageDealt bool

	// The stacks will only be granted as long as the trinket is active
	TrinketLimitsDuration bool
}

// Where the stacks actually live. A database-resolved stacking trinket keeps the window and the
// stacks in two auras, so the count, the per-stack stats and the stat aura's identity all come from
// the nested one rather than from the effect itself, and the window is then always what bounds it -
// whatever the config asked for. A flat trinket states all of it on the effect.
type stackingStats struct {
	actionID      core.ActionID
	maxStacks     int32
	perStack      map[int32]float64
	windowBounded bool
}

func resolveStackingStats(effect *proto.ItemEffect, effectActionID core.ActionID, trinketLimitsDuration bool) stackingStats {
	if stackingAura := effect.StackingAura; stackingAura != nil {
		return stackingStats{
			actionID:      core.ActionID{SpellID: stackingAura.BuffId},
			maxStacks:     stackingAura.MaxCumulativeStacks,
			perStack:      stackingAura.GetScalingOptions()[int32(0)].GetStats(),
			windowBounded: true,
		}
	}

	return stackingStats{
		actionID:      effectActionID,
		maxStacks:     effect.MaxCumulativeStacks,
		perStack:      effect.GetScalingOptions()[int32(0)].GetStats(),
		windowBounded: trinketLimitsDuration,
	}
}

// The aura the on-use itself applies. Effects that name no buff spell fall back to the item.
func stackingAuraID(effect *proto.ItemEffect, itemID int32) core.ActionID {
	if auraID := (core.ActionID{SpellID: effect.BuffId}); !auraID.IsEmptyAction() {
		return auraID
	}

	return core.ActionID{ItemID: itemID}
}

// The aura pair a stacking on-use drives. Where the window bounds the stacks the stat aura is given
// no duration of its own and a second aura ends it on expiry; otherwise the stat aura is its own
// window and both returns are the same object, which is what the caller's identity check keys off.
func buildStackingCDAuras(character *core.Character, config StackingStatBonusCD, effect *proto.ItemEffect, stacks stackingStats) (*core.StatBuffAura, *core.Aura) {
	auraDuration := time.Millisecond * time.Duration(effect.EffectDurationMs)

	statAura := core.MakeStackingAura(character, core.StackingStatAura{
		Aura: core.Aura{
			Label:     config.Name + " Proc",
			ActionID:  stacks.actionID,
			Duration:  core.TernaryDuration(stacks.windowBounded, core.NeverExpires, auraDuration),
			MaxStacks: stacks.maxStacks,
		},
		BonusPerStack: stats.FromProtoMap(stacks.perStack),
	})

	if !stacks.windowBounded {
		return statAura, statAura.Aura
	}

	return statAura, character.RegisterAura(core.Aura{
		Label:    fmt.Sprintf("%s Limit Aura %s", config.Name, effect.BuffName),
		ActionID: stackingAuraID(effect, config.ID),
		Duration: auraDuration,
		OnExpire: func(_ *core.Aura, sim *core.Simulation) {
			statAura.Deactivate(sim)
		},
	})
}

// What moves the stack count while the window is open. Attached to the window so it is live only
// then, and a decaying trinket spends a stack per event where the rest gain one.
func attachStackingCDTrigger(character *core.Character, config StackingStatBonusCD, effect *proto.ItemEffect, statAura *core.StatBuffAura, windowAura *core.Aura) {
	// Rate and lockout come off the stack proc, the same way the proc-item sibling reads them. Taken
	// from the config instead they were always zero - the generated on-use call states neither - and
	// a zero chance is normalised to 1, so a stack proc with a real chance or an ICD would have
	// stacked on every qualifying event with no cooldown. The getters are nil-safe: an item effect
	// that carries no stack proc keeps a plain always-on trigger.
	stackProc := effect.GetStackProc()

	var stackDPM *core.DynamicProcManager
	if stackProc != nil {
		stackDPM = stackTriggerDPM(character, stackProc, config.ProcMask)
	}

	windowAura.AttachProcTriggerCallback(&character.Unit, core.ProcTrigger{
		Name:               config.Name,
		Callback:           config.Callback,
		ProcMask:           config.ProcMask,
		SpellFlags:         config.SpellFlags,
		Outcome:            config.Outcome,
		RequireDamageDealt: config.RequireDamageDealt,
		ProcChance:         core.TernaryFloat64(stackDPM == nil, stackProc.GetProcChance(), 0),
		ICD:                time.Millisecond * time.Duration(stackProc.GetIcdMs()),
		DPM:                stackDPM,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			if !statAura.IsActive() {
				return
			}

			if effect.StacksDecay {
				statAura.RemoveStack(sim)
			} else {
				statAura.AddStack(sim)
			}
		},
	})
}

// Creates a new stacking stats bonus aura based on the configuration. If Bonus is not given, the ItemEffect of the item will be used
// to determine the correct values.
func NewStackingStatBonusCD(config StackingStatBonusCD) {
	core.NewItemEffect(config.ID, func(agent core.Agent) {
		character := agent.GetCharacter()
		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(config.ID)

		for _, itemEffect := range itemEffectsFor(config.ID) {
			stacks := resolveStackingStats(itemEffect, stackingAuraID(itemEffect, config.ID), config.TrinketLimitsDuration)
			statAura, procAura := buildStackingCDAuras(character, config, itemEffect, stacks)

			attachStackingCDTrigger(character, config, itemEffect, statAura, procAura)

			spell := character.RegisterSpell(core.SpellConfig{
				ActionID: core.ActionID{ItemID: config.ID},
				Flags:    core.SpellFlagNoOnCastComplete,

				Cast: core.CastConfig{
					CD: core.Cooldown{
						Timer:    character.NewTimer(),
						Duration: config.CD,
					},
					SharedCD: sharedCooldown(character, itemEffect),
				},

				ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
					statAura.Activate(sim)
					if procAura != statAura.Aura {
						procAura.Activate(sim)
					}
					if itemEffect.StacksDecay {
						statAura.SetStacks(sim, stacks.maxStacks)
					}
				},

				RelatedSelfBuff: statAura.Aura,
			})

			character.AddMajorCooldown(core.MajorCooldown{
				Spell:    spell,
				Type:     core.CooldownTypeDPS,
				BuffAura: statAura,
			})

			// The stat aura and not the window: the stacks are what an APL keys off. This is the
			// registry behind the "Item Stat Proc Check" value and the "Activate All Stat Buff Proc
			// Auras" action, so without it an APL asking after Insight of the Qiraji stacks matches
			// nothing. Registering the major cooldown alone leaves plain use-on-cooldown output
			// unchanged, which is why no fixture records the difference.
			character.AddStatProcBuff(config.ID, statAura, false, eligibleSlots)
		}
	})
}

func NewStackingStatBonusEffectWithVariants(config ProcStatBonusEffect, variants []ItemVariant) {
	forEachVariant(config, variants, func(config ProcStatBonusEffect) {
		factory_StatBonusEffect(config, nil)
	})
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
	// Set when the client bars the damage spell from critting, which the sim has no way to know:
	// it is a spell attribute, so only the database generator can see it.
	CannotCrit bool
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

		critMultiplier := damageCritMultiplier(character, config.School, config.IsMelee)

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
				spell.CalcAndDealDamage(sim, target, sim.Roll(minDmg, maxDmg), GetOutcome(spell, damageOutcome(config.School, config.IsMelee, config.CannotCrit, config.Outcome)))
			},
		})

		triggerConfig.TriggerImmediately = true

		// What result.Target means depends on the callback, so the callback has to decide whether it
		// may be read at all.
		callback := triggerConfig.Callback

		triggerConfig.Handler = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			target := character.CurrentTarget

			switch {
			case callback.Matches(core.CallbackOnSpellHitTaken):
				// Here result.Target is the wearer - core dispatches hit-taken through
				// result.Target.OnSpellHitTaken - so the retaliation goes to the attacker instead of
				// into the wearer's own health. This is the shield spike shape.
				if spell != nil && spell.Unit != nil {
					target = spell.Unit
				}

			case callback.Matches(core.CallbackOnSpellHitDealt | core.CallbackOnPeriodicDamageDealt):
				// Land the extra damage on whatever was hit, not on the primary target.
				if result != nil && result.Target != nil {
					target = result.Target
				}

			default:
				// The heal callbacks, cast complete and apply effects carry either no result or one
				// whose target is an ally, so nothing there can name what to damage and the current
				// target stands. Reading result.Target regardless is what would have a heal-triggered
				// damage proc hit the healed ally.
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
// needs none. The mask has to be a concrete one: a PPM manager built on ProcMaskUnknown matches
// nothing and would silently never proc. Every stack trigger the generator emits carries the mask
// it derived from the container's own proc flags, so that does not arise today - unlike procDPM,
// which handles the unknown case because a weapon or enchant proc can legitimately lack a mask.
func stackTriggerDPM(character *core.Character, stackProc *proto.ProcEffect, mask core.ProcMask) *core.DynamicProcManager {
	if stackProc.GetPpm() <= 0 {
		return nil
	}

	return character.NewLegacyPPMManager(stackProc.GetPpm(), mask)
}

///////////////////////////////////////////////////////////////////////////
//							Item and enchant plumbing
///////////////////////////////////////////////////////////////////////////

// Which of the two registries an effect belongs to, item or enchant, and its ID within it. Those
// are the only things the two differ in; every helper above treats them identically. Resolving it
// once keeps the same isEnchant branch from being written out at each of the six places that would
// otherwise need it.
type effectSource struct {
	id        int32
	isEnchant bool
}

func (config ProcStatBonusEffect) effectSource() effectSource {
	if config.EnchantID != 0 {
		return effectSource{id: config.EnchantID, isEnchant: true}
	}

	return effectSource{id: config.ItemID}
}

func (s effectSource) registerEffect(apply core.ApplyEffect) {
	if s.isEnchant {
		core.NewEnchantEffect(s.id, apply)
	} else {
		core.NewItemEffect(s.id, apply)
	}
}

// Whether a hand-written effect already covers this. That is the soft fail letting an override win
// over the generated registration, and it is why deleting one hands the generated version back.
func (s effectSource) isAlreadyImplemented() bool {
	if s.isEnchant {
		return core.HasEnchantEffect(s.id)
	}

	return core.HasItemEffect(s.id)
}

func (s effectSource) actionID() core.ActionID {
	if s.isEnchant {
		return core.ActionID{SpellID: s.id}
	}

	return core.ActionID{ItemID: s.id}
}

func (s effectSource) eligibleSlots(character *core.Character) []proto.ItemSlot {
	if s.isEnchant {
		return character.ItemSwap.EligibleSlotsForEffect(s.id)
	}

	return character.ItemSwap.EligibleSlotsForItem(s.id)
}

// The proc-carrying effects this item or enchant declares, keyed by the aura each one applies.
func (s effectSource) procEffects() map[int32]*proto.ItemEffect {
	var declared []*proto.ItemEffect
	if s.isEnchant {
		declared = core.GetEnchantByEffectID(s.id).EnchantEffects
	} else if item := core.GetItemByID(s.id); item != nil {
		declared = item.ItemEffects
	}

	procEffects := make(map[int32]*proto.ItemEffect)
	for _, effect := range declared {
		if effect.GetProc() != nil {
			procEffects[effect.BuffId] = effect
		}
	}

	return procEffects
}

// A weapon enchant's buff has to drop when the weapon carrying it is swapped out. AddStatProcBuff
// only flips IsSwapped, which gates the next proc but leaves a running buff up for the rest of its
// duration. Gated on the enchant being a weapon enchant: RegisterWeaponEnchantBuff watches the
// weapon slots, so handing it a cloak or shield enchant would deactivate that buff on any weapon
// swap.
func (s effectSource) registerWeaponEnchantBuff(character *core.Character, procAura *core.StatBuffAura) {
	if !s.isEnchant {
		return
	}

	if ench := core.GetEnchantByEffectID(s.id); ench == nil || ench.Type != proto.ItemType_ItemTypeWeapon {
		return
	}

	character.ItemSwap.RegisterWeaponEnchantBuff(procAura.Aura, s.id)
}

func (s effectSource) registerProc(character *core.Character, triggerAura *core.Aura, slots []proto.ItemSlot) {
	if s.isEnchant {
		character.ItemSwap.RegisterEnchantProcWithSlots(s.id, triggerAura, slots)
	} else {
		character.ItemSwap.RegisterProcWithSlots(s.id, triggerAura, slots)
	}
}

// The effects an on-use helper works from. A generated registration naming an item with no effect
// data is a database bug rather than a runtime case, so both callers fail loudly and identically.
func itemEffectsFor(itemID int32) []*proto.ItemEffect {
	item := core.GetItemByID(itemID)
	if item == nil {
		panic(fmt.Sprintf("No item with ID: %d", itemID))
	}

	if len(item.ItemEffects) == 0 {
		panic(fmt.Sprintf("No effects data for item with ID: %d", itemID))
	}

	return item.ItemEffects
}

// Share a cooldown only when the effect says it belongs to a category. One with no category shares
// nothing, and putting it on a generic trinket timer would gate it against unrelated items.
func sharedCooldown(character *core.Character, effect *proto.ItemEffect) core.Cooldown {
	onUse := effect.GetOnUse()
	if onUse == nil || onUse.CategoryId <= 0 {
		return core.Cooldown{}
	}

	duration := time.Millisecond * time.Duration(onUse.CategoryCooldownMs)
	if duration <= 0 {
		duration = time.Millisecond * time.Duration(effect.EffectDurationMs)
	}

	return core.Cooldown{
		Timer:    character.GetOrInitSpellCategoryTimer(onUse.CategoryId),
		Duration: duration,
	}
}
