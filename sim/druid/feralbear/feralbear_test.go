package feralbear

import (
	"testing"

	"github.com/wowsims/tbc/sim/common"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func init() {
	RegisterFeralBearDruid()
	common.RegisterAllEffects()
}

func TestFeralBear(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassDruid,
			Race:       proto.Race_RaceNightElf,
			OtherRaces: []proto.Race{proto.Race_RaceTauren},

			GearSet: core.GetGearSet("../../../ui/druid/feralbear/gear_sets", "p1"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/druid/feralbear/gear_sets", "preraid"),
			},

			Talents: DefaultTalents,
			OtherTalentSets: []core.TalentsCombo{
				{Label: "DemoRoar", Talents: DemoRoarTalents},
			},

			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: DefaultSpecOptions},

			Rotation: core.RotationCombo{
				Label: "Default",
				Rotation: &proto.APLRotation{
					Type: proto.APLRotation_TypeSimple,
				},
			},

			Consumables: DefaultConsumables,

			Profession1: proto.Profession_Engineering,
			Profession2: proto.Profession_Enchanting,

			IsTank:          true,
			InFrontOfTarget: true,

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeLeather,
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeStaff,
				},
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeIdol,
				},
			},

			EPReferenceStat: proto.Stat_StatAgility,
			StatsToWeigh: []proto.Stat{
				proto.Stat_StatHealth,
				proto.Stat_StatStamina,
				proto.Stat_StatAgility,
				proto.Stat_StatStrength,
				proto.Stat_StatAttackPower,
				proto.Stat_StatArmor,
				proto.Stat_StatBonusArmor,
				proto.Stat_StatDodgeRating,
				proto.Stat_StatDefenseRating,
				proto.Stat_StatMeleeHitRating,
				proto.Stat_StatExpertiseRating,
			},
		},
	}))
}

func BenchmarkSimulate(b *testing.B) {
	rsr := &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Class:         proto.Class_ClassDruid,
				Race:          proto.Race_RaceNightElf,
				TalentsString: DefaultTalents,
				Equipment:     core.GetGearSet("../../../ui/druid/feralbear/gear_sets", "p1").GearSet,
				Consumables:   DefaultConsumables,
				Spec:          DefaultSpecOptions,
				Rotation: &proto.APLRotation{
					Type: proto.APLRotation_TypeSimple,
				},
			},
			nil, nil, nil,
		),
		Encounter: &proto.Encounter{
			Duration: 300,
			Targets:  []*proto.Target{core.NewDefaultTarget()},
		},
		SimOptions: core.AverageDefaultSimTestOptions,
	}

	core.RaidBenchmark(b, rsr)
}

const DefaultTalents = "-503032132322105301251-05503301"
const DemoRoarTalents = "-553032132322105301051-05503001"

var DefaultSpecOptions = &proto.Player_FeralBearDruid{
	FeralBearDruid: &proto.FeralBearDruid{
		Options: &proto.FeralBearDruid_Options{
			StartingRage: 25,
		},
	},
}

var DefaultConsumables = &proto.ConsumesSpec{
	PotId:            22849, // Ironshield Potion
	BattleElixirId:   22831, // Elixir of Major Agility
	GuardianElixirId: 9088,  // Gift of Arthas
	FoodId:           27667, // Spicy Crawdad
	ConjuredId:       22105, // Healthstone
	SuperSapper:      true,
	GoblinSapper:     true,
	ScrollAgi:        true,
	ScrollStr:        true,
	ScrollArm:        true,
	NightmareSeed:    true,
}
