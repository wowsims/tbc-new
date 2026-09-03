package reforgeoptimizer

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// solver.go implements solveModel + checkCaps (the cap-refinement loop) and the thin adapter
// that runs the LP text through the shared HiGHS runtime. The loop is pure LP: caps are checked
// against the summed LP coefficients of the selected variables, so no sim recompute happens
// between passes.

// HiGHS model-status codes, shared by both runtime backends (the native wazero runner and the
// wasm bridge).
const (
	highsModelStatusOptimal    int32 = 7
	highsModelStatusInfeasible int32 = 8
	highsModelStatusTimeLimit  int32 = 13
)

// solveLPModel serializes model to LP text, runs HiGHS, and returns the selected variables
// (column primal >= 0.5), in x-index order.
func solveLPModel(model *lpModel, timeout time.Duration, mipRelGap float64) (lpSolution, error) {
	lpString, reverseNames := modelToLPFormat(model)
	values, modelStatus, err := runHiGHSLP(lpString, len(reverseNames), timeout, mipRelGap)
	if err != nil {
		return lpSolution{}, err
	}

	status := "unknown"
	switch modelStatus {
	case highsModelStatusOptimal:
		status = "optimal"
	case highsModelStatusTimeLimit:
		status = "timedout"
	case highsModelStatusInfeasible:
		status = "infeasible"
	}

	var selected []string
	for i := 0; i < len(values) && i < len(reverseNames); i++ {
		if values[i] >= 0.5 {
			selected = append(selected, reverseNames[i])
		}
	}

	result := math.NaN()
	if len(values) > 0 {
		result = 0
		for _, name := range selected {
			if coeffs, ok := model.variables.get(name); ok {
				result += coeffs[scoreCoeffKey]
			}
		}
	}

	return lpSolution{
		status:    status,
		result:    result,
		variables: selected,
	}, nil
}

// minSolveSeconds is the per-pass time_limit floor for cap-refinement re-solves, so an exhausted
// budget never yields a non-positive HiGHS time_limit. Tightened re-solves finish well under this.
const minSolveSeconds = 1.0

// solveModel scores the variables, solves, then checks caps and recurses with tightened
// constraints/weights until no cap is exceeded. Returns the selected variable names of the final
// solution and its objective value.
func (o *reforgeOptimizer) solveModel(
	weights core.UnitStats,
	reforgeCaps core.UnitStats,
	reforgeSoftCaps []*reforgeSoftCap,
	variables *lpVariables,
	constraints *lpConstraints,
	maxSeconds float64,
) ([]string, float64, error) {
	if o.signals.Abort.IsTriggered() {
		return nil, 0, context.Canceled
	}

	updatedVariables := o.updateReforgeScores(variables, weights)
	model := &lpModel{
		direction:   "maximize",
		objective:   scoreCoeffKey,
		constraints: constraints,
		variables:   updatedVariables,
		binaries:    true,
	}

	startedAt := time.Now()
	solution, err := solveLPModel(model, time.Duration(maxSeconds*float64(time.Second)), 0)
	if err != nil {
		return nil, 0, err
	}

	if math.IsNaN(solution.result) || math.IsInf(solution.result, 1) {
		switch solution.status {
		case "infeasible":
			return nil, 0, errors.New("The specified stat caps are impossible to achieve. Consider changing any upper bound stat caps to lower bounds instead.")
		case "timedout":
			return nil, 0, errors.New("Solver timed out before finding a feasible solution.")
		default:
			return nil, 0, errors.New(solution.status)
		}
	}

	elapsedSeconds := time.Since(startedAt).Seconds()

	anyCapsExceeded, updatedConstraints, updatedWeights, updatedSoftCaps := o.checkCaps(solution, reforgeCaps, reforgeSoftCaps, updatedVariables, constraints, weights)
	if !anyCapsExceeded {
		return solution.variables, solution.result, nil
	}
	// Cap refinement consumed part of the budget; keep a positive floor so the tightened re-solve
	// still gets a valid (non-negative) HiGHS time_limit even once the budget is spent.
	remainingSeconds := math.Max(maxSeconds-elapsedSeconds, minSolveSeconds)
	return o.solveModel(updatedWeights, reforgeCaps, updatedSoftCaps, updatedVariables, updatedConstraints, remainingSeconds)
}

