//go:build with_db

package bulk

import (
	"testing"

	"github.com/wowsims/tbc/sim"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

func TestBulkStatConstraints(t *testing.T) {
	stat := func(s proto.Stat, op proto.BulkStatConstraintOp, value float64) *proto.BulkStatConstraint {
		return &proto.BulkStatConstraint{UnitStat: &proto.BulkStatConstraint_Stat{Stat: s}, Op: op, Value: value}
	}
	pseudo := func(p proto.PseudoStat, op proto.BulkStatConstraintOp, value float64) *proto.BulkStatConstraint {
		return &proto.BulkStatConstraint{UnitStat: &proto.BulkStatConstraint_PseudoStat{PseudoStat: p}, Op: op, Value: value}
	}
	finalStats := func(fireRes float64, critReduction float64) *proto.UnitStats {
		stats := &proto.UnitStats{Stats: make([]float64, int(proto.Stat_StatPhysicalDamage)+1), PseudoStats: make([]float64, int(proto.PseudoStat_PseudoStatReducedCritTakenPercent)+1)}
		stats.Stats[proto.Stat_StatFireResistance] = fireRes
		stats.PseudoStats[proto.PseudoStat_PseudoStatReducedCritTakenPercent] = critReduction
		return stats
	}
	constraints := []*proto.BulkStatConstraint{
		stat(proto.Stat_StatFireResistance, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan, 175),
		pseudo(proto.PseudoStat_PseudoStatReducedCritTakenPercent, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, 5.6),
	}
	if !bulkFinalStatsPassConstraints(constraints, finalStats(200, 5.6)) {
		t.Fatal("200 fire res and 5.6 crit reduction should pass")
	}
	if bulkFinalStatsPassConstraints(constraints, finalStats(175, 5.6)) {
		t.Fatal("175 fire res should fail a > 175 constraint")
	}
	if bulkFinalStatsPassConstraints(constraints, finalStats(200, 5.2)) {
		t.Fatal("5.2 crit reduction should fail a >= 5.6 constraint")
	}
	if !bulkFinalStatsPassConstraints(nil, &proto.UnitStats{}) {
		t.Fatal("no constraints always pass")
	}
	// A constraint with no stat set is skipped, as the gem optimizer skips it, not judged as 0.
	noTarget := &proto.BulkStatConstraint{Op: proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan, Value: 100}
	if !bulkFinalStatsPassConstraints([]*proto.BulkStatConstraint{noTarget}, finalStats(1, 1)) {
		t.Fatal("a constraint with no stat set should be skipped")
	}
	if bulkFinalStatsPassConstraints(append([]*proto.BulkStatConstraint{noTarget}, constraints...), finalStats(175, 5.6)) {
		t.Fatal("the constraints after one with no stat set should still be judged")
	}
	// Missing stats read as 0.
	if bulkFinalStatsPassConstraints(constraints, nil) {
		t.Fatal("no stats should fail a > 175 constraint")
	}
}

// With no constraints the filter is a no-op that computes nothing, whatever the request looks like.
func TestFilterBulkSimCandidatesByConstraintsNoConstraints(t *testing.T) {
	candidates := []BulkSimCandidate{{Index: 0}, {Index: 1}}
	survivors, skipped, err := filterBulkSimCandidatesByConstraints(&proto.BulkSimRequest{}, candidates, nil, simsignals.CreateSignals())
	if err != nil || skipped != 0 || len(survivors) != 2 {
		t.Fatalf("no constraints: survivors %d skipped %d err %v", len(survivors), skipped, err)
	}
}

// Runs the batch sim on a real Mage with a stat constraint: of two candidate
// gear sets, only the one whose final fire resistance clears the constraint is
// simmed, and the result reports the other as skipped.
func TestBulkSimStatConstraints(t *testing.T) {
	sim.RegisterAll()
	const infernoweaveRobe = 30762 // +60 fire resistance

	player := core.WithSpec(&proto.Player{
		Class:          proto.Class_ClassMage,
		Race:           proto.Race_RaceTroll,
		Equipment:      core.GetGearSet("../../../ui/specs/mage/dps/gear_sets", "p1Arcane").GearSet,
		Consumables:    &proto.ConsumesSpec{},
		Buffs:          core.FullIndividualBuffs,
		TalentsString:  "2500052300030150330125--053500031003001",
		Profession1:    proto.Profession_Engineering,
		Rotation:       core.GetAplRotation("../../../ui/specs/mage/dps/apls", "arcane").Rotation,
		ReactionTimeMs: 100,
	}, &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{ClassOptions: &proto.MageOptions{DefaultMageArmor: proto.MageArmor_MageArmorMageArmor}}}})

	base := &proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, core.FullPartyBuffs, core.FullRaidBuffs, core.FullDebuffs),
		Encounter:  core.MakeDefaultEncounterCombos()[0].Encounter,
		SimOptions: &proto.SimOptions{Iterations: 50, RandomSeed: 1},
	}
	baseline := player.Equipment
	robeGear := googleProto.Clone(baseline).(*proto.EquipmentSpec)
	robeGear.Items[proto.ItemSlot_ItemSlotChest] = &proto.ItemSpec{Id: infernoweaveRobe}
	shoulderlessGear := googleProto.Clone(baseline).(*proto.EquipmentSpec)
	shoulderlessGear.Items[proto.ItemSlot_ItemSlotShoulder] = &proto.ItemSpec{}

	stats := core.ComputeStats(&proto.ComputeStatsRequest{Raid: base.Raid, Encounter: base.Encounter})
	baseFireRes := stats.RaidStats.Parties[0].Players[0].FinalStats.Stats[proto.Stat_StatFireResistance]

	request := &proto.BulkSimRequest{
		BaseRequest: base,
		Candidates: []*proto.BulkGearCandidate{
			{Index: 0, Gear: robeGear},
			{Index: 1, Gear: shoulderlessGear},
		},
		TopResults:          5,
		HighStageIterations: 50,
		BulkSettings: &proto.BulkSettings{
			UseLegacyBulkSim: true,
			StatConstraints: []*proto.BulkStatConstraint{{
				UnitStat: &proto.BulkStatConstraint_Stat{Stat: proto.Stat_StatFireResistance},
				Op:       proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan,
				Value:    baseFireRes + 55,
			}},
		},
	}

	result := BulkSim(request)
	if result.Error != nil {
		t.Fatalf("constrained batch failed: %s", result.Error.Message)
	}
	if result.SkippedByConstraints != 1 || result.CheckedByConstraints != 2 {
		t.Fatalf("skipped %d of %d candidates checked, want 1 of 2", result.SkippedByConstraints, result.CheckedByConstraints)
	}
	if len(result.TopResults) != 1 || result.TopResults[0].Gear.Items[proto.ItemSlot_ItemSlotChest].Id != infernoweaveRobe {
		t.Fatalf("only the robe should survive, got %d results: %+v", len(result.TopResults), result.TopResults)
	}
	if result.Baseline == nil || result.Baseline.DpsMetrics == nil || result.Baseline.DpsMetrics.Avg <= 0 {
		t.Fatalf("baseline missing: %+v", result.Baseline)
	}

	// Nothing passes: still a successful run, with the baseline and no results.
	request.BulkSettings.StatConstraints[0].Value = baseFireRes + 1000
	result = BulkSim(request)
	if result.Error != nil || result.SkippedByConstraints != 2 || len(result.TopResults) != 0 || result.Baseline == nil {
		t.Fatalf("all-skipped batch: err %v skipped %d results %d", result.Error, result.SkippedByConstraints, len(result.TopResults))
	}

	// No constraints: both candidates are simmed and nothing is skipped.
	request.BulkSettings.StatConstraints = nil
	result = BulkSim(request)
	if result.Error != nil || result.SkippedByConstraints != 0 || result.CheckedByConstraints != 0 || len(result.TopResults) != 2 {
		t.Fatalf("unconstrained batch: err %v skipped %d results %d", result.Error, result.SkippedByConstraints, len(result.TopResults))
	}
}

