package reforgeoptimizer

import (
	"errors"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// Batch stat constraints (ReforgeOptimizeRequest.stat_constraints) become rows of the
// model, in the same gap-to-base space as the caps: the chosen gems must move the stat
// from its base value (gems cleared) to the constrained side of the threshold. A stat no
// variable can move is decided on the base value before solving. Either way a candidate
// whose constraints cannot be met is reported as infeasible rather than gemmed as best
// effort; the batch sim drops it. The final stats of the gear the solver returns are
// checked again by the batch sim, so this steers the gems while that remains the gate.

// errStatConstraintsInfeasible reports that no gem choice can satisfy the constraints.
var errStatConstraintsInfeasible = errors.New("No gem choice satisfies the batch's stat constraints.")

// Strict comparisons are modelled as the inclusive bound moved by this much. Gem
// contributions are whole stat points or percentages far coarser than this.
const statConstraintStrictEpsilon = 1e-6

func statConstraintUnitStat(constraint *proto.BulkStatConstraint) (stats.UnitStat, bool) {
	switch target := constraint.GetUnitStat().(type) {
	case *proto.BulkStatConstraint_Stat:
		return stats.UnitStatFromStat(stats.Stat(target.Stat)), true
	case *proto.BulkStatConstraint_PseudoStat:
		return stats.UnitStatFromPseudoStat(target.PseudoStat), true
	}
	return 0, false
}

func statConstraintPasses(op proto.BulkStatConstraintOp, value float64, threshold float64) bool {
	switch op {
	case proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan:
		return value > threshold
	case proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual:
		return value >= threshold
	case proto.BulkStatConstraintOp_BulkStatConstraintOpEqual:
		return value == threshold
	case proto.BulkStatConstraintOp_BulkStatConstraintOpLessThanOrEqual:
		return value <= threshold
	case proto.BulkStatConstraintOp_BulkStatConstraintOpLessThan:
		return value < threshold
	}
	return false
}

// Tightens a row with another bound on the same stat: the larger of the minimums, the
// smaller of the maximums.
func tightenConstraint(row lpConstraint, bound lpConstraint) lpConstraint {
	if bound.hasMin && (!row.hasMin || bound.min > row.min) {
		row.min, row.hasMin = bound.min, true
	}
	if bound.hasMax && (!row.hasMax || bound.max < row.max) {
		row.max, row.hasMax = bound.max, true
	}
	return row
}

func statConstraintBound(op proto.BulkStatConstraintOp, gap float64) (lpConstraint, bool) {
	switch op {
	case proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan:
		return greaterEq(gap + statConstraintStrictEpsilon), true
	case proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual:
		return greaterEq(gap), true
	case proto.BulkStatConstraintOp_BulkStatConstraintOpEqual:
		return lpConstraint{min: gap, hasMin: true, max: gap, hasMax: true}, true
	case proto.BulkStatConstraintOp_BulkStatConstraintOpLessThanOrEqual:
		return lessEq(gap), true
	case proto.BulkStatConstraintOp_BulkStatConstraintOpLessThan:
		return lessEq(gap - statConstraintStrictEpsilon), true
	}
	return lpConstraint{}, false
}

// The character sheet floors defense to whole defense points, while the model credits defense
// rating linearly, so a stat with a defense term can end up to one defense point away from the
// model's value, in either direction. For those stats a bound is moved one point inward when gems
// can carry defense, so the floored value still meets it. Equality has no such room and is left
// as it is.
func defenseFloorMargin(unitStat stats.UnitStat, op proto.BulkStatConstraintOp, variables *lpVariables) float64 {
	if !unitStat.IsPseudoStat() || op == proto.BulkStatConstraintOp_BulkStatConstraintOpEqual {
		return 0
	}
	switch proto.PseudoStat(unitStat.PseudoStatIdx()) {
	case proto.PseudoStat_PseudoStatReducedCritTakenPercent, proto.PseudoStat_PseudoStatDodgePercent, proto.PseudoStat_PseudoStatParryPercent:
		if variablesCarryKey(variables, statCoeffKey(proto.Stat_StatDefenseRating)) {
			return core.MissDodgeParryBlockCritChancePerDefense
		}
	}
	return 0
}

// constrainedStatKeys returns the coefficient keys of the stats the request's stat constraints
// name, or nil when there are none.
func (o *reforgeOptimizer) constrainedStatKeys() []string {
	var keys []string
	for _, constraint := range o.request.GetStatConstraints() {
		if unitStat, ok := statConstraintUnitStat(constraint); ok {
			keys = append(keys, coeffKeyForUnitStat(unitStat))
		}
	}
	return keys
}

func gemMovesConstrainedStat(gem gemData, constrainedKeys []string) bool {
	for _, key := range constrainedKeys {
		if gem.capCoeffs[key] != 0 {
			return true
		}
	}
	return false
}

func variablesCarryKey(variables *lpVariables, key string) bool {
	found := false
	variables.each(func(_ string, coeffs map[string]float64) {
		if coeff, ok := coeffs[key]; ok && coeff != 0 {
			found = true
		}
	})
	return found
}

// The row key of a stat constraint on the stat with the given coefficient key. Constraint rows
// have their own keys, apart from the cap rows keyed by the stat itself, so cap refinement on the
// same stat adds its row as usual (pinning the stat and zeroing its value once a solution passes
// the cap) rather than skipping it, and soft-cap refinement does not overwrite the constraint.
// Every variable that carries the stat carries the same coefficient under this key.
func statConstraintRowKey(statKey string) string {
	return "StatConstraint_" + statKey
}

// statConstraintRows returns the model rows for the request's stat constraints, adding their
// coefficients to the variables. Returns errStatConstraintsInfeasible when a constraint on a
// stat the gems cannot move already fails on the base stats.
func (o *reforgeOptimizer) statConstraintRows(variables *lpVariables) (map[string]lpConstraint, error) {
	rows := make(map[string]lpConstraint)
	for _, constraint := range o.request.GetStatConstraints() {
		unitStat, ok := statConstraintUnitStat(constraint)
		if !ok {
			continue
		}
		bound, ok := statConstraintBound(constraint.GetOp(), 0)
		if !ok {
			continue
		}
		key := coeffKeyForUnitStat(unitStat)
		base := getUnitStat(o.capBaseStats, unitStat)
		if !variablesCarryKey(variables, key) {
			if !statConstraintPasses(constraint.GetOp(), base, constraint.GetValue()) {
				return nil, errStatConstraintsInfeasible
			}
			continue
		}
		bound, _ = statConstraintBound(constraint.GetOp(), constraint.GetValue()-base)
		margin := defenseFloorMargin(unitStat, constraint.GetOp(), variables)
		if bound.hasMin {
			bound.min += margin
		}
		if bound.hasMax {
			bound.max -= margin
		}
		rowKey := statConstraintRowKey(key)
		variables.each(func(_ string, coeffs map[string]float64) {
			if coeff, ok := coeffs[key]; ok && coeff != 0 {
				coeffs[rowKey] = coeff
			}
		})
		rows[rowKey] = tightenConstraint(rows[rowKey], bound)
	}
	return rows, nil
}

// statConstraintsCauseInfeasibility reports whether an infeasible model is infeasible because of
// the stat constraints' rows: the same model without them has a solution. When they are not the
// cause (a meta gem's colour rule, an upper-bound cap the gear already exceeds), the failure is
// reported as that, so the batch falls back to the candidate's own gear instead of dropping it.
// A relaxed solve that ends without an answer (a timeout) does not blame the constraints either.
func (o *reforgeOptimizer) statConstraintsCauseInfeasibility(model *lpModel, maxSeconds float64) bool {
	if len(o.statConstraintRowKeys) == 0 {
		return false
	}
	relaxed := newLPConstraints()
	model.constraints.each(func(name string, row lpConstraint) {
		if !o.statConstraintRowKeys[name] {
			relaxed.set(name, row)
		}
	})
	relaxedModel := *model
	relaxedModel.constraints = relaxed
	solution, err := solveLPModel(&relaxedModel, time.Duration(max(maxSeconds, minSolveSeconds)*float64(time.Second)), 0)
	return err == nil && solution.status == "optimal"
}
