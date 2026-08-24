package bulk

import (
	"fmt"
	"math"
	"slices"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

type bulkSimCandidateGenerator struct {
	settings           *proto.BulkSettings
	playerClass        proto.Class
	playerSpec         proto.Spec
	playerCanDualWield bool
	baseEquipment      core.Equipment
	selectedByBulkSlot map[BulkSimItemSlot][]bulkSimCandidateOption
	groupedPairsBySlot map[BulkSimItemSlot][][2]bulkSimCandidateOption
	frozenItems        map[BulkSimItemSlot]*core.Item
	frozenWeaponSlot   proto.ItemSlot
	weaponTypeFilters  map[proto.ItemSlot][]proto.WeaponType
	copyCounts         map[itemSpecCacheKey]int
}

func newBulkSimCandidateGenerator(request *proto.BulkSimRequest, player *proto.Player) (*bulkSimCandidateGenerator, error) {
	if player.GetEquipment() == nil {
		return nil, fmt.Errorf("bulk request is missing player equipment")
	}

	playerSpec, err := getPlayerSpec(player)
	if err != nil {
		return nil, err
	}

	generator := &bulkSimCandidateGenerator{
		settings:           request.GetBulkSettings(),
		playerClass:        player.GetClass(),
		playerSpec:         playerSpec,
		playerCanDualWield: core.SpecCanDualWieldCapabilities[playerSpec],
		baseEquipment:      core.ProtoToEquipment(player.GetEquipment()),
		selectedByBulkSlot: make(map[BulkSimItemSlot][]bulkSimCandidateOption),
		groupedPairsBySlot: make(map[BulkSimItemSlot][][2]bulkSimCandidateOption),
		frozenItems:        make(map[BulkSimItemSlot]*core.Item),
		weaponTypeFilters: map[proto.ItemSlot][]proto.WeaponType{
			proto.ItemSlot_ItemSlotMainHand: request.GetBulkSettings().GetFreezeMainhandWeaponSlots(),
			proto.ItemSlot_ItemSlotOffHand:  request.GetBulkSettings().GetFreezeOffhandWeaponSlots(),
		},
	}

	generator.initFrozenSettings()
	if err := generator.initSelectedItems(); err != nil {
		return nil, err
	}
	generator.initGroupedSlotPairs()
	if err := generator.validateGroupedSlots(); err != nil {
		return nil, err
	}

	return generator, nil
}

func (generator *bulkSimCandidateGenerator) validateGroupedSlots() error {
	for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotFinger, BulkSimItemSlotTrinket} {
		if len(generator.selectedByBulkSlot[bulkSlot]) == 0 {
			continue
		}
		if len(generator.groupedPairsBySlot[bulkSlot]) == 0 {
			return fmt.Errorf("no equippable pair of items available for grouped bulk slot %d", bulkSlot)
		}
	}
	return nil
}

func (generator *bulkSimCandidateGenerator) buildCandidates() ([]*proto.BulkGearCandidate, error) {
	rawCombinations := generator.rawCombinationsCount()
	// Never reserve the whole raw space: it is an unbounded product over the bulk slots,
	// so a few extra selections take it past anything that can be allocated for. Append
	// grows geometrically from here anyway.
	candidates := make([]*proto.BulkGearCandidate, 0, min(rawCombinations, maxBulkCandidatePreallocation))
	for comboIdx := 0; comboIdx < rawCombinations; comboIdx++ {
		gear, err := generator.buildGearForCombo(comboIdx)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, &proto.BulkGearCandidate{
			Index: int32(len(candidates)),
			Gear:  gear,
		})
	}
	return candidates, nil
}

