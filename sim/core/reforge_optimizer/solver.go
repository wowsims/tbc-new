package reforgeoptimizer

import (
	"context"
	"fmt"
	"log"
	"math"
	"slices"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/simsignals"
	"github.com/wowsims/tbc/sim/core/stats"
)

type mipVariable struct {
	slotIdx   int
	choiceIdx int
	objective float64
	upper     float64
	integer   bool
}

type mipConstraint struct {
	lower   float64
	upper   float64
	indices []int
	values  []float64
}

type mipModel struct {
	variables   []mipVariable
	constraints []mipConstraint
}

type mipSolution struct {
	values []float64
}

type mipStatConstraint struct {
	unitStat       stats.UnitStat
	lower          float64
	upper          float64
	actualLower    float64
	actualUpper    float64
	hasActualLower bool
	hasActualUpper bool
}

func reforgeDebug(search *reforgeSearchState) bool {
	return search != nil && search.request != nil && search.request.GetDebug()
}

func trySolveWithHiGHS(search *reforgeSearchState, signals simsignals.Signals) ([]reforgeChoice, float64, bool, error) {
	weights := search.weights
	softCaps := cloneSoftCaps(search.softCaps)
	statConstraints := make([]mipStatConstraint, 0, len(search.hardCaps)+len(search.softCaps))
	constrainedStats := make(map[stats.UnitStat]bool, len(search.hardCaps)+1)
	maxPasses := max(1, 2*(len(search.hardCaps)+countSoftCapBreakpoints(search.softCaps)+1))
	deadline := time.Now().Add(highsOptimizerTimeout(search))
	debug := reforgeDebug(search)

	for passIdx := 0; passIdx < maxPasses; passIdx++ {
		if signals.Abort.IsTriggered() {
			return nil, 0, false, context.Canceled
		}
		remainingTimeout := highsOptimizerPassTimeout(deadline)
		if remainingTimeout <= 0 {
			return nil, 0, false, nil
		}
		var passStartedAt time.Time
		var modelStartedAt time.Time
		if debug {
			passStartedAt = time.Now()
			modelStartedAt = time.Now()
		}
		model := buildChoiceMIPModel(search, weights, statConstraints)
		var modelDuration time.Duration
		var solveStartedAt time.Time
		if debug {
			modelDuration = time.Since(modelStartedAt)
			solveStartedAt = time.Now()
		}
		solution, ok, err := solveMIPWithHiGHS(model, remainingTimeout, highsOptimizerMIPRelGap(search))
		if signals.Abort.IsTriggered() {
			return nil, 0, false, context.Canceled
		}
		var solveDuration time.Duration
		if debug {
			solveDuration = time.Since(solveStartedAt)
		}
		if err != nil || !ok {
			if debug {
				log.Printf("[reforgeOptimize] HiGHS pass=%d failure vars=%d constraints=%d err=%v", passIdx+1, len(model.variables), len(model.constraints), err)
			}
			return nil, 0, ok, err
		}

		var selectStartedAt time.Time
		if debug {
			selectStartedAt = time.Now()
		}
		choices, err := choicesFromMIPSolution(search, model, solution)
		if err != nil {
			return nil, 0, false, err
		}
		if !selectedChoicesValid(search, choices) {
			return nil, 0, false, nil
		}
		objectiveDelta := selectedChoicesModelDelta(choices)
		// Constraint evaluation must use the actual resolved stat delta (coefficient
		// delta), not the weights-filtered objective delta. After a cap fires and
		// zeroes a stat's weight, objectiveDelta excludes that stat's contribution,
		// which would cause the cap tightening loop to spin forever.
		coefficientDelta := selectedChoicesCoefficientDelta(choices)
		var selectDuration time.Duration
		if debug {
			selectDuration = time.Since(selectStartedAt)
		}

		var capStartedAt time.Time
		if debug {
			capStartedAt = time.Now()
		}
		updated, nextWeights, nextSoftCaps, nextStatConstraints := updateHiGHSCapPass(search, passIdx, coefficientDelta, weights, softCaps, statConstraints, constrainedStats)
		var capDuration time.Duration
		if debug {
			capDuration = time.Since(capStartedAt)
			log.Printf("[reforgeOptimize] solver pass=%d vars=%d constraints=%d timings=model:%s solve:%s select:%s cap:%s total:%s", passIdx+1, len(model.variables), len(model.constraints), modelDuration, solveDuration, selectDuration, capDuration, time.Since(passStartedAt))
		}
		if !updated {
			choices, objectiveDelta = improveSelectedChoices(search, choices, objectiveDelta)
			score, ok := search.evaluate(objectiveDelta)
			return choices, score, ok, nil
		}
		prevConstraintCount := len(statConstraints)
		weights = nextWeights
		softCaps = nextSoftCaps
		statConstraints = nextStatConstraints
		// Rebuild slot choices whenever a new cap constraint is added.
		// Adding a constraint means a hard cap or soft cap breakpoint just fired,
		// which changes EP weights and which stats are capped — both affect
		// shouldForceSocketBonus and cappedGemStats. Constraint-tightening passes
		// update existing entries (no count change) and do not require a rebuild.
		if len(statConstraints) > prevConstraintCount {
			if newSlots, err := buildReforgeSlotChoices(search.request, search.baseRaid, search.baseGear, search.capBaseStats, weights, weights, search.hardCaps, softCaps, search.statDeps, statConstraints); err == nil {
				search.slots = newSlots
				search.uniqueGemIDs = buildUniqueGemLimitIDs(newSlots)
				search.choiceVarIdx = make([][]int, len(newSlots))
				for i, slot := range newSlots {
					search.choiceVarIdx[i] = make([]int, len(slot.choices))
				}
			}
		}
	}

	return nil, 0, false, fmt.Errorf("HiGHS optimizer reached cap refinement pass limit")
}

