package bulk

import (
	"fmt"
	"slices"
	"sync"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// Bulk candidate/count generation may mutate the shared item database from
// request-scoped player database payloads, so serialize this path.
var bulkCandidateDatabaseMu sync.Mutex

type bulkSimCandidateOption struct {
	spec *proto.ItemSpec
	item core.Item
}

type bulkSimRequiredSetBonusComboMatcher struct {
	baseCounts     []int
	requiredPieces []int
	dimensions     []bulkSimRequiredSetBonusDimension
}

type bulkSimRequiredSetBonusDimension struct {
	optionDeltas [][]int
}

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
}

type BulkSimItemSlot int

const (
	BulkSimItemSlotHead BulkSimItemSlot = iota
	BulkSimItemSlotNeck
	BulkSimItemSlotShoulder
	BulkSimItemSlotBack
	BulkSimItemSlotChest
	BulkSimItemSlotWrist
	BulkSimItemSlotHands
	BulkSimItemSlotWaist
	BulkSimItemSlotLegs
	BulkSimItemSlotFeet
	BulkSimItemSlotFinger
	BulkSimItemSlotTrinket
	BulkSimItemSlotMainHand
	BulkSimItemSlotOffHand
	BulkSimItemSlotRanged
	BulkSimItemSlotHandWeapon
)

// Exported generation constants for tools/database/gen_db/gen_bulksim_constants.ts.go.
var BulkSimItemSlotOrdered = []BulkSimItemSlot{
	BulkSimItemSlotHead,
	BulkSimItemSlotNeck,
	BulkSimItemSlotShoulder,
	BulkSimItemSlotBack,
	BulkSimItemSlotChest,
	BulkSimItemSlotWrist,
	BulkSimItemSlotHands,
	BulkSimItemSlotWaist,
	BulkSimItemSlotLegs,
	BulkSimItemSlotFeet,
	BulkSimItemSlotFinger,
	BulkSimItemSlotTrinket,
	BulkSimItemSlotMainHand,
	BulkSimItemSlotOffHand,
	BulkSimItemSlotRanged,
	BulkSimItemSlotHandWeapon,
}

var BulkSimItemSlotNames = map[BulkSimItemSlot]string{
	BulkSimItemSlotHead:       "ItemSlotHead",
	BulkSimItemSlotNeck:       "ItemSlotNeck",
	BulkSimItemSlotShoulder:   "ItemSlotShoulder",
	BulkSimItemSlotBack:       "ItemSlotBack",
	BulkSimItemSlotChest:      "ItemSlotChest",
	BulkSimItemSlotWrist:      "ItemSlotWrist",
	BulkSimItemSlotHands:      "ItemSlotHands",
	BulkSimItemSlotWaist:      "ItemSlotWaist",
	BulkSimItemSlotLegs:       "ItemSlotLegs",
	BulkSimItemSlotFeet:       "ItemSlotFeet",
	BulkSimItemSlotFinger:     "ItemSlotFinger",
	BulkSimItemSlotTrinket:    "ItemSlotTrinket",
	BulkSimItemSlotMainHand:   "ItemSlotMainHand",
	BulkSimItemSlotOffHand:    "ItemSlotOffHand",
	BulkSimItemSlotRanged:     "ItemSlotRanged",
	BulkSimItemSlotHandWeapon: "ItemSlotHandWeapon",
}

var BulkSimItemSlotToSingleItemSlot = map[BulkSimItemSlot]proto.ItemSlot{
	BulkSimItemSlotHead:     proto.ItemSlot_ItemSlotHead,
	BulkSimItemSlotNeck:     proto.ItemSlot_ItemSlotNeck,
	BulkSimItemSlotShoulder: proto.ItemSlot_ItemSlotShoulder,
	BulkSimItemSlotBack:     proto.ItemSlot_ItemSlotBack,
	BulkSimItemSlotChest:    proto.ItemSlot_ItemSlotChest,
	BulkSimItemSlotWrist:    proto.ItemSlot_ItemSlotWrist,
	BulkSimItemSlotHands:    proto.ItemSlot_ItemSlotHands,
	BulkSimItemSlotWaist:    proto.ItemSlot_ItemSlotWaist,
	BulkSimItemSlotLegs:     proto.ItemSlot_ItemSlotLegs,
	BulkSimItemSlotFeet:     proto.ItemSlot_ItemSlotFeet,
	BulkSimItemSlotMainHand: proto.ItemSlot_ItemSlotMainHand,
	BulkSimItemSlotOffHand:  proto.ItemSlot_ItemSlotOffHand,
	BulkSimItemSlotRanged:   proto.ItemSlot_ItemSlotRanged,
}

var BulkSimItemSlotToItemSlotPairs = map[BulkSimItemSlot][2]proto.ItemSlot{
	BulkSimItemSlotFinger:     {proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2},
	BulkSimItemSlotTrinket:    {proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2},
	BulkSimItemSlotHandWeapon: {proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand},
}