func (generator *bulkSimCandidateGenerator) initFrozenSettings() {
	if slot := generator.settings.GetFreezeRingSlot(); slot == int32(proto.ItemSlot_ItemSlotFinger1) || slot == int32(proto.ItemSlot_ItemSlotFinger2) {
		item := generator.baseEquipment.GetItemBySlot(proto.ItemSlot(slot))
		if item != nil && item.ID != 0 {
			itemCopy := *item
			generator.frozenItems[BulkSimItemSlotFinger] = &itemCopy
		}
	}
	if slot := generator.settings.GetFreezeTrinketSlot(); slot == int32(proto.ItemSlot_ItemSlotTrinket1) || slot == int32(proto.ItemSlot_ItemSlotTrinket2) {
		item := generator.baseEquipment.GetItemBySlot(proto.ItemSlot(slot))
		if item != nil && item.ID != 0 {
			itemCopy := *item
			generator.frozenItems[BulkSimItemSlotTrinket] = &itemCopy
		}
	}
	if slot := generator.settings.GetFreezeWeaponSlot(); slot == int32(proto.ItemSlot_ItemSlotMainHand) || slot == int32(proto.ItemSlot_ItemSlotOffHand) {
		generator.frozenWeaponSlot = proto.ItemSlot(slot)
	}
}

func (generator *bulkSimCandidateGenerator) initSelectedItems() error {
	equippedItemsBySlot := make(map[proto.ItemSlot]*core.Item)
	equippedCounts := make(map[itemSpecFingerprintKey]int)
	for slot := proto.ItemSlot_ItemSlotHead; slot < core.NumItemSlots; slot++ {
		equippedItem := generator.baseEquipment.GetItemBySlot(slot)
		if equippedItem == nil || equippedItem.ID == 0 {
			continue
		}
		itemCopy := *equippedItem
		equippedItemsBySlot[slot] = &itemCopy
		equippedCounts[buildItemSpecFingerprintKey(equippedItem.ToItemSpecProto())]++
	}

	for _, selectedItem := range generator.settings.GetItems() {
		if selectedItem == nil || selectedItem.GetId() == 0 {
			continue
		}
		selectedFingerprint := buildItemSpecFingerprintKey(selectedItem)
		if equippedCounts[selectedFingerprint] > 0 {
			// Items filling a pair of slots stack: 1 equipped + 1 added = 2 total, enabling
			// same-item combos like [Sp,Sp] or two Rings of Ancient Knowledge. For single-slot
			// items a user-added duplicate of an equipped item is redundant.
			baseItem := core.GetItemByID(selectedItem.GetId())
			if baseItem == nil || !canStackTwoCopies(*baseItem, generator.playerCanDualWield) {
				equippedCounts[selectedFingerprint]--
				continue
			}
			// paired slot: fall through to add it alongside the equipped copy
		}
		baseItem := core.GetItemByID(selectedItem.GetId())
		if baseItem == nil {
			return fmt.Errorf("selected bulk item %d is missing from the database", selectedItem.GetId())
		}

		option := bulkSimCandidateOption{
			spec: googleProto.Clone(selectedItem).(*proto.ItemSpec),
			item: core.NewItem(core.ItemSpec{
				ID:           selectedItem.GetId(),
				RandomSuffix: selectedItem.GetRandomSuffix(),
				Enchant:      selectedItem.GetEnchant(),
				Gems:         slices.Clone(selectedItem.GetGems()),
			}),
		}

		// Several physical slots can share one bulk slot (Finger1/Finger2, and both hands for a
		// dual-wielder), so dedupe on the bulk slot rather than skipping the secondary physical
		// slot: an off-hand-only item has no other eligible slot, and skipping it dropped the
		// item from the batch entirely.
		addedBulkSlots := make(map[BulkSimItemSlot]struct{}, 2)
		for _, slot := range getEligibleItemSlots(option.item) {
			if !canEquipItem(option.item, generator.playerClass, generator.playerSpec, slot) {
				continue
			}
			bulkSlot := getBulkItemSlotFromSlot(slot, generator.playerCanDualWield)
			if _, added := addedBulkSlots[bulkSlot]; added {
				continue
			}
			addedBulkSlots[bulkSlot] = struct{}{}
			generator.selectedByBulkSlot[bulkSlot] = append(generator.selectedByBulkSlot[bulkSlot], option)
		}
	}

	for slot := proto.ItemSlot_ItemSlotHead; slot < core.NumItemSlots; slot++ {
		equippedItem := equippedItemsBySlot[slot]
		if equippedItem == nil {
			continue
		}
		bulkSlot := getBulkItemSlotFromSlot(slot, generator.playerCanDualWield)
		generator.selectedByBulkSlot[bulkSlot] = append(generator.selectedByBulkSlot[bulkSlot], bulkSimCandidateOption{
			spec: equippedItem.ToItemSpecProto(),
			item: *equippedItem,
		})
	}

	generator.initCopyCounts()
	for bulkSlot, options := range generator.selectedByBulkSlot {
		generator.selectedByBulkSlot[bulkSlot] = dedupeCandidateOptions(options)
	}
	return nil
}

