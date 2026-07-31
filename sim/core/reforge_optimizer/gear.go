package reforgeoptimizer

import (
	"slices"
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

// applyLPSolution rebuilds the equipment from the solver's selected variables. Gem variables
// are keyed "<slot>_<socketIdx>_<gemID>"; every other selected variable (the SocketBonus_<slot>
// indicators) carries no gem and is skipped.
func (o *reforgeOptimizer) applyLPSolution(selectedVars []string) *proto.EquipmentSpec {
	gear := equipmentFromProto(o.baseStrippedGear)

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

	o.minimizeRegems(gear)
	return gear.ToEquipmentSpecProto()
}

// minimizeRegems cuts the number of gems the player must actually buy. For each socket the
// solver changed, it locates where that socket's original gem now lives and swaps the two gems
// back — reusing a gem the player already owns instead of buying a new one — unless doing so would
// drop a socket-color match the solver found.
func (o *reforgeOptimizer) minimizeRegems(newGear *core.Equipment) {
	if o.originalEquipment == nil {
		return
	}

	finalizedSocketKeys := map[reforgeSocketKey]bool{}
	for slotIdx := 0; slotIdx < int(core.NumItemSlots); slotIdx++ {
		slot := proto.ItemSlot(slotIdx)
		newItem := newGear.GetItemBySlot(slot)
		originalItem := o.originalEquipment.GetItemBySlot(slot)
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

			for _, loc := range o.findGem(newGear, originalGemID) {
				if o.frozenSlots[loc.slot] {
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
func (o *reforgeOptimizer) findGem(equipment *core.Equipment, gemID int32) []gemLocation {
	var locations []gemLocation
	for slotIdx := 0; slotIdx < int(core.NumItemSlots); slotIdx++ {
		slot := proto.ItemSlot(slotIdx)
		item := equipment.GetItemBySlot(slot)
		if item.ID == 0 {
			continue
		}
		var originalItem *core.Item
		if o.originalEquipment != nil {
			originalItem = o.originalEquipment.GetItemBySlot(slot)
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
		if !isColoredSocket(socketColor) {
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

type reforgeSocketKey struct {
	slot      proto.ItemSlot
	socketIdx int
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

func hasSocketBonus(item core.Item) bool {
	for _, value := range item.SocketBonus {
		if value != 0 {
			return true
		}
	}
	return false
}

func gemMatchesSocket(gemColor proto.GemColor, socketColor proto.GemColor) bool {
	if gemColor == socketColor {
		return true
	}
	switch socketColor {
	case proto.GemColor_GemColorBlue:
		return gemColor == proto.GemColor_GemColorPurple || gemColor == proto.GemColor_GemColorGreen || gemColor == proto.GemColor_GemColorPrismatic
	case proto.GemColor_GemColorRed:
		return gemColor == proto.GemColor_GemColorPurple || gemColor == proto.GemColor_GemColorOrange || gemColor == proto.GemColor_GemColorPrismatic
	case proto.GemColor_GemColorYellow:
		return gemColor == proto.GemColor_GemColorOrange || gemColor == proto.GemColor_GemColorGreen || gemColor == proto.GemColor_GemColorPrismatic
	case proto.GemColor_GemColorPrismatic:
		return gemColor == proto.GemColor_GemColorRed || gemColor == proto.GemColor_GemColorOrange || gemColor == proto.GemColor_GemColorYellow || gemColor == proto.GemColor_GemColorGreen || gemColor == proto.GemColor_GemColorBlue || gemColor == proto.GemColor_GemColorPurple
	default:
		return false
	}
}

func currentSocketColors(item core.Item) []proto.GemColor {
	return slices.Clone(item.GemSockets)
}

func clearGems(equipment *proto.EquipmentSpec, settings *proto.ReforgeSettings) {
	frozenSlots := frozenItemSlots(settings)
	for slotIdx, item := range equipment.Items {
		slot := proto.ItemSlot(slotIdx)
		if item == nil || frozenSlots[slot] {
			continue
		}

		for gemIdx, gemID := range item.Gems {
			if gemID == 0 {
				continue
			}
			if gem, ok := core.GetGemByID(gemID); ok && gem.Color == proto.GemColor_GemColorMeta {
				continue
			}
			if isHeadMetaSocket(item, slot, gemIdx) {
				continue
			}
			if gem, ok := core.GetGemByID(gemID); !ok || gem.Color != proto.GemColor_GemColorMeta {
				item.Gems[gemIdx] = 0
			}
		}
	}
}

func isHeadMetaSocket(item *proto.ItemSpec, slot proto.ItemSlot, gemIdx int) bool {
	if slot != proto.ItemSlot_ItemSlotHead {
		return false
	}
	if dbItem := core.GetItemByID(item.GetId()); dbItem != nil && gemIdx < len(dbItem.GemSockets) {
		return dbItem.GemSockets[gemIdx] == proto.GemColor_GemColorMeta
	}
	return gemIdx == 0
}
