package bulk

import (
	"slices"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func replaceItem(existing core.Item, option bulkSimCandidateOption) core.Item {
	itemSpec := existing.ToItemSpecProto()
	itemSpec.Id = option.spec.GetId()
	itemSpec.RandomSuffix = option.spec.GetRandomSuffix()

	if !enchantAppliesToItem(itemSpec.GetEnchant(), option.item) {
		itemSpec.Enchant = 0
	}
	itemSpec.Gems = applyMetaGem(existing, option.item)

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

// Gems are NOT carried over to the replacing item: the gem/reforge pre-pass re-gems every
// candidate anyway, and inheriting the old item's gems would only guess at socket colours it
// cannot satisfy. The head META gem is the exception - it is preserved across a head swap,
// matching what clearGems keeps in the reforge optimizer.
//
// Mirrors MoP. NOTE the shared consequence: specs with no ReforgeOptimizer (druid/restoration,
// paladin/holy, shaman/restoration here; five specs in MoP) get no pre-pass, so their bulk
// candidates are simmed with only the head meta gem.
func applyMetaGem(item core.Item, newItem core.Item) []int32 {
	newGems := make([]int32, len(newItem.GemSockets))

	if item.Type != proto.ItemType_ItemTypeHead || newItem.Type != proto.ItemType_ItemTypeHead {
		return newGems
	}

	metaGemID := int32(0)
	for _, gem := range item.Gems {
		if gem.ID != 0 && gem.Color == proto.GemColor_GemColorMeta {
			metaGemID = gem.ID
			break
		}
	}
	if metaGemID == 0 {
		return newGems
	}

	for socketIdx, socketColor := range newItem.GemSockets {
		if socketColor == proto.GemColor_GemColorMeta {
			newGems[socketIdx] = metaGemID
			break
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
	if !core.CheckSliceOverlap(getEligibleEnchantSlots(*enchant), getEligibleItemSlots(item, false)) {
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

func getEligibleEnchantSlots(enchant core.Enchant) []proto.ItemSlot {
	types := append([]proto.ItemType{enchant.Type}, enchant.ExtraTypes...)
	slots := make([]proto.ItemSlot, 0, len(types)*2)
	for _, itemType := range types {
		if typeSlots, ok := itemTypeToSlotsMap[itemType]; ok {
			slots = append(slots, typeSlots...)
			continue
		}
		if itemType == proto.ItemType_ItemTypeWeapon {
			slots = append(slots, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand)
		}
	}
	return slots
}

func getEligibleItemSlots(item core.Item, isFuryWarrior bool) []proto.ItemSlot {
	if slots, ok := itemTypeToSlotsMap[item.Type]; ok {
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
