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

// An arcane mage in its preset gear, and the batch's base request for it: 50 iterations on a
// fixed seed.
func mageBulkBaseRequest() (*proto.Player, *proto.RaidSimRequest) {
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
	return player, &proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, core.FullPartyBuffs, core.FullRaidBuffs, core.FullDebuffs),
		Encounter:  core.MakeDefaultEncounterCombos()[0].Encounter,
		SimOptions: &proto.SimOptions{Iterations: 50, RandomSeed: 1},
	}
}

func reforgeGemOption(t *testing.T, id int32) *proto.ReforgeGemOption {
	t.Helper()
	gem, ok := core.GetGemByID(id)
	if !ok {
		t.Fatalf("gem %d is not in the database", id)
	}
	return &proto.ReforgeGemOption{Id: gem.ID, Name: gem.Name, Color: gem.Color, Stats: gem.Stats[:], Quality: proto.ItemQuality_ItemQualityRare, Phase: 1}
}

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

	player, baseRequest := mageBulkBaseRequest()

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
			GemOptions: []*proto.ReforgeGemOption{reforgeGemOption(t, runedLivingRuby)},
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

// A batch on a fixed seed gives the same result every time it is run. Here the gem optimizer has
// constraints on three stats to meet and only spell damage to score, so many gem layouts tie; the
// solver breaks ties by the model it is handed, which must therefore be the same on every run.
func TestBulkSimWithStatConstraintsIsDeterministic(t *testing.T) {
	const (
		runedLivingRuby       = 24030 // +9 spell damage: the only stat with a weight
		solidStarOfElune      = 24033 // Stamina
		brilliantDawnstone    = 24047 // Intellect
		sparklingStarOfElune  = 24035 // Spirit
		infernoweaveRobe      = 30762
		constraintGemsPerStat = 2
	)
	player, baseRequest := mageBulkBaseRequest()

	// The floors are measured from the gear with only spell damage gems, so each takes gems of its
	// own stat to reach.
	rubyGear := func(gear *proto.EquipmentSpec) *proto.EquipmentSpec {
		gear = googleProto.Clone(gear).(*proto.EquipmentSpec)
		for _, item := range gear.Items {
			for i, gemID := range item.GetGems() {
				if gem, ok := core.GetGemByID(gemID); ok && gem.Color != proto.GemColor_GemColorMeta {
					item.Gems[i] = runedLivingRuby
				}
			}
		}
		return gear
	}
	finalStats := func(gear *proto.EquipmentSpec) []float64 {
		raid := googleProto.Clone(baseRequest.Raid).(*proto.Raid)
		raid.Parties[0].Players[0].Equipment = gear
		return core.ComputeStats(&proto.ComputeStatsRequest{Raid: raid, Encounter: baseRequest.Encounter}).RaidStats.Parties[0].Players[0].FinalStats.Stats
	}
	robeGear := googleProto.Clone(player.Equipment).(*proto.EquipmentSpec)
	robeGear.Items[proto.ItemSlot_ItemSlotChest] = &proto.ItemSpec{Id: infernoweaveRobe}
	ringlessGear := googleProto.Clone(player.Equipment).(*proto.EquipmentSpec)
	ringlessGear.Items[proto.ItemSlot_ItemSlotFinger2] = &proto.ItemSpec{}
	candidateGear := []*proto.EquipmentSpec{rubyGear(player.Equipment), rubyGear(robeGear), rubyGear(ringlessGear)}

	// Each floor is what the weakest candidate has, plus most of what two gems of the stat give.
	floor := func(stat proto.Stat, gemID int32) *proto.BulkStatConstraint {
		gem, _ := core.GetGemByID(gemID)
		if gem.Stats[stat] <= 0 {
			t.Fatalf("gem %d should carry %s", gemID, stat)
		}
		lowest := finalStats(candidateGear[0])[stat]
		for _, gear := range candidateGear[1:] {
			lowest = min(lowest, finalStats(gear)[stat])
		}
		return &proto.BulkStatConstraint{
			UnitStat: &proto.BulkStatConstraint_Stat{Stat: stat},
			Op:       proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual,
			Value:    lowest + gem.Stats[stat]*(constraintGemsPerStat-0.5),
		}
	}

	weights := make([]float64, int(proto.Stat_StatPhysicalDamage)+1)
	weights[proto.Stat_StatSpellDamage] = 1
	request := &proto.BulkSimRequest{
		BaseRequest:         baseRequest,
		TopResults:          5,
		HighStageIterations: 50,
		BulkSettings: &proto.BulkSettings{
			UseLegacyBulkSim: true,
			StatConstraints: []*proto.BulkStatConstraint{
				floor(proto.Stat_StatStamina, solidStarOfElune),
				floor(proto.Stat_StatIntellect, brilliantDawnstone),
				floor(proto.Stat_StatSpirit, sparklingStarOfElune),
			},
		},
		ReforgeRequest: &proto.ReforgeOptimizeRequest{
			PreCapEpWeights: &proto.UnitStats{Stats: weights, PseudoStats: make([]float64, int(proto.PseudoStat_PseudoStatReducedCritTakenPercent)+1)},
			Settings: &proto.ReforgeSettings{
				MaxGemPhase:   5,
				MaxGemQuality: proto.ItemQuality_ItemQualityEpic,
				EpStats:       []proto.Stat{proto.Stat_StatSpellDamage, proto.Stat_StatStamina, proto.Stat_StatIntellect, proto.Stat_StatSpirit},
			},
			GemOptions: []*proto.ReforgeGemOption{
				reforgeGemOption(t, runedLivingRuby),
				reforgeGemOption(t, solidStarOfElune),
				reforgeGemOption(t, brilliantDawnstone),
				reforgeGemOption(t, sparklingStarOfElune),
			},
		},
	}
	for i, gear := range candidateGear {
		request.Candidates = append(request.Candidates, &proto.BulkGearCandidate{Index: int32(i), Gear: gear})
	}

	run := func() *proto.BulkSimResult {
		request := googleProto.Clone(request).(*proto.BulkSimRequest)
		optimizeBulkSimReforgeCandidates(request, nil, simsignals.CreateSignals())
		request.ReforgeRequest = nil
		result := bulk.BulkSim(request)
		if result.Error != nil {
			t.Fatalf("batch failed: %s", result.Error.Message)
		}
		return result
	}

	first := run()
	if first.SkippedByConstraints != 0 || len(first.TopResults) != len(candidateGear) {
		t.Fatalf("the gems can meet every floor, so every candidate must be simmed: skipped %d, results %d", first.SkippedByConstraints, len(first.TopResults))
	}
	for repeat := 2; repeat <= 6; repeat++ {
		again := run()
		if len(again.TopResults) != len(first.TopResults) {
			t.Fatalf("run %d simmed %d gear sets, the first run %d", repeat, len(again.TopResults), len(first.TopResults))
		}
		for i, result := range again.TopResults {
			want := first.TopResults[i]
			if !googleProto.Equal(result.Gear, want.Gear) {
				t.Fatalf("run %d, result %d: the gear differs from the first run's\n got: %v\nwant: %v", repeat, i, result.Gear, want.Gear)
			}
			if result.DpsMetrics.Avg != want.DpsMetrics.Avg {
				t.Fatalf("run %d, result %d: %v DPS, the first run %v", repeat, i, result.DpsMetrics.Avg, want.DpsMetrics.Avg)
			}
		}
	}
}