var ItemSlotToBulkSimItemSlot = map[proto.ItemSlot]BulkSimItemSlot{
	proto.ItemSlot_ItemSlotHead:     BulkSimItemSlotHead,
	proto.ItemSlot_ItemSlotNeck:     BulkSimItemSlotNeck,
	proto.ItemSlot_ItemSlotShoulder: BulkSimItemSlotShoulder,
	proto.ItemSlot_ItemSlotBack:     BulkSimItemSlotBack,
	proto.ItemSlot_ItemSlotChest:    BulkSimItemSlotChest,
	proto.ItemSlot_ItemSlotWrist:    BulkSimItemSlotWrist,
	proto.ItemSlot_ItemSlotHands:    BulkSimItemSlotHands,
	proto.ItemSlot_ItemSlotWaist:    BulkSimItemSlotWaist,
	proto.ItemSlot_ItemSlotLegs:     BulkSimItemSlotLegs,
	proto.ItemSlot_ItemSlotFeet:     BulkSimItemSlotFeet,
	proto.ItemSlot_ItemSlotFinger1:  BulkSimItemSlotFinger,
	proto.ItemSlot_ItemSlotFinger2:  BulkSimItemSlotFinger,
	proto.ItemSlot_ItemSlotTrinket1: BulkSimItemSlotTrinket,
	proto.ItemSlot_ItemSlotTrinket2: BulkSimItemSlotTrinket,
	proto.ItemSlot_ItemSlotMainHand: BulkSimItemSlotMainHand,
	proto.ItemSlot_ItemSlotOffHand:  BulkSimItemSlotOffHand,
	proto.ItemSlot_ItemSlotRanged:   BulkSimItemSlotRanged,
}

var bulkSimItemTypeToSlots = map[proto.ItemType][]proto.ItemSlot{
	proto.ItemType_ItemTypeHead:     {proto.ItemSlot_ItemSlotHead},
	proto.ItemType_ItemTypeNeck:     {proto.ItemSlot_ItemSlotNeck},
	proto.ItemType_ItemTypeShoulder: {proto.ItemSlot_ItemSlotShoulder},
	proto.ItemType_ItemTypeBack:     {proto.ItemSlot_ItemSlotBack},
	proto.ItemType_ItemTypeChest:    {proto.ItemSlot_ItemSlotChest},
	proto.ItemType_ItemTypeWrist:    {proto.ItemSlot_ItemSlotWrist},
	proto.ItemType_ItemTypeHands:    {proto.ItemSlot_ItemSlotHands},
	proto.ItemType_ItemTypeWaist:    {proto.ItemSlot_ItemSlotWaist},
	proto.ItemType_ItemTypeLegs:     {proto.ItemSlot_ItemSlotLegs},
	proto.ItemType_ItemTypeFeet:     {proto.ItemSlot_ItemSlotFeet},
	proto.ItemType_ItemTypeFinger:   {proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2},
	proto.ItemType_ItemTypeTrinket:  {proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2},
	proto.ItemType_ItemTypeRanged:   {proto.ItemSlot_ItemSlotRanged},
}

func BulkCombinationCount(request *proto.BulkCombinationCountRequest) *proto.BulkCombinationCountResult {
	bulkCandidateDatabaseMu.Lock()
	defer bulkCandidateDatabaseMu.Unlock()

	if request == nil {
		return &proto.BulkCombinationCountResult{Error: &proto.ErrorOutcome{Message: "bulk combination count request is missing"}}
	}
	if request.GetBaseRequest() == nil {
		return &proto.BulkCombinationCountResult{Error: &proto.ErrorOutcome{Message: "bulk combination count request is missing base request"}}
	}
	if request.GetBulkSettings() == nil {
		return &proto.BulkCombinationCountResult{Error: &proto.ErrorOutcome{Message: "bulk combination count request is missing bulk settings"}}
	}

	bulkRequest := &proto.BulkSimRequest{
		BaseRequest:  request.GetBaseRequest(),
		BulkSettings: request.GetBulkSettings(),
	}
	player, playerErr := getPlayer(bulkRequest)
	if playerErr != nil {
		return &proto.BulkCombinationCountResult{Error: &proto.ErrorOutcome{Message: playerErr.Error()}}
	}
	if player.GetEquipment() == nil {
		return &proto.BulkCombinationCountResult{Error: &proto.ErrorOutcome{Message: "bulk combination count request is missing player equipment"}}
	}
	if player.GetDatabase() != nil {
		core.AddToDatabase(player.GetDatabase())
	}

	generator, err := newBulkSimCandidateGenerator(bulkRequest, player)
	if err != nil {
		return &proto.BulkCombinationCountResult{Error: &proto.ErrorOutcome{Message: err.Error()}}
	}

	rawCombinations := generator.rawCombinationsCount()
	matchingCombinations := rawCombinations
	if matcher := generator.buildRequiredSetBonusMatcher(generator.settings.GetRequiredSetBonuses()); matcher != nil {
		matchingCombinations = 0
		scratchCounts := make([]int, len(matcher.baseCounts))
		for comboIdx := 0; comboIdx < rawCombinations; comboIdx++ {
			if generator.comboMatchesRequiredSetBonusMatcher(comboIdx, matcher, scratchCounts) {
				matchingCombinations++
			}
		}
	}

	return &proto.BulkCombinationCountResult{
		RawCombinations:  int32(rawCombinations),
		Combinations:     int32(matchingCombinations),
		Iterations:       estimateIterationsForCountRequest(request.GetBulkSettings(), matchingCombinations),
		UseLegacyBulkSim: shouldUseLegacyBulkSimForCountRequest(request.GetBulkSettings(), matchingCombinations),
	}
}

