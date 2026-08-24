package bulk

import (
	"fmt"
	"sort"
	"testing"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

// Enchant, gem and item-rule helpers.

func addBulkTestEnchant(effectID int32, itemType proto.ItemType) {
	core.AddToDatabase(&proto.SimDatabase{
		Enchants: []*proto.SimEnchant{
			{
				EffectId: effectID,
				Type:     itemType,
			},
		},
	})
}

func TestBulkSimEnchantAppliesToItem_UsesWeaponTypeRules(t *testing.T) {
	weaponEffectID := int32(910001)

	addBulkTestEnchant(weaponEffectID, proto.ItemType_ItemTypeWeapon)

	twoHandSword := core.Item{
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeSword,
		HandType:   proto.HandType_HandTypeTwoHand,
	}
	oneHandSword := core.Item{
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeSword,
		HandType:   proto.HandType_HandTypeOneHand,
	}
	staff := core.Item{
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeStaff,
		HandType:   proto.HandType_HandTypeTwoHand,
	}
	shield := core.Item{
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeShield,
		HandType:   proto.HandType_HandTypeOffHand,
	}
	offHand := core.Item{
		Type:       proto.ItemType_ItemTypeWeapon,
		WeaponType: proto.WeaponType_WeaponTypeOffHand,
		HandType:   proto.HandType_HandTypeOffHand,
	}

	if !enchantAppliesToItem(weaponEffectID, twoHandSword) {
		t.Fatalf("expected weapon enchant to apply to two-handed weapon")
	}
	if !enchantAppliesToItem(weaponEffectID, oneHandSword) {
		t.Fatalf("expected weapon enchant to apply to one-handed weapon")
	}
	if !enchantAppliesToItem(weaponEffectID, staff) {
		t.Fatalf("expected weapon enchant to apply to staff")
	}
	if !enchantAppliesToItem(weaponEffectID, shield) {
		t.Fatalf("expected weapon enchant to apply to shield off-hand slot")
	}
	if !enchantAppliesToItem(weaponEffectID, offHand) {
		t.Fatalf("expected weapon enchant to apply to off-hand frill")
	}
}

func TestBulkSimEnchantAppliesToItem_UsesTypedRangedRules(t *testing.T) {
	rangedEffectID := int32(910005)
	weaponEffectID := int32(910006)

	addBulkTestEnchant(rangedEffectID, proto.ItemType_ItemTypeRanged)
	addBulkTestEnchant(weaponEffectID, proto.ItemType_ItemTypeWeapon)

	bow := core.Item{
		Type:             proto.ItemType_ItemTypeRanged,
		RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow,
	}
	wand := core.Item{
		Type:             proto.ItemType_ItemTypeRanged,
		RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
	}
	gun := core.Item{
		Type:             proto.ItemType_ItemTypeRanged,
		RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeGun,
	}

	if !enchantAppliesToItem(rangedEffectID, bow) {
		t.Fatalf("expected ranged enchant to apply to bow")
	}
	if enchantAppliesToItem(rangedEffectID, wand) {
		t.Fatalf("expected ranged enchant to not apply to wand")
	}
	if enchantAppliesToItem(weaponEffectID, gun) {
		t.Fatalf("expected non-ranged enchant to not apply to non-wand ranged weapon")
	}
}

func TestBulkSimEnchantAppliesToItem_RejectsNonMatchingItemTypes(t *testing.T) {
	extraTypeEffectID := int32(910007)
	addBulkTestEnchant(extraTypeEffectID, proto.ItemType_ItemTypeChest)

	wrist := core.Item{Type: proto.ItemType_ItemTypeWrist}
	legs := core.Item{Type: proto.ItemType_ItemTypeLegs}
	chest := core.Item{Type: proto.ItemType_ItemTypeChest}

	if !enchantAppliesToItem(extraTypeEffectID, chest) {
		t.Fatalf("expected enchant to apply to matching item type")
	}
	if enchantAppliesToItem(extraTypeEffectID, wrist) {
		t.Fatalf("expected chest enchant to not apply to wrist")
	}
	if enchantAppliesToItem(extraTypeEffectID, legs) {
		t.Fatalf("expected enchant to not apply to unrelated item type")
	}
}

func TestMergeGems_RehomesReplacedItemGems(t *testing.T) {
	existing := core.Item{
		Type:       proto.ItemType_ItemTypeHead,
		GemSockets: []proto.GemColor{proto.GemColor_GemColorMeta, proto.GemColor_GemColorRed},
		Gems: []core.Gem{
			{ID: 1001, Color: proto.GemColor_GemColorMeta},
			{ID: 1002, Color: proto.GemColor_GemColorRed},
		},
	}
	newItem := core.Item{
		Type:       proto.ItemType_ItemTypeHead,
		GemSockets: []proto.GemColor{proto.GemColor_GemColorMeta, proto.GemColor_GemColorBlue},
	}

	gems := mergeGems(existing, bulkSimCandidateOption{spec: &proto.ItemSpec{}}, newItem)
	if len(gems) != 2 {
		t.Fatalf("expected 2 gem slots, got %d", len(gems))
	}
	if gems[0] != 1001 {
		t.Fatalf("expected meta gem to persist in meta socket, got %d", gems[0])
	}
	if gems[1] != 1002 {
		t.Fatalf("expected red gem to be re-homed into the blue socket, got %d", gems[1])
	}
}

func TestMergeGems_KeepsNonHeadGems(t *testing.T) {
	existing := core.Item{
		Type:       proto.ItemType_ItemTypeHands,
		GemSockets: []proto.GemColor{proto.GemColor_GemColorRed},
		Gems: []core.Gem{
			{ID: 2001, Color: proto.GemColor_GemColorRed},
		},
	}
	newItem := core.Item{
		Type:       proto.ItemType_ItemTypeHands,
		GemSockets: []proto.GemColor{proto.GemColor_GemColorRed},
	}

	gems := mergeGems(existing, bulkSimCandidateOption{spec: &proto.ItemSpec{}}, newItem)
	if len(gems) != 1 {
		t.Fatalf("expected 1 gem slot, got %d", len(gems))
	}
	if gems[0] != 2001 {
		t.Fatalf("expected gem to be carried over into the matching socket, got %d", gems[0])
	}
}

func TestMergeGems_MetaGemNeedsMetaSocket(t *testing.T) {
	existing := core.Item{
		Type:       proto.ItemType_ItemTypeHead,
		GemSockets: []proto.GemColor{proto.GemColor_GemColorMeta},
		Gems:       []core.Gem{{ID: 3001, Color: proto.GemColor_GemColorMeta}},
	}
	newItem := core.Item{
		Type:       proto.ItemType_ItemTypeHead,
		GemSockets: []proto.GemColor{proto.GemColor_GemColorRed},
	}

	gems := mergeGems(existing, bulkSimCandidateOption{spec: &proto.ItemSpec{}}, newItem)
	if gems[0] != 0 {
		t.Fatalf("expected meta gem to be dropped when there is no meta socket, got %d", gems[0])
	}
}

func TestMergeGems_BulkItemGemsWin(t *testing.T) {
	existing := core.Item{
		Type:       proto.ItemType_ItemTypeHands,
		GemSockets: []proto.GemColor{proto.GemColor_GemColorRed},
		Gems:       []core.Gem{{ID: 4001, Color: proto.GemColor_GemColorRed}},
	}
	newItem := core.Item{
		Type:       proto.ItemType_ItemTypeHands,
		GemSockets: []proto.GemColor{proto.GemColor_GemColorRed},
	}
	option := bulkSimCandidateOption{spec: &proto.ItemSpec{Gems: []int32{4002}}}

	gems := mergeGems(existing, option, newItem)
	if gems[0] != 4002 {
		t.Fatalf("expected the gem picked for the bulk item to win, got %d", gems[0])
	}
}

func oneHandOption(id int32, gems ...int32) bulkSimCandidateOption {
	return bulkSimCandidateOption{
		spec: &proto.ItemSpec{Id: id, Gems: gems},
		item: core.Item{
			ID:         id,
			Type:       proto.ItemType_ItemTypeWeapon,
			HandType:   proto.HandType_HandTypeOneHand,
			WeaponType: proto.WeaponType_WeaponTypeSword,
			Gems:       make([]core.Gem, len(gems)),
		},
	}
}

func countSameWeaponBothHands(combos [][2]*bulkSimCandidateOption, id int32) int {
	matches := 0
	for _, combo := range combos {
		if combo[0] == nil || combo[1] == nil {
			continue
		}
		if combo[0].spec.GetId() == id && combo[1].spec.GetId() == id {
			matches++
		}
	}
	return matches
}

// Two copies of the same 1H weapon (one equipped, one user-added) must allow wielding it in
// both hands; a single copy must not. This pins the dual-wield stacking behaviour so the
// candidate-option dedup cannot silently collapse it.
func TestGetAllWeaponCombosSameWeaponBothHandsNeedsTwoCopies(t *testing.T) {
	testCases := []struct {
		name       string
		copies     int
		wantCombos int
	}{
		{name: "one copy", copies: 1, wantCombos: 0},
		{name: "two copies", copies: 2, wantCombos: 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			options := make([]bulkSimCandidateOption, 0, testCase.copies)
			for range testCase.copies {
				options = append(options, oneHandOption(100))
			}
			generator := &bulkSimCandidateGenerator{
				playerCanDualWield: true,
				selectedByBulkSlot: map[BulkSimItemSlot][]bulkSimCandidateOption{
					BulkSimItemSlotHandWeapon: options,
				},
			}
			generator.initCopyCounts()

			got := countSameWeaponBothHands(generator.getAllWeaponCombos(), 100)
			if got != testCase.wantCombos {
				t.Fatalf("same-weapon-both-hands combos for %d copies = %d, want %d", testCase.copies, got, testCase.wantCombos)
			}
		})
	}
}