// Constraints are judged on the values the stats panel shows, which include the raid debuffs the
// panel attributes to the character: Improved Seal of the Crusader's +3% crit and Improved Hunter's
// Mark's +110 attack power here. A candidate the panel shows as meeting a constraint must be simmed.
func TestBulkSimStatConstraintsSeeDebuffs(t *testing.T) {
	sim.RegisterAll()
	player := core.WithSpec(&proto.Player{
		Class:          proto.Class_ClassMage,
		Race:           proto.Race_RaceTroll,
		Equipment:      core.GetGearSet("../../../ui/specs/mage/dps/gear_sets", "p1Arcane").GearSet,
		Consumables:    &proto.ConsumesSpec{},
		Buffs:          core.FullIndividualBuffs,
		TalentsString:  "2500052300030150330125--053500031003001",
		Profession1:    proto.Profession_Engineering,
		Rotation:       core.GetAplRotation("../../../ui/specs/mage/dps/apls", "arcane").Rotation,
		ReactionTimeMs: 100,
	}, &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{ClassOptions: &proto.MageOptions{DefaultMageArmor: proto.MageArmor_MageArmorMageArmor}}}})
	debuffs := googleProto.Clone(core.FullDebuffs).(*proto.Debuffs)
	debuffs.ImprovedSealOfTheCrusader = proto.TristateEffect_TristateEffectImproved
	debuffs.HuntersMark = proto.TristateEffect_TristateEffectImproved
	base := &proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, core.FullPartyBuffs, core.FullRaidBuffs, debuffs),
		Encounter:  core.MakeDefaultEncounterCombos()[0].Encounter,
		SimOptions: &proto.SimOptions{Iterations: 50, RandomSeed: 1},
	}
	// A candidate that differs from the equipped gear only in its chest.
	candidateGear := googleProto.Clone(player.Equipment).(*proto.EquipmentSpec)
	candidateGear.Items[proto.ItemSlot_ItemSlotChest] = &proto.ItemSpec{Id: 30762}
	raid := googleProto.Clone(base.Raid).(*proto.Raid)
	raid.Parties[0].Players[0].Equipment = candidateGear
	raw := core.ComputeStats(&proto.ComputeStatsRequest{Raid: raid, Encounter: base.Encounter}).RaidStats.Parties[0].Players[0].FinalStats
	rawSpellCrit := raw.PseudoStats[proto.PseudoStat_PseudoStatSpellCritPercent]
	rawAttackPower := raw.Stats[proto.Stat_StatAttackPower]

	run := func(constraint *proto.BulkStatConstraint) *proto.BulkSimResult {
		request := &proto.BulkSimRequest{
			BaseRequest:         base,
			Candidates:          []*proto.BulkGearCandidate{{Index: 0, Gear: candidateGear}},
			TopResults:          5,
			HighStageIterations: 50,
			BulkSettings:        &proto.BulkSettings{UseLegacyBulkSim: true, StatConstraints: []*proto.BulkStatConstraint{constraint}},
		}
		result := BulkSim(request)
		if result.Error != nil {
			t.Fatalf("batch failed: %s", result.Error.Message)
		}
		return result
	}
	atLeast := proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual
	spellCrit := func(value float64) *proto.BulkStatConstraint {
		return &proto.BulkStatConstraint{UnitStat: &proto.BulkStatConstraint_PseudoStat{PseudoStat: proto.PseudoStat_PseudoStatSpellCritPercent}, Op: atLeast, Value: value}
	}
	attackPower := func(value float64) *proto.BulkStatConstraint {
		return &proto.BulkStatConstraint{UnitStat: &proto.BulkStatConstraint_Stat{Stat: proto.Stat_StatAttackPower}, Op: atLeast, Value: value}
	}

	for _, tc := range []struct {
		name       string
		constraint *proto.BulkStatConstraint
		simmed     bool
	}{
		{"spell crit as the panel shows it, with the seal's 3%", spellCrit(rawSpellCrit + 3), true},
		{"attack power as the panel shows it, with the mark's 110", attackPower(rawAttackPower + 110), true},
		{"spell crit beyond what the panel shows", spellCrit(rawSpellCrit + 3.5), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := run(tc.constraint)
			if simmed := len(result.TopResults) == 1 && result.SkippedByConstraints == 0; simmed != tc.simmed {
				t.Fatalf("simmed = %v, want %v (skipped %d, results %d)", simmed, tc.simmed, result.SkippedByConstraints, len(result.TopResults))
			}
		})
	}
}

