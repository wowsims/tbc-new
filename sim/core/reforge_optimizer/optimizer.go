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
	debug := request.GetDebug()
	logAbort := request.GetMode() != proto.ReforgeOptimizeMode_ReforgeOptimizeModeBulk || debug
	if debug {
		log.Printf("[reforgeOptimize:%d] started debug=%t", requestID, debug)
		logRequestInput(requestID, request)
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

	solveStartedAt := time.Now()
	optimizer, err := newReforgeOptimizer(request, signals)
	var optimizedGear *proto.EquipmentSpec
	var score float64
	if err == nil {
		optimizedGear, score, err = optimizer.optimizeReforges()
	}
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
	if debug {
		log.Printf("[reforgeOptimize:%d] HiGHS solved in %s score=%.3f", requestID, time.Since(solveStartedAt), score)
	}
	if signals.Abort.IsTriggered() {
		if logAbort {
			log.Printf("[reforgeOptimize:%d] aborted after solving in %s", requestID, time.Since(startedAt))
		}
		return optimizeAborted()
	}

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
		if debug {
			logOptimizedGearSummary(requestID, optimizedGear)
		}
	}
	if !isBulk {
		log.Printf("[Reforge Optimizer] Reforge optimization completed requestID=%d total=%s score=%.3f", requestID, time.Since(startedAt), score)
	}

	return &proto.ReforgeOptimizeResult{
		OptimizedGear:        optimizedGear,
		OptimizedPlayerStats: optimizedPlayerStats,
		Score:                score,
		PassesDone:           1,
	}
}

type reforgeOptimizer struct {
	request  *proto.ReforgeOptimizeRequest
	settings *proto.ReforgeSettings
	player   *proto.Player
	signals  simsignals.Signals

	isTankSpec bool

	frozenSlots    map[proto.ItemSlot]bool
	undershootCaps core.UnitStats
	gemOptions     []*proto.ReforgeGemOption

	// statDeps is the player's build-phase StatDependencyManager (ComputeStatDependencies). It
	// resolves every stat conversion the sim models and is used by resolveStatDelta to compute
	// each LP variable's cap-space coefficients (the FULL dependency graph), separately from the
	// EP-calibrated objective coefficients produced by applyReforgeStat.
	statDeps *stats.StatDependencyManager

	baseRaidProto     *proto.Raid
	baseStrippedGear  *proto.EquipmentSpec
	originalEquipment *core.Equipment
	baseStats         core.UnitStats
	// capBaseStats adds the raid's debuffs on top of baseStats; caps are evaluated against it.
	capBaseStats core.UnitStats
}

// newReforgeOptimizer builds the optimizer context from the request: strips gems for the
// baseline, computes base stats and the stat dependency manager, and derives the player flags.
func newReforgeOptimizer(request *proto.ReforgeOptimizeRequest, signals simsignals.Signals) (*reforgeOptimizer, error) {
	settings := request.GetSettings()
	if settings == nil {
		settings = &proto.ReforgeSettings{}
	} else {
		settings = googleProto.Clone(settings).(*proto.ReforgeSettings)
	}

	request = googleProto.Clone(request).(*proto.ReforgeOptimizeRequest)
	request.Settings = settings

	baseRaid := googleProto.Clone(request.Raid).(*proto.Raid)
	originalGear := cloneEquipmentSpec(baseRaid.Parties[0].Players[0].Equipment)
	baseStrippedGear := cloneEquipmentSpec(originalGear)
	clearGems(baseStrippedGear, settings)
	player := baseRaid.Parties[0].Players[0]
	player.Equipment = baseStrippedGear

	// One environment build yields both FinalStats and the finalized StatDependencyManager,
	// instead of building the character twice for the same base raid.
	baseResult, baseSDM := computeReforgeStatsAndDeps(&proto.ComputeStatsRequest{Raid: baseRaid})
	if baseResult.ErrorResult != "" {
		return nil, errors.New(baseResult.ErrorResult)
	}
	if signals.Abort.IsTriggered() {
		return nil, context.Canceled
	}

	baseStats := protoToCoreUnitStats(baseResult.RaidStats.Parties[0].Players[0].FinalStats)
	originalEquipment := core.ProtoToEquipment(originalGear)

	return &reforgeOptimizer{
		request:  request,
		settings: settings,
		player:   player,
		signals:  signals,

		isTankSpec: playerIsTankSpec(player),

		frozenSlots:    frozenItemSlots(settings),
		undershootCaps: protoToCoreUnitStats(request.GetUndershootCaps()),
		gemOptions:     request.GetGemOptions(),

		statDeps: baseSDM,

		baseRaidProto:     baseRaid,
		baseStrippedGear:  baseStrippedGear,
		originalEquipment: &originalEquipment,
		baseStats:         baseStats,
		capBaseStats:      addUnitStats(baseStats, buildDebuffUnitStats(request.Raid)),
	}, nil
}

// optimizeReforges runs the gem/socket-bonus optimization: convert the configured caps into
// gap-to-cap form, build the LP model, solve it with cap refinement, then apply the winning gems
// back onto the gear.
func (o *reforgeOptimizer) optimizeReforges() (*proto.EquipmentSpec, float64, error) {
	reforgeCaps := computeStatCapsDelta(o.capBaseStats, protoToCoreUnitStats(o.settings.GetStatCaps()))

	var softCapConfigs []*proto.StatCapConfig
	if o.settings.GetUseSoftCapBreakpoints() {
		softCapConfigs = o.request.GetSoftCaps()
	}
	reforgeSoftCaps := computeReforgeSoftCaps(o.capBaseStats, softCapConfigs)

	weights := checkWeights(protoToCoreUnitStats(o.request.GetPreCapEpWeights()), reforgeCaps, reforgeSoftCaps)

	equipment := core.ProtoToEquipment(o.baseStrippedGear)
	variables := o.buildYalpsVariables(equipment, weights, reforgeCaps, reforgeSoftCaps)
	constraints := o.buildYalpsConstraints(equipment)
	addStructuralConstraints(variables, constraints)

	timeoutSeconds := optimizerTimeout.Seconds()
	if o.request.GetMode() == proto.ReforgeOptimizeMode_ReforgeOptimizeModeBulk {
		timeoutSeconds /= 4
	}

	selectedVars, score, err := o.solveModel(weights, reforgeCaps, reforgeSoftCaps, variables, constraints, timeoutSeconds)
	if err != nil {
		return nil, 0, err
	}

	return o.applyLPSolution(selectedVars), score, nil
}

func computeReforgeStats(request *proto.ComputeStatsRequest) *proto.ComputeStatsResult {
	request.SkipRotation = true
	return core.ComputeStats(request)
}

func computeReforgeStatsAndDeps(request *proto.ComputeStatsRequest) (*proto.ComputeStatsResult, *stats.StatDependencyManager) {
	request.SkipRotation = true
	return core.ComputeStatsAndDeps(request)
}

func optimizeError(message string) *proto.ReforgeOptimizeResult {
	return &proto.ReforgeOptimizeResult{
		Error: &proto.ErrorOutcome{Message: message},
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

const (
	optimizerTimeout = 30 * time.Second
)
