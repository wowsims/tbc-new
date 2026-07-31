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
