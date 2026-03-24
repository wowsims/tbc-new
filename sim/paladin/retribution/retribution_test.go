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
			GearSet:  core.GetGearSet("../../../ui/paladin/retribution/gear_sets", "p1"),
			Talents:  "5-053201-0523005120033125331051",
			Rotation: core.GetAplRotation("../../../ui/paladin/retribution/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypePolearm,
					proto.WeaponType_WeaponTypeSword,
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeMace,
				},
				ArmorType: proto.ArmorType_ArmorTypePlate,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeLibram,
				},
			},
		},
	}))
}

/*
func TestArcane(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassMage,
			Race:       proto.Race_RaceTroll,
			OtherRaces: []proto.Race{proto.Race_RaceOrc},
			SpecOptions: core.SpecOptionsCombo{Label: "Arcane", SpecOptions: &proto.Player_Mage{
				Mage: &proto.Mage{
					Options: &proto.Mage_Options{
						ClassOptions: &proto.MageOptions{
							DefaultMageArmor: proto.MageArmor_MageArmorMageArmor,
						},
					},
				},
			}},
			GearSet:  core.GetGearSet("../../ui/mage/dps/gear_sets", "p1Arcane"),
			Talents:  "2500052300030150330125--053500031003001",
			Rotation: core.GetAplRotation("../../ui/mage/dps/apls", "arcane"),
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeSword,
					proto.WeaponType_WeaponTypeOffHand,
					proto.WeaponType_WeaponTypeStaff,
				},
				ArmorType: proto.ArmorType_ArmorTypeCloth,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeWand,
				},
				EnchantBlacklist: []int32{2673, 3225, 3273},
			},
		},
	}))
}
*/
