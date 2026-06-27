package reforgeoptimizer

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func TestBuildChoiceMIPModelKeepsSatisfiedMetaGemCompareConstraint(t *testing.T) {
	const blueGemID int32 = 91001
	const yellowGemID int32 = 91002
	originalBlueGem, blueGemExisted := core.GemsByID[blueGemID]
	originalYellowGem, yellowGemExisted := core.GemsByID[yellowGemID]
	core.GemsByID[blueGemID] = core.Gem{ID: blueGemID, Color: proto.GemColor_GemColorBlue}
	core.GemsByID[yellowGemID] = core.Gem{ID: yellowGemID, Color: proto.GemColor_GemColorYellow}
	t.Cleanup(func() {
		if blueGemExisted {
			core.GemsByID[blueGemID] = originalBlueGem
		} else {
			delete(core.GemsByID, blueGemID)
		}
		if yellowGemExisted {
			core.GemsByID[yellowGemID] = originalYellowGem
		} else {
			delete(core.GemsByID, yellowGemID)
		}
	})

	search := &reforgeSearchState{
		baseEquipment: core.Equipment{
			proto.ItemSlot_ItemSlotHead: core.Item{Gems: []core.Gem{{ID: 25893, Color: proto.GemColor_GemColorMeta}, {ID: blueGemID, Color: proto.GemColor_GemColorBlue}}},
		},
		slots: []reforgeSlotChoices{{
			slot: proto.ItemSlot_ItemSlotChest,
			choices: []reforgeChoice{
				{slot: proto.ItemSlot_ItemSlotChest, socketChoice: true, socketIdx: 0, gems: []reforgeGemChoice{{socketIdx: 0, gemID: yellowGemID}}},
			},
		}},
	}

	model := buildChoiceMIPModel(search, core.NewUnitStats(), nil)

	for _, constraint := range model.constraints {
		if constraint.lower == 0 && constraint.upper == 0 {
			continue
		}
		if constraint.lower == 0 && constraint.upper == math.Inf(1) && len(constraint.values) == 1 && constraint.values[0] == -1 {
			return
		}
	}

	t.Fatalf("meta gem compare constraint was not retained for satisfied base gear: %#v", model.constraints)
}

func TestSolveMIPWithHiGHSPrefersPurpleOverOrangeForShadowMeta(t *testing.T) {
	const purpleGemID int32 = 91003
	const orangeGemID int32 = 91004
	originalPurpleGem, purpleGemExisted := core.GemsByID[purpleGemID]
	originalOrangeGem, orangeGemExisted := core.GemsByID[orangeGemID]
	core.GemsByID[purpleGemID] = core.Gem{ID: purpleGemID, Color: proto.GemColor_GemColorPurple, Stats: stats.Stats{stats.Intellect: 1}}
	core.GemsByID[orangeGemID] = core.Gem{ID: orangeGemID, Color: proto.GemColor_GemColorOrange, Stats: stats.Stats{stats.Intellect: 10}}
	t.Cleanup(func() {
		if purpleGemExisted {
			core.GemsByID[purpleGemID] = originalPurpleGem
		} else {
			delete(core.GemsByID, purpleGemID)
		}
		if orangeGemExisted {
			core.GemsByID[orangeGemID] = originalOrangeGem
		} else {
			delete(core.GemsByID, orangeGemID)
		}
	})

	weights := core.NewUnitStats()
	weights = setUnitStat(weights, stats.UnitStatFromStat(stats.Intellect), 1)
	search := &reforgeSearchState{
		baseEquipment: core.Equipment{
			proto.ItemSlot_ItemSlotHead: core.Item{Gems: []core.Gem{{ID: 25893, Color: proto.GemColor_GemColorMeta}}},
		},
		weights: weights,
		slots: []reforgeSlotChoices{{
			slot: proto.ItemSlot_ItemSlotChest,
			choices: []reforgeChoice{
				{slot: proto.ItemSlot_ItemSlotChest, socketChoice: true, socketIdx: 0, gems: []reforgeGemChoice{{socketIdx: 0, gemID: purpleGemID}}, objectiveDelta: core.NewUnitStats()},
				{slot: proto.ItemSlot_ItemSlotChest, socketChoice: true, socketIdx: 0, gems: []reforgeGemChoice{{socketIdx: 0, gemID: orangeGemID}}, objectiveDelta: setUnitStat(core.NewUnitStats(), stats.UnitStatFromStat(stats.Intellect), 10)},
			},
		}},
	}

	model := buildChoiceMIPModel(search, weights, nil)
	solution, solved, err := solveMIPWithHiGHS(model, 5*time.Second, 0)
	if err != nil {
		t.Fatalf("solveMIPWithHiGHS returned error: %v", err)
	}
	if !solved {
		t.Fatal("solveMIPWithHiGHS did not solve shadow meta MIP")
	}
	if len(solution.values) != 2 {
		t.Fatalf("expected 2 solution values, got %d", len(solution.values))
	}
	if solution.values[0] < 0.5 || solution.values[1] >= 0.5 {
		t.Fatalf("expected purple gem to win over orange gem for 25893, got %v", solution.values)
	}
}

