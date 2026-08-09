package dbc

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// ItemEffect represents an item effect in the game.
type ItemEffect struct {
	ID                   int // Effect ID
	LegacySlotIndex      int // Legacy slot index
	TriggerType          int // Trigger type
	Charges              int // Number of charges
	CoolDownMSec         int // Cooldown in milliseconds
	CategoryCoolDownMSec int // Category cooldown in milliseconds
	SpellCategoryID      int // Spell category ID
	MaxCumulativeStacks  int // Max cumulative stacks
	SpellID              int // Spell ID
	ChrSpecializationID  int // Character specialization ID
	ParentItemID         int // Parent item ID
}

// ToMap returns a generic representation of the effect.
func (e *ItemEffect) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"ID":                   e.ID,
		"LegacySlotIndex":      e.LegacySlotIndex,
		"TriggerType":          e.TriggerType,
		"Charges":              e.Charges,
		"CoolDownMSec":         e.CoolDownMSec,
		"CategoryCoolDownMSec": e.CategoryCoolDownMSec,
		"MaxCumulativeStacks":  e.MaxCumulativeStacks,
		"SpellCategoryID":      e.SpellCategoryID,
		"SpellID":              e.SpellID,
		"ChrSpecializationID":  e.ChrSpecializationID,
		"ParentItemID":         e.ParentItemID,
	}
}

func GetItemEffect(effectId int) ItemEffect {
	return dbcInstance.ItemEffects[effectId]
}

func makeBaseProto(e *ItemEffect, statsSpellID int) *proto.ItemEffect {
	sp := dbcInstance.Spells[e.SpellID]
	base := &proto.ItemEffect{
		BuffId:           int32(e.SpellID),
		BuffName:         fmt.Sprintf("%s (%d)", sp.NameLang, e.SpellID),
		EffectDurationMs: int32(sp.Duration),
		ScalingOptions:   make(map[int32]*proto.ScalingItemEffectProperties),
	}
	// override duration if stats spell defines its own
	if dur := dbcInstance.Spells[statsSpellID].Duration; dur > 0 {
		base.EffectDurationMs = int32(dur)
	}
	return base
}

// The internal cooldown of a proc. Most spells carry it on the trigger's SpellAuraOptions, but a
// few leave that at zero and put it on the buff's spell category instead: Bulwark of Azzinoth's
// armor buff sits in category 30 with a 60s CategoryRecoveryTime while its trigger has nothing.
//
// Only a buff that is a *separate* spell from the trigger counts. Where the two are the same spell
// the recovery being read is that spell's own cast throttle rather than a gate on re-applying the
// buff, and in TBC that is almost always the generic 1s bucket - category 99 holds 170 spells, 101
// of them with exactly 1000, most of them quest-reward trinkets. Reading those as internal
// cooldowns perturbs every fixture that equips one and states nothing the game does not already do.
func procIcdMs(trigger Spell, buff Spell) int32 {
	if trigger.ProcCategoryRecovery > 0 {
		return trigger.ProcCategoryRecovery
	}

	if buff.ID != trigger.ID {
		return buff.CategoryRecoveryTime
	}

	return 0
}

func assignTrigger(e *ItemEffect, statsSpellID int, pe *proto.ItemEffect) {
	spTop := dbcInstance.Spells[e.SpellID]
	statsSP := dbcInstance.Spells[statsSpellID]

	switch resolveTriggerType(e.TriggerType, e.SpellID) {
	case ITEM_SPELLTRIGGER_ON_USE:
		pe.Effect = &proto.ItemEffect_OnUse{OnUse: &proto.OnUseEffect{
			CooldownMs:         int32(e.CoolDownMSec),
			CategoryId:         int32(e.SpellCategoryID),
			CategoryCooldownMs: int32(e.CategoryCoolDownMSec),
		}}
	case ITEM_SPELLTRIGGER_CHANCE_ON_HIT:
		proc := &proto.ProcEffect{
			IcdMs: procIcdMs(spTop, statsSP),
		}
		// If proc chance is above 100 it is most likely a PPM proc
		// Or if we manually assigned PPM
		ppm := getPPMForItemID(int32(e.ParentItemID))
		if spTop.ProcChance == 0 || spTop.ProcChance > 100 || ppm > 0 {
			if ppm > 0 {
				proc.ProcRate = &proto.ProcEffect_Ppm{
					Ppm: ppm,
				}
			} else {
				ReportMissingPPM(int32(e.ParentItemID), e.SpellID)
			}
		} else {
			proc.ProcRate = &proto.ProcEffect_ProcChance{
				ProcChance: float64(spTop.ProcChance) / 100,
			}
		}
		pe.BuffId = statsSP.ID
		pe.BuffName = fmt.Sprintf("%s (%d)", statsSP.NameLang, e.SpellID)
		pe.Effect = &proto.ItemEffect_Proc{Proc: proc}
	}

	// Stacks are a property of the buff aura, not of how it is triggered, so this sits outside
	// the switch: Pendant of the Violet Eye is an on-use whose buff, 35095 Enlightenment, has
	// CumulativeAura 20, and assigning only on the chance-on-hit path skipped it entirely.
	//
	// The count lives on whichever spell carries the aura, usually the stats spell rather than
	// the trigger - Blackened Naaru Sliver's trigger has none while its buff, 45041 Combat
	// Insight, has 10.
	maxCumulativeStacks := max(statsSP.MaxCumulativeStacks, spTop.MaxCumulativeStacks)
	if maxCumulativeStacks > 0 {
		pe.MaxCumulativeStacks = maxCumulativeStacks
	}
}