// Without the gem optimizer the server checks the same gear sets as with it, and as the browser
// does: duplicates and the equipped gear, which is simmed as the baseline, are not candidates. The
// candidate generator always includes the all-equipped combination, and it used to be counted as
// checked, and as skipped when it failed a constraint, while the results still showed it as the
// baseline.
func TestBulkSimWithoutGemOptimizerDoesNotCheckEquippedGear(t *testing.T) {
	const infernoweaveRobe = 30762
	player, baseRequest := mageBulkBaseRequest()
	robeGear := googleProto.Clone(player.Equipment).(*proto.EquipmentSpec)
	robeGear.Items[proto.ItemSlot_ItemSlotChest] = &proto.ItemSpec{Id: infernoweaveRobe}

	request := &proto.BulkSimRequest{
		BaseRequest: baseRequest,
		Candidates: []*proto.BulkGearCandidate{
			{Index: 0, Gear: googleProto.Clone(player.Equipment).(*proto.EquipmentSpec)},
			{Index: 1, Gear: robeGear},
			{Index: 2, Gear: googleProto.Clone(robeGear).(*proto.EquipmentSpec)},
		},
		TopResults:          5,
		HighStageIterations: 50,
		BulkSettings: &proto.BulkSettings{
			UseLegacyBulkSim: true,
			// Out of reach for every gear set here, the equipped one included.
			StatConstraints: []*proto.BulkStatConstraint{{
				UnitStat: &proto.BulkStatConstraint_Stat{Stat: proto.Stat_StatFireResistance},
				Op:       proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual,
				Value:    10000,
			}},
		},
	}

	progress := make(chan *proto.ProgressMetrics, 100)
	go runBulkSimAsync(request, progress, "bulk-constraints-test-no-gem-optimizer")
	var result *proto.BulkSimResult
	for update := range progress {
		if update.FinalBulkSimResult != nil {
			result = update.FinalBulkSimResult
		}
	}
	if result == nil {
		t.Fatal("the batch ended without a result")
	}
	if result.Error != nil {
		t.Fatalf("batch failed: %s", result.Error.Message)
	}
	if result.CheckedByConstraints != 1 || result.SkippedByConstraints != 1 {
		t.Fatalf("skipped %d of %d gear sets checked, want 1 of 1: the robe, once", result.SkippedByConstraints, result.CheckedByConstraints)
	}
	if result.Baseline == nil || result.Baseline.DpsMetrics.GetAvg() <= 0 {
		t.Fatalf("the equipped gear should still be simmed as the baseline, got %+v", result.Baseline)
	}
}

