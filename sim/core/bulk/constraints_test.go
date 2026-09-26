package bulk

import (
	"testing"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/simsignals"
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
	cases := []struct {
		op    proto.BulkStatConstraintOp
		value float64
		want  bool
	}{
		{proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan, 176, true},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan, 175, false},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, 175, true},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, 174, false},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpEqual, 175, true},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpEqual, 174, false},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpLessThanOrEqual, 175, true},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpLessThanOrEqual, 176, false},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpLessThan, 174, true},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpLessThan, 175, false},
	}
	for _, c := range cases {
		if got := bulkStatConstraintPasses(stat(proto.Stat_StatFireResistance, c.op, 175), c.value); got != c.want {
			t.Fatalf("%v with threshold 175 and value %v: got %v, want %v", c.op, c.value, got, c.want)
		}
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
	if got := bulkConstraintStatValue(&proto.BulkStatConstraint{}, finalStats(1, 1)); got != 0 {
		t.Fatalf("unset target reads as 0, got %v", got)
	}
	if got := bulkConstraintStatValue(constraints[0], nil); got != 0 {
		t.Fatalf("missing stats read as 0, got %v", got)
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