func improveSelectedChoices(search *reforgeSearchState, choices []reforgeChoice, delta core.UnitStats) ([]reforgeChoice, core.UnitStats) {
	bestChoices := slices.Clone(choices)
	bestDelta := delta
	bestScore, ok := search.evaluate(bestDelta)
	if !ok {
		return bestChoices, bestDelta
	}
	bestSocketBonusCount := selectedChoicesSocketBonusCount(bestChoices)
	if !selectedChoicesHardCapsValid(search, bestChoices) {
		return bestChoices, bestDelta
	}

	for {
		improved := false
		for slotIdx, slot := range search.slots {
			currentChoice := bestChoices[slotIdx]
			currentChoiceDelta := choiceObjectiveDelta(currentChoice)
			for _, alternative := range slot.choices {
				if sameReforgeChoice(currentChoice, alternative) {
					continue
				}

				candidateChoices := slices.Clone(bestChoices)
				candidateChoices[slotIdx] = alternative
				if !selectedChoicesValid(search, candidateChoices) {
					continue
				}
				if !selectedChoicesHardCapsValid(search, candidateChoices) {
					continue
				}

				candidateDelta := subtractUnitStats(bestDelta, currentChoiceDelta)
				candidateDelta = addUnitStats(candidateDelta, choiceObjectiveDelta(alternative))
				candidateScore, ok := search.evaluate(candidateDelta)
				if !ok {
					continue
				}
				candidateSocketBonusCount := selectedChoicesSocketBonusCount(candidateChoices)
				scoreImproved := candidateScore > bestScore+1e-9
				equalScore := math.Abs(candidateScore-bestScore) <= 1e-9
				bonusImproved := equalScore && candidateSocketBonusCount > bestSocketBonusCount
				if !scoreImproved && !bonusImproved {
					continue
				}

				bestChoices = candidateChoices
				bestDelta = candidateDelta
				bestScore = candidateScore
				bestSocketBonusCount = candidateSocketBonusCount
				improved = true
				break
			}
			if improved {
				break
			}
		}
		if !improved {
			return bestChoices, bestDelta
		}
	}
}

func sameReforgeChoice(left reforgeChoice, right reforgeChoice) bool {
	if left.slot != right.slot || left.socketChoice != right.socketChoice || left.socketIdx != right.socketIdx || left.socketBonus != right.socketBonus {
		return false
	}
	if !slices.Equal(left.bonusSocketIdxs, right.bonusSocketIdxs) || !slices.Equal(left.uniqueGemIDs, right.uniqueGemIDs) {
		return false
	}
	if len(left.gems) != len(right.gems) {
		return false
	}
	for idx := range left.gems {
		if left.gems[idx].socketIdx != right.gems[idx].socketIdx || left.gems[idx].gemID != right.gems[idx].gemID {
			return false
		}
	}
	return true
}

