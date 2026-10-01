package stats

import (
	"github.com/wowsims/tbc/sim/core/proto"
)

// Batch stat constraints (BulkSettings.stat_constraints) are read in two places: the gem
// optimizer, which turns them into rows of its model, and the batch sim's final-stats check. Both
// take the stat a constraint targets and the comparison it makes from here.

// BulkStatConstraintUnitStat returns the Stat or PseudoStat a batch stat constraint targets. A
// constraint with no target returns false; both readers skip it.
func BulkStatConstraintUnitStat(constraint *proto.BulkStatConstraint) (UnitStat, bool) {
	switch target := constraint.GetUnitStat().(type) {
	case *proto.BulkStatConstraint_Stat:
		return UnitStatFromStat(Stat(target.Stat)), true
	case *proto.BulkStatConstraint_PseudoStat:
		return UnitStatFromPseudoStat(target.PseudoStat), true
	}
	return 0, false
}

func BulkStatConstraintPasses(op proto.BulkStatConstraintOp, value float64, threshold float64) bool {
	switch op {
	case proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan:
		return value > threshold
	case proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual:
		return value >= threshold
	case proto.BulkStatConstraintOp_BulkStatConstraintOpLessThanOrEqual:
		return value <= threshold
	case proto.BulkStatConstraintOp_BulkStatConstraintOpLessThan:
		return value < threshold
	}
	return false
}

// ValueFromStatsProto reads this stat out of a stats proto, 0 when the array is too short.
func (s UnitStat) ValueFromStatsProto(p *proto.UnitStats) float64 {
	if s.IsStat() {
		if s.StatIdx() < len(p.GetStats()) {
			return p.Stats[s.StatIdx()]
		}
	} else if s.PseudoStatIdx() < len(p.GetPseudoStats()) {
		return p.PseudoStats[s.PseudoStatIdx()]
	}
	return 0
}
