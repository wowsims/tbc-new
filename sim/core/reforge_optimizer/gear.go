package reforgeoptimizer

import (
	"strconv"
	"strings"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// gemLocation identifies one socket of one equipped item.
type gemLocation struct {
	slot      proto.ItemSlot
	socketIdx int
}

// applyLPSolutionGear rebuilds the equipment from the solver's selected variables. Gem variables
// are keyed "<slot>_<socketIdx>_<gemID>"; every other selected variable (the SocketBonus_<slot>
// indicators) carries no gem and is skipped.
func applyLPSolutionGear(strippedGear *proto.EquipmentSpec, originalEquipment *core.Equipment, selectedVars []string, frozen map[proto.ItemSlot]bool) *proto.EquipmentSpec {
	gear := equipmentFromProto(strippedGear)

	for _, variableKey := range selectedVars {
		parts := strings.Split(variableKey, "_")
		if len(parts) <= 2 {
			continue
		}
		slotIdx, err := strconv.Atoi(parts[0])
		if err != nil || slotIdx < 0 || slotIdx >= int(core.NumItemSlots) {
			continue
		}
		item := gear.GetItemBySlot(proto.ItemSlot(slotIdx))
		if item.ID == 0 {
			continue
		}
		socketIdx, socketErr := strconv.Atoi(parts[1])
		gemID, gemErr := strconv.Atoi(parts[2])
		if socketErr != nil || gemErr != nil {
			continue
		}
		setGemIDAt(item, socketIdx, int32(gemID))
	}

	minimizeRegemsLP(gear, originalEquipment, frozen)
	return gear.ToEquipmentSpecProto()
}

// minimizeRegemsLP cuts the number of gems the player must actually buy. For each socket the
// solver changed, it locates where that socket's original gem now lives and swaps the two gems
// back — reusing a gem the player already owns instead of buying a new one — unless doing so would
// drop a socket-color match the solver found.
func minimizeRegemsLP(newGear *core.Equipment, originalGear *core.Equipment, frozen map[proto.ItemSlot]bool) {
	if originalGear == nil {
		return
	}

	finalizedSocketKeys := map[reforgeSocketKey]bool{}
	for slotIdx := 0; slotIdx < int(core.NumItemSlots); slotIdx++ {
		slot := proto.ItemSlot(slotIdx)
		newItem := newGear.GetItemBySlot(slot)
		originalItem := originalGear.GetItemBySlot(slot)
		if newItem.ID == 0 || originalItem.ID == 0 {
			continue
		}

		for socketIdx, socketColor := range currentSocketColors(*newItem) {
			socketKey := reforgeSocketKey{slot: slot, socketIdx: socketIdx}
			if finalizedSocketKeys[socketKey] {
				continue
			}
			finalizedSocketKeys[socketKey] = true

			newGemID := gemIDAt(newItem, socketIdx)
			originalGemID := gemIDAt(originalItem, socketIdx)
			if newGemID == 0 || originalGemID == 0 || newGemID == originalGemID {
				continue
			}
			newGem := gemFromID(newGemID)
			originalGem := gemFromID(originalGemID)

			for _, loc := range findGemLP(newGear, originalGear, originalGemID) {
				if frozen[loc.slot] {
					continue
				}
				matchedKey := reforgeSocketKey{slot: loc.slot, socketIdx: loc.socketIdx}
				if finalizedSocketKeys[matchedKey] {
					continue
				}
				matchedItem := newGear.GetItemBySlot(loc.slot)
				matchedColors := currentSocketColors(*matchedItem)
				if loc.socketIdx >= len(matchedColors) {
					continue
				}
				matchedSocketColor := matchedColors[loc.socketIdx]
				// Restore the original gem here only if it does not reduce the total socket-color
				// matches across BOTH sockets involved. Weighing both sockets (not just the one the
				// gem moved to) preserves a genuine color-match upgrade the solver found while still
				// undoing a match-neutral shuffle that would otherwise be a pointless regem.
				matchesIfSwapped := boolToInt(gemMatchesSocket(originalGem.Color, socketColor)) + boolToInt(gemMatchesSocket(newGem.Color, matchedSocketColor))
				matchesIfKept := boolToInt(gemMatchesSocket(newGem.Color, socketColor)) + boolToInt(gemMatchesSocket(originalGem.Color, matchedSocketColor))
				if matchesIfSwapped < matchesIfKept {
					continue
				}

				// A socket bonus is all-or-nothing per item, so a match-count-neutral swap can still
				// deactivate one item's bonus while gaining a match on another. Apply the swap, then
				// keep it only if no socket bonus was lost.
				bonusesBefore := boolToInt(socketBonusActive(newItem)) + boolToInt(socketBonusActive(matchedItem))
				setGemIDAt(newItem, socketIdx, originalGemID)
				setGemIDAt(matchedItem, loc.socketIdx, newGemID)
				if boolToInt(socketBonusActive(newItem))+boolToInt(socketBonusActive(matchedItem)) < bonusesBefore {
					setGemIDAt(newItem, socketIdx, newGemID)
					setGemIDAt(matchedItem, loc.socketIdx, originalGemID)
					continue
				}

				finalizedSocketKeys[matchedKey] = true
				break
			}
		}
	}
}

// findGemLP returns every socket into which the SOLVER moved gemID — i.e. a socket now holding
// gemID whose own original gem was something else. Sockets the solver never changed are skipped:
// they hold their rightful gem and must not be disturbed.
func findGemLP(equipment *core.Equipment, originalGear *core.Equipment, gemID int32) []gemLocation {
	var locations []gemLocation
	for slotIdx := 0; slotIdx < int(core.NumItemSlots); slotIdx++ {
		slot := proto.ItemSlot(slotIdx)
		item := equipment.GetItemBySlot(slot)
		if item.ID == 0 {
			continue
		}
		var originalItem *core.Item
		if originalGear != nil {
			originalItem = originalGear.GetItemBySlot(slot)
		}
		for socketIdx := range currentSocketColors(*item) {
			if gemIDAt(item, socketIdx) != gemID {
				continue
			}
			if originalItem != nil && gemIDAt(originalItem, socketIdx) == gemID {
				continue
			}
			locations = append(locations, gemLocation{slot: slot, socketIdx: socketIdx})
		}
	}
	return locations
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// socketBonusActive reports whether an item's socket bonus is currently earned: the item must have
// a socket bonus and every gemmable socket must hold a colour-matching gem. The bonus is
// all-or-nothing, which is why a single mismatched gem forfeits it entirely.
func socketBonusActive(item *core.Item) bool {
	if item == nil || item.ID == 0 || !hasSocketBonus(*item) {
		return false
	}
	for socketIdx, socketColor := range currentSocketColors(*item) {
		if !isGemmableSocketColor(socketColor) {
			continue
		}
		gemID := gemIDAt(item, socketIdx)
		if gemID == 0 || !gemMatchesSocket(gemFromID(gemID).Color, socketColor) {
			return false
		}
	}
	return true
}

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
	// non-meta gems to their original sockets — the player doesn't need to buy
	// anything, and we minimize physical gem swaps.
	// Exception: if the LP improved any socket's color matching (e.g. moved an
	// orange gem into a Yellow socket where the original had a non-matching Red),
	// keep the LP's full arrangement.  Partial restore would break multiset
	// consistency and undo the LP's intentional improvement.
	if editor.nonMetaGemMultisetUnchanged() {
		for slotIdx := range editor.gear {
			newItem := &editor.gear[slotIdx]
			originalItem := &editor.originalGear[slotIdx]
			if newItem.ID == 0 || originalItem.ID == 0 || editor.frozenSlots[proto.ItemSlot(slotIdx)] {
				continue
			}
			for socketIdx, socketColor := range currentSocketColors(*newItem) {
				if socketColor == proto.GemColor_GemColorMeta {
					continue
				}
				newGemID := gemIDAt(newItem, socketIdx)
				originalGemID := gemIDAt(originalItem, socketIdx)
				if newGemID == originalGemID {
					continue
				}
				newGem, newOk := core.GetGemByID(newGemID)
				originalGem, origOk := core.GetGemByID(originalGemID)
				if !newOk || !origOk {
					continue
				}
				if gemMatchesSocket(newGem.Color, socketColor) && !gemMatchesSocket(originalGem.Color, socketColor) {
					// LP placed a color-matching gem where original had a non-matching one.
					// Keep the LP's full arrangement to preserve the improvement.
					return
				}
			}
		}
		for slotIdx := range editor.gear {
			newItem := &editor.gear[slotIdx]
			originalItem := &editor.originalGear[slotIdx]
			if newItem.ID == 0 || originalItem.ID == 0 || editor.frozenSlots[proto.ItemSlot(slotIdx)] {
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
