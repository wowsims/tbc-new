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
	gear         *core.Equipment
	originalGear *core.Equipment
	player       *proto.Player
	settings     *proto.ReforgeSettings
	frozenSlots  map[proto.ItemSlot]bool
	gemOptions   map[int32]*proto.ReforgeGemOption
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
		gear:         equipmentFromProto(gear),
		originalGear: optionalEquipmentFromProto(originalGear),
		player:       player,
		settings:     settings,
		frozenSlots:  frozenItemSlots(settings),
		gemOptions:   gemOptionMap,
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

// Post-processes gem assignments to minimize unnecessary purchases.
//
// If the optimizer produced the same total multiset of non-meta gems as the
// input, every gem in the output was already present — the LP only permuted
// positions.  In that case we restore the original placement exactly, so the
// player doesn't need to buy or move any gem.
//
// When the multiset differs (the optimizer genuinely added or changed gems),
// we try to reduce regems by swapping gems between unfrozen sockets: for each
// socket where the output differs from the input, we look for a 2-cycle
// partner — another socket whose output gem is the original gem of the current
// socket AND whose original gem is the output gem of the current socket.
// Swapping the pair restores both to their original gems while keeping the
// total gem set unchanged.
func (editor *reforgeGearEditor) minimizeRegems() {
	if editor == nil || editor.gear == nil || editor.originalGear == nil || editor.player == nil {
		return
	}

	// Always restore meta gems first — the optimizer never changes them.
	for slotIdx := range editor.gear {
		newItem := &editor.gear[slotIdx]
		originalItem := &editor.originalGear[slotIdx]
		if newItem.ID == 0 || originalItem.ID == 0 {
			continue
		}
		for socketIdx, socketColor := range currentSocketColors(*newItem) {
			if socketColor == proto.GemColor_GemColorMeta {
				restoreMetaSocketGem(newItem, originalItem, socketIdx)
			}
		}
	}

	// If the optimizer only permuted gems (same total multiset), restore all
	// non-meta gems to their original sockets.  This handles arbitrary-length
	// permutation cycles without needing cycle decomposition.
	if editor.nonMetaGemMultisetUnchanged() {
		for slotIdx := range editor.gear {
			newItem := &editor.gear[slotIdx]
			originalItem := &editor.originalGear[slotIdx]
			if newItem.ID == 0 || originalItem.ID == 0 {
				continue
			}
			if editor.frozenSlots[proto.ItemSlot(slotIdx)] {
				continue
			}
			for socketIdx, socketColor := range currentSocketColors(*newItem) {
				if socketColor == proto.GemColor_GemColorMeta {
					continue
				}
				setGemIDAt(newItem, socketIdx, gemIDAt(originalItem, socketIdx))
			}
		}
		return
	}

	// Multisets differ: try to minimise regems with 2-cycle swaps.  For each
	// changed socket, look for a partner socket that holds a true 2-cycle swap
	// (each socket's output gem is the other's original gem).  This avoids the
	// greedy-matching bug where picking the wrong copy of a repeated gem ID
	// breaks a longer cycle elsewhere.
	finalizedSocketKeys := map[reforgeSocketKey]bool{}
	for slotIdx := range editor.gear {
		newItem := &editor.gear[slotIdx]
		originalItem := &editor.originalGear[slotIdx]
		if newItem.ID == 0 || originalItem.ID == 0 {
			continue
		}
		slot := proto.ItemSlot(slotIdx)
		for socketIdx, socketColor := range currentSocketColors(*newItem) {
			socketKey := reforgeSocketKey{slot: slot, socketIdx: socketIdx}
			if finalizedSocketKeys[socketKey] || socketColor == proto.GemColor_GemColorMeta {
				continue
			}
			newGemID := gemIDAt(newItem, socketIdx)
			originalGemID := gemIDAt(originalItem, socketIdx)
			if newGemID == 0 || originalGemID == 0 || newGemID == originalGemID {
				continue
			}
			newGem, newGemOk := core.GetGemByID(newGemID)
			originalGem, originalGemOk := core.GetGemByID(originalGemID)
			if !newGemOk || !originalGemOk {
				continue
			}
			// Don't swap away a gem that matches the socket better than the original.
			if gemMatchesSocket(newGem.Color, socketColor) && !gemMatchesSocket(originalGem.Color, socketColor) {
				continue
			}
			// Require a true 2-cycle partner: a socket where the current gem is
			// originalGemID and whose original gem is newGemID.
			matchedSlot, matchedSocketIdx, ok := editor.find2CyclePartner(originalGemID, newGemID, finalizedSocketKeys)
			if !ok {
				continue
			}
			finalizedSocketKeys[socketKey] = true
			finalizedSocketKeys[reforgeSocketKey{slot: matchedSlot, socketIdx: matchedSocketIdx}] = true
			setGemIDAt(newItem, socketIdx, originalGemID)
			setGemIDAt(editor.gear.GetItemBySlot(matchedSlot), matchedSocketIdx, newGemID)
		}
	}
}

