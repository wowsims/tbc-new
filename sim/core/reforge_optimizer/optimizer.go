package reforgeoptimizer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/simsignals"
	"github.com/wowsims/tbc/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

var reforgeOptimizeRequestID atomic.Uint64

func Optimize(request *proto.ReforgeOptimizeRequest) *proto.ReforgeOptimizeResult {
	return OptimizeAsync(request, simsignals.CreateSignals())
}

func OptimizeAsync(request *proto.ReforgeOptimizeRequest, signals simsignals.Signals) *proto.ReforgeOptimizeResult {
	requestID := reforgeOptimizeRequestID.Add(1)
	startedAt := time.Now()
	normalizedConfig, err := validateReforgeOptimizeSettings(request)
	if err != nil {
		log.Printf("[reforgeOptimize:%d] failed validating settings after %s: %s", requestID, time.Since(startedAt), err.Error())
		return optimizeError(err.Error())
	}
	debug := request.GetDebug()
	logAbort := request.GetMode() != proto.ReforgeOptimizeMode_ReforgeOptimizeModeBulk || debug
	if debug {
		log.Printf("[reforgeOptimize:%d] started debug=%t", requestID, debug)
		logRequestInput(requestID, request, normalizedConfig)
	}

	if request.Raid == nil || len(request.Raid.Parties) == 0 || len(request.Raid.Parties[0].Players) == 0 {
		log.Printf("[reforgeOptimize:%d] failed after %s: missing player", requestID, time.Since(startedAt))
		return optimizeError("Reforge optimizer requires a raid with player 0.")
	}
	if request.Raid.Parties[0].Players[0].Equipment == nil {
		log.Printf("[reforgeOptimize:%d] failed after %s: missing baseline gear", requestID, time.Since(startedAt))
		return optimizeError("Reforge optimizer requires baseline gear.")
	}
	if signals.Abort.IsTriggered() {
		return optimizeAborted()
	}

	optimization, err := newReforgeOptimization(request, normalizedConfig, signals)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			if logAbort {
				log.Printf("[reforgeOptimize:%d] aborted initializing after %s", requestID, time.Since(startedAt))
			}
			return optimizeAborted()
		}
		log.Printf("[reforgeOptimize:%d] failed initializing after %s: %s", requestID, time.Since(startedAt), err.Error())
		return optimizeError(err.Error())
	}
	if debug {
		log.Printf("[reforgeOptimize:%d] computed baseline stats in %s", requestID, time.Since(startedAt))
		log.Printf("[reforgeOptimize:%d] built %d choice groups / %d choices in %s", requestID, len(optimization.slotChoices), countReforgeChoices(optimization.slotChoices), time.Since(startedAt))
	}

	search := optimization.searchState()
	solveStartedAt := time.Now()
	choices, score, solved, err := trySolveWithHiGHS(search, signals)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			if logAbort {
				log.Printf("[reforgeOptimize:%d] aborted solving after %s", requestID, time.Since(startedAt))
			}
			return optimizeAborted()
		}
		gear := request.GetRaid().GetParties()[0].GetPlayers()[0].GetEquipment()
		gearJSON, _ := protojson.Marshal(gear)
		log.Printf("[reforgeOptimize:%d] HiGHS failed after %s: %s gear=%s", requestID, time.Since(solveStartedAt), err.Error(), gearJSON)
		return optimizeError(fmt.Sprintf("HiGHS reforge optimizer failed: %s", err.Error()))
	}
	if !solved {
		gear := request.GetRaid().GetParties()[0].GetPlayers()[0].GetEquipment()
		gearJSON, _ := protojson.Marshal(gear)
		log.Printf("[reforgeOptimize:%d] HiGHS did not return a solution after %s gear=%s", requestID, time.Since(solveStartedAt), gearJSON)
		return optimizeError("HiGHS reforge optimizer did not return a solution.")
	}
	if debug {
		log.Printf("[reforgeOptimize:%d] HiGHS solved in %s score=%.3f", requestID, time.Since(solveStartedAt), score)
	}
	if signals.Abort.IsTriggered() {
		if logAbort {
			log.Printf("[reforgeOptimize:%d] aborted after solving in %s", requestID, time.Since(startedAt))
		}
		return optimizeAborted()
	}

	optimizedGear := optimization.optimizedGear(choices)

	isBulk := request.GetMode() == proto.ReforgeOptimizeMode_ReforgeOptimizeModeBulk
	var optimizedPlayerStats *proto.PlayerStats
	if !isBulk || debug {
		optimizedRaid := googleProto.Clone(request.Raid).(*proto.Raid)
		optimizedRaid.Parties[0].Players[0].Equipment = optimizedGear
		optimizedResult := computeReforgeStats(&proto.ComputeStatsRequest{Raid: optimizedRaid})
		if optimizedResult.ErrorResult != "" {
			log.Printf("[reforgeOptimize:%d] failed computing optimized stats after %s: %s", requestID, time.Since(startedAt), optimizedResult.ErrorResult)
			return optimizeError(optimizedResult.ErrorResult)
		}
		optimizedPlayerStats = optimizedResult.RaidStats.Parties[0].Players[0]
		optimizedStats := protoToCoreUnitStats(optimizedPlayerStats.FinalStats)
		optimizedCapStats := optimizedStats
		optimizedDelta := subtractUnitStats(optimizedCapStats, optimization.capBaseStats)
		if debug {
			logOptimizedGearSummary(requestID, optimizedGear)
			logCapEvaluation(requestID, search.hardCaps, search.softCaps, optimizedDelta)
		}
	}
	if !isBulk {
		log.Printf("[Reforge Optimizer] Reforge optimization completed requestID=%d total=%s score=%.3f", requestID, time.Since(startedAt), score)
	}
	if debug {
		log.Printf("[reforgeOptimize:%d] selectedChoices=%d", requestID, len(choices))
		logSelectedChoices(requestID, choices, optimization.weights)
	}

	return &proto.ReforgeOptimizeResult{
		OptimizedGear:        optimizedGear,
		OptimizedPlayerStats: optimizedPlayerStats,
		Score:                score,
		PassesDone:           1,
	}
}