func estimateIterationsForCountRequest(settings *proto.BulkSettings, candidateCount int) float64 {
	iterations, _ := estimateBulkSimIterations(settings, settings.GetIterationsPerCombo(), candidateCount)
	return float64(iterations)
}

func shouldUseLegacyBulkSimForCountRequest(settings *proto.BulkSettings, candidateCount int) bool {
	useLegacyBulkSim := shouldUseLegacyBulkSim(settings, settings.GetIterationsPerCombo(), candidateCount)
	return useLegacyBulkSim
}

func BulkCandidates(request *proto.BulkCandidatesRequest) *proto.BulkCandidatesResult {
	bulkCandidateDatabaseMu.Lock()
	defer bulkCandidateDatabaseMu.Unlock()

	if request == nil {
		return &proto.BulkCandidatesResult{Error: &proto.ErrorOutcome{Message: "bulk candidates request is missing"}}
	}
	if request.GetBaseRequest() == nil {
		return &proto.BulkCandidatesResult{Error: &proto.ErrorOutcome{Message: "bulk candidates request is missing base request"}}
	}
	if request.GetBulkSettings() == nil {
		return &proto.BulkCandidatesResult{Error: &proto.ErrorOutcome{Message: "bulk candidates request is missing bulk settings"}}
	}

	bulkRequest := &proto.BulkSimRequest{
		BaseRequest:  request.GetBaseRequest(),
		BulkSettings: request.GetBulkSettings(),
	}
	player, playerErr := getPlayer(bulkRequest)
	if playerErr != nil {
		return &proto.BulkCandidatesResult{Error: &proto.ErrorOutcome{Message: playerErr.Error()}}
	}
	if player.GetEquipment() == nil {
		return &proto.BulkCandidatesResult{Error: &proto.ErrorOutcome{Message: "bulk candidates request is missing player equipment"}}
	}
	if player.GetDatabase() != nil {
		core.AddToDatabase(player.GetDatabase())
	}

	generator, err := newBulkSimCandidateGenerator(bulkRequest, player)
	if err != nil {
		return &proto.BulkCandidatesResult{Error: &proto.ErrorOutcome{Message: err.Error()}}
	}

	rawCombinations := generator.rawCombinationsCount()
	candidates, err := generator.buildCandidates()
	if err != nil {
		return &proto.BulkCandidatesResult{Error: &proto.ErrorOutcome{Message: err.Error()}}
	}

	return &proto.BulkCandidatesResult{
		Candidates:      candidates,
		RawCombinations: int32(rawCombinations),
		Combinations:    int32(len(candidates)),
	}
}

