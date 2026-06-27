package bulk

import (
	"testing"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

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

func TestReorganizeGems_PersistsHeadMetaAndReassignsOtherGems(t *testing.T) {
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

	gems := reorganizeGems(existing, newItem)
	if len(gems) != 2 {
		t.Fatalf("expected 2 gem slots, got %d", len(gems))
	}
	if gems[0] != 1001 {
		t.Fatalf("expected meta gem to persist in meta socket, got %d", gems[0])
	}
	if gems[1] != 1002 {
		t.Fatalf("expected non-meta gem to be reassigned to eligible socket, got %d", gems[1])
	}
}

func TestReorganizeGems_KeepsNonHeadGems(t *testing.T) {
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

	gems := reorganizeGems(existing, newItem)
	if len(gems) != 1 {
		t.Fatalf("expected 1 gem slot, got %d", len(gems))
	}
	if gems[0] != 2001 {
		t.Fatalf("expected non-head gem to be preserved in matching socket, got %d", gems[0])
	}
}