// Distinct 1H weapons must still pair in both orders regardless of copy counts.
func TestGetAllWeaponCombosPairsDistinctWeaponsBothOrders(t *testing.T) {
	generator := &bulkSimCandidateGenerator{
		playerCanDualWield: true,
		selectedByBulkSlot: map[BulkSimItemSlot][]bulkSimCandidateOption{
			BulkSimItemSlotHandWeapon: {oneHandOption(100), oneHandOption(200)},
		},
	}
	generator.initCopyCounts()

	forward, backward := 0, 0
	for _, combo := range generator.getAllWeaponCombos() {
		if combo[0] == nil || combo[1] == nil {
			continue
		}
		if combo[0].spec.GetId() == 100 && combo[1].spec.GetId() == 200 {
			forward++
		}
		if combo[0].spec.GetId() == 200 && combo[1].spec.GetId() == 100 {
			backward++
		}
	}
	if forward != 1 || backward != 1 {
		t.Fatalf("distinct weapon pairings: forward=%d backward=%d, want 1 and 1", forward, backward)
	}
}

func addBulkTestWeapon(id int32, handType proto.HandType, weaponType proto.WeaponType) {
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{
			{
				Id:             id,
				Type:           proto.ItemType_ItemTypeWeapon,
				WeaponType:     weaponType,
				HandType:       handType,
				ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
			},
		},
	})
}