func selectedChoicesSocketBonusCount(choices []reforgeChoice) int {
	count := 0
	for _, choice := range choices {
		if choice.socketBonus && len(choice.bonusSocketIdxs) > 0 {
			count++
		}
	}
	return count
}

func highsOptimizerTimeout(_ *reforgeSearchState) time.Duration { return optimizerTimeout }

func highsOptimizerPassTimeout(deadline time.Time) time.Duration {
	remaining := time.Until(deadline)
	if remaining < time.Second {
		return time.Second
	}
	return remaining
}

func highsOptimizerMIPRelGap(_ *reforgeSearchState) float64 { return 0 }

func choicesFromMIPSolution(search *reforgeSearchState, model mipModel, solution mipSolution) ([]reforgeChoice, error) {
	choices := make([]reforgeChoice, len(search.slots))
	selected := make([]bool, len(search.slots))
	for slotIdx, slot := range search.slots {
		if len(slot.choices) > 0 {
			choices[slotIdx] = slot.choices[0]
			selected[slotIdx] = true
		}
	}
	for varIdx, value := range solution.values {
		if value < 0.5 {
			continue
		}
		variable := model.variables[varIdx]
		if !variable.integer {
			continue
		}
		choices[variable.slotIdx] = search.slots[variable.slotIdx].choices[variable.choiceIdx]
		selected[variable.slotIdx] = true
	}

	for slotIdx := range choices {
		if !selected[slotIdx] {
			return nil, fmt.Errorf("HiGHS did not select a choice for slot %s", search.slots[slotIdx].slot.String())
		}
	}
	return choices, nil
}

func buildChoiceMIPModel(search *reforgeSearchState, weights core.UnitStats, statConstraints []mipStatConstraint) mipModel {
	variableCount := countMIPChoiceVariables(search.slots)
	uniqueGemIDs := search.uniqueGemIDs
	if uniqueGemIDs == nil {
		uniqueGemIDs = buildUniqueGemLimitIDs(search.slots)
	}
	metaGemConstraintCount := countMetaGemConstraints(search)
	model := mipModel{
		variables:   make([]mipVariable, 0, variableCount),
		constraints: make([]mipConstraint, 0, estimateMIPConstraintCount(search, statConstraints, len(uniqueGemIDs), metaGemConstraintCount)),
	}
	choiceVarIdx := search.choiceVarIdx
	if len(choiceVarIdx) != len(search.slots) {
		choiceVarIdx = make([][]int, len(search.slots))
		for i, slot := range search.slots {
			choiceVarIdx[i] = make([]int, len(slot.choices))
		}
	}
	for slotIdx, slot := range search.slots {
		for choiceIdx := range slot.choices {
			choiceVarIdx[slotIdx][choiceIdx] = -1
		}
		for choiceIdx, choice := range slot.choices {
			if !choiceMIPActive(choice) {
				continue
			}
			choiceVarIdx[slotIdx][choiceIdx] = len(model.variables)
			model.variables = append(model.variables, mipVariable{
				slotIdx:   slotIdx,
				choiceIdx: choiceIdx,
				objective: dotUnitStats(choiceObjectiveDelta(choice), weights),
				upper:     1,
				integer:   true,
			})
		}
	}

	for slotIdx := range search.slots {
		if reforgeSlotChoicesAreSocketBonus(search.slots[slotIdx]) {
			continue
		}
		constraint := newMIPConstraint(math.Inf(-1), 1, len(search.slots[slotIdx].choices))
		for choiceIdx := range search.slots[slotIdx].choices {
			if choiceVarIdx[slotIdx][choiceIdx] >= 0 {
				constraint.addCoefficient(choiceVarIdx[slotIdx][choiceIdx], 1)
			}
		}
		if constraint.coefficientCount() > 0 {
			model.constraints = append(model.constraints, constraint)
		}
	}
	addSocketBonusLinkConstraints(search, choiceVarIdx, &model)

	if constraint := buildChoiceLimitConstraint(search, choiceVarIdx, func(choice reforgeChoice) float64 { return float64(choice.jewelcraftingGems) }, 2); constraint.coefficientCount() > 0 {
		model.constraints = append(model.constraints, constraint)
	}
	for _, gemID := range uniqueGemIDs {
		constraint := buildChoiceLimitConstraint(search, choiceVarIdx, func(choice reforgeChoice) float64 {
			for _, choiceGemID := range choice.uniqueGemIDs {
				if choiceGemID == gemID {
					return 1
				}
			}
			return 0
		}, 1)
		if constraint.coefficientCount() > 0 {
			model.constraints = append(model.constraints, constraint)
		}
	}

	addMetaGemActivationConstraints(search, choiceVarIdx, &model)
	for _, statConstraint := range statConstraints {
		constraint := newMIPConstraint(statConstraint.lower, statConstraint.upper, variableCount)
		for slotIdx, slot := range search.slots {
			for choiceIdx, choice := range slot.choices {
				if choiceVarIdx[slotIdx][choiceIdx] < 0 {
					continue
				}
				if delta := getUnitStat(choiceCoefficientDelta(choice), statConstraint.unitStat); delta != 0 {
					constraint.addCoefficient(choiceVarIdx[slotIdx][choiceIdx], delta)
				}
			}
		}
		if constraint.coefficientCount() > 0 {
			model.constraints = append(model.constraints, constraint)
		}
	}

	return model
}