func EnsureBulkSimCandidatesGenerated(request *proto.BulkSimRequest) error {
	bulkCandidateDatabaseMu.Lock()
	defer bulkCandidateDatabaseMu.Unlock()

	if request == nil || request.GetBulkSettings() == nil || len(request.GetCandidates()) > 0 {
		return nil
	}
	if request.GetBaseRequest() == nil || request.GetBaseRequest().GetRaid() == nil {
		return fmt.Errorf("bulk sim request is missing base raid")
	}
	player, playerErr := getPlayer(request)
	if playerErr != nil {
		return playerErr
	}
	if player.GetEquipment() == nil {
		return fmt.Errorf("bulk sim request is missing player equipment")
	}
	if player.GetDatabase() != nil {
		core.AddToDatabase(player.GetDatabase())
	}
	generator, buildErr := newBulkSimCandidateGenerator(request, player)
	if buildErr != nil {
		return buildErr
	}
	candidates, buildErr := generator.buildCandidates()
	if buildErr != nil {
		return buildErr
	}
	request.Candidates = candidates
	return nil
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
	matcher := generator.buildRequiredSetBonusMatcher(generator.settings.GetRequiredSetBonuses())
	candidates := make([]*proto.BulkGearCandidate, 0, rawCombinations)
	var scratchCounts []int
	if matcher != nil {
		scratchCounts = make([]int, len(matcher.baseCounts))
	}
	for comboIdx := 0; comboIdx < rawCombinations; comboIdx++ {
		if !generator.comboMatchesRequiredSetBonusMatcher(comboIdx, matcher, scratchCounts) {
			continue
		}
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
	equippedCounts := make(map[int32]int)
	for slot := proto.ItemSlot_ItemSlotHead; slot < core.NumItemSlots; slot++ {
		equippedItem := generator.baseEquipment.GetItemBySlot(slot)
		if equippedItem == nil || equippedItem.ID == 0 {
			continue
		}
		itemCopy := *equippedItem
		equippedItemsBySlot[slot] = &itemCopy
		equippedCounts[equippedItem.ID]++
	}

	for _, selectedItem := range generator.settings.GetItems() {
		if selectedItem == nil || selectedItem.GetId() == 0 {
			continue
		}
		if equippedCounts[selectedItem.GetId()] > 0 {
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
				equippedCounts[selectedItem.GetId()]--
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

	return nil
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
	for _, bulkSlot := range []BulkSimItemSlot{
		BulkSimItemSlotHead,
		BulkSimItemSlotNeck,
		BulkSimItemSlotShoulder,
		BulkSimItemSlotBack,
		BulkSimItemSlotChest,
		BulkSimItemSlotWrist,
		BulkSimItemSlotHands,
		BulkSimItemSlotWaist,
		BulkSimItemSlotLegs,
		BulkSimItemSlotFeet,
		BulkSimItemSlotFinger,
		BulkSimItemSlotTrinket,
		BulkSimItemSlotRanged,
	} {
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
	slotItems, err := generator.itemsForCombo(comboIdx)
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

func (generator *bulkSimCandidateGenerator) itemsForCombo(comboIdx int) (map[proto.ItemSlot]bulkSimCandidateOption, error) {
	itemsForCombo := make(map[proto.ItemSlot]bulkSimCandidateOption)

	allWeaponPairs := generator.getAllWeaponCombos()
	if len(allWeaponPairs) > 0 {
		weaponPairIdx := comboIdx % len(allWeaponPairs)
		comboIdx = comboIdx / len(allWeaponPairs)
		weaponPair := allWeaponPairs[weaponPairIdx]
		if weaponPair[0] != nil {
			itemsForCombo[proto.ItemSlot_ItemSlotMainHand] = *weaponPair[0]
		}
		if weaponPair[1] != nil {
			itemsForCombo[proto.ItemSlot_ItemSlotOffHand] = *weaponPair[1]
		}
	}

	for _, bulkSlot := range []BulkSimItemSlot{
		BulkSimItemSlotHead,
		BulkSimItemSlotNeck,
		BulkSimItemSlotShoulder,
		BulkSimItemSlotBack,
		BulkSimItemSlotChest,
		BulkSimItemSlotWrist,
		BulkSimItemSlotHands,
		BulkSimItemSlotWaist,
		BulkSimItemSlotLegs,
		BulkSimItemSlotFeet,
		BulkSimItemSlotFinger,
		BulkSimItemSlotTrinket,
		BulkSimItemSlotRanged,
	} {
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
			itemsForCombo[slots[0]] = pairs[pairIdx][0]
			itemsForCombo[slots[1]] = pairs[pairIdx][1]
			continue
		}

		optionIdx := comboIdx % len(options)
		comboIdx = comboIdx / len(options)
		slot := BulkSimItemSlotToSingleItemSlot[bulkSlot]
		itemsForCombo[slot] = options[optionIdx]
	}

	return itemsForCombo, nil
}

func (generator *bulkSimCandidateGenerator) getAllWeaponCombos() [][2]*bulkSimCandidateOption {
	allWeaponCombos := make([][2]*bulkSimCandidateOption, 0)
	all2HWeapons := make([]bulkSimCandidateOption, 0)
	for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotMainHand, BulkSimItemSlotHandWeapon} {
		for _, option := range generator.selectedByBulkSlot[bulkSlot] {
			if option.item.HandType == proto.HandType_HandTypeTwoHand {
				all2HWeapons = append(all2HWeapons, option)
			}
		}
	}

	if generator.playerIsFuryWarrior {
		for i := range all2HWeapons {
			if optionsContainEquivalent(all2HWeapons[:i], all2HWeapons[i]) {
				continue
			}
			allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&all2HWeapons[i], nil})
			for j := i + 1; j < len(all2HWeapons); j++ {
				if optionsContainEquivalent(all2HWeapons[i+1:j], all2HWeapons[j]) {
					continue
				}
				allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&all2HWeapons[i], &all2HWeapons[j]})
				if !candidateOptionsEqual(all2HWeapons[i], all2HWeapons[j]) {
					allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&all2HWeapons[j], &all2HWeapons[i]})
				}
			}
		}
	} else {
		for i := range all2HWeapons {
			allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&all2HWeapons[i], nil})
		}
	}

	mhOptions := generator.selectedByBulkSlot[BulkSimItemSlotMainHand]
	ohOptions := generator.selectedByBulkSlot[BulkSimItemSlotOffHand]
	if len(mhOptions) > 0 {
		for i := range mhOptions {
			if optionsContainEquivalent(all2HWeapons, mhOptions[i]) {
				continue
			}
			if len(ohOptions) > 0 {
				for j := range ohOptions {
					allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&mhOptions[i], &ohOptions[j]})
				}
			} else {
				allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&mhOptions[i], nil})
			}
		}
	} else if len(ohOptions) > 0 {
		for i := range ohOptions {
			allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{nil, &ohOptions[i]})
		}
	}

	oneHandOptions := generator.selectedByBulkSlot[BulkSimItemSlotHandWeapon]
	if len(oneHandOptions) > 0 {
		type weaponEntry struct {
			option bulkSimCandidateOption
			count  int
		}
		unique := make([]weaponEntry, 0, len(oneHandOptions))
		for _, option := range oneHandOptions {
			if optionsContainEquivalent(all2HWeapons, option) {
				continue
			}
			found := false
			for k := range unique {
				if candidateOptionsEqual(unique[k].option, option) {
					unique[k].count++
					found = true
					break
				}
			}
			if !found {
				unique = append(unique, weaponEntry{option: option, count: 1})
			}
		}

		for i := range unique {
			iCanMH := unique[i].option.item.HandType != proto.HandType_HandTypeOffHand
			iCanOH := unique[i].option.item.HandType != proto.HandType_HandTypeMainHand
			if unique[i].count >= 2 && iCanMH && iCanOH {
				allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&unique[i].option, &unique[i].option})
			}
			for j := i + 1; j < len(unique); j++ {
				jCanMH := unique[j].option.item.HandType != proto.HandType_HandTypeOffHand
				jCanOH := unique[j].option.item.HandType != proto.HandType_HandTypeMainHand
				if iCanMH && jCanOH {
					allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&unique[i].option, &unique[j].option})
				}
				if jCanMH && iCanOH {
					allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&unique[j].option, &unique[i].option})
				}
			}
		}
	}

	filteredCombos := make([][2]*bulkSimCandidateOption, 0, len(allWeaponCombos))
	for _, combo := range allWeaponCombos {
		if generator.weaponComboMatchesSettings(combo[0], combo[1]) {
			filteredCombos = append(filteredCombos, combo)
		}
	}

	return filteredCombos
}

