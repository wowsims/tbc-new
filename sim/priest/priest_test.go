package priest

import (
	"testing"

	"github.com/wowsims/tbc/sim/common"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func init() {
	RegisterPriest()
	common.RegisterAllEffects()
}

const defaultTalents = "500230013--503250510240103051451"

func TestShadowPriest(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassPriest,
			Race:       proto.Race_RaceTroll,
			OtherRaces: []proto.Race{proto.Race_RaceUndead, proto.Race_RaceDwarf},

			SpecOptions: core.SpecOptionsCombo{
				Label: "Shadow",
				SpecOptions: &proto.Player_Priest{
					Priest: &proto.Priest{
						Options: &proto.Priest_Options{
							ClassOptions: &proto.PriestOptions{
								// Begin the sim already in Shadowform so the opener
								// doesn't spend a GCD casting it.
								PreShadowform: true,
							},
						},
					},
				},
			},

			// Primary gear set — update path when higher phase sets are added.
			GearSet: core.GetGearSet("../../ui/priest/dps/gear_sets", "pre_raid"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../ui/priest/dps/gear_sets", "p1"),
			},

			Talents: defaultTalents,

			// Primary rotation
			Rotation: core.GetAplRotation("../../ui/priest/dps/apls", "default"),

			// Secondary rotation: casts every implemented spell
			OtherRotations: []core.RotationCombo{
				core.GetAplRotation("../../ui/priest/dps/apls", "test"),
			},

			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeStaff,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeOffHand,
				},
				ArmorType: proto.ArmorType_ArmorTypeCloth,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeWand,
				},
				// Blacklist melee enchants that appear on cloth-relevant slots but
				// are never used by casters.
				EnchantBlacklist: []int32{2673, 3225, 3273},
			},
		},
	}))
}