func countMIPChoiceVariables(slots []reforgeSlotChoices) int {
	count := 0
	for _, slot := range slots {
		for _, choice := range slot.choices {
			if choiceMIPActive(choice) {
				count++
			}
		}
	}
	return count
}

func choiceMIPActive(choice reforgeChoice) bool {
	if choice.socketChoice {
		return len(choice.gems) > 0 && choice.gems[0].gemID != 0
	}
	if choice.socketBonus {
		return len(choice.bonusSocketIdxs) > 0
	}
	return false
}

func reforgeSlotChoicesAreSocketBonus(slot reforgeSlotChoices) bool {
	return len(slot.choices) > 0 && slot.choices[0].socketBonus
}

func estimateMIPConstraintCount(search *reforgeSearchState, statConstraints []mipStatConstraint, uniqueGemLimitCount int, metaGemConstraintCount int) int {
	count := len(search.slots) + countSocketBonusLinkConstraints(search) + len(statConstraints) + uniqueGemLimitCount + metaGemConstraintCount
	count += 2
	return count
}

func countMetaGemConstraints(search *reforgeSearchState) int {
	constraint, ok := equippedMetaGemConstraint(search.baseEquipment)
	if !ok {
		return 0
	}

	count := 0
	if constraint.minBlue > 0 {
		count++
	}
	if constraint.minRed > 0 {
		count++
	}
	if constraint.minYellow > 0 {
		count++
	}
	if constraint.compareColorGreater != proto.GemColor_GemColorUnknown && constraint.compareColorLesser != proto.GemColor_GemColorUnknown {
		count++
	}
	return count
}

func countSocketBonusLinkConstraints(search *reforgeSearchState) int {
	count := 0
	for _, group := range search.slots {
		for _, choice := range group.choices {
			if choice.socketBonus {
				count += len(choice.bonusSocketIdxs)
			}
		}
	}
	return count
}

func buildUniqueGemLimitIDs(slots []reforgeSlotChoices) []int32 {
	uniqueGemIDs := make([]int32, 0)
	seen := map[int32]bool{}
	for _, slot := range slots {
		for _, choice := range slot.choices {
			for _, gemID := range choice.uniqueGemIDs {
				if seen[gemID] {
					continue
				}
				seen[gemID] = true
				uniqueGemIDs = append(uniqueGemIDs, gemID)
			}
		}
	}
	return uniqueGemIDs
}

func newMIPConstraint(lower float64, upper float64, capacity int) mipConstraint {
	return mipConstraint{
		lower:   lower,
		upper:   upper,
		indices: make([]int, 0, capacity),
		values:  make([]float64, 0, capacity),
	}
}

func (constraint *mipConstraint) addCoefficient(index int, value float64) {
	constraint.indices = append(constraint.indices, index)
	constraint.values = append(constraint.values, value)
}

func (constraint mipConstraint) coefficientCount() int {
	return len(constraint.indices)
}

