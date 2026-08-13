package bulk

import (
	"testing"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

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
			generator.initWeaponCopyCounts()

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
	generator.initWeaponCopyCounts()

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
	generator.initWeaponCopyCounts()

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
	generator.initWeaponCopyCounts()

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
