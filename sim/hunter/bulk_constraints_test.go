//go:build with_db

package hunter

import (
	"testing"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/bulk"
	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// The stats panel credits a hunter who applies Expose Weakness with their own talent their own
// agility, not the agility configured for the debuff. A batch stat constraint on attack power is
// judged the same way.
func TestBulkSimStatConstraintsCreditOwnExposeWeakness(t *testing.T) {
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassHunter,
		Race:          proto.Race_RaceOrc,
		Equipment:     core.GetGearSet("../../ui/specs/hunter/dps/gear_sets/phase_2/bm", "2h_6p").GearSet,
		Consumables:   DefaultConsumables,
		Buffs:         core.FullIndividualBuffs,
		TalentsString: DefaultSVTalents, // Takes Expose Weakness.
		Profession1:   proto.Profession_Engineering,
		Rotation:      core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default").Rotation,
	}, DefaultOptions)
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
		result := bulk.BulkSim(&proto.BulkSimRequest{
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
