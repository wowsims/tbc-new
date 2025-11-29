package dbc

import "github.com/wowsims/tbc/sim/core/proto"

func MapResistanceToStat(index int) (proto.Stat, bool) {
	switch index {
	case 0:
		return proto.Stat_StatBonusArmor, true
	}
	return proto.Stat_StatBonusArmor, false
}

var MapArmorSubclassToArmorType = map[int]proto.ArmorType{
	ITEM_SUBCLASS_ARMOR_CLOTH:   proto.ArmorType_ArmorTypeCloth,
	ITEM_SUBCLASS_ARMOR_LEATHER: proto.ArmorType_ArmorTypeLeather,
	ITEM_SUBCLASS_ARMOR_MAIL:    proto.ArmorType_ArmorTypeMail,
	ITEM_SUBCLASS_ARMOR_PLATE:   proto.ArmorType_ArmorTypePlate,
	0:                           proto.ArmorType_ArmorTypeUnknown,
}

// Ref: https://wowdev.wiki/Stat_Types
func MapMainStatToStat(index int) (proto.Stat, bool) {
	switch index {
	case 0:
		return proto.Stat_StatStrength, true
	case 1:
		return proto.Stat_StatAgility, true
	case 2:
		return proto.Stat_StatStamina, true
	case 3:
		return proto.Stat_StatIntellect, true
	case 4:
		return proto.Stat_StatSpirit, true
	}
	return 0, false
}
func MapBonusStatIndexToStat(index int) (proto.Stat, bool) {
	switch index {
	case 0: // Mana
		return proto.Stat_StatMana, true
	case 1: // Health
		return proto.Stat_StatHealth, true
	case 7: // Stamina
		return proto.Stat_StatStamina, true
	case 3: // Agility
		return proto.Stat_StatAgility, true
	case 4: // Strength
		return proto.Stat_StatStrength, true
	case 5: // Intellect
		return proto.Stat_StatIntellect, true
	case 6: // Spirit
		return proto.Stat_StatSpirit, true

	case 12:
		return proto.Stat_StatDefenseRating, true
	case 13:
		return proto.Stat_StatDodgeRating, true
	case 14:
		return proto.Stat_StatParryRating, true
	case 15:
		return proto.Stat_StatBlockRating, true

	// Secondary ratings
	case 16, 17:
		return proto.Stat_StatMeleeHitRating, true
	case 18:
		return proto.Stat_StatSpellHitRating, true
	case 31:
		return proto.Stat_StatAllHitRating, true
	case 19, 20:
		return proto.Stat_StatMeleeCritRating, true
	case 21:
		return proto.Stat_StatSpellCritRating, true
	case 32:
		return proto.Stat_StatAllCritRating, true
	case 28, 29:
		return proto.Stat_StatMeleeHasteRating, true
	case 30:
		return proto.Stat_StatSpellHasteRating, true
	case 36:
		return proto.Stat_StatAllHasteRating, true
	case 37:
		return proto.Stat_StatExpertiseRating, true

	case 38: // AttackPower
		return proto.Stat_StatAttackPower, true
	case 39: // RangedAttackPower
		return proto.Stat_StatRangedAttackPower, true
	case 40:
		return proto.Stat_StatFeralAttackPower, true
	case 41:
		return proto.Stat_StatHealingPower, true
	case 42:
		return proto.Stat_StatSpellDamage, true
	case 45:
		return proto.Stat_StatSpellPower, true
	case 50: // ExtraArmor maps to BonusArmor (green armor)
		return proto.Stat_StatBonusArmor, true
	case 43: // ManaRegeneration
		return proto.Stat_StatMP5, true
	case 47:
		return proto.Stat_StatSpellPenetration, true
	case 48:
		return proto.Stat_StatBlockValue, true
	default:
		return 0, false
	}
}

var MapProfessionIdToProfession = map[int]proto.Profession{
	0:   proto.Profession_ProfessionUnknown,
	164: proto.Profession_Blacksmithing,
	165: proto.Profession_Leatherworking,
	171: proto.Profession_Alchemy,
	182: proto.Profession_Herbalism,
	186: proto.Profession_Mining,
	197: proto.Profession_Tailoring,
	202: proto.Profession_Engineering,
	333: proto.Profession_Enchanting,
	393: proto.Profession_Skinning,
	755: proto.Profession_Jewelcrafting,
	773: proto.Profession_Inscription,
}