// nonMetaGemMultisetUnchanged reports whether the optimizer's output contains
// the same multiset of non-meta gem IDs as the original gear.
func (editor *reforgeGearEditor) nonMetaGemMultisetUnchanged() bool {
	counts := make(map[int32]int)
	for slotIdx := range editor.gear {
		newItem := &editor.gear[slotIdx]
		originalItem := &editor.originalGear[slotIdx]
		if newItem.ID == 0 || originalItem.ID == 0 {
			continue
		}
		if editor.frozenSlots[proto.ItemSlot(slotIdx)] {
			continue
		}
		for socketIdx, socketColor := range currentSocketColors(*newItem) {
			if socketColor == proto.GemColor_GemColorMeta {
				continue
			}
			counts[gemIDAt(newItem, socketIdx)]++
			counts[gemIDAt(originalItem, socketIdx)]--
		}
	}
	for _, v := range counts {
		if v != 0 {
			return false
		}
	}
	return true
}

// find2CyclePartner finds an unfrozen, non-finalized socket whose current gem
// is wantCurrentGemID and whose original gem is wantOriginalGemID.  This is
// the exact 2-cycle partner for a swap that restores both sockets to their
// original gems without changing any other socket.
func (editor *reforgeGearEditor) find2CyclePartner(wantCurrentGemID, wantOriginalGemID int32, finalizedSocketKeys map[reforgeSocketKey]bool) (proto.ItemSlot, int, bool) {
	for slotIdx, item := range editor.gear {
		if item.ID == 0 {
			continue
		}
		slot := proto.ItemSlot(slotIdx)
		if editor.frozenSlots[slot] {
			continue
		}
		originalItem := &editor.originalGear[slotIdx]
		for socketIdx, socketColor := range currentSocketColors(item) {
			if socketColor == proto.GemColor_GemColorMeta {
				continue
			}
			if finalizedSocketKeys[reforgeSocketKey{slot: slot, socketIdx: socketIdx}] {
				continue
			}
			if gemIDAt(&item, socketIdx) == wantCurrentGemID && gemIDAt(originalItem, socketIdx) == wantOriginalGemID {
				return slot, socketIdx, true
			}
		}
	}
	return proto.ItemSlot_ItemSlotHead, 0, false
}

// Restores the original meta gem; meta sockets are never modified by the optimizer so the
// original gem is always correct.
func restoreMetaSocketGem(newItem *core.Item, originalItem *core.Item, socketIdx int) {
	originalGemID := gemIDAt(originalItem, socketIdx)
	if originalGemID != 0 || socketIdx < len(newItem.Gems) {
		setGemIDAt(newItem, socketIdx, originalGemID)
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
	if gem, ok := core.GetGemByID(gemID); ok {
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
