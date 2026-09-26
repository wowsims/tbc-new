//go:build with_db

package main

import (
	"testing"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/bulk"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

// A candidate the gem optimizer cannot bring within the stat constraints must still be judged on
// its own gear: the optimizer starts from the gear with its gems removed and only offers the gems
// in its pool, so "no gem choice meets the constraints" says nothing about the gems the candidate
// already has. Here a mage's candidate carries Stamina gems that meet a Stamina floor, while the
// pool offers only a spell damage gem; the candidate must be simmed on its own gems, not dropped.
func TestBulkSimInfeasibleGemModelKeepsCandidateGear(t *testing.T) {
	const (
		runedLivingRuby  = 24030 // +9 spell damage: the only gem in the pool
		solidStarOfElune = 24033 // Stamina: the candidate's own gems
	)
	for _, id := range []int32{runedLivingRuby, solidStarOfElune} {
		if _, ok := core.GetGemByID(id); !ok {
			t.Fatalf("gem %d is not in the database", id)
		}
	}
	if gem, _ := core.GetGemByID(solidStarOfElune); gem.Stats[proto.Stat_StatStamina] <= 0 {
		t.Fatalf("gem %d should carry Stamina", solidStarOfElune)
	}

	player := core.WithSpec(&proto.Player{
		Class:          proto.Class_ClassMage,
		Race:           proto.Race_RaceTroll,
		Equipment:      core.GetGearSet("../../ui/specs/mage/dps/gear_sets", "p1Arcane").GearSet,
		Consumables:    &proto.ConsumesSpec{},
		Buffs:          core.FullIndividualBuffs,
		TalentsString:  "2500052300030150330125--053500031003001",
		Profession1:    proto.Profession_Engineering,
		Rotation:       core.GetAplRotation("../../ui/specs/mage/dps/apls", "arcane").Rotation,
		ReactionTimeMs: 100,
	}, &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{ClassOptions: &proto.MageOptions{DefaultMageArmor: proto.MageArmor_MageArmorMageArmor}}}})
	baseRequest := &proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, core.FullPartyBuffs, core.FullRaidBuffs, core.FullDebuffs),
		Encounter:  core.MakeDefaultEncounterCombos()[0].Encounter,
		SimOptions: &proto.SimOptions{Iterations: 50, RandomSeed: 1},
	}

	// The candidate: the equipped gear with every socketed gem swapped for a Stamina gem.
	candidateGear := googleProto.Clone(player.Equipment).(*proto.EquipmentSpec)
	stamGems := 0
	for _, item := range candidateGear.Items {
		for i, gem := range item.GetGems() {
			if gem != 0 {
				if metaGem, ok := core.GetGemByID(gem); ok && metaGem.Color == proto.GemColor_GemColorMeta {
					continue
				}
				item.Gems[i] = solidStarOfElune
				stamGems++
			}
		}
	}
	if stamGems == 0 {
		t.Fatal("the gear set has no gems to swap; the case is unsuitable")
	}
	finalStamina := func(gear *proto.EquipmentSpec) float64 {
		raid := googleProto.Clone(baseRequest.Raid).(*proto.Raid)
		raid.Parties[0].Players[0].Equipment = gear
		return core.ComputeStats(&proto.ComputeStatsRequest{Raid: raid, Encounter: baseRequest.Encounter}).RaidStats.Parties[0].Players[0].FinalStats.Stats[proto.Stat_StatStamina]
	}
	// Exactly what the candidate's own gems give, well beyond anything a spell damage gem can.
	floor := finalStamina(candidateGear)

	ruby, _ := core.GetGemByID(runedLivingRuby)
	weights := make([]float64, int(proto.Stat_StatPhysicalDamage)+1)
	weights[proto.Stat_StatSpellDamage] = 1
	request := &proto.BulkSimRequest{
		BaseRequest:         baseRequest,
		Candidates:          []*proto.BulkGearCandidate{{Index: 0, Gear: candidateGear}},
		TopResults:          5,
		HighStageIterations: 50,
		BulkSettings: &proto.BulkSettings{
			UseLegacyBulkSim: true,
			StatConstraints: []*proto.BulkStatConstraint{{
				UnitStat: &proto.BulkStatConstraint_Stat{Stat: proto.Stat_StatStamina},
				Op:       proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual,
				Value:    floor,
			}},
		},
		ReforgeRequest: &proto.ReforgeOptimizeRequest{
			PreCapEpWeights: &proto.UnitStats{Stats: weights, PseudoStats: make([]float64, int(proto.PseudoStat_PseudoStatReducedCritTakenPercent)+1)},
			Settings: &proto.ReforgeSettings{
				MaxGemPhase:   5,
				MaxGemQuality: proto.ItemQuality_ItemQualityEpic,
				EpStats:       []proto.Stat{proto.Stat_StatSpellDamage, proto.Stat_StatStamina},
			},
			GemOptions: []*proto.ReforgeGemOption{{
				Id: ruby.ID, Name: ruby.Name, Color: ruby.Color, Stats: ruby.Stats[:], Quality: proto.ItemQuality_ItemQualityRare, Phase: 1,
			}},
		},
	}

	optimizeBulkSimReforgeCandidates(request, nil, simsignals.CreateSignals())
	request.ReforgeRequest = nil
	result := bulk.BulkSim(request)
	if result.Error != nil {
		t.Fatalf("batch failed: %s", result.Error.Message)
	}
	if result.SkippedByConstraints != 0 || len(result.TopResults) != 1 {
		t.Fatalf("the candidate's own gems meet the constraint, so it must be simmed: skipped %d, results %d", result.SkippedByConstraints, len(result.TopResults))
	}
	if got := finalStamina(result.TopResults[0].Gear); got < floor {
		t.Fatalf("the simmed gear has %.0f Stamina, below the constraint's %.0f", got, floor)
	}

	// The same gear without its Stamina gems, the floor is out of reach: the final-stats check
	// drops it, and the result counts it as skipped.
	request.Candidates = []*proto.BulkGearCandidate{{Index: 0, Gear: googleProto.Clone(candidateGear).(*proto.EquipmentSpec)}}
	for _, item := range request.Candidates[0].Gear.Items {
		for i, gem := range item.GetGems() {
			if gem == solidStarOfElune {
				item.Gems[i] = runedLivingRuby
			}
		}
	}
	request.BulkSettings.StatConstraints[0].Value = floor
	result = bulk.BulkSim(request)
	if result.Error != nil || result.SkippedByConstraints != 1 || len(result.TopResults) != 0 {
		t.Fatalf("gear below the floor must be skipped: err %v skipped %d results %d", result.Error, result.SkippedByConstraints, len(result.TopResults))
	}
}