func newReforgeOptimization(request *proto.ReforgeOptimizeRequest, normalizedConfig *normalizedReforgeOptimizeConfig, signals simsignals.Signals) (*reforgeOptimization, error) {
	request = googleProto.Clone(request).(*proto.ReforgeOptimizeRequest)
	request.Settings = normalizedConfig.settings
	settings := normalizedConfig.settings
	baseRaid := googleProto.Clone(request.Raid).(*proto.Raid)
	originalGear := cloneEquipmentSpec(baseRaid.Parties[0].Players[0].Equipment)
	baseGear := cloneEquipmentSpec(originalGear)
	clearGems(baseGear, settings)
	player := baseRaid.Parties[0].Players[0]
	player.Equipment = baseGear

	baseResult, statDeps := computeReforgeStatsAndDeps(&proto.ComputeStatsRequest{Raid: baseRaid})
	if baseResult.ErrorResult != "" {
		return nil, errors.New(baseResult.ErrorResult)
	}
	if signals.Abort.IsTriggered() {
		return nil, context.Canceled
	}

	basePlayerStats := baseResult.RaidStats.Parties[0].Players[0]
	baseStats := protoToCoreUnitStats(basePlayerStats.FinalStats)
	capBaseStats := addUnitStats(baseStats, buildDebuffUnitStats(request.Raid))
	weights := validateReforgeWeights(protoToCoreUnitStats(request.PreCapEpWeights), settings, normalizedConfig.softCaps)

	hardCaps := buildReforgeHardCaps(capBaseStats, settings, protoToCoreUnitStats(request.UndershootCaps))
	softCaps := buildReforgeSoftCaps(capBaseStats, normalizedConfig.softCaps)
	gemSortWeights := weights

	slotChoices, err := buildReforgeSlotChoices(request, baseRaid, baseGear, baseStats, weights, gemSortWeights, hardCaps, softCaps, statDeps, nil)
	if err != nil {
		return nil, err
	}

	return &reforgeOptimization{
		request:      request,
		settings:     settings,
		player:       player,
		baseRaid:     baseRaid,
		originalGear: originalGear,
		baseGear:     baseGear,
		capBaseStats: capBaseStats,
		weights:      weights,
		hardCaps:     hardCaps,
		softCaps:     softCaps,
		slotChoices:  slotChoices,
		statDeps:     statDeps,
	}, nil
}

func computeReforgeStats(request *proto.ComputeStatsRequest) *proto.ComputeStatsResult {
	request.SkipRotation = true
	return core.ComputeStats(request)
}

func computeReforgeStatsAndDeps(request *proto.ComputeStatsRequest) (*proto.ComputeStatsResult, *stats.StatDependencyManager) {
	request.SkipRotation = true
	return core.ComputeStatsAndDeps(request)
}

func (optimization *reforgeOptimization) searchState() *reforgeSearchState {
	choiceVarIdx := make([][]int, len(optimization.slotChoices))
	for i, slot := range optimization.slotChoices {
		choiceVarIdx[i] = make([]int, len(slot.choices))
	}
	return &reforgeSearchState{
		request:        optimization.request,
		baseRaid:       optimization.baseRaid,
		baseEquipment:  core.ProtoToEquipment(optimization.baseGear),
		baseGear:       optimization.baseGear,
		capBaseStats:   optimization.capBaseStats,
		statDeps:       optimization.statDeps,
		slots:          optimization.slotChoices,
		weights:        optimization.weights,
		hardCaps:       optimization.hardCaps,
		hardCapsByStat: reforgeHardCapsByStat(optimization.hardCaps),
		softCaps:       optimization.softCaps,
		softCapsByStat: reforgeSoftCapsByStat(optimization.softCaps),
		choiceVarIdx:   choiceVarIdx,
		uniqueGemIDs:   buildUniqueGemLimitIDs(optimization.slotChoices),
	}
}

func (optimization *reforgeOptimization) optimizedGear(choices []reforgeChoice) *proto.EquipmentSpec {
	gearEditor := newReforgeGearEditor(optimization.baseGear, optimization.originalGear, optimization.player, optimization.settings, optimization.request.GetGemOptions())
	gearEditor.applyChoices(choices)
	// minimizeRegems is temporarily disabled until the swap logic is fixed to
	// never change which gems are present, only permute their socket positions.
	gearEditor.minimizeRegems()
	return gearEditor.equipment()
}

func countReforgeChoices(slots []reforgeSlotChoices) int {
	count := 0
	for _, slot := range slots {
		count += len(slot.choices)
	}
	return count
}

func optimizeError(message string) *proto.ReforgeOptimizeResult {
	return &proto.ReforgeOptimizeResult{
		Error: &proto.ErrorOutcome{
			Message: message,
		},
	}
}

func optimizeAborted() *proto.ReforgeOptimizeResult {
	return &proto.ReforgeOptimizeResult{
		Error: &proto.ErrorOutcome{
			Type:    proto.ErrorOutcomeType_ErrorOutcomeAborted,
			Message: "Reforge optimization aborted.",
		},
	}
}