// The stats an effect grants. A stacking effect grants nothing through the aura its trigger
// applies - the stats it is named after live on the aura that accumulates - so an empty scaling
// option is not the same as no stats, and no caller may read ScalingOptions directly to answer
// this question.
func EffectStats(effect *proto.ItemEffect) map[int32]float64 {
	if stats := effect.GetScalingOptions()[int32(0)].GetStats(); len(stats) > 0 {
		return stats
	}

	return effect.GetStackingAura().GetScalingOptions()[int32(0)].GetStats()
}

// On a stacking effect the aura the trigger applies grants nothing itself, so its scaling options
// resolve to an empty message. ScalingItemEffectProperties holds only a stats map, so such an entry
// carries no information at all - drop the whole map rather than ship an empty entry per state.
// Readers reach the map through GetStats(), which is nil-safe.
//
// A no-op when anything did resolve, which keeps this honest if an effect ever carries stats on
// both the trigger aura and the aura it accumulates.
func dropEmptyScalingOptions(pe *proto.ItemEffect) {
	for _, opt := range pe.ScalingOptions {
		if len(opt.GetStats()) > 0 {
			return
		}
	}

	pe.ScalingOptions = nil
}

func (e *ItemEffect) ToProto(itemLevel int) (*proto.ItemEffect, bool) {
	statsSpellID := resolveStatsSpell(e.SpellID)

	pe := makeBaseProto(e, statsSpellID)
	assignTrigger(e, statsSpellID, pe)

	pe.ScalingOptions[int32(0)] = buildBaseStatScalingProps(statsSpellID, e.SpellID)

	// The stats may live on the accumulating aura rather than on the one the trigger applies, in
	// which case the effect is real even though the scaling options above resolved to nothing.
	if stacking := buildStackingAura(e.SpellID, statsSpellID, itemLevel, e.ParentItemID); stacking != nil {
		stacking.Aura.ScalingOptions[int32(0)] = buildItemEffectScalingProps(int(stacking.Aura.BuffId), itemLevel)
		applyStackingAura(pe, stacking)
	}

	dropEmptyScalingOptions(pe)

	if len(EffectStats(pe)) == 0 {
		return nil, false
	}

	return pe, true
}

func resolveStatsSpell(spellID int) int {
	return newChainWalker().resolveStatsSpell(spellID)
}

func (w *chainWalker) resolveStatsSpell(spellID int) int {
	effects := w.effects(spellID)
	for _, se := range effects {
		if se.GrantsStats() {
			return spellID
		}
	}

	// If we cant resolve the spell in the first loop, we follow proc triggers downwards
	for _, se := range effects {
		if se.IsProcTrigger() {
			return w.resolveStatsSpell(se.EffectTriggerSpell)
		}
	}
	return spellID
}

func resolveTriggerType(topType, spellID int) int {
	if topType == ITEM_SPELLTRIGGER_ON_USE || topType == ITEM_SPELLTRIGGER_CHANCE_ON_HIT {
		return topType
	}

	// Carrying A_PROC_TRIGGER_SPELL is not on its own a proc. The client also uses that aura to hang
	// a permanent sub-aura off an equip effect: Leggings of Beast Mastery and Void Star Talisman
	// grant pet stats that way, with no proc mask, no chance and no duration, and promoting them
	// emitted a proc with no rate at all. Something has to say when it would fire.
	//
	// The weapons whose rate lives only in MapItemIdToPPM are unaffected: the database types those
	// CHANCE_ON_HIT itself, so they return above without being promoted here.
	sp := dbcInstance.Spells[spellID]
	statesAProc := slices.ContainsFunc(sp.ProcTypeMask, func(bits int) bool { return bits != 0 }) || sp.ProcChance != 0
	if !statesAProc {
		return topType
	}

	for _, se := range dbcInstance.SpellEffectsInOrder(spellID) {
		if se.IsProcTrigger() {
			return ITEM_SPELLTRIGGER_CHANCE_ON_HIT
		}
	}

	return topType
}