// canStackTwoCopies reports whether both of an item's slots can hold it at once, which is what
// makes a same-item combo (two identical rings, or one weapon in each hand) a valid input.
func canStackTwoCopies(item core.Item, playerCanDualWield bool) bool {
	if item.Unique || item.LimitCategory != 0 {
		return false
	}
	switch item.Type {
	case proto.ItemType_ItemTypeFinger, proto.ItemType_ItemTypeTrinket:
		return true
	case proto.ItemType_ItemTypeWeapon:
		return playerCanDualWield &&
			item.HandType != proto.HandType_HandTypeTwoHand &&
			item.HandType != proto.HandType_HandTypeMainHand &&
			item.HandType != proto.HandType_HandTypeOffHand
	default:
		return false
	}
}

// Capture how many copies of each item are available (selected + equipped) before dedup
// collapses them. An item may only fill both of its slots when at least two copies exist.
func (generator *bulkSimCandidateGenerator) initCopyCounts() {
	generator.copyCounts = make(map[itemSpecCacheKey]int)
	for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotHandWeapon, BulkSimItemSlotFinger, BulkSimItemSlotTrinket} {
		for _, option := range generator.selectedByBulkSlot[bulkSlot] {
			generator.copyCounts[buildItemSpecKey(option.spec)]++
		}
	}
}

// hasTwoCopies reports whether this item may occupy both of its slots at once: the item allows it
// and the batch actually holds two copies.
func (generator *bulkSimCandidateGenerator) hasTwoCopies(option bulkSimCandidateOption, item core.Item) bool {
	return canStackTwoCopies(item, generator.playerCanDualWield) && generator.copyCounts[buildItemSpecKey(option.spec)] >= 2
}

func (generator *bulkSimCandidateGenerator) initGroupedSlotPairs() {
	for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotFinger, BulkSimItemSlotTrinket} {
		options := generator.selectedByBulkSlot[bulkSlot]
		if len(options) == 0 {
			continue
		}
		var pairs [][2]bulkSimCandidateOption
		if frozenItem := generator.frozenItems[bulkSlot]; frozenItem != nil {
			pairs = make([][2]bulkSimCandidateOption, 0, len(options))
			frozenSpec := frozenItem.ToItemSpecProto()
			for _, option := range options {
				if candidateOptionEqualsItem(option, *frozenItem) &&
					!generator.hasTwoCopies(option, *frozenItem) {
					continue
				}
				if frozenItem.Unique && frozenItem.ID == option.item.ID {
					continue
				}
				if frozenItem.LimitCategory != 0 && frozenItem.LimitCategory == option.item.LimitCategory {
					continue
				}
				pairs = append(pairs, [2]bulkSimCandidateOption{{spec: frozenSpec, item: *frozenItem}, option})
			}
		} else {
			pairs = make([][2]bulkSimCandidateOption, 0, len(options)*(len(options)-1)/2)
			for i := 0; i < len(options); i++ {
				// Wearing two copies of the same item is a valid input, but only when two copies
				// were actually selected.
				if generator.hasTwoCopies(options[i], options[i].item) {
					pairs = append(pairs, [2]bulkSimCandidateOption{options[i], options[i]})
				}
				for j := i + 1; j < len(options); j++ {
					if options[i].item.Unique && options[i].item.ID == options[j].item.ID {
						continue
					}
					lc := options[i].item.LimitCategory
					if lc != 0 && lc == options[j].item.LimitCategory {
						continue
					}
					pairs = append(pairs, [2]bulkSimCandidateOption{options[i], options[j]})
				}
			}
		}
		generator.groupedPairsBySlot[bulkSlot] = pairs
	}
}

// The raw combination space is a plain product over the bulk slots and nothing bounds it.
// The product is clamped only to keep it inside the int32 result fields and out of
// overflow - never to refuse the request.
const maxBulkRawCombinations = math.MaxInt32
const maxBulkCandidatePreallocation = 1 << 16

