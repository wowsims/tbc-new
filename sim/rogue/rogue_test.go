package rogue

import (
	"testing"

	_ "github.com/wowsims/tbc/sim/common" // imported to get item effects included.
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func init() {
	RegisterRogue()
}

func TestRogue(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:         proto.Class_ClassRogue,
			Race:          proto.Race_RaceHuman,
			OtherRaces:    []proto.Race{proto.Race_RaceOrc},
			GearSet:       core.GetGearSet("../../ui/rogue/dps/gear_sets", "preraid"),
			OtherGearSets: []core.GearSetCombo{
				//core.GetGearSet("../../../ui/rogue/combat/gear_sets", "p3_combat"),
				//core.GetGearSet("../../../ui/rogue/combat/gear_sets", "p4_combat"),
			},
			Talents:     DefaultTalents,
			Consumables: DefaultConsumables,
			SpecOptions: core.SpecOptionsCombo{Label: "Rogue", SpecOptions: DefaultOptions},

			Rotation:       core.GetAplRotation("../../ui/rogue/dps/apls", "sinister"),
			OtherRotations: []core.RotationCombo{},
			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeLeather,
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeSword,
				},
				HandTypes: []proto.HandType{
					proto.HandType_HandTypeMainHand,
					proto.HandType_HandTypeOffHand,
					proto.HandType_HandTypeOneHand,
				},
			},
		},
	}))
}

var DefaultOptions = &proto.Player_Rogue{
	Rogue: &proto.Rogue{
		Options: &proto.Rogue_Options{
			ClassOptions: &proto.RogueOptions{},
		},
	},
}

var DefaultTalents = "00532012502-023305200005015002321151"

var DefaultConsumables = &proto.ConsumesSpec{
	FlaskId:    22854,
	FoodId:     33872,
	PotId:      22838,
	ConjuredId: 7676,
	OhImbueId:  27186,
	DrumsId:    351355,
}
