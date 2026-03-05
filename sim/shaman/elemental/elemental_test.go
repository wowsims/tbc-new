package elemental

import (
	"testing"

	"github.com/wowsims/tbc/sim/common"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func init() {
	RegisterElementalShaman()
	common.RegisterAllEffects()
}

func TestElemental(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassShaman,
			Race:       proto.Race_RaceTroll,
			OtherRaces: []proto.Race{proto.Race_RaceOrc, proto.Race_RaceDraenei},
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: &proto.Player_ElementalShaman{
				ElementalShaman: &proto.ElementalShaman{
					Options: &proto.ElementalShaman_Options{
						ClassOptions: &proto.ShamanOptions{
							ShieldProcrate: 0.0,
						},
					},
				},
			}},
			GearSet:  core.GetGearSet("../../../ui/shaman/elemental/gear_sets", "p1"),
			Talents:  "55003105100213351051--05105301005",
			Rotation: core.GetAplRotation("../../../ui/shaman/elemental/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeShield,
				},
				ArmorType: proto.ArmorType_ArmorTypeMail,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeTotem,
				},
			},
		},
	}))
}