func (generator *bulkSimCandidateGenerator) getFrozenWeaponItem() *core.Item {
	if generator.frozenWeaponSlot != proto.ItemSlot_ItemSlotMainHand && generator.frozenWeaponSlot != proto.ItemSlot_ItemSlotOffHand {
		return nil
	}
	item := generator.baseEquipment.GetItemBySlot(generator.frozenWeaponSlot)
	if item == nil || item.ID == 0 {
		return nil
	}
	itemCopy := *item
	return &itemCopy
}

func (generator *bulkSimCandidateGenerator) matchesWeaponTypeFilter(option *bulkSimCandidateOption, slot proto.ItemSlot) bool {
	filter := generator.weaponTypeFilters[slot]
	if len(filter) == 0 {
		return true
	}
	if option == nil {
		return false
	}
	return option.item.WeaponType > proto.WeaponType_WeaponTypeUnknown && slices.Contains(filter, option.item.WeaponType)
}

func (generator *bulkSimCandidateGenerator) weaponComboMatchesSettings(mhItem *bulkSimCandidateOption, ohItem *bulkSimCandidateOption) bool {
	frozenWeaponItem := generator.getFrozenWeaponItem()
	if generator.frozenWeaponSlot == proto.ItemSlot_ItemSlotMainHand && frozenWeaponItem != nil && !candidateOptionEqualsItemPtr(mhItem, frozenWeaponItem) {
		return false
	}
	if generator.frozenWeaponSlot == proto.ItemSlot_ItemSlotOffHand && frozenWeaponItem != nil && !candidateOptionEqualsItemPtr(ohItem, frozenWeaponItem) {
		return false
	}
	return generator.matchesWeaponTypeFilter(mhItem, proto.ItemSlot_ItemSlotMainHand) && generator.matchesWeaponTypeFilter(ohItem, proto.ItemSlot_ItemSlotOffHand)
}

func (generator *bulkSimCandidateGenerator) buildRequiredSetBonusMatcher(requiredSetBonuses []*proto.BulkRequiredSetBonus) *bulkSimRequiredSetBonusComboMatcher {
	if len(requiredSetBonuses) == 0 {
		return nil
	}

	requiredIndexes := make(map[int32]int, len(requiredSetBonuses))
	for idx, required := range requiredSetBonuses {
		requiredIndexes[required.GetSetId()] = idx
	}

	baseCounts := make([]int, len(requiredSetBonuses))
	for slot := proto.ItemSlot_ItemSlotHead; slot < core.NumItemSlots; slot++ {
		generator.addItemToRequiredSetBonusCounts(baseCounts, requiredIndexes, generator.baseEquipment.GetItemBySlot(slot), 1)
	}

	dimensions := make([]bulkSimRequiredSetBonusDimension, 0)

	weaponPairs := generator.getAllWeaponCombos()
	if len(weaponPairs) > 0 {
		optionDeltas := make([][]int, 0, len(weaponPairs))
		for _, pair := range weaponPairs {
			optionDeltas = append(optionDeltas, generator.getRequiredSetBonusOptionDeltas(requiredIndexes, [][2]any{{proto.ItemSlot_ItemSlotMainHand, pair[0]}, {proto.ItemSlot_ItemSlotOffHand, pair[1]}}))
		}
		dimensions = append(dimensions, bulkSimRequiredSetBonusDimension{optionDeltas: optionDeltas})
	}

	for _, bulkSlot := range []BulkSimItemSlot{
		BulkSimItemSlotHead,
		BulkSimItemSlotNeck,
		BulkSimItemSlotShoulder,
		BulkSimItemSlotBack,
		BulkSimItemSlotChest,
		BulkSimItemSlotWrist,
		BulkSimItemSlotHands,
		BulkSimItemSlotWaist,
		BulkSimItemSlotLegs,
		BulkSimItemSlotFeet,
		BulkSimItemSlotFinger,
		BulkSimItemSlotTrinket,
		BulkSimItemSlotRanged,
	} {
		options := generator.selectedByBulkSlot[bulkSlot]
		if len(options) == 0 {
			continue
		}

		if bulkSlot == BulkSimItemSlotFinger || bulkSlot == BulkSimItemSlotTrinket {
			pairs := generator.groupedPairsBySlot[bulkSlot]
			slots := BulkSimItemSlotToItemSlotPairs[bulkSlot]
			optionDeltas := make([][]int, 0, len(pairs))
			for _, pair := range pairs {
				optionDeltas = append(optionDeltas, generator.getRequiredSetBonusOptionDeltas(requiredIndexes, [][2]any{{slots[0], &pair[0]}, {slots[1], &pair[1]}}))
			}
			dimensions = append(dimensions, bulkSimRequiredSetBonusDimension{optionDeltas: optionDeltas})
		} else {
			slot := BulkSimItemSlotToSingleItemSlot[bulkSlot]
			optionDeltas := make([][]int, 0, len(options))
			for idx := range options {
				optionDeltas = append(optionDeltas, generator.getRequiredSetBonusOptionDeltas(requiredIndexes, [][2]any{{slot, &options[idx]}}))
			}
			dimensions = append(dimensions, bulkSimRequiredSetBonusDimension{optionDeltas: optionDeltas})
		}
	}

	requiredPieces := make([]int, len(requiredSetBonuses))
	for idx, required := range requiredSetBonuses {
		requiredPieces[idx] = int(required.GetPieces())
	}

	return &bulkSimRequiredSetBonusComboMatcher{baseCounts: baseCounts, requiredPieces: requiredPieces, dimensions: dimensions}
}

