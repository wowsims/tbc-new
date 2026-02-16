package core_test

import (
	"strconv"
	"testing"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/druid/feral"
	"github.com/wowsims/mop/sim/hunter/marksmanship"
)

func getTestPlayerMM() *proto.Player {
	var MMTalents = "032002-2302320032120231221-03"

	var FerocityTalents = &proto.HunterPetTalents{
		SerpentSwiftness: 2,
		Dive:             true,
		SpikedCollar:     3,
		Bloodthirsty:     1,
		CullingTheHerd:   3,
		SpidersBite:      3,
		Rabid:            true,
		CallOfTheWild:    true,
		SharkAttack:      2,
	}

	var PlayerOptionsBasic = &proto.Player_MarksmanshipHunter{
		MarksmanshipHunter: &proto.MarksmanshipHunter{
			Options: &proto.MarksmanshipHunter_Options{
				ClassOptions: &proto.HunterOptions{
					PetType:           proto.HunterOptions_Wolf,
					PetTalents:        FerocityTalents,
					PetUptime:         0.9,
					TimeToTrapWeaveMs: 0,
				},
			},
		},
	}

	marksmanship.RegisterMarksmanshipHunter()

	return &proto.Player{
		Race:           proto.Race_RaceOrc,
		Class:          proto.Class_ClassHunter,
		Equipment:      core.GetGearSet("../../ui/hunter/marksmanship/gear_sets", "preraid_mm").GearSet,
		Rotation:       core.GetAplRotation("../../ui/hunter/marksmanship/apls", "mm").Rotation,
		Spec:           PlayerOptionsBasic,
		TalentsString:  MMTalents,
		Buffs:          core.FullIndividualBuffs,
		ReactionTimeMs: 100,
	}
}

func getTestPlayerFeralCat() *proto.Player {
	var StandardTalents = "-2320322312012121202301-020301"

	var PlayerOptionsMonoCat = &proto.Player_FeralDruid{
		FeralDruid: &proto.FeralDruid{
			Options: &proto.FeralDruid_Options{
				AssumeBleedActive: true,
			},
		},
	}

	var FullConsumes = &proto.Consumes{
		Flask:         proto.Flask_FlaskOfTheWinds,
		Food:          proto.Food_FoodSkeweredEel,
		DefaultPotion: proto.Potions_PotionOfTheTolvir,
		PrepopPotion:  proto.Potions_PotionOfTheTolvir,
	}

	feral.RegisterFeralDruid()

	return &proto.Player{
		Race:           proto.Race_RaceTauren,
		Class:          proto.Class_ClassDruid,
		Equipment:      core.GetGearSet("../../ui/druid/feral/gear_sets", "preraid").GearSet,
		Rotation:       core.GetAplRotation("../../ui/druid/feral/apls", "default").Rotation,
		Consumes:       FullConsumes,
		Spec:           PlayerOptionsMonoCat,
		TalentsString:  StandardTalents,
		Buffs:          core.FullIndividualBuffs,
		ReactionTimeMs: 100,
	}
}

func makeTestCase(player *proto.Player) *proto.RaidSimRequest {
	return &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			player,
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter: &proto.Encounter{
			Duration: 180,
			Targets: []*proto.Target{
				core.NewDefaultTarget(),
			},
		},
		SimOptions: &proto.SimOptions{
			Iterations:    200,
			IsTest:        true,
			Debug:         false,
			RandomSeed:    123,
			SaveAllValues: true,
		},
	}
}

func TestConcurrentRaidSim(t *testing.T) {
	players := []*proto.Player{
		getTestPlayerMM(),
		getTestPlayerBloodDk(),
		getTestPlayerFeralCat(),
	}

	for i, player := range players {
		rsr := makeTestCase(player)
		stRes := core.RunRaidSim(rsr)
		mtRes := core.RunRaidSimConcurrent(rsr)
		core.CompareConcurrentSimResultsTest(t, strconv.Itoa(i), stRes, mtRes, 0.00001)
	}
}