func TestSolveMIPWithHiGHSCountsGemReplacementsForMetaActivation(t *testing.T) {
	const blueGemID int32 = 91005
	const yellowGemID int32 = 91006
	originalBlueGem, blueGemExisted := core.GemsByID[blueGemID]
	originalYellowGem, yellowGemExisted := core.GemsByID[yellowGemID]
	core.GemsByID[blueGemID] = core.Gem{ID: blueGemID, Color: proto.GemColor_GemColorBlue, Stats: stats.Stats{stats.Intellect: 1}}
	core.GemsByID[yellowGemID] = core.Gem{ID: yellowGemID, Color: proto.GemColor_GemColorYellow, Stats: stats.Stats{stats.Intellect: 10}}
	t.Cleanup(func() {
		if blueGemExisted {
			core.GemsByID[blueGemID] = originalBlueGem
		} else {
			delete(core.GemsByID, blueGemID)
		}
		if yellowGemExisted {
			core.GemsByID[yellowGemID] = originalYellowGem
		} else {
			delete(core.GemsByID, yellowGemID)
		}
	})

	weights := core.NewUnitStats()
	weights = setUnitStat(weights, stats.UnitStatFromStat(stats.Intellect), 1)
	search := &reforgeSearchState{
		baseEquipment: core.Equipment{
			proto.ItemSlot_ItemSlotHead:  core.Item{Gems: []core.Gem{{ID: 25893, Color: proto.GemColor_GemColorMeta}, {ID: blueGemID, Color: proto.GemColor_GemColorBlue}}},
			proto.ItemSlot_ItemSlotChest: core.Item{Gems: []core.Gem{{ID: blueGemID, Color: proto.GemColor_GemColorBlue}}},
		},
		weights: weights,
		slots: []reforgeSlotChoices{{
			slot: proto.ItemSlot_ItemSlotChest,
			choices: []reforgeChoice{
				{slot: proto.ItemSlot_ItemSlotChest, socketChoice: true, socketIdx: 0, gems: []reforgeGemChoice{{socketIdx: 0, gemID: blueGemID}}, objectiveDelta: setUnitStat(core.NewUnitStats(), stats.UnitStatFromStat(stats.Intellect), 1)},
				{slot: proto.ItemSlot_ItemSlotChest, socketChoice: true, socketIdx: 0, gems: []reforgeGemChoice{{socketIdx: 0, gemID: yellowGemID}}, objectiveDelta: setUnitStat(core.NewUnitStats(), stats.UnitStatFromStat(stats.Intellect), 10)},
			},
		}},
	}

	model := buildChoiceMIPModel(search, weights, nil)
	solution, solved, err := solveMIPWithHiGHS(model, 5*time.Second, 0)
	if err != nil {
		t.Fatalf("solveMIPWithHiGHS returned error: %v", err)
	}
	if !solved {
		t.Fatal("solveMIPWithHiGHS did not solve meta replacement MIP")
	}
	if len(solution.values) != 2 {
		t.Fatalf("expected 2 solution values, got %d", len(solution.values))
	}
	if solution.values[0] < 0.5 || solution.values[1] >= 0.5 {
		t.Fatalf("expected blue gem to win over yellow gem when replacing a socket gem for 25893, got %v", solution.values)
	}
}