func (generator *bulkSimCandidateGenerator) addItemToRequiredSetBonusCounts(counts []int, requiredIndexes map[int32]int, item *core.Item, delta int) {
	if item == nil || item.SetID == 0 {
		return
	}
	idx, ok := requiredIndexes[item.SetID]
	if !ok {
		return
	}
	counts[idx] += delta
}

func (generator *bulkSimCandidateGenerator) getRequiredSetBonusOptionDeltas(requiredIndexes map[int32]int, slotItems [][2]any) []int {
	deltas := make([]int, len(requiredIndexes))
	for _, slotItem := range slotItems {
		slot := slotItem[0].(proto.ItemSlot)
		generator.addItemToRequiredSetBonusCounts(deltas, requiredIndexes, generator.baseEquipment.GetItemBySlot(slot), -1)
		switch option := slotItem[1].(type) {
		case *bulkSimCandidateOption:
			if option != nil {
				generator.addItemToRequiredSetBonusCounts(deltas, requiredIndexes, &option.item, 1)
			}
		}
	}
	return deltas
}

func (generator *bulkSimCandidateGenerator) comboMatchesRequiredSetBonusMatcher(comboIdx int, matcher *bulkSimRequiredSetBonusComboMatcher, scratchCounts []int) bool {
	if matcher == nil {
		return true
	}

	counts := scratchCounts
	if len(counts) != len(matcher.baseCounts) {
		counts = make([]int, len(matcher.baseCounts))
	}
	copy(counts, matcher.baseCounts)

	for _, dimension := range matcher.dimensions {
		if len(dimension.optionDeltas) == 0 {
			return false
		}

		optionIdx := comboIdx % len(dimension.optionDeltas)
		comboIdx = comboIdx / len(dimension.optionDeltas)

		deltas := dimension.optionDeltas[optionIdx]
		for idx, delta := range deltas {
			counts[idx] += delta
		}
	}

	for idx, count := range counts {
		if count < matcher.requiredPieces[idx] {
			return false
		}
	}

	return true
}

func replaceItem(existing core.Item, option bulkSimCandidateOption) core.Item {
	itemSpec := existing.ToItemSpecProto()
	itemSpec.Id = option.spec.GetId()
	itemSpec.RandomSuffix = option.spec.GetRandomSuffix()

	if !enchantAppliesToItem(itemSpec.GetEnchant(), option.item) {
		itemSpec.Enchant = 0
	}
	itemSpec.Gems = reorganizeGems(existing, option.item)

	return core.NewItem(core.ItemSpec{
		ID:           itemSpec.GetId(),
		RandomSuffix: itemSpec.GetRandomSuffix(),
		Enchant:      itemSpec.GetEnchant(),
		Gems:         slices.Clone(itemSpec.GetGems()),
	})
}

func createSelectedItem(option bulkSimCandidateOption) core.Item {
	return core.NewItem(core.ItemSpec{
		ID:           option.spec.GetId(),
		RandomSuffix: option.spec.GetRandomSuffix(),
		Enchant:      option.spec.GetEnchant(),
		Gems:         slices.Clone(option.spec.GetGems()),
	})
}

func reorganizeGems(existing core.Item, newItem core.Item) []int32 {
	newGems := make([]int32, len(newItem.GemSockets))
	for _, gem := range existing.Gems {
		if gem.ID == 0 {
			continue
		}
		firstMatching := -1
		firstEligible := -1
		for socketIdx, socketColor := range newItem.GemSockets {
			if newGems[socketIdx] != 0 {
				continue
			}
			if firstMatching == -1 && gemMatchesSocket(gem.Color, socketColor) {
				firstMatching = socketIdx
			}
			if firstEligible == -1 && gemEligibleForSocket(gem.Color, socketColor) {
				firstEligible = socketIdx
			}
		}
		if firstMatching != -1 {
			newGems[firstMatching] = gem.ID
		} else if firstEligible != -1 {
			newGems[firstEligible] = gem.ID
		}
	}
	return newGems
}

