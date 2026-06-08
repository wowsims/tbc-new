package reforgeoptimizer

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

func cloneEquipmentSpec(equipment *proto.EquipmentSpec) *proto.EquipmentSpec {
	if equipment == nil {
		return &proto.EquipmentSpec{}
	}
	return googleProto.Clone(equipment).(*proto.EquipmentSpec)
}

type reforgeGearEditor struct {
	gear          *core.Equipment
	originalGear  *core.Equipment
	player        *proto.Player
	settings      *proto.ReforgeSettings
	frozenSlots   map[proto.ItemSlot]bool
	gemOptions    map[int32]*proto.ReforgeGemOption
	maxGemPhase   int32
	maxGemQuality proto.ItemQuality
}

type reforgeSocketKey struct {
	slot      proto.ItemSlot
	socketIdx int
}

func newReforgeGearEditor(gear *proto.EquipmentSpec, originalGear *proto.EquipmentSpec, player *proto.Player, settings *proto.ReforgeSettings, gemOptions []*proto.ReforgeGemOption) *reforgeGearEditor {
	gemOptionMap := make(map[int32]*proto.ReforgeGemOption, len(gemOptions))
	for _, gemOption := range gemOptions {
		if gemOption == nil {
			continue
		}
		gemOptionMap[gemOption.GetId()] = gemOption
	}

	editor := &reforgeGearEditor{
		gear:          equipmentFromProto(gear),
		originalGear:  optionalEquipmentFromProto(originalGear),
		player:        player,
		settings:      settings,
		frozenSlots:   frozenItemSlots(settings),
		gemOptions:    gemOptionMap,
		maxGemPhase:   settings.GetMaxGemPhase(),
		maxGemQuality: settings.GetMaxGemQuality(),
	}
	return editor
}

func (editor *reforgeGearEditor) equipment() *proto.EquipmentSpec {
	if editor == nil || editor.gear == nil {
		return &proto.EquipmentSpec{}
	}
	return editor.gear.ToEquipmentSpecProto()
}

func (editor *reforgeGearEditor) applyChoice(choice reforgeChoice) {
	if editor == nil || editor.gear == nil || int(choice.slot) < 0 || int(choice.slot) >= int(core.NumItemSlots) {
		return
	}
	item := editor.gear.GetItemBySlot(choice.slot)
	if item.ID == 0 {
		return
	}

	for _, gemChoice := range choice.gems {
		for len(item.Gems) <= gemChoice.socketIdx {
			item.Gems = append(item.Gems, core.Gem{})
		}
		item.Gems[gemChoice.socketIdx] = gemFromID(gemChoice.gemID)
	}
}

func (editor *reforgeGearEditor) applyChoices(choices []reforgeChoice) {
	for _, choice := range choices {
		editor.applyChoice(choice)
	}
}

func (editor *reforgeGearEditor) minimizeRegems() {
	if editor == nil || editor.gear == nil || editor.originalGear == nil {
		return
	}

	finalized := make(map[reforgeSocketKey]struct{})
	for slotIdx, item := range editor.gear {
		slot := proto.ItemSlot(slotIdx)
		if item.ID == 0 {
			continue
		}

		currentItem := editor.gear.GetItemBySlot(slot)
		originalItem := editor.originalGear.GetItemBySlot(slot)
		if originalItem == nil || originalItem.ID == 0 {
			continue
		}

		newGemIDs := make([]int32, len(currentItem.Gems))
		newGemColors := make([]proto.GemColor, len(currentItem.Gems))
		for idx := range currentItem.Gems {
			newGemIDs[idx] = currentItem.Gems[idx].ID
			newGemColors[idx] = currentItem.Gems[idx].Color
		}
		originalGemIDs := make([]int32, len(originalItem.Gems))
		originalGemColors := make([]proto.GemColor, len(originalItem.Gems))
		for idx := range originalItem.Gems {
			originalGemIDs[idx] = originalItem.Gems[idx].ID
			originalGemColors[idx] = originalItem.Gems[idx].Color
		}

		socketColors := currentSocketColors(item)
		for socketIdx, socketColor := range socketColors {
			socketKey := reforgeSocketKey{slot: slot, socketIdx: socketIdx}
			if _, ok := finalized[socketKey]; ok {
				continue
			}
			finalized[socketKey] = struct{}{}

			if socketColor == proto.GemColor_GemColorMeta {
				continue
			}

			desiredGemID := int32(0)
			if socketIdx < len(originalGemIDs) {
				desiredGemID = originalGemIDs[socketIdx]
			}
			currentGemID := int32(0)
			currentGemColor := proto.GemColor_GemColorUnknown
			if socketIdx < len(newGemIDs) {
				currentGemID = newGemIDs[socketIdx]
				currentGemColor = newGemColors[socketIdx]
			}
			desiredGemColor := proto.GemColor_GemColorUnknown
			if socketIdx < len(originalGemColors) {
				desiredGemColor = originalGemColors[socketIdx]
			}
			if desiredGemID == 0 || currentGemID == 0 || desiredGemID == currentGemID {
				continue
			}

			editor.forEachGemSocketWithCurrentGem(desiredGemID, slot, socketIdx, func(matched reforgeSocketKey) bool {
				matchedKey := reforgeSocketKey{slot: matched.slot, socketIdx: matched.socketIdx}
				if _, ok := finalized[matchedKey]; ok {
					return false
				}

				matchedItem := editor.gear.GetItemBySlot(matched.slot)
				if matchedItem == nil {
					return false
				}
				otherGemID := gemIDAt(matchedItem, matched.socketIdx)
				if otherGemID == 0 {
					return false
				}

				if !swapPreservesSocketMatches(currentGemColor, desiredGemColor, socketColor, matchedItem, matched.socketIdx) {
					return false
				}

				finalized[matchedKey] = struct{}{}
				setGemIDAt(currentItem, socketIdx, desiredGemID)
				setGemIDAt(matchedItem, matched.socketIdx, currentGemID)
				return true
			})
		}
	}

	editor.restoreNonHitOriginalGems()
}

