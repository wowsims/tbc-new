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
	"google.golang.org/protobuf/encoding/protojson"
	protopkg "google.golang.org/protobuf/proto"
)

func TestReforgerOptimizer(t *testing.T) {
	sim.RegisterAll()

	testCases := []struct {
		name     string
		fileName string
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
		{name: "reference-10", fileName: "reference-10.test.json"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

			if !protopkg.Equal(expectedGear, optimizedGear) {
				t.Fatalf("optimized gear does not match expected gear\n%s", formatGearDiff(expectedGear, optimizedGear))
			}
		})
	}
}

func formatGearDiff(expectedGear *proto.EquipmentSpec, optimizedGear *proto.EquipmentSpec) string {
	maxItems := len(expectedGear.GetItems())
	if len(optimizedGear.GetItems()) > maxItems {
		maxItems = len(optimizedGear.GetItems())
	}

	out := ""
	for i := 0; i < maxItems; i++ {
		var expectedItem *proto.ItemSpec
		var optimizedItem *proto.ItemSpec
		if i < len(expectedGear.GetItems()) {
			expectedItem = expectedGear.GetItems()[i]
		}
		if i < len(optimizedGear.GetItems()) {
			optimizedItem = optimizedGear.GetItems()[i]
		}

		if protopkg.Equal(expectedItem, optimizedItem) {
			continue
		}

		expectedJSON := protojson.Format(expectedItem)
		optimizedJSON := protojson.Format(optimizedItem)
		out += fmt.Sprintf("slot %d:\nexpected: %s\nactual:   %s\n", i, expectedJSON, optimizedJSON)
	}

	return out
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