// A 2H mainhand combo carries a nil offhand, so the base gear's offhand has to be cleared:
// without the reset the equipped shield stays wielded alongside the 2H weapon.
func TestBuildGearForComboClearsOffHandForTwoHandMainHand(t *testing.T) {
	const twoHandID, oneHandID, shieldID = int32(910201), int32(910202), int32(910203)
	addBulkTestWeapon(twoHandID, proto.HandType_HandTypeTwoHand, proto.WeaponType_WeaponTypeSword)
	addBulkTestWeapon(oneHandID, proto.HandType_HandTypeOneHand, proto.WeaponType_WeaponTypeSword)
	addBulkTestWeapon(shieldID, proto.HandType_HandTypeOffHand, proto.WeaponType_WeaponTypeShield)

	var baseEquipment core.Equipment
	baseEquipment[proto.ItemSlot_ItemSlotMainHand] = core.NewItem(core.ItemSpec{ID: oneHandID})
	baseEquipment[proto.ItemSlot_ItemSlotOffHand] = core.NewItem(core.ItemSpec{ID: shieldID})

	generator := &bulkSimCandidateGenerator{
		baseEquipment: baseEquipment,
		selectedByBulkSlot: map[BulkSimItemSlot][]bulkSimCandidateOption{
			BulkSimItemSlotMainHand: {{
				spec: &proto.ItemSpec{Id: twoHandID},
				item: core.NewItem(core.ItemSpec{ID: twoHandID}),
			}},
		},
	}
	generator.initCopyCounts()

	if combos := generator.getAllWeaponCombos(); len(combos) != 1 {
		t.Fatalf("expected 1 weapon combo, got %d", len(combos))
	}

	gear, err := generator.buildGearForCombo(0)
	if err != nil {
		t.Fatalf("buildGearForCombo failed: %v", err)
	}
	items := gear.GetItems()
	if got := items[proto.ItemSlot_ItemSlotMainHand].GetId(); got != twoHandID {
		t.Fatalf("expected 2H weapon in mainhand, got %d", got)
	}
	if got := items[proto.ItemSlot_ItemSlotOffHand].GetId(); got != 0 {
		t.Fatalf("expected offhand to be cleared for 2H mainhand, got %d", got)
	}
}

