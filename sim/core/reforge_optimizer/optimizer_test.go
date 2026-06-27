//go:build with_db

package reforgeoptimizer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	assetsdb "github.com/wowsims/tbc/assets/database"
	"github.com/wowsims/tbc/sim"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	protopkg "google.golang.org/protobuf/proto"
)

func TestReforgerOptimizer(t *testing.T) {
	sim.RegisterAll()

	testCases := []struct {
		name     string
		fileName string
		skip     bool
	}{
		{name: "reference-1", fileName: "reference-1.test.json"},
		{name: "reference-2", fileName: "reference-2.test.json"},
		{name: "reference-3", fileName: "reference-3.test.json"},
		{name: "reference-4", fileName: "reference-4.test.json"},
		{name: "reference-5", fileName: "reference-5.test.json"},
		{name: "reference-6", fileName: "reference-6.test.json"},
		{name: "reference-7", fileName: "reference-7.test.json"},
		{name: "reference-8", fileName: "reference-8.test.json"},
		{name: "reference-9", fileName: "reference-9.test.json"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip("skipping test case")
			}
			request := loadPreset(t, tc.fileName)
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

			if os.Getenv("UPDATE_FIXTURES") != "" {
				updateFixture(t, tc.fileName, request, optimizedGear)
				return
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
			statsDiffer := !isEmptyUnitStats(diff)
			if statsDiffer {
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

func updateFixture(t testing.TB, fileName string, request *proto.ReforgeOptimizeRequest, optimizedGear *proto.EquipmentSpec) {
	t.Helper()

	updated := protopkg.Clone(request).(*proto.ReforgeOptimizeRequest)
	updated.Raid.Parties[0].Players[0].Equipment = optimizedGear

	out, err := (protojson.MarshalOptions{Multiline: true, Indent: "\t", EmitUnpopulated: false}).Marshal(updated)
	if err != nil {
		t.Fatalf("failed marshalling updated fixture %s: %v", fileName, err)
	}
	if err := os.WriteFile(filepath.Join(".", fileName), out, 0644); err != nil {
		t.Fatalf("failed writing updated fixture %s: %v", fileName, err)
	}
	t.Logf("updated fixture %s", fileName)
}

// loadReforgeGemOptionsFromDB loads gem options from the embedded asset database,
// mirroring the UI's getReforgeGemOptions: quality >= Rare, no "Perfect" gems,
// all non-meta socket colors (Red/Blue/Yellow/Orange/Green/Purple/Prismatic).
func loadReforgeGemOptionsFromDB() []*proto.ReforgeGemOption {
	uiDB := assetsdb.Load()
	seen := make(map[int32]struct{})
	var options []*proto.ReforgeGemOption
	for _, gem := range uiDB.GetGems() {
		if gem.GetId() == 0 {
			continue
		}
		if gem.GetQuality() < proto.ItemQuality_ItemQualityRare {
			continue
		}
		if strings.Contains(gem.GetName(), "Perfect") {
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

	data, err := os.ReadFile(filepath.Join(".", fileName))
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
