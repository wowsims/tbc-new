package reforgeoptimizer

import (
	"cmp"
	"math"
	"slices"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func buildReforgeSlotChoices(request *proto.ReforgeOptimizeRequest, baseRaid *proto.Raid, baseGear *proto.EquipmentSpec, baseStats core.UnitStats, weights core.UnitStats, gemSortWeights core.UnitStats, hardCaps []reforgeHardCap, softCaps []reforgeSoftCap, statDeps *stats.StatDependencyManager, statConstraints []mipStatConstraint) ([]reforgeSlotChoices, error) {
	frozenSlots := frozenItemSlots(request.GetSettings())
	player := request.Raid.Parties[0].Players[0]
	if statDeps == nil {
		statDeps = core.ComputeStatDependencies(&proto.ComputeStatsRequest{Raid: baseRaid})
	}
	gemOptions := buildReforgeGemOptions(request, player, gemSortWeights, hardCaps, softCaps, statConstraints)
	baseEquipment := core.ProtoToEquipment(baseGear)

	allSlots := make([]reforgeSlotChoices, 0, int(core.NumItemSlots))
	for slotIdx := 0; slotIdx < int(core.NumItemSlots); slotIdx++ {
		slot := proto.ItemSlot(slotIdx)
		item := baseEquipment.GetItemBySlot(slot)
		if item.ID == 0 || frozenSlots[slot] {
			continue
		}

		socketColors := currentSocketColors(*item)
		forceSocketBonus := shouldForceSocketBonus(*item, socketColors, gemOptions, weights, hardCaps, softCaps)
		socketBonusSocketCount := socketBonusNormalization(socketColors)
		distributedSocketBonusDelta := core.NewUnitStats()
		distributedSocketBonusObjectiveDelta := core.NewUnitStats()
		if forceSocketBonus && socketBonusSocketCount > 0 {
			distributedSocketBonus := item.SocketBonus.Multiply(1 / float64(socketBonusSocketCount))
			distributedSocketBonusDelta = resolveStatDelta(statDeps, baseStats, rawUnitStatsFromStats(distributedSocketBonus))
			distributedSocketBonusObjectiveDelta = unitStatsFromStats(distributedSocketBonus, weights)
		}
		variableSocketIdxs := make([]int, 0, len(socketColors))
		for socketIdx, socketColor := range socketColors {
			if socketColor == proto.GemColor_GemColorMeta {
				continue
			}
			gemChoices := []reforgeChoice{{slot: slot, gems: []reforgeGemChoice{{socketIdx: socketIdx, gemID: 0}}, socketChoice: true, socketIdx: socketIdx}}
			forEachGemOptionForSocket(gemOptions, socketColor, forceSocketBonus, func(gemOption reforgeGemOption) {
				if !gemEligibleForSocket(gemOption.color, socketColor) {
					return
				}
				choice := reforgeChoice{
					slot:           slot,
					gems:           []reforgeGemChoice{{socketIdx: socketIdx, gemID: gemOption.id, rawDelta: gemOption.rawDelta}},
					socketChoice:   true,
					socketIdx:      socketIdx,
					socketMatches:  gemMatchesSocket(gemOption.color, socketColor),
					objectiveDelta: gemOption.objectiveDelta,
				}
				if forceSocketBonus && choice.socketMatches {
					choice.forcedBonusDelta = distributedSocketBonusDelta
					choice.objectiveDelta = addUnitStats(choice.objectiveDelta, distributedSocketBonusObjectiveDelta)
				}
				choice.score = dotUnitStats(choice.objectiveDelta, weights)
				if gemOption.isJewelcrafting {
					choice.jewelcraftingGems = 1
				}
				if gemOption.unique {
					choice.uniqueGemIDs = []int32{gemOption.id}
				}
				gemChoices = append(gemChoices, choice)
			})
			if len(gemChoices) > 1 {
				allSlots = append(allSlots, reforgeSlotChoices{slot: slot, choices: gemChoices})
				variableSocketIdxs = append(variableSocketIdxs, socketIdx)
			}
		}
		if !forceSocketBonus && len(variableSocketIdxs) > 0 && hasSocketBonus(*item) {
			socketBonusDelta := resolveStatDelta(statDeps, baseStats, rawUnitStatsFromStats(item.SocketBonus))
			socketBonusObjectiveDelta := unitStatsFromStats(item.SocketBonus, weights)
			allSlots = append(allSlots, reforgeSlotChoices{slot: slot, choices: []reforgeChoice{
				{slot: slot, socketBonus: true},
				{slot: slot, socketBonus: true, bonusSocketIdxs: variableSocketIdxs, delta: socketBonusDelta, objectiveDelta: socketBonusObjectiveDelta, score: dotUnitStats(socketBonusObjectiveDelta, weights)},
			}})
		}
	}

	computeChoiceDeltas(baseGear, allSlots, statDeps, baseStats)

	slices.SortFunc(allSlots, func(a, b reforgeSlotChoices) int {
		return cmp.Compare(maxChoiceScore(b.choices), maxChoiceScore(a.choices))
	})
	return allSlots, nil
}

func allowedReforgeDestinationStats(weights *proto.UnitStats) map[stats.Stat]bool {
	allowedStats := map[stats.Stat]bool{}
	if weights == nil {
		return allowedStats
	}
	for statIdx, weight := range weights.GetStats() {
		if weight != 0 {
			allowedStats[stats.Stat(statIdx)] = true
		}
	}
	return allowedStats
}

func computeChoiceDeltas(baseGear *proto.EquipmentSpec, allSlots []reforgeSlotChoices, sdm *stats.StatDependencyManager, baseStats core.UnitStats) {
	for slotIdx := range allSlots {
		for choiceIdx := range allSlots[slotIdx].choices {
			choice := &allSlots[slotIdx].choices[choiceIdx]
			if choice.socketBonus || len(choice.gems) == 0 || (len(choice.gems) == 1 && choice.gems[0].gemID == 0) {
				continue
			}
			choice.delta = resolveStatDelta(sdm, baseStats, rawChoiceDelta(choice))
			if !isEmptyUnitStats(choice.forcedBonusDelta) {
				choice.delta = addUnitStats(choice.delta, choice.forcedBonusDelta)
			}
		}
	}
}

func rawChoiceDelta(choice *reforgeChoice) core.UnitStats {
	rawDelta := core.NewUnitStats()
	for _, gemChoice := range choice.gems {
		if gemChoice.gemID == 0 {
			continue
		}
		rawDelta = addUnitStats(rawDelta, gemChoice.rawDelta)
	}
	return rawDelta
}

func equipmentSpecWithChoice(baseEquipment core.Equipment, choice reforgeChoice) *proto.EquipmentSpec {
	gear := baseEquipment
	if int(choice.slot) >= 0 && int(choice.slot) < int(core.NumItemSlots) {
		gear[choice.slot].Gems = slices.Clone(gear[choice.slot].Gems)
	}
	gearEditor := &reforgeGearEditor{gear: &gear}
	gearEditor.applyChoice(choice)
	return gearEditor.equipment()
}

func equipmentSpecWithChoices(baseEquipment core.Equipment, choices []reforgeChoice) *proto.EquipmentSpec {
	gear := baseEquipment
	clonedGemSlots := [core.NumItemSlots]bool{}
	for _, choice := range choices {
		if int(choice.slot) < 0 || int(choice.slot) >= int(core.NumItemSlots) || clonedGemSlots[choice.slot] {
			continue
		}
		gear[choice.slot].Gems = slices.Clone(gear[choice.slot].Gems)
		clonedGemSlots[choice.slot] = true
	}
	gearEditor := &reforgeGearEditor{gear: &gear}
	gearEditor.applyChoices(choices)
	return gearEditor.equipment()
}

func hasSocketBonus(item core.Item) bool {
	for _, value := range item.SocketBonus {
		if value != 0 {
			return true
		}
	}
	return false
}

func shouldForceSocketBonus(item core.Item, socketColors []proto.GemColor, gemOptions map[proto.GemColor][]reforgeGemOption, weights core.UnitStats, hardCaps []reforgeHardCap, softCaps []reforgeSoftCap) bool {
	if !hasSocketBonus(item) {
		return false
	}
	normalization := socketBonusNormalization(socketColors)
	if normalization == 0 {
		return false
	}
	distributedSocketBonus := item.SocketBonus.Multiply(1 / float64(normalization))
	socketBonusDelta := unitStatsFromStats(distributedSocketBonus, weights)
	if isEmptyUnitStats(socketBonusDelta) {
		return false
	}

	matchedDelta := core.NewUnitStats()
	unmatchedDelta := core.NewUnitStats()
	for _, socketColor := range socketColors {
		if socketColor != proto.GemColor_GemColorRed && socketColor != proto.GemColor_GemColorBlue && socketColor != proto.GemColor_GemColorYellow && socketColor != proto.GemColor_GemColorPrismatic {
			break
		}

		matchedOptions := gemOptions[socketColor]
		unmatchedOptions := gemOptions[proto.GemColor_GemColorPrismatic]
		if len(matchedOptions) == 0 || len(unmatchedOptions) == 0 {
			return false
		}

		matchedDelta = addUnitStats(matchedDelta, matchedOptions[0].objectiveDelta)
		matchedDelta = addUnitStats(matchedDelta, socketBonusDelta)

		unmatchedDelta = addUnitStats(unmatchedDelta, unmatchedOptions[0].objectiveDelta)
	}

	if dotUnitStats(matchedDelta, weights) > dotUnitStats(unmatchedDelta, weights) && (normalization > 1 || (includesStatWithCap(socketBonusDelta, hardCaps, softCaps) && !includesCappedStat(socketBonusDelta, hardCaps))) {
		return true
	}
	return false
}

func socketBonusNormalization(socketColors []proto.GemColor) int {
	normalization := len(socketColors)
	if normalization == 0 {
		return 1
	}
	if normalization > 1 && socketColors[0] == proto.GemColor_GemColorMeta {
		normalization--
	}
	return normalization
}

func includesStatWithCap(delta core.UnitStats, hardCaps []reforgeHardCap, softCaps []reforgeSoftCap) bool {
	for _, hardCap := range hardCaps {
		if getUnitStat(delta, hardCap.unitStat) != 0 {
			return true
		}
	}
	for _, softCap := range softCaps {
		if getUnitStat(delta, softCap.unitStat) != 0 {
			return true
		}
	}
	return false
}

func includesCappedStat(delta core.UnitStats, hardCaps []reforgeHardCap) bool {
	for _, hardCap := range hardCaps {
		if hardCap.undershoot && getUnitStat(delta, hardCap.unitStat) != 0 {
			return true
		}
	}
	return false
}

func gemMatchesSocket(gemColor proto.GemColor, socketColor proto.GemColor) bool {
	if gemColor == socketColor {
		return true
	}
	switch socketColor {
	case proto.GemColor_GemColorBlue:
		return gemColor == proto.GemColor_GemColorPurple || gemColor == proto.GemColor_GemColorGreen || gemColor == proto.GemColor_GemColorPrismatic
	case proto.GemColor_GemColorRed:
		return gemColor == proto.GemColor_GemColorPurple || gemColor == proto.GemColor_GemColorOrange || gemColor == proto.GemColor_GemColorPrismatic
	case proto.GemColor_GemColorYellow:
		return gemColor == proto.GemColor_GemColorOrange || gemColor == proto.GemColor_GemColorGreen || gemColor == proto.GemColor_GemColorPrismatic
	case proto.GemColor_GemColorPrismatic:
		return gemColor == proto.GemColor_GemColorRed || gemColor == proto.GemColor_GemColorOrange || gemColor == proto.GemColor_GemColorYellow || gemColor == proto.GemColor_GemColorGreen || gemColor == proto.GemColor_GemColorBlue || gemColor == proto.GemColor_GemColorPurple
	default:
		return false
	}
}

func rawUnitStatsFromStats(statValues stats.Stats) core.UnitStats {
	unitStats := core.NewUnitStats()
	for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
		amount := statValues[statIdx]
		if amount == 0 {
			continue
		}
		unitStats.Stats[statIdx] += amount
	}
	return unitStats
}