// TBC has no Titan's Grip, so a 2H weapon never pairs with an offhand for any spec, DPS
// warriors included.
func TestGetAllWeaponCombosNeverPairsTwoHandWithOffHand(t *testing.T) {
	const firstTwoHandID, secondTwoHandID = int32(910301), int32(910302)
	addBulkTestWeapon(firstTwoHandID, proto.HandType_HandTypeTwoHand, proto.WeaponType_WeaponTypeSword)
	addBulkTestWeapon(secondTwoHandID, proto.HandType_HandTypeTwoHand, proto.WeaponType_WeaponTypeAxe)

	twoHandOption := func(id int32) bulkSimCandidateOption {
		return bulkSimCandidateOption{
			spec: &proto.ItemSpec{Id: id},
			item: core.NewItem(core.ItemSpec{ID: id}),
		}
	}

	generator := &bulkSimCandidateGenerator{
		playerSpec:         proto.Spec_SpecDpsWarrior,
		playerCanDualWield: true,
		selectedByBulkSlot: map[BulkSimItemSlot][]bulkSimCandidateOption{
			BulkSimItemSlotHandWeapon: {twoHandOption(firstTwoHandID), twoHandOption(secondTwoHandID)},
		},
	}
	generator.initCopyCounts()

	combos := generator.getAllWeaponCombos()
	if len(combos) != 2 {
		t.Fatalf("expected 2 weapon combos (one per 2H weapon), got %d", len(combos))
	}
	for _, combo := range combos {
		if combo[1] != nil {
			t.Fatalf("expected nil offhand for 2H combo, got item %d", combo[1].spec.GetId())
		}
	}
}

// A differently-gemmed copy of the same item id is a distinct candidate, so the fingerprint
// key must not collapse it the way an id-only key would.
func TestItemSpecFingerprintKeyDistinguishesGems(t *testing.T) {
	plain := buildItemSpecFingerprintKey(&proto.ItemSpec{Id: 100})
	gemmed := buildItemSpecFingerprintKey(&proto.ItemSpec{Id: 100, Gems: []int32{42}})
	if plain == gemmed {
		t.Fatal("fingerprint key ignored gems")
	}

	reordered := buildItemSpecFingerprintKey(&proto.ItemSpec{Id: 100, Gems: []int32{42, 7}})
	swapped := buildItemSpecFingerprintKey(&proto.ItemSpec{Id: 100, Gems: []int32{7, 42}})
	if reordered == swapped {
		t.Fatal("fingerprint key ignored gem order")
	}
}

// Per-spec mainhand/offhand combination matrix.

// Feeds every spec a pool of synthetic weapons covering all (weaponType, handType) shapes the
// class can equip, then asserts the generated MH/OH pairs are all wearable in game. Run with
// -v to print the per-spec combination table.

var comboTableWeaponTypes = []proto.WeaponType{
	proto.WeaponType_WeaponTypeAxe,
	proto.WeaponType_WeaponTypeDagger,
	proto.WeaponType_WeaponTypeFist,
	proto.WeaponType_WeaponTypeMace,
	proto.WeaponType_WeaponTypeOffHand,
	proto.WeaponType_WeaponTypePolearm,
	proto.WeaponType_WeaponTypeShield,
	proto.WeaponType_WeaponTypeStaff,
	proto.WeaponType_WeaponTypeSword,
}

var comboTableHandTypes = []proto.HandType{
	proto.HandType_HandTypeMainHand,
	proto.HandType_HandTypeOneHand,
	proto.HandType_HandTypeOffHand,
	proto.HandType_HandTypeTwoHand,
}

// (weaponType, handType) pairs that actually occur in assets/database/db.json, so the synthetic
// pool cannot invent shapes like a one-handed off-hand frill.
var comboTableRealPairs = map[proto.WeaponType][]proto.HandType{
	proto.WeaponType_WeaponTypeAxe:     {proto.HandType_HandTypeMainHand, proto.HandType_HandTypeOneHand, proto.HandType_HandTypeOffHand, proto.HandType_HandTypeTwoHand},
	proto.WeaponType_WeaponTypeDagger:  {proto.HandType_HandTypeMainHand, proto.HandType_HandTypeOneHand, proto.HandType_HandTypeOffHand},
	proto.WeaponType_WeaponTypeFist:    {proto.HandType_HandTypeMainHand, proto.HandType_HandTypeOneHand, proto.HandType_HandTypeOffHand},
	proto.WeaponType_WeaponTypeMace:    {proto.HandType_HandTypeMainHand, proto.HandType_HandTypeOneHand, proto.HandType_HandTypeOffHand, proto.HandType_HandTypeTwoHand},
	proto.WeaponType_WeaponTypeOffHand: {proto.HandType_HandTypeOffHand},
	proto.WeaponType_WeaponTypePolearm: {proto.HandType_HandTypeTwoHand},
	proto.WeaponType_WeaponTypeShield:  {proto.HandType_HandTypeOffHand},
	proto.WeaponType_WeaponTypeStaff:   {proto.HandType_HandTypeTwoHand},
	proto.WeaponType_WeaponTypeSword:   {proto.HandType_HandTypeMainHand, proto.HandType_HandTypeOneHand, proto.HandType_HandTypeOffHand, proto.HandType_HandTypeTwoHand},
}