func saturatingCombinationsMul(rawCombinations int, factor int) int {
	if factor == 0 {
		return 0
	}
	if rawCombinations > maxBulkRawCombinations/factor {
		return maxBulkRawCombinations
	}
	return rawCombinations * factor
}

func (generator *bulkSimCandidateGenerator) rawCombinationsCount() int {
	rawCombinations := len(generator.getAllWeaponCombos())
	if rawCombinations == 0 {
		rawCombinations = 1
	}
	for _, bulkSlot := range bulkSimSelectedOrder {
		if bulkSlot == BulkSimItemSlotMainHand || bulkSlot == BulkSimItemSlotOffHand || bulkSlot == BulkSimItemSlotHandWeapon {
			continue
		}
		numOptions := len(generator.selectedByBulkSlot[bulkSlot])
		if numOptions == 0 {
			continue
		}
		if bulkSlot == BulkSimItemSlotFinger || bulkSlot == BulkSimItemSlotTrinket {
			rawCombinations = saturatingCombinationsMul(rawCombinations, len(generator.groupedPairsBySlot[bulkSlot]))
		} else {
			rawCombinations = saturatingCombinationsMul(rawCombinations, numOptions)
		}
	}
	return rawCombinations
}

func (generator *bulkSimCandidateGenerator) buildGearForCombo(comboIdx int) (*proto.EquipmentSpec, error) {
	gear := generator.baseEquipment
	slotItems, err := generator.populateItemsForCombo(comboIdx)
	if err != nil {
		return nil, err
	}

	for slot, option := range slotItems {
		existingItem := gear.GetItemBySlot(slot)
		if existingItem != nil && existingItem.ID != 0 {
			gear[slot] = replaceItem(*existingItem, option)
		} else {
			gear[slot] = createSelectedItem(option)
		}
	}

	if mh := gear.GetItemBySlot(proto.ItemSlot_ItemSlotMainHand); mh != nil && mh.HandType == proto.HandType_HandTypeTwoHand {
		gear[proto.ItemSlot_ItemSlotOffHand] = core.Item{}
	}

	return gear.ToEquipmentSpecProto(), nil
}

func (generator *bulkSimCandidateGenerator) populateItemsForCombo(comboIdx int) (map[proto.ItemSlot]bulkSimCandidateOption, error) {
	itemsBySlot := make(map[proto.ItemSlot]bulkSimCandidateOption)

	allWeaponPairs := generator.getAllWeaponCombos()
	if len(allWeaponPairs) > 0 {
		weaponPairIdx := comboIdx % len(allWeaponPairs)
		comboIdx = comboIdx / len(allWeaponPairs)
		weaponPair := allWeaponPairs[weaponPairIdx]
		if weaponPair[0] != nil {
			itemsBySlot[proto.ItemSlot_ItemSlotMainHand] = *weaponPair[0]
		}
		if weaponPair[1] != nil {
			itemsBySlot[proto.ItemSlot_ItemSlotOffHand] = *weaponPair[1]
		}
	}

	for _, bulkSlot := range bulkSimSelectedOrder {
		if bulkSlot == BulkSimItemSlotMainHand || bulkSlot == BulkSimItemSlotOffHand || bulkSlot == BulkSimItemSlotHandWeapon {
			continue
		}
		options := generator.selectedByBulkSlot[bulkSlot]
		if len(options) == 0 {
			continue
		}

		if bulkSlot == BulkSimItemSlotFinger || bulkSlot == BulkSimItemSlotTrinket {
			pairs := generator.groupedPairsBySlot[bulkSlot]
			if len(pairs) == 0 {
				return nil, fmt.Errorf("at least 2 items must be selected for grouped bulk slot %d", bulkSlot)
			}
			pairIdx := comboIdx % len(pairs)
			comboIdx = comboIdx / len(pairs)
			slots := BulkSimItemSlotToItemSlotPairs[bulkSlot]
			itemsBySlot[slots[0]] = pairs[pairIdx][0]
			itemsBySlot[slots[1]] = pairs[pairIdx][1]
			continue
		}

		optionIdx := comboIdx % len(options)
		comboIdx = comboIdx / len(options)
		slot := BulkSimItemSlotToSingleItemSlot[bulkSlot]
		itemsBySlot[slot] = options[optionIdx]
	}

	return itemsBySlot, nil
}