func choiceObjectiveDelta(choice reforgeChoice) core.UnitStats {
	if !isEmptyUnitStats(choice.objectiveDelta) {
		return choice.objectiveDelta
	}
	return choice.delta
}

func choiceCoefficientDelta(choice reforgeChoice) core.UnitStats {
	return choice.delta
}

func addSocketBonusLinkConstraints(search *reforgeSearchState, choiceVarIdx [][]int, model *mipModel) {
	for groupIdx, group := range search.slots {
		for choiceIdx, choice := range group.choices {
			if !choice.socketBonus || len(choice.bonusSocketIdxs) == 0 {
				continue
			}
			bonusVarIdx := choiceVarIdx[groupIdx][choiceIdx]
			if bonusVarIdx < 0 {
				continue
			}
			for _, socketIdx := range choice.bonusSocketIdxs {
				constraint := newMIPConstraint(math.Inf(-1), 0, 1)
				constraint.addCoefficient(bonusVarIdx, 1)
				for socketGroupIdx, socketGroup := range search.slots {
					for socketChoiceIdx, socketChoice := range socketGroup.choices {
						if socketChoice.slot == choice.slot && socketChoice.socketChoice && socketChoice.socketIdx == socketIdx && socketChoice.socketMatches && choiceVarIdx[socketGroupIdx][socketChoiceIdx] >= 0 {
							constraint.addCoefficient(choiceVarIdx[socketGroupIdx][socketChoiceIdx], -1)
						}
					}
				}
				model.constraints = append(model.constraints, constraint)
			}
		}
	}
}

func addMetaGemActivationConstraints(search *reforgeSearchState, choiceVarIdx [][]int, model *mipModel) {
	constraint, ok := equippedMetaGemConstraint(search.baseEquipment)
	if !ok {
		return
	}

	fixedCounts := metaGemColorCounts(search.baseEquipment)
	addMetaGemColorConstraint := func(color proto.GemColor, required int) {
		remaining := required - metaGemCountForColor(fixedCounts, color)
		if remaining <= 0 {
			return
		}
		row := newMIPConstraint(float64(remaining), math.Inf(1), 0)
		for slotIdx, slot := range search.slots {
			for choiceIdx, choice := range slot.choices {
				varIdx := choiceVarIdx[slotIdx][choiceIdx]
				if varIdx < 0 {
					continue
				}
				coefficient := 0.0
				for _, gemChoice := range choice.gems {
					selectedGem, ok := core.GetGemByID(gemChoice.gemID)
					if !ok || selectedGem.ID == 0 {
						continue
					}
					baseGem := gemIDAt(search.baseEquipment.GetItemBySlot(choice.slot), gemChoice.socketIdx)
					baseGemColor := proto.GemColor_GemColorUnknown
					if baseGem != 0 {
						if gem, ok := core.GetGemByID(baseGem); ok {
							baseGemColor = gem.Color
						}
					}
					selectedRed, selectedYellow, selectedBlue := metaGemActivationColorContribution(selectedGem.Color)
					baseRed, baseYellow, baseBlue := metaGemActivationColorContribution(baseGemColor)
					switch color {
					case proto.GemColor_GemColorRed:
						coefficient += float64(selectedRed - baseRed)
					case proto.GemColor_GemColorYellow:
						coefficient += float64(selectedYellow - baseYellow)
					case proto.GemColor_GemColorBlue:
						coefficient += float64(selectedBlue - baseBlue)
					}
				}
				if coefficient != 0 {
					row.addCoefficient(varIdx, coefficient)
				}
			}
		}
		model.constraints = append(model.constraints, row)
	}

	addMetaGemColorConstraint(proto.GemColor_GemColorBlue, constraint.minBlue)
	addMetaGemColorConstraint(proto.GemColor_GemColorRed, constraint.minRed)
	addMetaGemColorConstraint(proto.GemColor_GemColorYellow, constraint.minYellow)

	if constraint.compareColorGreater != proto.GemColor_GemColorUnknown && constraint.compareColorLesser != proto.GemColor_GemColorUnknown {
		fixedGreater := metaGemCountForColor(fixedCounts, constraint.compareColorGreater)
		fixedLesser := metaGemCountForColor(fixedCounts, constraint.compareColorLesser)
		row := newMIPConstraint(float64(1-(fixedGreater-fixedLesser)), math.Inf(1), 0)
		for slotIdx, slot := range search.slots {
			for choiceIdx, choice := range slot.choices {
				varIdx := choiceVarIdx[slotIdx][choiceIdx]
				if varIdx < 0 {
					continue
				}
				coefficient := 0.0
				for _, gemChoice := range choice.gems {
					selectedGem, ok := core.GetGemByID(gemChoice.gemID)
					if !ok || selectedGem.ID == 0 {
						continue
					}
					baseGem := gemIDAt(search.baseEquipment.GetItemBySlot(choice.slot), gemChoice.socketIdx)
					baseGemColor := proto.GemColor_GemColorUnknown
					if baseGem != 0 {
						if gem, ok := core.GetGemByID(baseGem); ok {
							baseGemColor = gem.Color
						}
					}
					selectedRed, selectedYellow, selectedBlue := metaGemActivationColorContribution(selectedGem.Color)
					baseRed, baseYellow, baseBlue := metaGemActivationColorContribution(baseGemColor)
					switch constraint.compareColorGreater {
					case proto.GemColor_GemColorRed:
						coefficient += float64(selectedRed - baseRed)
					case proto.GemColor_GemColorYellow:
						coefficient += float64(selectedYellow - baseYellow)
					case proto.GemColor_GemColorBlue:
						coefficient += float64(selectedBlue - baseBlue)
					}
					switch constraint.compareColorLesser {
					case proto.GemColor_GemColorRed:
						coefficient -= float64(selectedRed - baseRed)
					case proto.GemColor_GemColorYellow:
						coefficient -= float64(selectedYellow - baseYellow)
					case proto.GemColor_GemColorBlue:
						coefficient -= float64(selectedBlue - baseBlue)
					}
				}
				if coefficient != 0 {
					row.addCoefficient(varIdx, coefficient)
				}
			}
		}
		model.constraints = append(model.constraints, row)
	}
}