func enchantAppliesToItem(effectID int32, item core.Item) bool {
	if effectID == 0 {
		return false
	}
	enchant := core.GetEnchantByEffectID(effectID)
	if enchant == nil {
		return false
	}
	sharedSlots := sharedSlots(eligibleEnchantSlots(*enchant), getEligibleItemSlots(item, false))
	if len(sharedSlots) == 0 {
		return false
	}
	if enchant.Type == proto.ItemType_ItemTypeRanged {
		return item.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeBow || item.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeCrossbow || item.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeGun
	}
	if item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeUnknown && item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeWand && enchant.Type != proto.ItemType_ItemTypeRanged {
		return false
	}
	return true
}

func eligibleEnchantSlots(enchant core.Enchant) []proto.ItemSlot {
	if slots, ok := bulkSimItemTypeToSlots[enchant.Type]; ok {
		return slots
	}
	if enchant.Type == proto.ItemType_ItemTypeWeapon {
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}
	}
	return nil
}

func sharedSlots(left []proto.ItemSlot, right []proto.ItemSlot) []proto.ItemSlot {
	shared := make([]proto.ItemSlot, 0, min(len(left), len(right)))
	for _, slot := range left {
		if slices.Contains(right, slot) {
			shared = append(shared, slot)
		}
	}
	return shared
}

func gemMatchesSocket(gemColor proto.GemColor, socketColor proto.GemColor) bool {
	if gemColor == socketColor {
		return true
	}
	switch socketColor {
	case proto.GemColor_GemColorMeta:
		return gemColor == proto.GemColor_GemColorMeta
	case proto.GemColor_GemColorBlue:
		return gemColor == proto.GemColor_GemColorBlue || gemColor == proto.GemColor_GemColorPurple || gemColor == proto.GemColor_GemColorGreen || gemColor == proto.GemColor_GemColorPrismatic
	case proto.GemColor_GemColorRed:
		return gemColor == proto.GemColor_GemColorRed || gemColor == proto.GemColor_GemColorPurple || gemColor == proto.GemColor_GemColorOrange || gemColor == proto.GemColor_GemColorPrismatic
	case proto.GemColor_GemColorYellow:
		return gemColor == proto.GemColor_GemColorYellow || gemColor == proto.GemColor_GemColorOrange || gemColor == proto.GemColor_GemColorGreen || gemColor == proto.GemColor_GemColorPrismatic
	case proto.GemColor_GemColorPrismatic:
		return gemColor != proto.GemColor_GemColorMeta
	default:
		return false
	}
}

func gemEligibleForSocket(gemColor proto.GemColor, socketColor proto.GemColor) bool {
	switch socketColor {
	case proto.GemColor_GemColorMeta:
		return gemColor == proto.GemColor_GemColorMeta
	default:
		return gemColor != proto.GemColor_GemColorMeta
	}
}

func getEligibleItemSlots(item core.Item, isFuryWarrior bool) []proto.ItemSlot {
	if slots, ok := bulkSimItemTypeToSlots[item.Type]; ok {
		return slots
	}
	if item.Type == proto.ItemType_ItemTypeWeapon {
		if isFuryWarrior {
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}
		}
		switch item.HandType {
		case proto.HandType_HandTypeMainHand:
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}
		case proto.HandType_HandTypeOffHand:
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}
		default:
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}
		}
	}
	return nil
}

func canEquipItem(item core.Item, playerClass proto.Class, playerSpec proto.Spec, slot proto.ItemSlot) bool {
	if item.Type == proto.ItemType_ItemTypeFinger || item.Type == proto.ItemType_ItemTypeTrinket {
		return true
	}
	if item.Type == proto.ItemType_ItemTypeWeapon {
		eligibleWeaponTypes := core.ClassWeaponTypeCapabilities[playerClass]
		eligibleWeaponType, ok := eligibleWeaponTypes[item.WeaponType]
		if !ok {
			return false
		}
		if (item.HandType == proto.HandType_HandTypeOffHand || (item.HandType == proto.HandType_HandTypeOneHand && slot == proto.ItemSlot_ItemSlotOffHand)) && item.WeaponType != proto.WeaponType_WeaponTypeShield && item.WeaponType != proto.WeaponType_WeaponTypeOffHand && !core.SpecCanDualWieldCapabilities[playerSpec] {
			return false
		}
		if item.HandType == proto.HandType_HandTypeTwoHand && !eligibleWeaponType.CanUseTwoHand {
			return false
		}
		if item.HandType == proto.HandType_HandTypeTwoHand && slot == proto.ItemSlot_ItemSlotOffHand {
			return false
		}
		return true
	}
	if item.Type == proto.ItemType_ItemTypeRanged {
		return slices.Contains(core.ClassRangedWeaponTypeCapabilities[playerClass], item.RangedWeaponType)
	}
	classArmorTypes := core.ClassArmorTypeCapabilities[playerClass]
	if len(classArmorTypes) == 0 {
		return false
	}
	maxArmorType := classArmorTypes[0]
	return maxArmorType >= item.ArmorType
}