// The solver reads the stat constraints from the gem optimizer's own request, the copy the
// client's gem cache key is made from, so the two cannot drift apart. The batch settings' copy is
// a fallback for a request that set none there.
func TestBulkSimReforgeOptimizerStatConstraintSource(t *testing.T) {
	staminaAtLeast := func(value float64) *proto.BulkStatConstraint {
		return &proto.BulkStatConstraint{
			UnitStat: &proto.BulkStatConstraint_Stat{Stat: proto.Stat_StatStamina},
			Op:       proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual,
			Value:    value,
		}
	}
	solverConstraints := func(reforgeConstraints []*proto.BulkStatConstraint, settingsConstraints []*proto.BulkStatConstraint) []*proto.BulkStatConstraint {
		return newBulkSimReforgeOptimizer(&proto.BulkSimRequest{
			BaseRequest:    &proto.RaidSimRequest{Raid: &proto.Raid{}},
			ReforgeRequest: &proto.ReforgeOptimizeRequest{StatConstraints: reforgeConstraints},
			BulkSettings:   &proto.BulkSettings{StatConstraints: settingsConstraints},
		}).templateRequest.StatConstraints
	}

	own := solverConstraints([]*proto.BulkStatConstraint{staminaAtLeast(500)}, []*proto.BulkStatConstraint{staminaAtLeast(1000)})
	if len(own) != 1 || own[0].Value != 500 {
		t.Fatalf("the solver should be given the gem optimizer request's own constraints, got %v", own)
	}
	fallback := solverConstraints(nil, []*proto.BulkStatConstraint{staminaAtLeast(1000)})
	if len(fallback) != 1 || fallback[0].Value != 1000 {
		t.Fatalf("a request with none of its own should be given the batch settings' constraints, got %v", fallback)
	}
}
