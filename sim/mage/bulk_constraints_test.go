//go:build with_db

package mage

import (
	"testing"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/bulk"
	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// Runs the batch sim on a real Mage with a stat constraint: of two candidate
// gear sets, only the one whose final fire resistance clears the constraint is
// simmed, and the result reports the other as skipped.
func TestBulkSimStatConstraints(t *testing.T) {
	const infernoweaveRobe = 30762 // +60 fire resistance

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

	result := bulk.BulkSim(request)
	if result.Error != nil {
		t.Fatalf("constrained batch failed: %s", result.Error.Message)
	}
	if result.SkippedByConstraints != 1 {
		t.Fatalf("skipped %d candidates, want 1", result.SkippedByConstraints)
	}
	if len(result.TopResults) != 1 || result.TopResults[0].Gear.Items[proto.ItemSlot_ItemSlotChest].Id != infernoweaveRobe {
		t.Fatalf("only the robe should survive, got %d results: %+v", len(result.TopResults), result.TopResults)
	}
	if result.Baseline == nil || result.Baseline.DpsMetrics == nil || result.Baseline.DpsMetrics.Avg <= 0 {
		t.Fatalf("baseline missing: %+v", result.Baseline)
	}

	// Nothing passes: still a successful run, with the baseline and no results.
	request.BulkSettings.StatConstraints[0].Value = baseFireRes + 1000
	result = bulk.BulkSim(request)
	if result.Error != nil || result.SkippedByConstraints != 2 || len(result.TopResults) != 0 || result.Baseline == nil {
		t.Fatalf("all-skipped batch: err %v skipped %d results %d", result.Error, result.SkippedByConstraints, len(result.TopResults))
	}

	// No constraints: both candidates are simmed and nothing is skipped.
	request.BulkSettings.StatConstraints = nil
	result = bulk.BulkSim(request)
	if result.Error != nil || result.SkippedByConstraints != 0 || len(result.TopResults) != 2 {
		t.Fatalf("unconstrained batch: err %v skipped %d results %d", result.Error, result.SkippedByConstraints, len(result.TopResults))
	}
}

// Constraints are judged on the values the stats panel shows, which include the raid debuffs the
// panel attributes to the character: Improved Seal of the Crusader's +3% crit and Improved Hunter's
// Mark's +110 attack power here. A candidate the panel shows as meeting a constraint must be simmed.
func TestBulkSimStatConstraintsSeeDebuffs(t *testing.T) {
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
		result := bulk.BulkSim(request)
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