// The stats panel credits a hunter who applies Expose Weakness with their own talent their own
// agility, not the agility configured for the debuff. A batch stat constraint on attack power is
// judged the same way.
func TestBulkSimStatConstraintsCreditOwnExposeWeakness(t *testing.T) {
	sim.RegisterAll()
	player := core.WithSpec(&proto.Player{
		Class:     proto.Class_ClassHunter,
		Race:      proto.Race_RaceOrc,
		Equipment: core.GetGearSet("../../../ui/specs/hunter/dps/gear_sets/phase_2/bm", "2h_6p").GearSet,
		Consumables: &proto.ConsumesSpec{
			BattleElixirId:   22831,
			GuardianElixirId: 22840,
			FoodId:           27659,
			PotId:            22838,
			ConjuredId:       12662,
			ExplosiveId:      30217,
			PetFoodId:        33874,
			PetScrollAgi:     true,
			PetScrollStr:     true,
			SuperSapper:      true,
			GoblinSapper:     true,
			ScrollAgi:        true,
			ScrollStr:        true,
		},
		Buffs:         core.FullIndividualBuffs,
		TalentsString: "502-0550201205-333200022003223005103", // Survival, with Expose Weakness.
		Profession1:   proto.Profession_Engineering,
		Rotation:      core.GetAplRotation("../../../ui/specs/hunter/dps/apls", "default").Rotation,
	}, &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{ClassOptions: &proto.HunterOptions{
		Ammo:        proto.HunterOptions_AdamantiteStinger,
		PetType:     proto.HunterOptions_Ravager,
		PetUptime:   100.0,
		QuiverBonus: proto.HunterOptions_Speed15,
	}}}})
	// The debuff's configured agility is far below the hunter's own, so the two are told apart.
	debuffs := &proto.Debuffs{ExposeWeaknessUptime: 1, ExposeWeaknessHunterAgility: 100}
	base := &proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, core.FullPartyBuffs, core.FullRaidBuffs, debuffs),
		Encounter:  core.MakeDefaultEncounterCombos()[0].Encounter,
		SimOptions: &proto.SimOptions{Iterations: 50, RandomSeed: 1},
	}
	// A candidate that differs from the equipped gear: one ring fewer.
	candidateGear := googleProto.Clone(player.Equipment).(*proto.EquipmentSpec)
	candidateGear.Items[proto.ItemSlot_ItemSlotFinger2] = &proto.ItemSpec{}
	raid := googleProto.Clone(base.Raid).(*proto.Raid)
	raid.Parties[0].Players[0].Equipment = candidateGear
	raw := core.ComputeStats(&proto.ComputeStatsRequest{Raid: raid, Encounter: base.Encounter}).RaidStats.Parties[0].Players[0].FinalStats
	rawAttackPower := raw.Stats[proto.Stat_StatAttackPower]
	ownAgility := raw.Stats[proto.Stat_StatAgility]
	if ownAgility < 200 {
		t.Fatalf("the hunter has only %.0f agility; the case cannot tell own from configured", ownAgility)
	}
	sheetAttackPower := rawAttackPower + ownAgility*0.25

	run := func(minAttackPower float64) *proto.BulkSimResult {
		result := BulkSim(&proto.BulkSimRequest{
			BaseRequest:         base,
			Candidates:          []*proto.BulkGearCandidate{{Index: 0, Gear: candidateGear}},
			TopResults:          5,
			HighStageIterations: 50,
			BulkSettings: &proto.BulkSettings{
				UseLegacyBulkSim: true,
				StatConstraints: []*proto.BulkStatConstraint{{
					UnitStat: &proto.BulkStatConstraint_Stat{Stat: proto.Stat_StatAttackPower},
					Op:       proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual,
					Value:    minAttackPower,
				}},
			},
		})
		if result.Error != nil {
			t.Fatalf("batch failed: %s", result.Error.Message)
		}
		return result
	}
	for _, tc := range []struct {
		name           string
		minAttackPower float64
		simmed         bool
	}{
		{"more than the configured agility gives, less than the hunter's own", rawAttackPower + 100*0.25 + 1, true},
		{"exactly what the panel shows", sheetAttackPower, true},
		{"beyond what the panel shows", sheetAttackPower + 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := run(tc.minAttackPower)
			if simmed := len(result.TopResults) == 1 && result.SkippedByConstraints == 0; simmed != tc.simmed {
				t.Fatalf("simmed = %v, want %v (skipped %d, results %d)", simmed, tc.simmed, result.SkippedByConstraints, len(result.TopResults))
			}
		})
	}
}
