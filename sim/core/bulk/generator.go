package bulk

import (
	"fmt"
	"slices"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

type bulkSimCandidateGenerator struct {
	settings            *proto.BulkSettings
	playerClass         proto.Class
	playerSpec          proto.Spec
	playerCanDualWield  bool
	playerIsFuryWarrior bool
	baseEquipment       core.Equipment
	selectedByBulkSlot  map[BulkSimItemSlot][]bulkSimCandidateOption
	groupedPairsBySlot  map[BulkSimItemSlot][][2]bulkSimCandidateOption
	frozenItems         map[BulkSimItemSlot]*core.Item
	frozenWeaponSlot    proto.ItemSlot
	weaponTypeFilters   map[proto.ItemSlot][]proto.WeaponType
	weaponCopyCounts    map[itemSpecCacheKey]int
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
		settings:            request.GetBulkSettings(),
		playerClass:         player.GetClass(),
		playerSpec:          playerSpec,
		playerCanDualWield:  core.SpecCanDualWieldCapabilities[playerSpec],
		playerIsFuryWarrior: playerSpec == proto.Spec_SpecDpsWarrior,
		baseEquipment:       core.ProtoToEquipment(player.GetEquipment()),
		selectedByBulkSlot:  make(map[BulkSimItemSlot][]bulkSimCandidateOption),
		groupedPairsBySlot:  make(map[BulkSimItemSlot][][2]bulkSimCandidateOption),
		frozenItems:         make(map[BulkSimItemSlot]*core.Item),
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

	return generator, nil
}

func (generator *bulkSimCandidateGenerator) buildCandidates() ([]*proto.BulkGearCandidate, error) {
	rawCombinations := generator.rawCombinationsCount()
	candidates := make([]*proto.BulkGearCandidate, 0, rawCombinations)
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
			// For dual-wield weapons, equipped and user-added copies stack:
			// 1 equipped + 1 added = 2 total, enabling same-weapon combos like [Sp,Sp].
			// For all other slots a user-added duplicate of an equipped item is redundant.
			// Fast path: non-dual-wield players can never benefit from stacking, skip lookup.
			skip := !generator.playerCanDualWield
			if !skip {
				baseItem := core.GetItemByID(selectedItem.GetId())
				skip = baseItem == nil ||
					baseItem.Type != proto.ItemType_ItemTypeWeapon ||
					baseItem.HandType == proto.HandType_HandTypeTwoHand
			}
			if skip {
				equippedCounts[selectedFingerprint]--
				continue
			}
			// dual-wield 1H weapon: fall through to add it alongside the equipped copy
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

		for _, slot := range getEligibleItemSlots(option.item, generator.playerIsFuryWarrior) {
			if isSecondaryItemSlot(slot, generator.playerCanDualWield) {
				continue
			}
			if !canEquipItem(option.item, generator.playerClass, generator.playerSpec, slot) {
				continue
			}
			bulkSlot := getBulkItemSlotFromSlot(slot, generator.playerCanDualWield)
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

	generator.initWeaponCopyCounts()
	for bulkSlot, options := range generator.selectedByBulkSlot {
		generator.selectedByBulkSlot[bulkSlot] = dedupeCandidateOptions(options)
	}
	return nil
}

// Capture how many copies of each 1H weapon are available (selected + equipped)
// before dedup collapses them. A weapon may only occupy both hands when at least
// two copies exist.
func (generator *bulkSimCandidateGenerator) initWeaponCopyCounts() {
	generator.weaponCopyCounts = make(map[itemSpecCacheKey]int)
	for _, option := range generator.selectedByBulkSlot[BulkSimItemSlotHandWeapon] {
		generator.weaponCopyCounts[buildItemSpecKey(option.spec)]++
	}
}

func (generator *bulkSimCandidateGenerator) initGroupedSlotPairs() {
	for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotFinger, BulkSimItemSlotTrinket} {
		options := generator.selectedByBulkSlot[bulkSlot]
		if len(options) < 2 {
			continue
		}
		var pairs [][2]bulkSimCandidateOption
		if frozenItem := generator.frozenItems[bulkSlot]; frozenItem != nil {
			pairs = make([][2]bulkSimCandidateOption, 0, len(options))
			frozenSpec := frozenItem.ToItemSpecProto()
			for _, option := range options {
				if candidateOptionEqualsItem(option, *frozenItem) {
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
		if numOptions > 1 && (bulkSlot == BulkSimItemSlotFinger || bulkSlot == BulkSimItemSlotTrinket) {
			rawCombinations *= len(generator.groupedPairsBySlot[bulkSlot])
		} else if numOptions > 0 {
			rawCombinations *= numOptions
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