func comboTableItemID(wt proto.WeaponType, ht proto.HandType, copyIdx int) int32 {
	return 990000 + int32(wt)*100 + int32(ht)*10 + int32(copyIdx)
}

func comboTableItemName(wt proto.WeaponType, ht proto.HandType, copyIdx int) string {
	wtNames := map[proto.WeaponType]string{
		proto.WeaponType_WeaponTypeAxe:     "Axe",
		proto.WeaponType_WeaponTypeDagger:  "Dagger",
		proto.WeaponType_WeaponTypeFist:    "Fist",
		proto.WeaponType_WeaponTypeMace:    "Mace",
		proto.WeaponType_WeaponTypeOffHand: "Frill",
		proto.WeaponType_WeaponTypePolearm: "Polearm",
		proto.WeaponType_WeaponTypeShield:  "Shield",
		proto.WeaponType_WeaponTypeStaff:   "Staff",
		proto.WeaponType_WeaponTypeSword:   "Sword",
	}
	htNames := map[proto.HandType]string{
		proto.HandType_HandTypeMainHand: "MHonly",
		proto.HandType_HandTypeOneHand:  "1H",
		proto.HandType_HandTypeOffHand:  "OHonly",
		proto.HandType_HandTypeTwoHand:  "2H",
	}
	return fmt.Sprintf("%s-%s#%d", wtNames[wt], htNames[ht], copyIdx)
}

func registerComboTableItems() {
	items := make([]*proto.SimItem, 0, len(comboTableWeaponTypes)*len(comboTableHandTypes)*2)
	for _, wt := range comboTableWeaponTypes {
		for _, ht := range comboTableRealPairs[wt] {
			for copyIdx := range 2 {
				items = append(items, &proto.SimItem{
					Id:             comboTableItemID(wt, ht, copyIdx),
					Name:           comboTableItemName(wt, ht, copyIdx),
					Type:           proto.ItemType_ItemTypeWeapon,
					WeaponType:     wt,
					HandType:       ht,
					ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
				})
			}
		}
	}
	core.AddToDatabase(&proto.SimDatabase{Items: items})
}

var comboTableSpecs = []struct {
	spec  proto.Spec
	class proto.Class
	label string
}{
	{proto.Spec_SpecBalanceDruid, proto.Class_ClassDruid, "Balance Druid"},
	{proto.Spec_SpecFeralCatDruid, proto.Class_ClassDruid, "Feral Cat Druid"},
	{proto.Spec_SpecFeralBearDruid, proto.Class_ClassDruid, "Feral Bear Druid"},
	{proto.Spec_SpecRestorationDruid, proto.Class_ClassDruid, "Restoration Druid"},
	{proto.Spec_SpecHunter, proto.Class_ClassHunter, "Hunter"},
	{proto.Spec_SpecMage, proto.Class_ClassMage, "Mage"},
	{proto.Spec_SpecHolyPaladin, proto.Class_ClassPaladin, "Holy Paladin"},
	{proto.Spec_SpecProtectionPaladin, proto.Class_ClassPaladin, "Protection Paladin"},
	{proto.Spec_SpecRetributionPaladin, proto.Class_ClassPaladin, "Retribution Paladin"},
	{proto.Spec_SpecPriest, proto.Class_ClassPriest, "Priest"},
	{proto.Spec_SpecRogue, proto.Class_ClassRogue, "Rogue"},
	{proto.Spec_SpecElementalShaman, proto.Class_ClassShaman, "Elemental Shaman"},
	{proto.Spec_SpecEnhancementShaman, proto.Class_ClassShaman, "Enhancement Shaman"},
	{proto.Spec_SpecRestorationShaman, proto.Class_ClassShaman, "Restoration Shaman"},
	{proto.Spec_SpecWarlock, proto.Class_ClassWarlock, "Warlock"},
	{proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, "DPS Warrior"},
	{proto.Spec_SpecProtectionWarrior, proto.Class_ClassWarrior, "Protection Warrior"},
}