var MapItemSubclassNames = map[ItemSubClass]string{
	OneHandedAxes:    "One-Handed Axes",
	TwoHandedAxes:    "Two-Handed Axes",
	Bows:             "Bows",
	Guns:             "Guns",
	OneHandedMaces:   "One-Handed Maces",
	TwoHandedMaces:   "Two-Handed Maces",
	Polearms:         "Polearms",
	OneHandedSwords:  "One-Handed Swords",
	TwoHandedSwords:  "Two-Handed Swords",
	Staves:           "Staves",
	OneHandedExotics: "One-Handed Exotics",
	TwoHandedExotics: "Two-Handed Exotics",
	FistWeapons:      "Fist Weapons",
	Daggers:          "Daggers",
}

var MapSocketTypeToGemColor = map[int]proto.GemColor{
	0: proto.GemColor_GemColorUnknown,
	1: proto.GemColor_GemColorMeta,
	2: proto.GemColor_GemColorRed,
	3: proto.GemColor_GemColorYellow,
	4: proto.GemColor_GemColorBlue,
	7: proto.GemColor_GemColorPrismatic,
}

var MapInventoryTypeToItemType = map[int]proto.ItemType{
	0:                      proto.ItemType_ItemTypeUnknown,
	INVTYPE_HEAD:           proto.ItemType_ItemTypeHead,
	INVTYPE_NECK:           proto.ItemType_ItemTypeNeck,
	INVTYPE_SHOULDERS:      proto.ItemType_ItemTypeShoulder,
	INVTYPE_CHEST:          proto.ItemType_ItemTypeChest,
	INVTYPE_WAIST:          proto.ItemType_ItemTypeWaist,
	INVTYPE_LEGS:           proto.ItemType_ItemTypeLegs,
	INVTYPE_FEET:           proto.ItemType_ItemTypeFeet,
	INVTYPE_WRISTS:         proto.ItemType_ItemTypeWrist,
	INVTYPE_HANDS:          proto.ItemType_ItemTypeHands,
	INVTYPE_FINGER:         proto.ItemType_ItemTypeFinger,
	INVTYPE_TRINKET:        proto.ItemType_ItemTypeTrinket,
	INVTYPE_WEAPON:         proto.ItemType_ItemTypeWeapon,
	INVTYPE_SHIELD:         proto.ItemType_ItemTypeWeapon,
	INVTYPE_RANGED:         proto.ItemType_ItemTypeRanged,
	INVTYPE_CLOAK:          proto.ItemType_ItemTypeBack,
	INVTYPE_2HWEAPON:       proto.ItemType_ItemTypeWeapon,
	INVTYPE_BAG:            proto.ItemType_ItemTypeUnknown,
	INVTYPE_TABARD:         proto.ItemType_ItemTypeUnknown,
	INVTYPE_ROBE:           proto.ItemType_ItemTypeChest,
	INVTYPE_WEAPONMAINHAND: proto.ItemType_ItemTypeWeapon,
	INVTYPE_WEAPONOFFHAND:  proto.ItemType_ItemTypeWeapon,
	INVTYPE_HOLDABLE:       proto.ItemType_ItemTypeWeapon,
	INVTYPE_AMMO:           proto.ItemType_ItemTypeUnknown,
	INVTYPE_THROWN:         proto.ItemType_ItemTypeRanged,
	INVTYPE_RANGEDRIGHT:    proto.ItemType_ItemTypeRanged,
	INVTYPE_QUIVER:         proto.ItemType_ItemTypeRanged,
	INVTYPE_RELIC:          proto.ItemType_ItemTypeRanged,
}
var MapInventoryTypeFlagToItemType = map[InventoryTypeFlag]proto.ItemType{
	0:                proto.ItemType_ItemTypeUnknown,
	HEAD:             proto.ItemType_ItemTypeHead,
	NECK:             proto.ItemType_ItemTypeNeck,
	SHOULDER:         proto.ItemType_ItemTypeShoulder,
	CHEST:            proto.ItemType_ItemTypeChest,
	WAIST:            proto.ItemType_ItemTypeWaist,
	LEGS:             proto.ItemType_ItemTypeLegs,
	FEET:             proto.ItemType_ItemTypeFeet,
	WRIST:            proto.ItemType_ItemTypeWrist,
	HAND:             proto.ItemType_ItemTypeHands,
	FINGER:           proto.ItemType_ItemTypeFinger,
	TRINKET:          proto.ItemType_ItemTypeTrinket,
	MAIN_HAND:        proto.ItemType_ItemTypeWeapon,
	OFF_HAND:         proto.ItemType_ItemTypeWeapon,
	RANGED:           proto.ItemType_ItemTypeRanged,
	CLOAK:            proto.ItemType_ItemTypeBack,
	TWO_H_WEAPON:     proto.ItemType_ItemTypeWeapon,
	BAG:              proto.ItemType_ItemTypeUnknown,
	TABARD:           proto.ItemType_ItemTypeUnknown,
	ROBE:             proto.ItemType_ItemTypeChest,
	WEAPON_MAIN_HAND: proto.ItemType_ItemTypeWeapon,
	WEAPON_OFF_HAND:  proto.ItemType_ItemTypeWeapon,
	HOLDABLE:         proto.ItemType_ItemTypeWeapon,
	AMMO:             proto.ItemType_ItemTypeUnknown,
	THROWN:           proto.ItemType_ItemTypeRanged,
	RANGED_RIGHT:     proto.ItemType_ItemTypeRanged,
	QUIVER:           proto.ItemType_ItemTypeRanged,
	RELIC:            proto.ItemType_ItemTypeRanged,
}

