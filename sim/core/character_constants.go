package core

import (
	"github.com/wowsims/tbc/sim/core/proto"
)

// EligibleWeaponType describes whether a weapon type can be used as a two-hand weapon.
type EligibleWeaponType struct {
	CanUseTwoHand bool
}

var ClassArmorTypeCapabilities = map[proto.Class][]proto.ArmorType{
	proto.Class_ClassWarrior: {proto.ArmorType_ArmorTypePlate, proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassPaladin: {proto.ArmorType_ArmorTypePlate, proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassHunter:  {proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassRogue:   {proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassPriest:  {proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassShaman:  {proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassMage:    {proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassWarlock: {proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassDruid:   {proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
}

var ClassWeaponTypeCapabilities = map[proto.Class]map[proto.WeaponType]EligibleWeaponType{
	proto.Class_ClassWarrior: {
		proto.WeaponType_WeaponTypeAxe:     {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeFist:    {},
		proto.WeaponType_WeaponTypeMace:    {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypePolearm: {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeShield:  {},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeSword:   {CanUseTwoHand: true},
	},
	proto.Class_ClassPaladin: {
		proto.WeaponType_WeaponTypeAxe:     {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeMace:    {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypePolearm: {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeShield:  {},
		proto.WeaponType_WeaponTypeSword:   {CanUseTwoHand: true},
	},
	proto.Class_ClassRogue: {
		proto.WeaponType_WeaponTypeAxe:     {},
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeFist:    {},
		proto.WeaponType_WeaponTypeMace:    {},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypeSword:   {},
	},
	proto.Class_ClassDruid: {
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeFist:    {},
		proto.WeaponType_WeaponTypeMace:    {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypePolearm: {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
	},
	proto.Class_ClassShaman: {
		proto.WeaponType_WeaponTypeAxe:     {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeFist:    {},
		proto.WeaponType_WeaponTypeMace:    {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypeShield:  {},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
	},
	proto.Class_ClassMage: {
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeSword:   {},
	},
	proto.Class_ClassPriest: {
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeMace:    {},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
	},
	proto.Class_ClassWarlock: {
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeSword:   {},
	},
	proto.Class_ClassHunter: {},
}

var ClassRangedWeaponTypeCapabilities = map[proto.Class][]proto.RangedWeaponType{
	proto.Class_ClassWarrior: {proto.RangedWeaponType_RangedWeaponTypeBow, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeGun, proto.RangedWeaponType_RangedWeaponTypeThrown},
	proto.Class_ClassRogue:   {proto.RangedWeaponType_RangedWeaponTypeBow, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeGun, proto.RangedWeaponType_RangedWeaponTypeThrown},
	proto.Class_ClassHunter:  {proto.RangedWeaponType_RangedWeaponTypeBow, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeGun},
	proto.Class_ClassMage:    {proto.RangedWeaponType_RangedWeaponTypeWand},
	proto.Class_ClassPriest:  {proto.RangedWeaponType_RangedWeaponTypeWand},
	proto.Class_ClassWarlock: {proto.RangedWeaponType_RangedWeaponTypeWand},
	proto.Class_ClassDruid:   {proto.RangedWeaponType_RangedWeaponTypeIdol},
	proto.Class_ClassPaladin: {proto.RangedWeaponType_RangedWeaponTypeLibram},
	proto.Class_ClassShaman:  {proto.RangedWeaponType_RangedWeaponTypeTotem},
}

var ClassRaceCapabilities = map[proto.Class][]proto.Race{
	proto.Class_ClassWarrior: {
		proto.Race_RaceHuman,
		proto.Race_RaceDwarf,
		proto.Race_RaceNightElf,
		proto.Race_RaceGnome,
		proto.Race_RaceDraenei,
		proto.Race_RaceOrc,
		proto.Race_RaceUndead,
		proto.Race_RaceTauren,
		proto.Race_RaceTroll,
	},
	proto.Class_ClassPaladin: {
		proto.Race_RaceHuman,
		proto.Race_RaceDwarf,
		proto.Race_RaceDraenei,
		proto.Race_RaceBloodElf,
	},
	proto.Class_ClassHunter: {
		proto.Race_RaceDwarf,
		proto.Race_RaceNightElf,
		proto.Race_RaceDraenei,
		proto.Race_RaceOrc,
		proto.Race_RaceTauren,
		proto.Race_RaceTroll,
		proto.Race_RaceBloodElf,
	},
	proto.Class_ClassRogue: {
		proto.Race_RaceHuman,
		proto.Race_RaceDwarf,
		proto.Race_RaceNightElf,
		proto.Race_RaceGnome,
		proto.Race_RaceOrc,
		proto.Race_RaceUndead,
		proto.Race_RaceTroll,
		proto.Race_RaceBloodElf,
	},
	proto.Class_ClassPriest: {
		proto.Race_RaceHuman,
		proto.Race_RaceDwarf,
		proto.Race_RaceNightElf,
		proto.Race_RaceDraenei,
		proto.Race_RaceUndead,
		proto.Race_RaceTroll,
		proto.Race_RaceBloodElf,
	},
	proto.Class_ClassShaman: {
		proto.Race_RaceDraenei,
		proto.Race_RaceOrc,
		proto.Race_RaceTauren,
		proto.Race_RaceTroll,
	},
	proto.Class_ClassMage: {
		proto.Race_RaceHuman,
		proto.Race_RaceGnome,
		proto.Race_RaceDraenei,
		proto.Race_RaceUndead,
		proto.Race_RaceTroll,
		proto.Race_RaceBloodElf,
	},
	proto.Class_ClassWarlock: {
		proto.Race_RaceHuman,
		proto.Race_RaceGnome,
		proto.Race_RaceOrc,
		proto.Race_RaceUndead,
		proto.Race_RaceBloodElf,
	},
	proto.Class_ClassDruid: {
		proto.Race_RaceNightElf,
		proto.Race_RaceTauren,
	},
}

var SpecCanDualWieldCapabilities = map[proto.Spec]bool{
	proto.Spec_SpecHunter:            true,
	proto.Spec_SpecRogue:             true,
	proto.Spec_SpecEnhancementShaman: true,
	proto.Spec_SpecDpsWarrior:        true,
	proto.Spec_SpecProtectionWarrior: true,
}
