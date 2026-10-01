package stats

import (
	"testing"

	"github.com/wowsims/tbc/sim/core/proto"
)

func TestBulkStatConstraintPasses(t *testing.T) {
	for _, c := range []struct {
		op    proto.BulkStatConstraintOp
		value float64
		want  bool
	}{
		{proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan, 176, true},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan, 175, false},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, 175, true},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, 174, false},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpLessThanOrEqual, 175, true},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpLessThanOrEqual, 176, false},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpLessThan, 174, true},
		{proto.BulkStatConstraintOp_BulkStatConstraintOpLessThan, 175, false},
		// The reserved value 2 (the removed = operator) and anything unknown pass nothing.
		{proto.BulkStatConstraintOp(2), 175, false},
	} {
		if got := BulkStatConstraintPasses(c.op, c.value, 175); got != c.want {
			t.Fatalf("%v with threshold 175 and value %v: got %v, want %v", c.op, c.value, got, c.want)
		}
	}
}

func TestBulkStatConstraintUnitStat(t *testing.T) {
	stat, ok := BulkStatConstraintUnitStat(&proto.BulkStatConstraint{UnitStat: &proto.BulkStatConstraint_Stat{Stat: proto.Stat_StatFireResistance}})
	if !ok || !stat.EqualsStat(FireResistance) {
		t.Fatalf("stat target: got %v, %v", stat, ok)
	}
	pseudoStat, ok := BulkStatConstraintUnitStat(&proto.BulkStatConstraint{UnitStat: &proto.BulkStatConstraint_PseudoStat{PseudoStat: proto.PseudoStat_PseudoStatDodgePercent}})
	if !ok || !pseudoStat.EqualsPseudoStat(proto.PseudoStat_PseudoStatDodgePercent) {
		t.Fatalf("pseudo-stat target: got %v, %v", pseudoStat, ok)
	}
	if _, ok := BulkStatConstraintUnitStat(&proto.BulkStatConstraint{}); ok {
		t.Fatal("a constraint with no target should report none")
	}
}

func TestUnitStatValueFromStatsProto(t *testing.T) {
	fireRes := UnitStatFromStat(FireResistance)
	dodge := UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatDodgePercent)
	statsProto := &proto.UnitStats{Stats: make([]float64, ProtoStatsLen), PseudoStats: make([]float64, PseudoStatsLen)}
	statsProto.Stats[FireResistance] = 93
	statsProto.PseudoStats[proto.PseudoStat_PseudoStatDodgePercent] = 21.5
	if got := fireRes.ValueFromStatsProto(statsProto); got != 93 {
		t.Fatalf("stat: got %v, want 93", got)
	}
	if got := dodge.ValueFromStatsProto(statsProto); got != 21.5 {
		t.Fatalf("pseudo-stat: got %v, want 21.5", got)
	}
	// Arrays too short to hold the stat, and no stats at all, read as 0.
	if got := fireRes.ValueFromStatsProto(&proto.UnitStats{}) + dodge.ValueFromStatsProto(&proto.UnitStats{}) + fireRes.ValueFromStatsProto(nil); got != 0 {
		t.Fatalf("missing values: got %v, want 0", got)
	}
}
