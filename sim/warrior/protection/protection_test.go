package protection

import (
	_ "github.com/wowsims/tbc/sim/common" // imported to get item effects included.
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"

	"testing"
)

func init() {
	RegisterProtectionWarrior()
}

func TestProtectionWarrior(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassWarrior,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceHuman},
			GearSet:    core.GetGearSet("../../../ui/warrior/protection/gear_sets", "preraid"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/warrior/protection/gear_sets", "p1_bis"),
			},
			Talents:          DefaultProtectionTalents,
			Consumables:      DefaultConsumables,
			SpecOptions:      core.SpecOptionsCombo{Label: "Protection", SpecOptions: DefaultOptions},
			StartingDistance: 0,
			Profession1:      proto.Profession_Engineering,
			Profession2:      proto.Profession_Blacksmithing,

			Rotation: core.GetAplRotation("../../../ui/warrior/protection/apls", "default"),

			IsTank:          true,
			InFrontOfTarget: true,

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypePlate,
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeSword,
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeShield,
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

var DefaultOptions = &proto.Player_ProtectionWarrior{
	ProtectionWarrior: &proto.ProtectionWarrior{
		Options: &proto.ProtectionWarrior_Options{
			ClassOptions: &proto.WarriorOptions{
				StartingRage:  100,
				DefaultShout:  proto.WarriorShout_WarriorShoutCommanding,
				DefaultStance: proto.WarriorStance_WarriorStanceDefensive,
			},
		},
	},
}

var DefaultProtectionTalents = "35000301302-03-0055511033001101501351"

var DefaultConsumables = &proto.ConsumesSpec{
	PotId:            22849,
	FoodId:           27667,
	ConjuredId:       22105,
	ExplosiveId:      30217,
	SuperSapper:      true,
	OhImbueId:        29453,
	DrumsId:          351355,
	ScrollAgi:        true,
	ScrollStr:        true,
	ScrollArm:        true,
	BattleElixirId:   22831,
	GuardianElixirId: 9088,
	NightmareSeed:    true,
}
