package reforgeoptimizer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

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
		{name: "normal", fileName: "normal.test.json"},
		{name: "reference-1", fileName: "reforge-reference-1.json"},
		{name: "reference-2", fileName: "reforge-reference-2.json"},
		{name: "reference-3", fileName: "reforge-reference-3.json"},
		{name: "reference-4", fileName: "reforge-reference-4.json"},
		{name: "reference-5", fileName: "reforge-reference-5.json"},
		{name: "reference-6", fileName: "reforge-reference-6.json"},
		{name: "reference-7", fileName: "reforge-reference-7.json"},
		{name: "reference-8", fileName: "reforge-reference-8.json"},
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

func TestReforgerOptimizerReference9RespectsMaxGemPhase(t *testing.T) {
	sim.RegisterAll()

	request := loadPreset(t, "reforge-reference-9.json")
	originalGear := request.GetRaid().GetParties()[0].GetPlayers()[0].GetEquipment()
	if originalGear == nil {
		t.Fatal("reference-9 fixture has no input gear")
	}

	result := Optimize(request)
	if err := result.GetError(); err != nil {
		t.Fatalf("Optimize returned error: %s", err.GetMessage())
	}
	optimizedGear := result.GetOptimizedGear()
	if optimizedGear == nil {
		t.Fatal("Optimize returned no optimized gear")
	}

	if protopkg.Equal(originalGear, optimizedGear) {
		t.Fatal("reference-9 optimization unexpectedly made no gear changes")
	}

	maxGemPhase := request.GetSettings().GetMaxGemPhase()
	gemPhaseByID := map[int32]int32{}
	for _, gemOption := range request.GetGemOptions() {
		gemPhaseByID[gemOption.GetId()] = gemOption.GetPhase()
	}

	for slotIdx, item := range optimizedGear.GetItems() {
		if item == nil {
			continue
		}
		for socketIdx, gemID := range item.GetGems() {
			if gemID == 0 {
				continue
			}
			phase, ok := gemPhaseByID[gemID]
			if !ok {
				continue
			}
			if phase > maxGemPhase {
				t.Fatalf("slot %d socket %d gem %d has phase %d > maxGemPhase %d", slotIdx, socketIdx, gemID, phase, maxGemPhase)
			}
		}
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

func loadPreset(t *testing.T, fileName string) *proto.ReforgeOptimizeRequest {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(".", fileName))
	if err != nil {
		data, err = os.ReadFile(filepath.Join("..", "..", "..", fileName))
		if err != nil {
			t.Fatalf("failed reading preset %s: %v", fileName, err)
		}
	}

	request := &proto.ReforgeOptimizeRequest{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, request); err != nil {
		t.Fatalf("failed unmarshalling fixture %s: %v", fileName, err)
	}
	return request
}