func TestShouldForceSocketBonusDoesNotAutoForceCappedHitBonus(t *testing.T) {
	const blueGemID int32 = 91007
	const yellowGemID int32 = 91008
	const redGemID int32 = 91009

	item := core.Item{GemSockets: []proto.GemColor{proto.GemColor_GemColorBlue, proto.GemColor_GemColorYellow}}
	item.SocketBonus = stats.Stats{}
	item.SocketBonus[stats.SpellHitRating] = 3

	weights := core.NewUnitStats()
	weights = setUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSchoolHitPercentShadow), 1)
	weights = setUnitStat(weights, stats.UnitStatFromStat(stats.Intellect), 0.1)

	gemOptions := map[proto.GemColor][]reforgeGemOption{
		proto.GemColor_GemColorBlue:      {{id: blueGemID, objectiveDelta: unitStatsFromStats(stats.Stats{stats.Intellect: 1}, weights)}},
		proto.GemColor_GemColorYellow:    {{id: yellowGemID, objectiveDelta: unitStatsFromStats(stats.Stats{stats.Intellect: 1}, weights)}},
		proto.GemColor_GemColorPrismatic: {{id: redGemID, objectiveDelta: unitStatsFromStats(stats.Stats{stats.Intellect: 100}, weights)}},
	}

	if got := shouldForceSocketBonus(item, item.GemSockets, gemOptions, weights, []reforgeHardCap{{unitStat: stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSchoolHitPercentShadow), cap: 1}}, nil); got {
		t.Fatal("expected capped hit socket bonus to rely on EP comparison rather than be auto-forced")
	}
}

func TestBuildChoiceMIPModelUsesAnalyticChoiceDeltaForObjective(t *testing.T) {
	exactDelta := core.NewUnitStats()
	exactDelta = setUnitStat(exactDelta, stats.UnitStatFromStat(stats.SpellHasteRating), 10)
	analyticDelta := core.NewUnitStats()
	analyticDelta = setUnitStat(analyticDelta, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSpellCritPercent), 10)
	weights := core.NewUnitStats()
	weights = setUnitStat(weights, stats.UnitStatFromStat(stats.SpellHasteRating), 1)
	weights = setUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSpellCritPercent), 100)
	search := &reforgeSearchState{
		slots: []reforgeSlotChoices{{
			slot: proto.ItemSlot_ItemSlotHead,
			choices: []reforgeChoice{
				{slot: proto.ItemSlot_ItemSlotHead, socketChoice: true, socketIdx: 0, gems: []reforgeGemChoice{{socketIdx: 0, gemID: 0}}},
				{slot: proto.ItemSlot_ItemSlotHead, socketChoice: true, socketIdx: 0, gems: []reforgeGemChoice{{socketIdx: 0, gemID: 113}}, delta: exactDelta, objectiveDelta: analyticDelta},
			},
		}},
	}

	model := buildChoiceMIPModel(search, weights, nil)

	if len(model.variables) != 1 {
		t.Fatalf("expected 1 MIP variable, got %d", len(model.variables))
	}
	if model.variables[0].objective != 1000 {
		t.Fatalf("expected analytic delta objective 1000, got %v", model.variables[0].objective)
	}
}

func TestBuildChoiceMIPModelUsesExactChoiceDeltaForConstraints(t *testing.T) {
	unitStat := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSpellCritPercent)
	exactDelta := core.NewUnitStats()
	exactDelta = setUnitStat(exactDelta, unitStat, 10)
	analyticDelta := core.NewUnitStats()
	analyticDelta = setUnitStat(analyticDelta, unitStat, 20)
	search := &reforgeSearchState{
		slots: []reforgeSlotChoices{{
			slot: proto.ItemSlot_ItemSlotHead,
			choices: []reforgeChoice{
				{slot: proto.ItemSlot_ItemSlotHead, socketChoice: true, socketIdx: 0, gems: []reforgeGemChoice{{socketIdx: 0, gemID: 0}}},
				{slot: proto.ItemSlot_ItemSlotHead, socketChoice: true, socketIdx: 0, gems: []reforgeGemChoice{{socketIdx: 0, gemID: 113}}, delta: exactDelta, objectiveDelta: analyticDelta},
			},
		}},
	}
	constraints := []mipStatConstraint{{unitStat: unitStat, lower: 5, upper: 100}}

	model := buildChoiceMIPModel(search, core.NewUnitStats(), constraints)

	if len(model.constraints) != 2 {
		t.Fatalf("expected choice and stat constraints, got %d", len(model.constraints))
	}
	statConstraint := model.constraints[1]
	if len(statConstraint.values) != 1 || statConstraint.values[0] != 10 {
		t.Fatalf("expected exact delta constraint coefficient 10, got %v", statConstraint.values)
	}
}