// checkCaps sums the selected variables' stat contributions, then adds a hard-cap constraint for
// every unconstrained stat that exceeds its cap. If no hard cap fired, it consumes soft caps from
// the front of the list until one of their breakpoints is exceeded. Returns whether any cap was
// newly enforced along with the tightened constraints/weights/soft-caps for the next pass.
func (o *reforgeOptimizer) checkCaps(
	solution lpSolution,
	reforgeCaps core.UnitStats,
	reforgeSoftCaps []*reforgeSoftCap,
	variables *lpVariables,
	constraints *lpConstraints,
	currentWeights core.UnitStats,
) (bool, *lpConstraints, core.UnitStats, []*reforgeSoftCap) {
	reforgeStatContribution := core.NewUnitStats()
	for _, variableKey := range solution.variables {
		coeffs, ok := variables.get(variableKey)
		if !ok {
			continue
		}
		for key, value := range coeffs {
			if unitStat, ok := unitStatFromCoeffKey(key); ok {
				reforgeStatContribution = setUnitStat(reforgeStatContribution, unitStat, getUnitStat(reforgeStatContribution, unitStat)+value)
			}
		}
	}

	anyCapsExceeded := false
	updatedConstraints := constraints.clone()
	updatedWeights := currentWeights

	eachUnitStat(reforgeStatContribution, func(unitStat stats.UnitStat, value float64) {
		cap := getUnitStat(reforgeCaps, unitStat)
		statName := coeffKeyForUnitStat(unitStat)
		if cap != 0 && value > cap && !constraints.has(statName) {
			anyCapsExceeded = true
			// Stats treated as upper bounds get a ceiling; the rest are pinned at the cap and lose
			// their EP so the solver stops chasing them.
			if getUnitStat(o.undershootCaps, unitStat) != 0 {
				updatedConstraints.set(statName, lessEq(cap))
			} else {
				updatedConstraints.set(statName, greaterEq(cap))
				updatedWeights = setUnitStat(updatedWeights, unitStat, 0)
			}
		}
	})

	// Soft caps are consumed from the front of the list, one per pass, and dropped once their
	// breakpoints are used up (threshold caps are dropped after a single pass).
	updatedSoftCaps := reforgeSoftCaps
	for !anyCapsExceeded && len(updatedSoftCaps) > 0 {
		softCap := updatedSoftCaps[0]
		unitStat := softCap.unitStat
		statName := coeffKeyForUnitStat(unitStat)
		currentValue := getUnitStat(reforgeStatContribution, unitStat)

		idx := 0
		for _, breakpoint := range softCap.breakpoints {
			if currentValue > breakpoint {
				postCapEP := 0.0
				if idx < len(softCap.postCapEPs) {
					postCapEP = softCap.postCapEPs[idx]
				}
				updatedConstraints.set(statName, greaterEq(breakpoint))
				updatedWeights = setUnitStat(updatedWeights, unitStat, postCapEP)
				anyCapsExceeded = true
				break
			}
			idx++
		}

		if softCap.capType == proto.StatCapType_TypeSoftCap {
			if idx+1 <= len(softCap.breakpoints) {
				softCap.breakpoints = softCap.breakpoints[idx+1:]
			} else {
				softCap.breakpoints = nil
			}
			if idx+1 <= len(softCap.postCapEPs) {
				softCap.postCapEPs = softCap.postCapEPs[idx+1:]
			} else {
				softCap.postCapEPs = nil
			}
		}

		if softCap.capType == proto.StatCapType_TypeThreshold || len(softCap.breakpoints) == 0 {
			updatedSoftCaps = updatedSoftCaps[1:]
		}
	}

	return anyCapsExceeded, updatedConstraints, updatedWeights, updatedSoftCaps
}
