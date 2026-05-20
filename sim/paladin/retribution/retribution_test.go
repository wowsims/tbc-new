package retribution

import (
	"testing"

	"github.com/wowsims/tbc/sim/common" // imported to get item effects included.
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func init() {
	RegisterRetributionPaladin()
	common.RegisterAllEffects()
}

func TestRetribution(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassPaladin,
			Race:       proto.Race_RaceBloodElf,
			OtherRaces: []proto.Race{proto.Race_RaceHuman},
			SpecOptions: core.SpecOptionsCombo{Label: "Default", SpecOptions: &proto.Player_RetributionPaladin{
				RetributionPaladin: &proto.RetributionPaladin{
					Options: &proto.RetributionPaladin_Options{
						ClassOptions: &proto.PaladinOptions{},
					},
				},
			}},
			Consumables: DefaultConsumables,
			Profession1: proto.Profession_Engineering,
			Profession2: proto.Profession_Blacksmithing,
			GearSet:     core.GetGearSet("../../../ui/paladin/retribution/gear_sets", "p1"),
			Talents:     "5-053201-0523005120033125331051",
			Rotation:    core.GetAplRotation("../../../ui/paladin/retribution/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypePolearm,
					proto.WeaponType_WeaponTypeSword,
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeMace,
				},
				ArmorType: proto.ArmorType_ArmorTypePlate,
				HandTypes: []proto.HandType{proto.HandType_HandTypeTwoHand},
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeLibram,
				},
			},
		},
	}))
}

var DefaultConsumables = &proto.ConsumesSpec{
	PotId:        22838,
	FlaskId:      22854,
	FoodId:       27658,
	ConjuredId:   12662,
	SuperSapper:  true,
	GoblinSapper: true,
	ScrollAgi:    true,
	ScrollStr:    true,
	ExplosiveId:  30217,
}