func swapPreservesSocketMatches(currentGemColor proto.GemColor, desiredGemColor proto.GemColor, currentSocketColor proto.GemColor, matchedItem *core.Item, matchedSocketIdx int) bool {

	matchedSocketColor := proto.GemColor_GemColorUnknown
	matchedSocketColors := currentSocketColors(*matchedItem)
	if matchedSocketIdx < len(matchedSocketColors) {
		matchedSocketColor = matchedSocketColors[matchedSocketIdx]
	}

	currentBefore := gemMatchesSocket(currentGemColor, currentSocketColor)
	currentAfter := gemMatchesSocket(desiredGemColor, currentSocketColor)
	matchedBefore := gemMatchesSocket(desiredGemColor, matchedSocketColor)
	matchedAfter := gemMatchesSocket(currentGemColor, matchedSocketColor)

	return currentBefore == currentAfter && matchedBefore == matchedAfter
}

func (editor *reforgeGearEditor) restoreNonHitOriginalGems() {
	if editor == nil || editor.gear == nil || editor.originalGear == nil {
		return
	}

	for slotIdx, item := range editor.gear {
		slot := proto.ItemSlot(slotIdx)
		if editor.frozenSlots[slot] || item.ID == 0 {
			continue
		}
		currentItem := editor.gear.GetItemBySlot(slot)
		originalItem := editor.originalGear.GetItemBySlot(slot)
		socketColors := currentSocketColors(item)
		for socketIdx, socketColor := range socketColors {
			if socketColor == proto.GemColor_GemColorMeta {
				continue
			}
			currentGemID := gemIDAt(currentItem, socketIdx)
			originalGemID := gemIDAt(originalItem, socketIdx)
			if originalGemID == 0 {
				continue
			}
			if currentGemID == originalGemID {
				continue
			}
			if !editor.gemAllowedBySettings(originalGemID) {
				continue
			}

			setGemIDAt(currentItem, socketIdx, originalGemID)
		}
	}
}

func (editor *reforgeGearEditor) gemAllowedBySettings(gemID int32) bool {
	if editor == nil || editor.settings == nil || gemID == 0 {
		return true
	}
	gemOption, ok := editor.gemOptions[gemID]
	if !ok || gemOption == nil {
		return true
	}
	if gemOption.GetPhase() > editor.maxGemPhase {
		return false
	}
	if gemOption.GetQuality() > editor.maxGemQuality {
		return false
	}
	return true
}

func hasSpellHitStatCap(settings *proto.ReforgeSettings) bool {
	if settings == nil || settings.StatCaps == nil {
		return false
	}
	pseudoStats := settings.StatCaps.GetPseudoStats()
	spellHitIdx := int(proto.PseudoStat_PseudoStatSpellHitPercent)
	return spellHitIdx < len(pseudoStats) && pseudoStats[spellHitIdx] > 0
}

func restoreMetaSocketGem(newItem *core.Item, originalItem *core.Item, socketIdx int) {
	originalGemID := gemIDAt(originalItem, socketIdx)
	if originalGemID != 0 || socketIdx < len(newItem.Gems) {
		setGemIDAt(newItem, socketIdx, originalGemID)
	}
}

func (editor *reforgeGearEditor) forEachGemSocketWithCurrentGem(gemID int32, skipSlot proto.ItemSlot, skipSocketIdx int, visit func(reforgeSocketKey) bool) {
	for slotIdx, item := range editor.gear {
		slot := proto.ItemSlot(slotIdx)
		if item.ID == 0 || editor.frozenSlots[slot] {
			continue
		}
		socketColors := currentSocketColors(item)
		for socketIdx := range socketColors {
			if slot == skipSlot && socketIdx == skipSocketIdx {
				continue
			}
			if gemIDAt(&item, socketIdx) == gemID {
				if visit(reforgeSocketKey{slot: slot, socketIdx: socketIdx}) {
					return
				}
			}
		}
	}
}

func gemIDAt(item *core.Item, socketIdx int) int32 {
	if item == nil || socketIdx >= len(item.Gems) {
		return 0
	}
	return item.Gems[socketIdx].ID
}

func setGemIDAt(item *core.Item, socketIdx int, gemID int32) {
	if item == nil {
		return
	}
	for len(item.Gems) <= socketIdx {
		item.Gems = append(item.Gems, core.Gem{})
	}
	item.Gems[socketIdx] = gemFromID(gemID)
}

func equipmentFromProto(equipment *proto.EquipmentSpec) *core.Equipment {
	if equipment == nil {
		return &core.Equipment{}
	}
	coreEquipment := core.ProtoToEquipment(equipment)
	return &coreEquipment
}

func optionalEquipmentFromProto(equipment *proto.EquipmentSpec) *core.Equipment {
	if equipment == nil {
		return nil
	}
	return equipmentFromProto(equipment)
}

func gemFromID(gemID int32) core.Gem {
	if gemID == 0 {
		return core.Gem{}
	}
	if gem, ok := core.GemsByID[gemID]; ok {
		return gem
	}
	return core.Gem{ID: gemID}
}

func frozenItemSlots(settings *proto.ReforgeSettings) map[proto.ItemSlot]bool {
	frozen := map[proto.ItemSlot]bool{}
	if settings == nil || !settings.GetFreezeItemSlots() {
		return frozen
	}
	for _, item := range settings.GetFrozenItemSlots() {
		frozen[item] = true
	}
	return frozen
}