var MapWeaponSubClassToWeaponType = map[int]proto.WeaponType{
	ITEM_SUBCLASS_WEAPON_AXE:          proto.WeaponType_WeaponTypeAxe,
	ITEM_SUBCLASS_WEAPON_AXE2:         proto.WeaponType_WeaponTypeAxe,
	ITEM_SUBCLASS_WEAPON_BOW:          proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_GUN:          proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_MACE:         proto.WeaponType_WeaponTypeMace,
	ITEM_SUBCLASS_WEAPON_MACE2:        proto.WeaponType_WeaponTypeMace,
	ITEM_SUBCLASS_WEAPON_POLEARM:      proto.WeaponType_WeaponTypePolearm,
	ITEM_SUBCLASS_WEAPON_SWORD:        proto.WeaponType_WeaponTypeSword,
	ITEM_SUBCLASS_WEAPON_SWORD2:       proto.WeaponType_WeaponTypeSword,
	ITEM_SUBCLASS_WEAPON_WARGLAIVE:    proto.WeaponType_WeaponTypePolearm, // assuming polearm idk
	ITEM_SUBCLASS_WEAPON_STAFF:        proto.WeaponType_WeaponTypeStaff,
	ITEM_SUBCLASS_WEAPON_EXOTIC:       proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_EXOTIC2:      proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_FIST:         proto.WeaponType_WeaponTypeFist,
	ITEM_SUBCLASS_WEAPON_MISC:         proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_DAGGER:       proto.WeaponType_WeaponTypeDagger,
	ITEM_SUBCLASS_WEAPON_THROWN:       proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_SPEAR:        proto.WeaponType_WeaponTypePolearm,
	ITEM_SUBCLASS_WEAPON_CROSSBOW:     proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_WAND:         proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_FISHING_POLE: proto.WeaponType_WeaponTypeUnknown,
}

var MapMinReputationToRepLevel = map[int]proto.RepLevel{
	0: proto.RepLevel_RepLevelUnknown,
	1: proto.RepLevel_RepLevelHostile,
	2: proto.RepLevel_RepLevelUnfriendly,
	3: proto.RepLevel_RepLevelNeutral,
	4: proto.RepLevel_RepLevelFriendly,
	5: proto.RepLevel_RepLevelHonored,
	6: proto.RepLevel_RepLevelRevered,
	7: proto.RepLevel_RepLevelExalted,
}

type EnchantMetaType struct {
	ItemType   proto.ItemType
	WeaponType proto.WeaponType
}