func updateHiGHSCapPass(search *reforgeSearchState, passIdx int, delta core.UnitStats, weights core.UnitStats, softCaps []reforgeSoftCap, statConstraints []mipStatConstraint, constrainedStats map[stats.UnitStat]bool) (bool, core.UnitStats, []reforgeSoftCap, []mipStatConstraint) {
	for idx, constraint := range statConstraints {
		if constraint.hasActualLower && constraint.hasActualUpper && math.Abs(constraint.actualUpper-constraint.actualLower) < 1e-9 {
			continue
		}
		value := getUnitStat(delta, constraint.unitStat)
		if constraint.hasActualLower && value < constraint.actualLower-1e-6 {
			missing := constraint.actualLower - value
			// If the actual value also violates the tightened LP lower bound, the LP
			// approximation has a systematic error. Tighten by the larger LP violation
			// to avoid many small increments before the LP produces a different solution.
			if lpMissing := constraint.lower - value; lpMissing > missing+1e-6 {
				statConstraints[idx].lower += lpMissing + 1e-6
			} else {
				statConstraints[idx].lower += missing + 1e-6
			}
			if reforgeDebug(search) {
				log.Printf("[reforgeOptimize] HiGHS pass=%d tightening min cap stat=%s valueDelta=%.3f requiredDelta=%.3f adjustedDelta=%.3f", passIdx+1, unitStatName(constraint.unitStat), value, constraint.actualLower, statConstraints[idx].lower)
			}
			return true, weights, softCaps, statConstraints
		}
		if constraint.hasActualUpper && value > constraint.actualUpper+1e-6 {
			excess := value - constraint.actualUpper
			statConstraints[idx].upper -= excess + 1e-6
			if reforgeDebug(search) {
				log.Printf("[reforgeOptimize] HiGHS pass=%d tightening max cap stat=%s valueDelta=%.3f requiredDelta=%.3f adjustedDelta=%.3f", passIdx+1, unitStatName(constraint.unitStat), value, constraint.actualUpper, statConstraints[idx].upper)
			}
			return true, weights, softCaps, statConstraints
		}
	}

	for _, hardCap := range search.hardCaps {
		value := getUnitStat(delta, hardCap.unitStat)
		if hardCap.cap == 0 || value <= hardCap.cap+1e-9 || constrainedStats[hardCap.unitStat] {
			continue
		}
		if hardCap.undershoot {
			statConstraints = append(statConstraints, mipStatConstraint{unitStat: hardCap.unitStat, lower: math.Inf(-1), upper: hardCap.cap, actualUpper: hardCap.cap, hasActualUpper: true})
			if reforgeDebug(search) {
				log.Printf("[reforgeOptimize] HiGHS pass=%d adding max cap stat=%s valueDelta=%.3f capDelta=%.3f", passIdx+1, unitStatName(hardCap.unitStat), value, hardCap.cap)
			}
		} else {
			statConstraints = append(statConstraints, mipStatConstraint{unitStat: hardCap.unitStat, lower: hardCap.cap, upper: math.Inf(1), actualLower: hardCap.cap, hasActualLower: true})
			weights = setUnitStat(weights, hardCap.unitStat, 0)
			if reforgeDebug(search) {
				log.Printf("[reforgeOptimize] HiGHS pass=%d adding min cap stat=%s valueDelta=%.3f capDelta=%.3f newWeight=0", passIdx+1, unitStatName(hardCap.unitStat), value, hardCap.cap)
			}
		}
		constrainedStats[hardCap.unitStat] = true
		return true, weights, softCaps, statConstraints
	}

	remainingSoftCaps := make([]reforgeSoftCap, 0, len(softCaps))
	for softCapIdx, softCap := range softCaps {
		value := getUnitStat(delta, softCap.unitStat)
		exceededBreakpointIdx := -1
		for idx, breakpoint := range softCap.breakpoints {
			if value > breakpoint+1e-9 {
				exceededBreakpointIdx = idx
				break
			}
		}
		if exceededBreakpointIdx == -1 {
			remainingSoftCaps = append(remainingSoftCaps, softCap)
			continue
		}

		statConstraints = append(statConstraints, mipStatConstraint{unitStat: softCap.unitStat, lower: softCap.breakpoints[exceededBreakpointIdx], upper: math.Inf(1), actualLower: softCap.breakpoints[exceededBreakpointIdx], hasActualLower: true})
		if exceededBreakpointIdx < len(softCap.postCapEPs) {
			weights = setUnitStat(weights, softCap.unitStat, softCap.postCapEPs[exceededBreakpointIdx])
		}
		if reforgeDebug(search) {
			log.Printf("[reforgeOptimize] HiGHS pass=%d adding breakpoint stat=%s valueDelta=%.3f breakpointDelta=%.3f newWeight=%.3f", passIdx+1, unitStatName(softCap.unitStat), value, softCap.breakpoints[exceededBreakpointIdx], getUnitStat(weights, softCap.unitStat))
		}
		if softCap.capType == proto.StatCapType_TypeSoftCap {
			softCap.breakpoints = softCap.breakpoints[exceededBreakpointIdx+1:]
			softCap.postCapEPs = softCap.postCapEPs[min(exceededBreakpointIdx+1, len(softCap.postCapEPs)):]
			if len(softCap.breakpoints) > 0 {
				remainingSoftCaps = append(remainingSoftCaps, softCap)
			}
			remainingSoftCaps = append(remainingSoftCaps, softCaps[softCapIdx+1:]...)
		} else {
			remainingSoftCaps = append(remainingSoftCaps, softCaps[softCapIdx+1:]...)
		}
		return true, weights, remainingSoftCaps, statConstraints
	}

	return false, weights, softCaps, statConstraints
}