func isSecondaryItemSlot(slot proto.ItemSlot, playerCanDualWield bool) bool {
	return slot == proto.ItemSlot_ItemSlotFinger2 || slot == proto.ItemSlot_ItemSlotTrinket2 || (playerCanDualWield && slot == proto.ItemSlot_ItemSlotOffHand)
}

func GetBulkSimItemSlotFromSlot(slot proto.ItemSlot, playerCanDualWield bool) BulkSimItemSlot {
	return getBulkItemSlotFromSlot(slot, playerCanDualWield)
}

func getBulkItemSlotFromSlot(slot proto.ItemSlot, playerCanDualWield bool) BulkSimItemSlot {
	if playerCanDualWield && (slot == proto.ItemSlot_ItemSlotMainHand || slot == proto.ItemSlot_ItemSlotOffHand) {
		return BulkSimItemSlotHandWeapon
	}
	for bulkSlot, singleSlot := range BulkSimItemSlotToSingleItemSlot {
		if singleSlot == slot {
			return bulkSlot
		}
	}
	if slot == proto.ItemSlot_ItemSlotFinger1 || slot == proto.ItemSlot_ItemSlotFinger2 {
		return BulkSimItemSlotFinger
	}
	if slot == proto.ItemSlot_ItemSlotTrinket1 || slot == proto.ItemSlot_ItemSlotTrinket2 {
		return BulkSimItemSlotTrinket
	}
	return BulkSimItemSlotHead
}

func optionsContainEquivalent(options []bulkSimCandidateOption, target bulkSimCandidateOption) bool {
	for _, option := range options {
		if candidateOptionsEqual(option, target) {
			return true
		}
	}
	return false
}

func candidateOptionsEqual(left bulkSimCandidateOption, right bulkSimCandidateOption) bool {
	return itemSpecKey(left.spec) == itemSpecKey(right.spec)
}

func candidateOptionEqualsItem(option bulkSimCandidateOption, item core.Item) bool {
	return itemSpecKey(option.spec) == itemSpecKey(item.ToItemSpecProto())
}

func candidateOptionEqualsItemPtr(option *bulkSimCandidateOption, item *core.Item) bool {
	if option == nil || item == nil {
		return option == nil && item == nil
	}
	return candidateOptionEqualsItem(*option, *item)
}

func itemSpecKey(itemSpec *proto.ItemSpec) string {
	if itemSpec == nil {
		return ""
	}
	return fmt.Sprintf("%d:%d", itemSpec.GetId(), itemSpec.GetRandomSuffix())
}

func getPlayerSpec(player *proto.Player) (proto.Spec, error) {
	switch {
	case player.GetBalanceDruid() != nil:
		return proto.Spec_SpecBalanceDruid, nil
	case player.GetFeralCatDruid() != nil:
		return proto.Spec_SpecFeralCatDruid, nil
	case player.GetFeralBearDruid() != nil:
		return proto.Spec_SpecFeralBearDruid, nil
	case player.GetRestorationDruid() != nil:
		return proto.Spec_SpecRestorationDruid, nil
	case player.GetHunter() != nil:
		return proto.Spec_SpecHunter, nil
	case player.GetMage() != nil:
		return proto.Spec_SpecMage, nil
	case player.GetHolyPaladin() != nil:
		return proto.Spec_SpecHolyPaladin, nil
	case player.GetProtectionPaladin() != nil:
		return proto.Spec_SpecProtectionPaladin, nil
	case player.GetRetributionPaladin() != nil:
		return proto.Spec_SpecRetributionPaladin, nil
	case player.GetPriest() != nil:
		return proto.Spec_SpecPriest, nil
	case player.GetRogue() != nil:
		return proto.Spec_SpecRogue, nil
	case player.GetElementalShaman() != nil:
		return proto.Spec_SpecElementalShaman, nil
	case player.GetEnhancementShaman() != nil:
		return proto.Spec_SpecEnhancementShaman, nil
	case player.GetRestorationShaman() != nil:
		return proto.Spec_SpecRestorationShaman, nil
	case player.GetWarlock() != nil:
		return proto.Spec_SpecWarlock, nil
	case player.GetDpsWarrior() != nil:
		return proto.Spec_SpecDpsWarrior, nil
	case player.GetProtectionWarrior() != nil:
		return proto.Spec_SpecProtectionWarrior, nil
	default:
		return proto.Spec_SpecUnknown, fmt.Errorf("unsupported player spec for backend bulk candidate generation")
	}
}

func getPlayer(request *proto.BulkSimRequest) (*proto.Player, error) {
	if request == nil || request.GetBaseRequest() == nil || request.GetBaseRequest().GetRaid() == nil {
		return nil, fmt.Errorf("bulk request is missing base raid")
	}
	parties := request.GetBaseRequest().GetRaid().GetParties()
	if len(parties) == 0 || parties[0] == nil {
		return nil, fmt.Errorf("bulk request raid is missing parties")
	}
	players := parties[0].GetPlayers()
	if len(players) == 0 || players[0] == nil {
		return nil, fmt.Errorf("bulk request raid is missing player")
	}
	return players[0], nil
}
