package feralcat

import (
	"testing"

	"github.com/wowsims/tbc/sim/common"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func init() {
	RegisterFeralCatDruid()
	common.RegisterAllEffects()
}

func TestFeralCat(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassDruid,
			Race:       proto.Race_RaceNightElf,
			OtherRaces: []proto.Race{proto.Race_RaceTauren},

			GearSet: core.GetGearSet("../../../ui/druid/feralcat/gear_sets", "p1"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/druid/feralcat/gear_sets", "p2"),
				core.GetGearSet("../../../ui/druid/feralcat/gear_sets", "p3"),
				core.GetGearSet("../../../ui/druid/feralcat/gear_sets", "p4"),
				core.GetGearSet("../../../ui/druid/feralcat/gear_sets", "p5"),
			},

			Talents: DefaultTalents,
			OtherTalentSets: []core.TalentsCombo{
				{Label: "Monocat", Talents: MonocatTalents},
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

			EPReferenceStat: proto.Stat_StatAttackPower,
			StatsToWeigh: []proto.Stat{
				proto.Stat_StatAgility,
				proto.Stat_StatStrength,
				proto.Stat_StatAttackPower,
				proto.Stat_StatFeralAttackPower,
				proto.Stat_StatMeleeHitRating,
				proto.Stat_StatExpertiseRating,
				proto.Stat_StatMeleeCritRating,
				proto.Stat_StatMeleeHasteRating,
				proto.Stat_StatArmorPenetration,
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
				Equipment:     core.GetGearSet("../../../ui/druid/feralcat/gear_sets", "p1").GearSet,
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
const MonocatTalents = "-553002132322105301051-05503301"

var DefaultSpecOptions = &proto.Player_FeralCatDruid{
	FeralCatDruid: &proto.FeralCatDruid{
		Rotation: &proto.FeralCatDruid_Rotation{
			FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
			Biteweave:          true,
			RipMinComboPoints:  5,
			BiteMinComboPoints: 5,
			MangleTrick:        true,
			MaintainFaerieFire: false,
		},
		Options: &proto.FeralCatDruid_Options{},
	},
}

var DefaultConsumables = &proto.ConsumesSpec{
	PotId:            22838, // Haste Potion
	BattleElixirId:   22831, // Elixir of Major Agility
	GuardianElixirId: 32067, // Elixir of Draenic Wisdom
	FoodId:           27664, // Grilled Mudfish
	MhImbueId:        34340, // Adamantite Weightstone
	ConjuredId:       12662, // Demonic Rune
	SuperSapper:      true,
	GoblinSapper:     true,
	ScrollAgi:        true,
	ScrollStr:        true,
}