func selectedChoicesValid(search *reforgeSearchState, choices []reforgeChoice) bool {
	jewelcraftingGems := 0
	uniqueGemIDs := map[int32]bool{}
	for _, choice := range choices {
		if !canAddChoice(choice, jewelcraftingGems, uniqueGemIDs) {
			return false
		}
		jewelcraftingGems += choice.jewelcraftingGems
		for _, gemID := range choice.uniqueGemIDs {
			uniqueGemIDs[gemID] = true
		}
	}
	if !selectedChoicesSocketBonusLinksValid(choices) {
		return false
	}
	return selectedChoicesMetaGemValid(search, choices)
}

func selectedChoicesHardCapsValid(search *reforgeSearchState, choices []reforgeChoice) bool {
	delta := selectedChoicesModelDelta(choices)
	for _, hardCap := range search.hardCaps {
		if hardCap.cap == 0 {
			continue
		}
		value := getUnitStat(delta, hardCap.unitStat)
		if hardCap.undershoot {
			if value > hardCap.cap+1e-9 {
				return false
			}
			continue
		}
		if value < hardCap.cap-1e-9 {
			return false
		}
	}
	return true
}

func selectedChoicesSocketBonusLinksValid(choices []reforgeChoice) bool {
	matchedSockets := make(map[reforgeSocketKey]bool)
	for _, choice := range choices {
		if !choice.socketChoice || !choice.socketMatches {
			continue
		}
		matchedSockets[reforgeSocketKey{slot: choice.slot, socketIdx: choice.socketIdx}] = true
	}

	for _, choice := range choices {
		if !choice.socketBonus || len(choice.bonusSocketIdxs) == 0 {
			continue
		}
		for _, socketIdx := range choice.bonusSocketIdxs {
			if !matchedSockets[reforgeSocketKey{slot: choice.slot, socketIdx: socketIdx}] {
				return false
			}
		}
	}

	return true
}

