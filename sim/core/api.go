// Proto-based function interface for the simulator
package core

import (
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/simsignals"
	"github.com/wowsims/tbc/sim/core/stats"
)

/**
 * Returns character stats taking into account gear / buffs / consumes / etc
 */
func ComputeStats(csr *proto.ComputeStatsRequest) *proto.ComputeStatsResult {
	encounter := csr.Encounter
	if encounter == nil {
		encounter = &proto.Encounter{}
	}

	_, raidStats, encounterStats := NewEnvironment(csr.Raid, encounter, !csr.SkipRotation, csr.SkipRotation)

	return &proto.ComputeStatsResult{
		RaidStats:      raidStats,
		EncounterStats: encounterStats,
	}
}

// ComputeStatDependencies builds a character from the request and returns its finalized
// StatDependencyManager. Lightweight compared to ComputeStats — builds the character and
// resolves all stat dependencies but does not run the simulation.
func ComputeStatDependencies(request *proto.ComputeStatsRequest) *stats.StatDependencyManager {
	_, sdm := ComputeStatsAndDeps(request)
	return sdm
}

// ComputeStatsAndDeps combines a skip-rotation ComputeStats with ComputeStatDependencies
// in a single NewEnvironment call. Use this when both are needed for the same raid to
// avoid building the character environment twice.
func ComputeStatsAndDeps(request *proto.ComputeStatsRequest) (*proto.ComputeStatsResult, *stats.StatDependencyManager) {
	encounter := request.Encounter
	if encounter == nil {
		encounter = &proto.Encounter{}
	}
	env, raidStats, encounterStats := NewEnvironment(request.Raid, encounter, false, true)
	result := &proto.ComputeStatsResult{
		RaidStats:      raidStats,
		EncounterStats: encounterStats,
	}
	if len(env.Raid.Parties) == 0 || len(env.Raid.Parties[0].Players) == 0 {
		return result, &stats.StatDependencyManager{}
	}
	character := env.Raid.Parties[0].Players[0].GetCharacter()
	// FillPlayerStats activates build-phase auras to compute FinalStats then clears them.
	// Re-apply base-phase auras so that starting-form multipliers are active in the
	// returned SDM without also enabling talent/gear/buff deps already handled analytically.
	character.applyBuildPhaseAuras(CharacterBuildPhaseBase)
	sdm := character.StatDependencyManager
	return result, &sdm
}

/**
 * Returns stat weights and EP values, with standard deviations, for all stats.
 */
func StatWeights(request *proto.StatWeightsRequest) *proto.StatWeightsResult {
	return runStatWeights(request, nil, simsignals.CreateSignals())
}

func StatWeightsAsync(request *proto.StatWeightsRequest, progress chan *proto.ProgressMetrics, requestId string) {
	signals, err := simsignals.RegisterWithId(requestId)
	if err != nil {
		progress <- &proto.ProgressMetrics{
			FinalWeightResult: &proto.StatWeightsResult{
				Error: &proto.ErrorOutcome{
					Message: "Couldn't register for signal API: " + err.Error(),
				},
			},
		}
		return
	}
	go func() {
		defer simsignals.UnregisterId(requestId)
		result := runStatWeights(request, progress, signals)
		progress <- &proto.ProgressMetrics{
			FinalWeightResult: result,
		}
	}()
}

// Get data for all requests needed for stat weights.
func StatWeightRequests(request *proto.StatWeightsRequest) *proto.StatWeightRequestsData {
	return buildStatWeightRequests(request)
}

func StatWeightCompute(request *proto.StatWeightsCalcRequest) *proto.StatWeightsResult {
	return computeStatWeights(request)
}

/**
 * Runs multiple iterations of the sim with a full raid.
 */
func RunRaidSim(request *proto.RaidSimRequest) *proto.RaidSimResult {
	return RunSim(request, nil, simsignals.CreateSignals())
}

func RunRaidSimAsync(request *proto.RaidSimRequest, progress chan *proto.ProgressMetrics, requestId string) {
	signals, err := simsignals.RegisterWithId(requestId)
	if err != nil {
		progress <- &proto.ProgressMetrics{
			FinalRaidResult: &proto.RaidSimResult{
				Error: &proto.ErrorOutcome{
					Message: "Couldn't register for signal API: " + err.Error(),
				},
			},
		}
		return
	}
	go func() {
		defer simsignals.UnregisterId(requestId)
		RunSim(request, progress, signals)
	}()
}

// Threading does not work in WASM!
func RunRaidSimConcurrent(request *proto.RaidSimRequest) *proto.RaidSimResult {
	return runSimConcurrent(request, nil, simsignals.CreateSignals())
}

// Threading does not work in WASM!
func RunRaidSimConcurrentWithSignals(request *proto.RaidSimRequest, signals simsignals.Signals) *proto.RaidSimResult {
	return runSimConcurrent(request, nil, signals)
}

// Threading does not work in WASM!
func RunRaidSimConcurrentAsync(request *proto.RaidSimRequest, progress chan *proto.ProgressMetrics, requestId string) {
	signals, err := simsignals.RegisterWithId(requestId)
	if err != nil {
		progress <- &proto.ProgressMetrics{
			FinalRaidResult: &proto.RaidSimResult{
				Error: &proto.ErrorOutcome{
					Message: "Couldn't register for signal API: " + err.Error(),
				},
			},
		}
		return
	}
	go func() {
		defer simsignals.UnregisterId(requestId)
		runSimConcurrent(request, progress, signals)
	}()
}

var runningInWasm = false

func SetRunningInWasm() {
	runningInWasm = true
}

func IsRunningInWasm() bool {
	return runningInWasm
}