func unitStatsFromStats(statValues stats.Stats, weights core.UnitStats) core.UnitStats {
	unitStats := core.NewUnitStats()
	for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
		amount := statValues[statIdx]
		if amount == 0 {
			continue
		}
		stat := stats.Stat(statIdx)
		if weights.Stats[statIdx] != 0 {
			unitStats.Stats[statIdx] += amount
			continue
		}
		switch stat {
		case stats.MeleeHitRating:
			if getUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeHitPercent)) != 0 {
				unitStats = addPseudoStat(unitStats, proto.PseudoStat_PseudoStatMeleeHitPercent, amount/core.PhysicalHitRatingPerHitPercent)
			}
			if getUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent)) != 0 {
				unitStats = addPseudoStat(unitStats, proto.PseudoStat_PseudoStatRangedHitPercent, amount/core.PhysicalHitRatingPerHitPercent)
			}
		case stats.SpellHitRating:
			if getUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSpellHitPercent)) != 0 {
				unitStats = addPseudoStat(unitStats, proto.PseudoStat_PseudoStatSpellHitPercent, amount/core.SpellHitRatingPerHitPercent)
			}
			for _, schoolHit := range []proto.PseudoStat{
				proto.PseudoStat_PseudoStatSchoolHitPercentArcane,
				proto.PseudoStat_PseudoStatSchoolHitPercentFire,
				proto.PseudoStat_PseudoStatSchoolHitPercentFrost,
				proto.PseudoStat_PseudoStatSchoolHitPercentHoly,
				proto.PseudoStat_PseudoStatSchoolHitPercentNature,
				proto.PseudoStat_PseudoStatSchoolHitPercentShadow,
			} {
				if getUnitStat(weights, stats.UnitStatFromPseudoStat(schoolHit)) != 0 {
					unitStats = addPseudoStat(unitStats, schoolHit, amount/core.SpellHitRatingPerHitPercent)
				}
			}
		case stats.MeleeCritRating:
			if getUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeCritPercent)) != 0 {
				unitStats = addPseudoStat(unitStats, proto.PseudoStat_PseudoStatMeleeCritPercent, amount/core.PhysicalCritRatingPerCritPercent)
			}
			if getUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedCritPercent)) != 0 {
				unitStats = addPseudoStat(unitStats, proto.PseudoStat_PseudoStatRangedCritPercent, amount/core.PhysicalCritRatingPerCritPercent)
			}
		case stats.SpellCritRating:
			if getUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSpellCritPercent)) != 0 {
				unitStats = addPseudoStat(unitStats, proto.PseudoStat_PseudoStatSpellCritPercent, amount/core.SpellCritRatingPerCritPercent)
			}
		case stats.MeleeHasteRating:
			if getUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeHastePercent)) != 0 {
				unitStats = addPseudoStat(unitStats, proto.PseudoStat_PseudoStatMeleeHastePercent, amount/core.PhysicalHasteRatingPerHastePercent)
			}
			if getUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedHastePercent)) != 0 {
				unitStats = addPseudoStat(unitStats, proto.PseudoStat_PseudoStatRangedHastePercent, amount/core.PhysicalHasteRatingPerHastePercent)
			}
		case stats.SpellHasteRating:
			if getUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSpellHastePercent)) != 0 {
				unitStats = addPseudoStat(unitStats, proto.PseudoStat_PseudoStatSpellHastePercent, amount/core.SpellHasteRatingPerHastePercent)
			}
		}
	}
	return unitStats
}

func addPseudoStat(unitStats core.UnitStats, pseudoStat proto.PseudoStat, value float64) core.UnitStats {
	unitStat := stats.UnitStatFromPseudoStat(pseudoStat)
	return setUnitStat(unitStats, unitStat, getUnitStat(unitStats, unitStat)+value)
}

func maxChoiceScore(choices []reforgeChoice) float64 {
	best := math.Inf(-1)
	for _, choice := range choices {
		best = math.Max(best, choice.score)
	}
	return best
}
