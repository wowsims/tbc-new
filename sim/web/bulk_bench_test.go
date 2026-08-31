//go:build with_db

package main

// End-to-end BulkSim benchmark exercising the full production flow:
//   EnsureBulkSimCandidatesGenerated → parallel reforge optimize (with gear cache)
//   → deduplicate by optimized gear → multistage bulk sim
//
// The request is assembled from the gem-pool-wide reforge fixture (a P1 warlock) plus a
// hand-picked warlock candidate item selection, mirroring the MoP repo's mage bench.
//
// Usage:
//
//	taskset -c 0-9 go test -tags with_db ./sim/web/ \
//	    -bench BenchmarkBulkSimWarlockFullFlow -benchtime 3x -benchmem -v -timeout 7200s

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/wowsims/tbc/sim"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

const warlockReforgeFixturePath = "../core/reforge_optimizer/test-fixtures/gem-pool-wide.test.json"

// Caster cloth epics (phase <= 2) the fixture warlock can equip, one per armor slot plus
// neck/ring picks, so candidate generation produces a real combination matrix.
var warlockBulkCandidateItemIDs = []int32{29990, 31976, 30107, 29918, 29987, 32799, 31975, 32787, 31924, 17108}

func loadWarlockBulkSimRequest(tb testing.TB) *proto.BulkSimRequest {
	tb.Helper()
	data, err := os.ReadFile(warlockReforgeFixturePath)
	if err != nil {
		tb.Skipf("reforge fixture not found at %q — skipping", warlockReforgeFixturePath)
		return nil
	}
	reforgeRequest := &proto.ReforgeOptimizeRequest{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, reforgeRequest); err != nil {
		tb.Fatalf("protojson.Unmarshal fixture: %v", err)
	}
	if reforgeRequest.GetRaid() == nil {
		tb.Fatalf("fixture is missing raid")
	}

	// The reforge fixture carries no rotation; give the warlock the destruction APL and
	// matching talents so the bulk stages sim a real workload.
	player := reforgeRequest.Raid.Parties[0].Players[0]
	player.Rotation = core.GetAplRotation("../../ui/warlock/dps/apls", "destruction").Rotation
	player.TalentsString = "-20500301332101-50500051220051053105"

	baseRequest := &proto.RaidSimRequest{
		Raid:      reforgeRequest.Raid,
		Encounter: core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{
			Iterations: 12500,
			RandomSeed: 83674828,
		},
	}

	items := make([]*proto.ItemSpec, 0, len(warlockBulkCandidateItemIDs))
	for _, id := range warlockBulkCandidateItemIDs {
		items = append(items, &proto.ItemSpec{Id: id})
	}

	// The bulk pre-pass owns the raid; the reforge template must not carry its own copy.
	bulkReforgeRequest := googleProto.Clone(reforgeRequest).(*proto.ReforgeOptimizeRequest)
	bulkReforgeRequest.Raid = nil

	return &proto.BulkSimRequest{
		RequestId:   "bench-bulk-warlock",
		BaseRequest: baseRequest,
		BulkSettings: &proto.BulkSettings{
			Items:              items,
			IterationsPerCombo: 12500,
		},
		ReforgeRequest:      bulkReforgeRequest,
		TopResults:          5,
		HighStageIterations: 12500,
	}
}

// BenchmarkBulkSimWarlockFullFlow runs the complete production pipeline via BulkSimAsync.
// Reports reforge_s, bulksim_s, sim_candidates, and top_dps per operation.
func BenchmarkBulkSimWarlockFullFlow(b *testing.B) {
	sim.RegisterAll()

	req := loadWarlockBulkSimRequest(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		iterReq := googleProto.Clone(req).(*proto.BulkSimRequest)

		progress := make(chan *proto.ProgressMetrics, 256)
		requestID := fmt.Sprintf("bench-bulk-warlock-%d", i)

		phaseDurations := make(map[proto.BulkSimStage]time.Duration)
		var prevStage proto.BulkSimStage
		prevTime := time.Now()

		type phaseResult struct {
			final   *proto.BulkSimResult
			timings map[proto.BulkSimStage]time.Duration
		}
		resultCh := make(chan phaseResult, 1)
		go func() {
			for msg := range progress {
				stage := msg.GetBulkStage()
				if stage != prevStage {
					now := time.Now()
					if prevStage != proto.BulkSimStage_BulkSimStageUnknown {
						phaseDurations[prevStage] = now.Sub(prevTime)
					}
					prevTime = now
					prevStage = stage
				}
				if msg.GetFinalBulkSimResult() != nil {
					resultCh <- phaseResult{final: msg.GetFinalBulkSimResult(), timings: phaseDurations}
					return
				}
			}
			resultCh <- phaseResult{timings: phaseDurations}
		}()

		BulkSimAsync(iterReq, progress, requestID)
		res := <-resultCh

		if err := res.final.GetError(); err != nil {
			b.Errorf("iter %d: BulkSim error: %s", i, err.GetMessage())
			continue
		}

		topDPS := 0.0
		for _, r := range res.final.GetTopResults() {
			if dps := r.GetDpsMetrics().GetAvg(); dps > topDPS {
				topDPS = dps
			}
		}

		reforgeDur := res.timings[proto.BulkSimStage_BulkSimStageReforge]
		simDur := res.timings[proto.BulkSimStage_BulkSimStageLow] +
			res.timings[proto.BulkSimStage_BulkSimStageMedium] +
			res.timings[proto.BulkSimStage_BulkSimStageHigh]

		simCandidates := len(iterReq.GetCandidates())
		b.Logf("iter %d: optimizedCandidates=%d simCandidates=%d reforge=%s bulksim=%s topDPS=%.1f results=%d",
			i, len(iterReq.GetOptimizedCandidates()), simCandidates,
			reforgeDur, simDur, topDPS, len(res.final.GetTopResults()))

		b.ReportMetric(reforgeDur.Seconds(), "reforge_s/op")
		b.ReportMetric(simDur.Seconds(), "bulksim_s/op")
		b.ReportMetric(float64(simCandidates), "sim_candidates/op")
		b.ReportMetric(topDPS, "top_dps/op")
	}
}
