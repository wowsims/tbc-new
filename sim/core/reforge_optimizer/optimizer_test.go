//go:build with_db

package reforgeoptimizer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	assetsdb "github.com/wowsims/tbc/assets/database"
	"github.com/wowsims/tbc/sim"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	protopkg "google.golang.org/protobuf/proto"
)

// fixturesDir holds the committed spec-based parity fixtures (<spec>.test.json), each
// generated from master gear by TestGenerateReforgeFixtures (fixture_support_test.go).
// Enumerating by glob keeps this suite self-contained: it does not depend on the
// fixture_*_test.go generators. If no fixtures are present, skip rather than fail so the
// package's other tests still run.
const fixturesDir = "test-fixtures"

func TestReforgerOptimizer(t *testing.T) {
	sim.RegisterAll()

	paths, err := filepath.Glob(filepath.Join(fixturesDir, "*.test.json"))
	if err != nil {
		t.Fatalf("globbing fixtures: %v", err)
	}
	if len(paths) == 0 {
		t.Skip("no fixtures in " + fixturesDir)
	}
	sort.Strings(paths)

	for _, path := range paths {
		fileName := filepath.Base(path)
		t.Run(strings.TrimSuffix(fileName, ".test.json"), func(t *testing.T) {
			request := loadPreset(t, fileName)
			expectedGear := request.GetRaid().GetParties()[0].GetPlayers()[0].GetEquipment()
			if expectedGear == nil {
				t.Fatal("preset has no player equipment to compare against")
			}

			result := Optimize(request)
			if err := result.GetError(); err != nil {
				t.Fatalf("Optimize returned error: %s", err.GetMessage())
			}
			optimizedGear := result.GetOptimizedGear()
			if optimizedGear == nil {
				t.Fatal("Optimize returned no optimized gear")
			}

			expectedRaid := protopkg.Clone(request.Raid).(*proto.Raid)
			expectedRaid.Parties[0].Players[0].Equipment = expectedGear
			expectedResult := computeReforgeStats(&proto.ComputeStatsRequest{Raid: expectedRaid})
			if expectedResult.ErrorResult != "" {
				t.Fatalf("ComputeStats on expected gear failed: %s", expectedResult.ErrorResult)
			}
			expStats := protoToCoreUnitStats(expectedResult.RaidStats.Parties[0].Players[0].FinalStats)
			optStats := protoToCoreUnitStats(result.GetOptimizedPlayerStats().GetFinalStats())
			diff := subtractUnitStats(optStats, expStats)
			if !isEmptyUnitStats(diff) {
				for i, expItem := range expectedGear.GetItems() {
					var optItem *proto.ItemSpec
					if i < len(optimizedGear.GetItems()) {
						optItem = optimizedGear.GetItems()[i]
					}
					if !protopkg.Equal(expItem, optItem) {
						expJSON, _ := protojson.Marshal(expItem)
						optJSON, _ := protojson.Marshal(optItem)
						t.Logf("slot %d: expected %s", i, expJSON)
						t.Logf("slot %d: got      %s", i, optJSON)
					}
				}
				for statIdx, d := range diff.Stats {
					if d != 0 {
						t.Logf("stat %-24s expected=%8.2f got=%8.2f diff=%+.2f", stats.Stat(statIdx).StatName(), expStats.Stats[statIdx], optStats.Stats[statIdx], d)
					}
				}
				for psIdx, d := range diff.PseudoStats {
					if d != 0 {
						name := proto.PseudoStat_name[int32(psIdx)]
						if name == "" {
							name = fmt.Sprintf("PseudoStat(%d)", psIdx)
						}
						t.Logf("stat %-24s expected=%8.4f got=%8.4f diff=%+.4f", name, expStats.PseudoStats[psIdx], optStats.PseudoStats[psIdx], d)
					}
				}
				t.Fatal("optimized stats do not match expected stats")
			}
		})
	}
}

// loadReforgeGemOptionsFromDB loads gem options from the embedded asset database, mirroring the
// gem list the UI hands the reforger: every non-meta socket colour, with NO lower quality bound.
// Quality is gated inside buildGemOptions via the request's max_gem_quality (an upper bound), so
// filtering out Uncommon gems here would starve the solver of candidates the reference reforger
// considers and change which equally-scoring solution it settles on.
func loadReforgeGemOptionsFromDB() []*proto.ReforgeGemOption {
	uiDB := assetsdb.Load()
	seen := make(map[int32]struct{})
	var options []*proto.ReforgeGemOption
	for _, gem := range uiDB.GetGems() {
		if gem.GetId() == 0 {
			continue
		}
		c := gem.GetColor()
		if c == proto.GemColor_GemColorUnknown || c == proto.GemColor_GemColorMeta {
			continue
		}
		if _, dup := seen[gem.GetId()]; dup {
			continue
		}
		seen[gem.GetId()] = struct{}{}
		options = append(options, &proto.ReforgeGemOption{
			Id:                 gem.GetId(),
			Name:               gem.GetName(),
			Color:              gem.GetColor(),
			Stats:              gem.GetStats(),
			Unique:             gem.GetUnique(),
			RequiredProfession: gem.GetRequiredProfession(),
			Icon:               gem.GetIcon(),
			Phase:              gem.GetPhase(),
			Quality:            gem.GetQuality(),
		})
	}
	return options
}

func loadPreset(t *testing.T, fileName string) *proto.ReforgeOptimizeRequest {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(fixturesDir, fileName))
	if err != nil {
		t.Fatalf("failed reading preset %s: %v", fileName, err)
	}

	request := &proto.ReforgeOptimizeRequest{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, request); err != nil {
		t.Fatalf("failed unmarshalling fixture %s: %v", fileName, err)
	}

	// Auto-populate gem options from the embedded DB when the fixture has none
	// but gem optimization is enabled (maxGemPhase > 0).
	if len(request.GemOptions) == 0 && request.GetSettings().GetMaxGemPhase() > 0 {
		request.GemOptions = loadReforgeGemOptionsFromDB()
	}

	return request
}

// subtractUnitStats is used by the fixture parity test to diff expected against optimized stats.
func subtractUnitStats(a core.UnitStats, b core.UnitStats) core.UnitStats {
	result := a
	result.Stats = a.Stats.Subtract(b.Stats)
	maxLen := max(len(a.PseudoStats), len(b.PseudoStats))
	result.PseudoStats = make([]float64, maxLen)
	copy(result.PseudoStats, a.PseudoStats)
	for idx, value := range b.PseudoStats {
		result.PseudoStats[idx] -= value
	}
	return result
}