func buildItemEffectScalingProps(spellID int, itemLevel int) *proto.ScalingItemEffectProperties {
	return &proto.ScalingItemEffectProperties{Stats: collectStats(spellID, itemLevel).ToProtoMap()}
}

func buildBaseStatScalingProps(spellID int, itemSpellID int) *proto.ScalingItemEffectProperties {
	var total stats.Stats

	// check if spell is procced by a SPELL_WITH_VALUE
	if effects := dbcInstance.SpellEffectsInOrder(itemSpellID); len(effects) > 0 {
		for _, se := range effects {
			// TBC ANNI: Items can have "static" ItemEffects that don't have a duration.
			// We need to parse these into stats just as is done for ItemSparse data.
			stat := ConvertEffectAuraToStatIndex(se.EffectAura, se.EffectMiscValues[0])
			if stat >= 0 || stat == -2 {
				value := float64(se.EffectBasePoints + 1)
				// Make sure it's not Feral AP
				if strings.Contains(dbcInstance.Spells[se.SpellID].Description, "forms only") {
					stat = proto.Stat_StatFeralAttackPower
				}
				if stat == proto.Stat_StatArmorPenetration || stat == proto.Stat_StatSpellPenetration {
					// Make these not negative
					value = math.Abs(value)
				}

				if se.EffectAura == A_MOD_RESISTANCE && stat == -2 {
					// All Resists
					total[stats.ArcaneResistance] += value
					total[stats.FireResistance] += value
					total[stats.FrostResistance] += value
					total[stats.NatureResistance] += value
					total[stats.ShadowResistance] += value
				} else {
					total[int32(stat)] += value
				}

				continue
			}

			if se.EffectAura == A_PROC_TRIGGER_SPELL_WITH_VALUE && spellID == se.EffectTriggerSpell {
				for idx := range total {
					if total[idx] == 0 {
						continue
					}

					total[idx] = float64(se.EffectBasePoints)
				}
			}
		}
	}

	return &proto.ScalingItemEffectProperties{Stats: total.ToProtoMap()}
}

func collectStats(spellID, itemLevel int) stats.Stats {
	var total stats.Stats
	newChainWalker().collectStats(spellID, itemLevel, &total)
	return total
}

func (w *chainWalker) collectStats(spellID, itemLevel int, total *stats.Stats) {
	sp := dbcInstance.Spells[spellID]
	for _, se := range w.effects(spellID) {
		if s, resolved := se.ParseStatEffect(sp.ScalesWithItemLevel(), itemLevel); resolved {
			total.AddInplace(&s)
		} else if se.EffectAura == A_PROC_TRIGGER_SPELL {
			// Deliberately narrower than IsProcTrigger: descending through an
			// A_PROC_TRIGGER_SPELL_WITH_VALUE would collect the triggered spell's own amounts,
			// past the point where the caller can still override them with the value the
			// trigger carries.
			w.collectStats(se.EffectTriggerSpell, itemLevel, total)
		}
	}
}

func ParseItemEffects(itemID, itemLevel int) []*proto.ItemEffect {
	raw := dbcInstance.ItemEffectsByParentID[itemID]
	out := make([]*proto.ItemEffect, 0, len(raw))
	for _, ie := range raw {
		if pe, ok := ie.ToProto(itemLevel); ok {
			out = append(out, pe)
		}
	}
	return out
}

func GetItemEffectSpellTooltip(itemID int, buffId int) (string, int) {
	raw := dbcInstance.ItemEffectsByParentID[itemID]
	var spellID int

	for _, effect := range raw {
		spellID = effect.SpellID
		if effect.SpellID == buffId {
			spellID = effect.SpellID
			break
		} else {
			triggerEffects := dbcInstance.SpellEffects[effect.SpellID]
			if len(triggerEffects) == 0 {
				continue
			}
			if spellEffect := GetSpellEffectRecursive(buffId, triggerEffects); spellEffect != nil {
				if spellEffect.EffectTriggerSpell == buffId {
					spellID = effect.SpellID
				}
				break
			}
		}
	}
	spell := dbcInstance.Spells[spellID]
	return spell.Description, spellID
}