var SpellSchoolToStat = map[SpellSchool]proto.Stat{
	FIRE:     -1,
	ARCANE:   -1,
	NATURE:   -1,
	FROST:    -1,
	SHADOW:   -1,
	PHYSICAL: proto.Stat_StatArmor,
}
var MapInventoryTypeToEnchantMetaType = map[InventoryTypeFlag]EnchantMetaType{
	HEAD:     {ItemType: proto.ItemType_ItemTypeHead, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	NECK:     {ItemType: proto.ItemType_ItemTypeNeck, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	SHOULDER: {ItemType: proto.ItemType_ItemTypeShoulder, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	CHEST:    {ItemType: proto.ItemType_ItemTypeChest, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	WAIST:    {ItemType: proto.ItemType_ItemTypeWaist, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	LEGS:     {ItemType: proto.ItemType_ItemTypeLegs, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	FEET:     {ItemType: proto.ItemType_ItemTypeFeet, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	WRIST:    {ItemType: proto.ItemType_ItemTypeWrist, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	HAND:     {ItemType: proto.ItemType_ItemTypeHands, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	FINGER:   {ItemType: proto.ItemType_ItemTypeFinger, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	TRINKET:  {ItemType: proto.ItemType_ItemTypeTrinket, WeaponType: proto.WeaponType_WeaponTypeUnknown},

	WEAPON_MAIN_HAND: {ItemType: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeUnknown}, // One-Hand
	WEAPON_OFF_HAND:  {ItemType: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeShield},  // Off Hand
	RANGED:           {ItemType: proto.ItemType_ItemTypeRanged, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	CLOAK:            {ItemType: proto.ItemType_ItemTypeBack, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	TWO_H_WEAPON:     {ItemType: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeUnknown},
}
var consumableClassToProto = map[ConsumableClass]proto.ConsumableType{
	EXPLOSIVES_AND_DEVICES: proto.ConsumableType_ConsumableTypeExplosive,
	POTION:                 proto.ConsumableType_ConsumableTypePotion,
	FLASK:                  proto.ConsumableType_ConsumableTypeFlask,
	SCROLL:                 proto.ConsumableType_ConsumableTypeScroll,
	FOOD:                   proto.ConsumableType_ConsumableTypeFood,
	BANDAGE:                proto.ConsumableType_ConsumableTypeUnknown,
	OTHER:                  proto.ConsumableType_ConsumableTypeUnknown,
}

var MapPowerTypeEnumToResourceType = map[int32]proto.ResourceType{
	0: proto.ResourceType_ResourceTypeMana,
	1: proto.ResourceType_ResourceTypeRage,
	2: proto.ResourceType_ResourceTypeFocus,
	3: proto.ResourceType_ResourceTypeEnergy,
	4: proto.ResourceType_ResourceTypeComboPoints,
}

func ClassNameFromDBC(dbc DbcClass) string {
	switch dbc.ID {
	case 1:
		return "Warrior"
	case 2:
		return "Paladin"
	case 3:
		return "Hunter"
	case 4:
		return "Rogue"
	case 5:
		return "Priest"
	case 7:
		return "Shaman"
	case 8:
		return "Mage"
	case 9:
		return "Warlock"
	case 11:
		return "Druid"
	default:
		return "Unknown"
	}
}
func getMatchingRatingMods(value int) []RatingModType {
	allMods := []RatingModType{
		RATING_MOD_DODGE,
		RATING_MOD_PARRY,
		RATING_MOD_HIT_MELEE,
		RATING_MOD_HIT_RANGED,
		RATING_MOD_HIT_SPELL,
		RATING_MOD_CRIT_MELEE,
		RATING_MOD_CRIT_RANGED,
		RATING_MOD_CRIT_SPELL,
		RATING_MOD_MULTISTRIKE,
		RATING_MOD_READINESS,
		RATING_MOD_SPEED,
		RATING_MOD_RESILIENCE,
		RATING_MOD_LEECH,
		RATING_MOD_HASTE_MELEE,
		RATING_MOD_HASTE_RANGED,
		RATING_MOD_HASTE_SPELL,
		RATING_MOD_AVOIDANCE,
		RATING_MOD_EXPERTISE,
		RATING_MOD_MASTERY,
		RATING_MOD_PVP_POWER,
		RATING_MOD_VERS_DAMAGE,
		RATING_MOD_VERS_HEAL,
		RATING_MOD_VERS_MITIG,
	}

	var result []RatingModType
	for _, mod := range allMods {
		if value&int(mod) != 0 {
			result = append(result, mod)
		}
	}
	return result
}

var RatingModToStat = map[RatingModType]proto.Stat{
	RATING_MOD_DODGE:        proto.Stat_StatDodgeRating,
	RATING_MOD_PARRY:        proto.Stat_StatParryRating,
	RATING_MOD_HIT_MELEE:    proto.Stat_StatMeleeHitRating,
	RATING_MOD_HIT_RANGED:   proto.Stat_StatMeleeHitRating,
	RATING_MOD_HIT_SPELL:    proto.Stat_StatSpellHitRating,
	RATING_MOD_CRIT_MELEE:   proto.Stat_StatMeleeCritRating,
	RATING_MOD_CRIT_RANGED:  proto.Stat_StatMeleeCritRating,
	RATING_MOD_CRIT_SPELL:   proto.Stat_StatSpellCritRating,
	RATING_MOD_MULTISTRIKE:  -1,
	RATING_MOD_READINESS:    -1,
	RATING_MOD_SPEED:        -1,
	RATING_MOD_RESILIENCE:   proto.Stat_StatResilience,
	RATING_MOD_LEECH:        -1,
	RATING_MOD_HASTE_MELEE:  proto.Stat_StatMeleeHasteRating,
	RATING_MOD_HASTE_RANGED: proto.Stat_StatMeleeHasteRating,
	RATING_MOD_HASTE_SPELL:  proto.Stat_StatSpellHasteRating,
	RATING_MOD_AVOIDANCE:    -1,
	RATING_MOD_EXPERTISE:    proto.Stat_StatExpertiseRating,
	RATING_MOD_MASTERY:      -1,
	RATING_MOD_PVP_POWER:    -1,

	RATING_MOD_VERS_DAMAGE: -1,
	RATING_MOD_VERS_HEAL:   -1,
	RATING_MOD_VERS_MITIG:  -1,
}

type DbcClass struct {
	ProtoClass proto.Class
	ID         int
}

var Classes = []DbcClass{
	{proto.Class_ClassWarrior, 1},
	{proto.Class_ClassPaladin, 2},
	{proto.Class_ClassHunter, 3},
	{proto.Class_ClassRogue, 4},
	{proto.Class_ClassPriest, 5},
	{proto.Class_ClassShaman, 7},
	{proto.Class_ClassMage, 8},
	{proto.Class_ClassWarlock, 9},
	{proto.Class_ClassDruid, 11},
}

// SpecByID maps the ChrSpecialization.DB2 ID to proto.Spec
var SpecByID = map[int32]proto.Spec{
	// Druid
	102: proto.Spec_SpecBalanceDruid,
	103: proto.Spec_SpecFeralDruid,
	104: proto.Spec_SpecGuardianDruid,
	105: proto.Spec_SpecRestorationDruid,

	// Hunter
	253: proto.Spec_SpecBeastMasteryHunter,
	254: proto.Spec_SpecMarksmanshipHunter,
	255: proto.Spec_SpecSurvivalHunter,

	// Mage
	62: proto.Spec_SpecArcaneMage,
	63: proto.Spec_SpecFireMage,
	64: proto.Spec_SpecFrostMage,

	// Paladin
	65: proto.Spec_SpecHolyPaladin,
	66: proto.Spec_SpecProtectionPaladin,
	70: proto.Spec_SpecRetributionPaladin,

	// Priest
	256: proto.Spec_SpecDisciplinePriest,
	257: proto.Spec_SpecHolyPriest,
	258: proto.Spec_SpecShadowPriest,

	// Rogue
	259: proto.Spec_SpecAssassinationRogue,
	260: proto.Spec_SpecCombatRogue,
	261: proto.Spec_SpecSubtletyRogue,

	// Shaman
	262: proto.Spec_SpecElementalShaman,
	263: proto.Spec_SpecEnhancementShaman,
	264: proto.Spec_SpecRestorationShaman,

	// Warlock
	265: proto.Spec_SpecAfflictionWarlock,
	266: proto.Spec_SpecDemonologyWarlock,
	267: proto.Spec_SpecDestructionWarlock,

	// Warrior
	71: proto.Spec_SpecArmsWarrior,
	72: proto.Spec_SpecFuryWarrior,
	73: proto.Spec_SpecProtectionWarrior,
}

func SpecFromID(id int32) proto.Spec {
	if s, ok := SpecByID[id]; ok {
		return s
	}
	return proto.Spec_SpecUnknown
}