// classComboTableItems returns every synthetic weapon the class can actually equip, so the
// selection fed to the generator mirrors what the item search would let through.
func classComboTableItems(class proto.Class) []*proto.ItemSpec {
	specs := make([]*proto.ItemSpec, 0)
	for _, wt := range comboTableWeaponTypes {
		if _, ok := core.ClassWeaponTypeCapabilities[class][wt]; !ok {
			continue
		}
		for _, ht := range comboTableRealPairs[wt] {
			for copyIdx := range 2 {
				specs = append(specs, &proto.ItemSpec{Id: comboTableItemID(wt, ht, copyIdx)})
			}
		}
	}
	return specs
}

func comboLabel(option *bulkSimCandidateOption) string {
	if option == nil {
		return "(keeps base gear)"
	}
	return option.item.Name
}

func TestWeaponComboMatrixPerSpec(t *testing.T) {
	registerComboTableItems()

	for _, testCase := range comboTableSpecs {
		selected := classComboTableItems(testCase.class)
		generator := &bulkSimCandidateGenerator{
			settings:           &proto.BulkSettings{Items: selected},
			playerClass:        testCase.class,
			playerSpec:         testCase.spec,
			playerCanDualWield: core.SpecCanDualWieldCapabilities[testCase.spec],
			baseEquipment:      core.Equipment{},
			selectedByBulkSlot: make(map[BulkSimItemSlot][]bulkSimCandidateOption),
			groupedPairsBySlot: make(map[BulkSimItemSlot][][2]bulkSimCandidateOption),
			frozenItems:        make(map[BulkSimItemSlot]*core.Item),
			weaponTypeFilters:  make(map[proto.ItemSlot][]proto.WeaponType),
		}
		if err := generator.initSelectedItems(); err != nil {
			t.Fatalf("%s: initSelectedItems: %v", testCase.label, err)
		}

		shapes := make(map[string]int)
		illegal := make([]string, 0)
		for _, combo := range generator.getAllWeaponCombos() {
			mh, oh := comboLabel(combo[0]), comboLabel(combo[1])
			shapes[shapeOf(combo)]++
			if isIllegalCombo(combo) {
				illegal = append(illegal, fmt.Sprintf("%s + %s", mh, oh))
			}
		}

		shapeKeys := make([]string, 0, len(shapes))
		for key := range shapes {
			shapeKeys = append(shapeKeys, key)
		}
		sort.Strings(shapeKeys)

		t.Logf("=== %s (dualWield=%v)", testCase.label, generator.playerCanDualWield)
		t.Logf("    bulkSlots: MainHand=%d OffHand=%d HandWeapon=%d",
			len(generator.selectedByBulkSlot[BulkSimItemSlotMainHand]),
			len(generator.selectedByBulkSlot[BulkSimItemSlotOffHand]),
			len(generator.selectedByBulkSlot[BulkSimItemSlotHandWeapon]))
		for _, key := range shapeKeys {
			t.Logf("    %-22s x%d", key, shapes[key])
		}
		if len(illegal) > 0 {
			t.Errorf("%s: %d unwearable combos, e.g. %v", testCase.label, len(illegal), illegal[:min(len(illegal), 8)])
		}
		if !generator.playerCanDualWield {
			for _, combo := range generator.getAllWeaponCombos() {
				if combo[0] != nil && combo[1] != nil && combo[1].item.WeaponType != proto.WeaponType_WeaponTypeShield &&
					combo[1].item.WeaponType != proto.WeaponType_WeaponTypeOffHand {
					t.Errorf("%s cannot dual wield but got %s + %s", testCase.label, comboLabel(combo[0]), comboLabel(combo[1]))
					break
				}
			}
		}
	}
}

func handTypeShape(option *bulkSimCandidateOption) string {
	if option == nil {
		return "none"
	}
	switch option.item.HandType {
	case proto.HandType_HandTypeMainHand:
		return "MHonly"
	case proto.HandType_HandTypeOneHand:
		return "1H"
	case proto.HandType_HandTypeOffHand:
		if option.item.WeaponType == proto.WeaponType_WeaponTypeShield {
			return "Shield"
		}
		return "OHonly"
	case proto.HandType_HandTypeTwoHand:
		return "2H"
	}
	return "unknown"
}