func GetItemEffectForBuffID(itemID int, buffId int) *ItemEffect {
	raw := dbcInstance.ItemEffectsByParentID[itemID]
	var itemEffect *ItemEffect
	for _, effect := range raw {
		if effect.SpellID == buffId {
			itemEffect = &effect
			break
		} else {
			triggerEffects := dbcInstance.SpellEffects[effect.SpellID]
			if len(triggerEffects) == 0 {
				continue
			}
			if spellEffect := GetSpellEffectRecursive(buffId, triggerEffects); spellEffect != nil {
				if spellEffect.EffectTriggerSpell == buffId {
					return &effect
				}
				break
			}
		}
	}
	return itemEffect
}

// Walks the trigger chain below the given effects and returns the effect that triggers
// spellIDToMatch, or nil. Effects are visited in index order and a branch that does not contain
// the match no longer aborts the search of its siblings, which is what hid any buff reachable
// only through a later effect.
//
// Takes the effects rather than a spell ID because that is what the call sites hold; everything
// below the first level goes through the chain walker.
func GetSpellEffectRecursive(spellIDToMatch int, spellEffects map[int]SpellEffect) *SpellEffect {
	w := newChainWalker()

	indices := make([]int, 0, len(spellEffects))
	for idx := range spellEffects {
		indices = append(indices, idx)
	}
	slices.Sort(indices)

	for _, idx := range indices {
		spellEffect := spellEffects[idx]
		if spellEffect.EffectTriggerSpell == 0 {
			continue
		}

		if spellEffect.EffectTriggerSpell == spellIDToMatch {
			return &spellEffect
		}

		if match := w.findTriggerOf(spellIDToMatch, spellEffect.EffectTriggerSpell); match != nil {
			return match
		}
	}
	return nil
}

func (w *chainWalker) findTriggerOf(spellIDToMatch int, spellID int) *SpellEffect {
	for _, spellEffect := range w.effects(spellID) {
		if spellEffect.EffectTriggerSpell == 0 {
			continue
		}

		if spellEffect.EffectTriggerSpell == spellIDToMatch {
			return &spellEffect
		}

		if match := w.findTriggerOf(spellIDToMatch, spellEffect.EffectTriggerSpell); match != nil {
			return match
		}
	}
	return nil
}

// Parses a UIItem and loops through Scaling Options for that item.
func MergeItemEffectsForAllStates(parsed *proto.UIItem) []*proto.ItemEffect {
	var effects []*proto.ItemEffect

	for i := range dbcInstance.ItemEffectsByParentID[int(parsed.Id)] {
		// pick a base effect that has stats if there is more than one effect on the item
		var baseEff *ItemEffect

		e := &dbcInstance.ItemEffectsByParentID[int(parsed.Id)][i]
		statsSpell := resolveStatsSpell(e.SpellID)
		props := buildBaseStatScalingProps(statsSpell, e.SpellID)

		hasStats := len(props.Stats) > 0

		if e.TriggerType == ITEM_SPELLTRIGGER_ON_EQUIP && hasStats {
			for stat, value := range props.Stats {
				parsed.ScalingOptions[0].Stats[int32(stat)] += value
			}
			continue
		} else if (e.TriggerType == ITEM_SPELLTRIGGER_ON_EQUIP) || (e.TriggerType == ITEM_SPELLTRIGGER_CHANCE_ON_HIT && getPPMForItemID(parsed.Id) > 0) || e.CoolDownMSec > 0 {
			baseEff = e
		} else {
			continue
		}

		statsSpellID := resolveStatsSpell(baseEff.SpellID)
		pe := makeBaseProto(baseEff, statsSpellID)
		assignTrigger(baseEff, statsSpellID, pe)

		// add scaling for each saved state
		for state, opt := range parsed.ScalingOptions {
			ilvl := int(opt.Ilvl)
			pe.ScalingOptions[state] = buildItemEffectScalingProps(baseEff.SpellID, ilvl)
		}

		// A container aura that accumulates a separate stat aura resolves its amounts at every
		// state too, and per stack rather than in total.
		if stacking := buildStackingAura(baseEff.SpellID, statsSpellID, int(parsed.ScalingOptions[0].Ilvl), baseEff.ParentItemID); stacking != nil {
			for state, opt := range parsed.ScalingOptions {
				stacking.Aura.ScalingOptions[state] = buildItemEffectScalingProps(int(stacking.Aura.BuffId), int(opt.Ilvl))
			}
			applyStackingAura(pe, stacking)
		}

		dropEmptyScalingOptions(pe)

		// Appended once per effect. Inside the loop above it appended the same pointer once
		// per scaling state, which only stays invisible while items carry a single state.
		effects = append(effects, pe)
	}

	return effects
}
