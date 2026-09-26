package bulk

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

// Stat constraints (BulkSettings.stat_constraints) drop candidates whose final
// stats fail any constraint, after gem optimization and before the staged sims.
// Final stats are the ones the character sheet shows, so a surviving candidate
// displays matching numbers once equipped. Mirrors ui/core/wasm/bulk_sim/constraints.ts.

func bulkStatConstraintPasses(constraint *proto.BulkStatConstraint, value float64) bool {
	switch constraint.Op {
	case proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan:
		return value > constraint.Value
	case proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual:
		return value >= constraint.Value
	case proto.BulkStatConstraintOp_BulkStatConstraintOpEqual:
		return value == constraint.Value
	case proto.BulkStatConstraintOp_BulkStatConstraintOpLessThanOrEqual:
		return value <= constraint.Value
	case proto.BulkStatConstraintOp_BulkStatConstraintOpLessThan:
		return value < constraint.Value
	}
	return false
}

// The constrained Stat or PseudoStat read out of a final-stats proto. A
// constraint with no target reads as 0, as in the UI.
func bulkConstraintStatValue(constraint *proto.BulkStatConstraint, finalStats *proto.UnitStats) float64 {
	if finalStats == nil {
		return 0
	}
	switch target := constraint.UnitStat.(type) {
	case *proto.BulkStatConstraint_Stat:
		if idx := int(target.Stat); idx < len(finalStats.Stats) {
			return finalStats.Stats[idx]
		}
	case *proto.BulkStatConstraint_PseudoStat:
		if idx := int(target.PseudoStat); idx < len(finalStats.PseudoStats) {
			return finalStats.PseudoStats[idx]
		}
	}
	return 0
}

func bulkFinalStatsPassConstraints(constraints []*proto.BulkStatConstraint, finalStats *proto.UnitStats) bool {
	for _, constraint := range constraints {
		if !bulkStatConstraintPasses(constraint, bulkConstraintStatValue(constraint, finalStats)) {
			return false
		}
	}
	return true
}

// Final stats of the bulk player wearing the candidate's gear, as the stats panel shows them:
// ComputeStats' final stats plus the raid debuffs the panel attributes to the character.
func bulkSimCandidateFinalStats(request *proto.BulkSimRequest, candidate BulkSimCandidate) (finalStats *proto.UnitStats, errOutcome *proto.ErrorOutcome) {
	defer func() {
		if r := recover(); r != nil {
			errOutcome = &proto.ErrorOutcome{Message: fmt.Sprintf("[Bulk Sim] Computing stats for constraint check: %v", r)}
		}
	}()
	raid := googleProto.Clone(request.BaseRequest.Raid).(*proto.Raid)
	player, err := getBulkSimPlayer(raid)
	if err != "" {
		return nil, &proto.ErrorOutcome{Message: err}
	}
	player.Equipment = googleProto.Clone(candidate.Gear).(*proto.EquipmentSpec)
	adjustCandidateImbues(player)
	// Read before computing: building the character rewrites the raid's debuffs (a hunter with
	// Expose Weakness clears the raid's copy of it), and the panel shows the configured ones.
	debuffs := googleProto.Clone(raid.GetDebuffs()).(*proto.Debuffs)
	result := core.ComputeStats(&proto.ComputeStatsRequest{Raid: raid, Encounter: request.BaseRequest.Encounter})
	if result.ErrorResult != "" {
		return nil, &proto.ErrorOutcome{Message: result.ErrorResult}
	}
	return core.WithCharacterSheetDebuffs(result.RaidStats.Parties[0].Players[0].FinalStats, debuffs, player), nil
}

// Keeps the candidates whose final stats satisfy every constraint, in order.
// With no constraints every candidate survives without any stats being computed.
// Progress is reported as the constraints stage; an abort or the first stats
// error ends the check.
func filterBulkSimCandidatesByConstraints(request *proto.BulkSimRequest, candidates []BulkSimCandidate, progress chan *proto.ProgressMetrics, signals simsignals.Signals) (survivors []BulkSimCandidate, skipped int, errOutcome *proto.ErrorOutcome) {
	constraints := request.GetBulkSettings().GetStatConstraints()
	if len(constraints) == 0 || len(candidates) == 0 {
		return candidates, 0, nil
	}

	startedAt := time.Now()
	log.Printf("[Bulk Sim] Stat constraint check started candidates=%d constraints=%d", len(candidates), len(constraints))
	emitBulkSimStageProgress(progress, proto.BulkSimStage_BulkSimStageConstraints, 0, len(candidates), 0, 0, 0)

	passes := make([]bool, len(candidates))
	var (
		next      atomic.Int64
		failed    atomic.Bool
		mu        sync.Mutex
		firstErr  *proto.ErrorOutcome
		completed int
		lastEmit  time.Time
		wg        sync.WaitGroup
	)
	next.Store(-1)
	workers := max(1, min(runtime.NumCPU(), len(candidates)))
	for w := 0; w < workers; w++ {
		wg.Go(func() {
			for {
				i := int(next.Add(1))
				if i >= len(candidates) || failed.Load() || signals.Abort.IsTriggered() {
					return
				}
				finalStats, err := bulkSimCandidateFinalStats(request, candidates[i])
				if err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					failed.Store(true)
					return
				}
				passes[i] = bulkFinalStatsPassConstraints(constraints, finalStats)

				mu.Lock()
				completed++
				done := completed
				emit := done == len(candidates) || time.Since(lastEmit) >= BulkSimProgressThrottle
				if emit {
					lastEmit = time.Now()
				}
				mu.Unlock()
				if emit {
					emitBulkSimStageProgress(progress, proto.BulkSimStage_BulkSimStageConstraints, done, len(candidates), 0, 0, 0)
				}
			}
		})
	}
	wg.Wait()

	if signals.Abort.IsTriggered() {
		return nil, 0, bulkSimAbortedError()
	}
	if firstErr != nil {
		return nil, 0, firstErr
	}
	survivors = make([]BulkSimCandidate, 0, len(candidates))
	for i, candidate := range candidates {
		if passes[i] {
			survivors = append(survivors, candidate)
		}
	}
	skipped = len(candidates) - len(survivors)
	log.Printf("[Bulk Sim] Stat constraint check completed total=%s candidates=%d survivors=%d skipped=%d", time.Since(startedAt), len(candidates), len(survivors), skipped)
	return survivors, skipped, nil
}