func shapeOf(combo [2]*bulkSimCandidateOption) string {
	return fmt.Sprintf("MH=%s OH=%s", handTypeShape(combo[0]), handTypeShape(combo[1]))
}

// Any pairing WoW itself would reject: a 2H in either hand alongside another weapon, an
// off-hand-only item in the mainhand, or a mainhand-only item in the offhand.
func isIllegalCombo(combo [2]*bulkSimCandidateOption) bool {
	mh, oh := combo[0], combo[1]
	if mh != nil && oh != nil {
		if mh.item.HandType == proto.HandType_HandTypeTwoHand || oh.item.HandType == proto.HandType_HandTypeTwoHand {
			return true
		}
	}
	if mh != nil && mh.item.HandType == proto.HandType_HandTypeOffHand {
		return true
	}
	if oh != nil && oh.item.HandType == proto.HandType_HandTypeMainHand {
		return true
	}
	return false
}

// Per-spec paired-slot (ring / trinket / both hands) matrix.

// Covers the paired bulk slots (ring, trinket, both hands) for every spec: which same-slot
// pairings the candidate generator offers when the batch holds two copies of one item, two items
// sharing a limit category, unique items, and plain distinct items. Run with -v to print the
// table.

const (
	pairRingPlainID   int32 = 992001
	pairRingOtherID   int32 = 992002
	pairRingUniqueID  int32 = 992003
	pairRingLimitAID  int32 = 992004
	pairRingLimitBID  int32 = 992005
	pairTrinketPlain  int32 = 992011
	pairTrinketOther  int32 = 992012
	pairTrinketUnique int32 = 992013
	pairWeaponPlainID int32 = 992021
	pairWeaponUniqID  int32 = 992022
)

const pairLimitCategory int32 = 4242

func registerPairedSlotItems() {
	item := func(id int32, itemType proto.ItemType, unique bool, limitCategory int32) *proto.SimItem {
		return &proto.SimItem{
			Id:             id,
			Name:           fmt.Sprintf("PairTest-%d", id),
			Type:           itemType,
			Unique:         unique,
			LimitCategory:  limitCategory,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
		}
	}
	weapon := func(id int32, unique bool) *proto.SimItem {
		simItem := item(id, proto.ItemType_ItemTypeWeapon, unique, 0)
		simItem.WeaponType = proto.WeaponType_WeaponTypeDagger
		simItem.HandType = proto.HandType_HandTypeOneHand
		return simItem
	}

	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
		item(pairRingPlainID, proto.ItemType_ItemTypeFinger, false, 0),
		item(pairRingOtherID, proto.ItemType_ItemTypeFinger, false, 0),
		item(pairRingUniqueID, proto.ItemType_ItemTypeFinger, true, 0),
		item(pairRingLimitAID, proto.ItemType_ItemTypeFinger, false, pairLimitCategory),
		item(pairRingLimitBID, proto.ItemType_ItemTypeFinger, false, pairLimitCategory),
		item(pairTrinketPlain, proto.ItemType_ItemTypeTrinket, false, 0),
		item(pairTrinketOther, proto.ItemType_ItemTypeTrinket, false, 0),
		item(pairTrinketUnique, proto.ItemType_ItemTypeTrinket, true, 0),
		weapon(pairWeaponPlainID, false),
		weapon(pairWeaponUniqID, true),
	}})
}

// Two copies of everything that may legitimately be doubled, one copy of the rest.
func pairedSlotSelection() []*proto.ItemSpec {
	ids := []int32{
		pairRingPlainID, pairRingPlainID,
		pairRingOtherID,
		pairRingUniqueID, pairRingUniqueID,
		pairRingLimitAID, pairRingLimitBID,
		pairTrinketPlain, pairTrinketPlain,
		pairTrinketOther,
		pairTrinketUnique, pairTrinketUnique,
		pairWeaponPlainID, pairWeaponPlainID,
		pairWeaponUniqID, pairWeaponUniqID,
	}
	specs := make([]*proto.ItemSpec, 0, len(ids))
	for _, id := range ids {
		specs = append(specs, &proto.ItemSpec{Id: id})
	}
	return specs
}

type pairKey struct {
	first  int32
	second int32
}

func newPairKey(first int32, second int32) pairKey {
	if second < first {
		first, second = second, first
	}
	return pairKey{first: first, second: second}
}

