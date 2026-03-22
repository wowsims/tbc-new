package dps

import (
	"github.com/wowsims/tbc/sim/common"
	_ "github.com/wowsims/tbc/sim/common" // imported to get item effects included.
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"

	"testing"
)

func init() {
	RegisterDpsWarrior()
	common.RegisterAllEffects()
}

func TestDpsWarrior(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassWarrior,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceHuman},
			GearSet:    core.GetGearSet("../../../ui/warrior/dps/gear_sets", "preraid_fury"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/warrior/dps/gear_sets", "p1_fury"),
				core.GetGearSet("../../../ui/warrior/dps/gear_sets", "p1_arms"),
			},
			Talents: DefaultFuryTalents,
			OtherTalentSets: []core.TalentsCombo{
				{Label: "Arms", Talents: DefaultArmsTalents},
			},
			Consumables:      DefaultConsumables,
			SpecOptions:      core.SpecOptionsCombo{Label: "Fury", SpecOptions: DefaultOptions},
			StartingDistance: 25,
			Profession1:      proto.Profession_Engineering,
			Profession2:      proto.Profession_Blacksmithing,

			Rotation: core.GetAplRotation("../../../ui/warrior/dps/apls", "fury"),
			OtherRotations: []core.RotationCombo{
				core.GetAplRotation("../../../ui/warrior/dps/apls", "arms"),
			},

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeLeather,
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeSword,
					proto.WeaponType_WeaponTypeAxe,
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

var DefaultOptions = &proto.Player_DpsWarrior{
	DpsWarrior: &proto.DpsWarrior{
		Options: &proto.DpsWarrior_Options{
			ClassOptions: &proto.WarriorOptions{
				DefaultShout:  proto.WarriorShout_WarriorShoutBattle,
				DefaultStance: proto.WarriorStance_WarriorStanceBerserker,
			},
		},
	},
}

var DefaultFuryTalents = "3500501130201-05050005505012050115"
var DefaultArmsTalents = "32005020352010500221-0550000500521203"

var DefaultConsumables = &proto.ConsumesSpec{
	PotId:       22838,
	FlaskId:     22854,
	FoodId:      27658,
	ConjuredId:  22788,
	ExplosiveId: 30217,
	SuperSapper: true,
	OhImbueId:   29453,
	ScrollAgi:   true,
	ScrollStr:   true,
}