func selectedChoicesMetaGemValid(search *reforgeSearchState, choices []reforgeChoice) bool {
	constraint, ok := equippedMetaGemConstraint(search.baseEquipment)
	if !ok {
		return true
	}

	counts := metaGemColorCounts(search.baseEquipment)
	for _, choice := range choices {
		for _, gemChoice := range choice.gems {
			gem, ok := core.GetGemByID(gemChoice.gemID)
			if !ok || gem.ID == 0 || gem.Color == proto.GemColor_GemColorMeta {
				continue
			}
			red, yellow, blue := metaGemActivationColorContribution(gem.Color)
			if red != 0 {
				counts[proto.GemColor_GemColorRed] += red
			}
			if yellow != 0 {
				counts[proto.GemColor_GemColorYellow] += yellow
			}
			if blue != 0 {
				counts[proto.GemColor_GemColorBlue] += blue
			}
		}
	}

	if counts[proto.GemColor_GemColorRed] < constraint.minRed || counts[proto.GemColor_GemColorYellow] < constraint.minYellow || counts[proto.GemColor_GemColorBlue] < constraint.minBlue {
		return false
	}
	if constraint.compareColorGreater != proto.GemColor_GemColorUnknown && constraint.compareColorLesser != proto.GemColor_GemColorUnknown {
		if counts[constraint.compareColorGreater] <= counts[constraint.compareColorLesser] {
			return false
		}
	}
	return true
}

func selectedChoicesModelDelta(choices []reforgeChoice) core.UnitStats {
	total := core.NewUnitStats()
	for _, choice := range choices {
		total = addUnitStats(total, choiceObjectiveDelta(choice))
	}
	return total
}

func selectedChoicesCoefficientDelta(choices []reforgeChoice) core.UnitStats {
	total := core.NewUnitStats()
	for _, choice := range choices {
		total = addUnitStats(total, choiceCoefficientDelta(choice))
	}
	return total
}

func cloneSoftCaps(softCaps []reforgeSoftCap) []reforgeSoftCap {
	cloned := make([]reforgeSoftCap, len(softCaps))
	for idx, softCap := range softCaps {
		cloned[idx] = reforgeSoftCap{
			unitStat:    softCap.unitStat,
			breakpoints: slices.Clone(softCap.breakpoints),
			postCapEPs:  slices.Clone(softCap.postCapEPs),
			capType:     softCap.capType,
		}
	}
	return cloned
}

func countSoftCapBreakpoints(softCaps []reforgeSoftCap) int {
	count := 0
	for _, softCap := range softCaps {
		count += len(softCap.breakpoints)
	}
	return count
}

func buildChoiceLimitConstraint(search *reforgeSearchState, choiceVarIdx [][]int, coefficient func(reforgeChoice) float64, upper float64) mipConstraint {
	constraint := newMIPConstraint(math.Inf(-1), upper, 0)
	for slotIdx, slot := range search.slots {
		for choiceIdx, choice := range slot.choices {
			if choiceVarIdx[slotIdx][choiceIdx] < 0 {
				continue
			}
			if value := coefficient(choice); value != 0 {
				constraint.addCoefficient(choiceVarIdx[slotIdx][choiceIdx], value)
			}
		}
	}
	return constraint
}
