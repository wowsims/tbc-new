package hunter

import (
	"github.com/wowsims/tbc/sim/common"
	_ "github.com/wowsims/tbc/sim/common" // imported to get item effects included.
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"

	"testing"
)

func init() {
	RegisterHunter()
	common.RegisterAllEffects()
}

func TestHunter(t *testing.T) {
	weaveRotation := core.GetAplRotation("../../ui/hunter/dps/apls", "default")
	weaveRotation.Label = "weave"

	turretRotation := core.GetAplRotation("../../ui/hunter/dps/apls", "default").Rotation
	turretRotation.ValueVariables[2] = &proto.APLValueVariable{
		Name: "Melee weave",
		Value: &proto.APLValue{
			Value: &proto.APLValue_Const{
				Const: &proto.APLValueConst{
					Val: "false",
				},
			},
		},
	}

	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassHunter,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceNightElf},
			GearSet:    core.GetGearSet("../../ui/hunter/dps/gear_sets", "p1_bm_2h_6p"),
			Talents:    DefaultBMTalents,
			OtherTalentSets: []core.TalentsCombo{
				{Label: "SV", Talents: DefaultSVTalents},
			},
			Consumables:      DefaultConsumables,
			SpecOptions:      core.SpecOptionsCombo{Label: "Default", SpecOptions: DefaultOptions},
			StartingDistance: 7,
			Profession1:      proto.Profession_Engineering,
			Profession2:      proto.Profession_Blacksmithing,

			Rotation: weaveRotation,
			OtherRotations: []core.RotationCombo{
				{Label: "Turret", Rotation: turretRotation},
			},

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeMail,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeBow,
					proto.RangedWeaponType_RangedWeaponTypeCrossbow,
					proto.RangedWeaponType_RangedWeaponTypeGun,
				},
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypePolearm,
					proto.WeaponType_WeaponTypeStaff,
					proto.WeaponType_WeaponTypeSword,
				},
				HandTypes: []proto.HandType{
					proto.HandType_HandTypeMainHand,
					proto.HandType_HandTypeOffHand,
					proto.HandType_HandTypeOneHand,
					proto.HandType_HandTypeTwoHand,
				},
			},
		},
	}))
}

var DefaultOptions = &proto.Player_Hunter{
	Hunter: &proto.Hunter{
		Options: &proto.Hunter_Options{
			ClassOptions: &proto.HunterOptions{
				Ammo:             proto.HunterOptions_WardensArrow,
				PetSingleAbility: false,
				PetType:          proto.HunterOptions_Ravager,
				PetUptime:        100.0,
				QuiverBonus:      proto.HunterOptions_Speed15,
			},
		},
	},
}

var DefaultBMTalents = "512002005250122431051-0505201205"
var DefaultSVTalents = "502-0550201205-333200022003223005103"

var DefaultConsumables = &proto.ConsumesSpec{
	BattleElixirId:   22831, // Elixir of Major Agility
	GuardianElixirId: 22840, // Elixir of Major Mageblood
	FoodId:           27659, // Warp Burger
	PotId:            22838, // Haste Potion
	ConjuredId:       12662, // Demonic Rune
	ExplosiveId:      30217, // Adamantite Grenade
	PetFoodId:        33874, // Kibler's Bits
	PetScrollAgi:     true,
	PetScrollStr:     true,
	SuperSapper:      true,
	GoblinSapper:     true,
	ScrollAgi:        true,
	ScrollStr:        true,
}