func TestPairedSlotMatrixPerSpec(t *testing.T) {
	registerPairedSlotItems()

	cases := []struct {
		label string
		pair  pairKey
		want  bool
	}{
		{label: "ring: two copies of a plain item", pair: newPairKey(pairRingPlainID, pairRingPlainID), want: true},
		{label: "ring: one copy of a plain item", pair: newPairKey(pairRingOtherID, pairRingOtherID), want: false},
		{label: "ring: two copies of a unique item", pair: newPairKey(pairRingUniqueID, pairRingUniqueID), want: false},
		{label: "ring: two items sharing a limit category", pair: newPairKey(pairRingLimitAID, pairRingLimitBID), want: false},
		{label: "ring: two distinct plain items", pair: newPairKey(pairRingPlainID, pairRingOtherID), want: true},
		{label: "ring: distinct plain + unique items", pair: newPairKey(pairRingPlainID, pairRingUniqueID), want: true},
		{label: "trinket: two copies of a plain item", pair: newPairKey(pairTrinketPlain, pairTrinketPlain), want: true},
		{label: "trinket: two copies of a unique item", pair: newPairKey(pairTrinketUnique, pairTrinketUnique), want: false},
		{label: "trinket: two distinct plain items", pair: newPairKey(pairTrinketPlain, pairTrinketOther), want: true},
	}

	for _, testCase := range comboTableSpecs {
		generator := &bulkSimCandidateGenerator{
			settings:           &proto.BulkSettings{Items: pairedSlotSelection()},
			playerClass:        testCase.class,
			playerSpec:         testCase.spec,
			playerCanDualWield: core.SpecCanDualWieldCapabilities[testCase.spec],
			baseEquipment:      core.Equipment{},
			selectedByBulkSlot: make(map[BulkSimItemSlot][]bulkSimCandidateOption),
			groupedPairsBySlot: make(map[BulkSimItemSlot][][2]bulkSimCandidateOption),
			frozenItems:        make(map[BulkSimItemSlot]*core.Item),
			weaponTypeFilters:  make(map[proto.ItemSlot][]proto.WeaponType),
		}
		if err := generator.initSelectedItems(); err != nil {
			t.Fatalf("%s: initSelectedItems: %v", testCase.label, err)
		}
		generator.initGroupedSlotPairs()

		got := make(map[pairKey]bool)
		for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotFinger, BulkSimItemSlotTrinket} {
			for _, pair := range generator.groupedPairsBySlot[bulkSlot] {
				got[newPairKey(pair[0].spec.GetId(), pair[1].spec.GetId())] = true
			}
		}
		// Both hands are a paired slot too, just built by getAllWeaponCombos instead.
		weaponPairs := make(map[pairKey]bool)
		for _, combo := range generator.getAllWeaponCombos() {
			if combo[0] == nil || combo[1] == nil {
				continue
			}
			weaponPairs[newPairKey(combo[0].spec.GetId(), combo[1].spec.GetId())] = true
		}

		t.Logf("=== %s (dualWield=%v)", testCase.label, generator.playerCanDualWield)
		for _, expectation := range cases {
			if got[expectation.pair] != expectation.want {
				t.Errorf("%s: %s = %v, want %v", testCase.label, expectation.label, got[expectation.pair], expectation.want)
			}
			t.Logf("    %-46s %v", expectation.label, got[expectation.pair])
		}

		// A weapon may fill both hands only for a dual-wielder, and never when it is unique.
		wantSameWeapon := generator.playerCanDualWield
		samePlain := weaponPairs[newPairKey(pairWeaponPlainID, pairWeaponPlainID)]
		sameUnique := weaponPairs[newPairKey(pairWeaponUniqID, pairWeaponUniqID)]
		if samePlain != wantSameWeapon {
			t.Errorf("%s: two copies of a plain 1H weapon in both hands = %v, want %v", testCase.label, samePlain, wantSameWeapon)
		}
		if sameUnique {
			t.Errorf("%s: two copies of a unique 1H weapon wielded in both hands", testCase.label)
		}
		t.Logf("    %-46s %v", "weapon: two copies of a plain 1H item", samePlain)
		t.Logf("    %-46s %v", "weapon: two copies of a unique 1H item", sameUnique)

		pairCounts := make([]string, 0, 3)
		for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotFinger, BulkSimItemSlotTrinket} {
			pairCounts = append(pairCounts, fmt.Sprintf("%s=%d", BulkSimItemSlotNames[bulkSlot], len(generator.groupedPairsBySlot[bulkSlot])))
		}
		sort.Strings(pairCounts)
		t.Logf("    pairs: %v, weaponCombos=%d", pairCounts, len(generator.getAllWeaponCombos()))
	}
}
