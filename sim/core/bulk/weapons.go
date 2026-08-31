package bulk

import (
	"slices"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

const (
	adamantiteSharpeningStoneID int32 = 29453
	adamantiteWeightstoneID     int32 = 34340
)

func isSharpWeaponType(wt proto.WeaponType) bool {
	switch wt {
	case proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypePolearm, proto.WeaponType_WeaponTypeSword:
		return true
	default:
		return false
	}
}

func isBluntWeaponType(wt proto.WeaponType) bool {
	switch wt {
	case proto.WeaponType_WeaponTypeFist, proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeStaff:
		return true
	default:
		return false
	}
}

// adjustWeaponImbueID rewrites the Adamantite sharpening/weightstone pair to match the equipped
// weapon type; all other imbue ids pass through unchanged. Returns 0 when neither stone family is
// valid for the weapon (no weapon, or a shield / offhand-only item).
func adjustWeaponImbueID(imbueID int32, weapon *proto.ItemSpec) int32 {
	if imbueID != adamantiteSharpeningStoneID && imbueID != adamantiteWeightstoneID {
		return imbueID
	}
	if weapon == nil || weapon.Id == 0 {
		return 0
	}
	item := core.GetItemByID(weapon.Id)
	if item == nil {
		return 0
	}
	if isSharpWeaponType(item.WeaponType) {
		return adamantiteSharpeningStoneID
	}
	if isBluntWeaponType(item.WeaponType) {
		return adamantiteWeightstoneID
	}
	return 0
}

// adjustCandidateImbues keeps the MH/OH weapon stone imbues in sync with the candidate's equipped
// weapon types, mirroring the frontend auto-switch so bulk sim combos use the correct stone.
func adjustCandidateImbues(player *proto.Player) {
	consumables := player.GetConsumables()
	if consumables == nil {
		return
	}
	if consumables.MhImbueId == 0 && consumables.OhImbueId == 0 {
		return
	}
	items := player.GetEquipment().GetItems()
	var mhWeapon, ohWeapon *proto.ItemSpec
	if int(proto.ItemSlot_ItemSlotMainHand) < len(items) {
		mhWeapon = items[proto.ItemSlot_ItemSlotMainHand]
	}
	if int(proto.ItemSlot_ItemSlotOffHand) < len(items) {
		ohWeapon = items[proto.ItemSlot_ItemSlotOffHand]
	}
	consumables.MhImbueId = adjustWeaponImbueID(consumables.MhImbueId, mhWeapon)
	consumables.OhImbueId = adjustWeaponImbueID(consumables.OhImbueId, ohWeapon)
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

	for i := range all2HWeapons {
		allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&all2HWeapons[i], nil})
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
		unique := make([]bulkSimCandidateOption, 0, len(oneHandOptions))
		for _, option := range oneHandOptions {
			if optionsContainEquivalent(all2HWeapons, option) {
				continue
			}
			if optionsContainEquivalent(unique, option) {
				continue
			}
			unique = append(unique, option)
		}

		for i := range unique {
			iCanMH := unique[i].item.HandType != proto.HandType_HandTypeOffHand
			iCanOH := unique[i].item.HandType != proto.HandType_HandTypeMainHand
			// Only wield the same 1H weapon in both hands when two copies exist and the weapon
			// itself allows it (not unique, no limit category).
			if generator.hasTwoCopies(unique[i], unique[i].item) && iCanMH && iCanOH {
				allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&unique[i], &unique[i]})
			}
			for j := i + 1; j < len(unique); j++ {
				jCanMH := unique[j].item.HandType != proto.HandType_HandTypeOffHand
				jCanOH := unique[j].item.HandType != proto.HandType_HandTypeMainHand
				if iCanMH && jCanOH {
					allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&unique[i], &unique[j]})
				}
				if jCanMH && iCanOH {
					allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&unique[j], &unique[i]})
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
	frozenWeaponItem := generator.frozenWeaponItem
	if generator.frozenWeaponSlot == proto.ItemSlot_ItemSlotMainHand && frozenWeaponItem != nil && !candidateOptionEqualsItemPtr(mhItem, frozenWeaponItem) {
		return false
	}
	if generator.frozenWeaponSlot == proto.ItemSlot_ItemSlotOffHand && frozenWeaponItem != nil && !candidateOptionEqualsItemPtr(ohItem, frozenWeaponItem) {
		return false
	}
	return generator.matchesWeaponTypeFilter(mhItem, proto.ItemSlot_ItemSlotMainHand) && generator.matchesWeaponTypeFilter(ohItem, proto.ItemSlot_ItemSlotOffHand)
}
