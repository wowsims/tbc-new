package reforgeoptimizer

import (
	"fmt"
	"log"
	"strings"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func logRequestInput(requestID uint64, request *proto.ReforgeOptimizeRequest) {
	settings := request.GetSettings()
	softCaps := request.GetSoftCaps()
	log.Printf("[reforgeOptimize:%d] request", requestID)
	log.Printf("[reforgeOptimize:%d]   id=%q", requestID, request.GetRequestId())
	log.Printf("[reforgeOptimize:%d]   settings", requestID)
	log.Printf("[reforgeOptimize:%d]     useSoftCapBreakpoints=%t freezeItemSlots=%t", requestID, settings.GetUseSoftCapBreakpoints(), settings.GetFreezeItemSlots())
	log.Printf("[reforgeOptimize:%d]     frozenSlots=%s", requestID, formatItemSlots(settings.GetFrozenItemSlots()))
	log.Printf("[reforgeOptimize:%d]   inputs gemOptions=%d baselineItems=%d softCaps=%d", requestID, len(request.GetGemOptions()), baselineItemCount(request), len(softCaps))
	log.Printf("[reforgeOptimize:%d]   caps", requestID)
	logProtoUnitStats(requestID, "    hard", settings.GetStatCaps())
	logProtoUnitStats(requestID, "    undershoot", request.GetUndershootCaps())
	logProtoUnitStats(requestID, "    breakpoints", settings.GetBreakpointLimits())
	log.Printf("[reforgeOptimize:%d]   weights preCap", requestID)
	logProtoUnitStats(requestID, "    ", request.GetPreCapEpWeights())
	for idx, softCap := range softCaps {
		if idx == 0 {
			log.Printf("[reforgeOptimize:%d]   softCaps", requestID)
		}
		log.Printf("[reforgeOptimize:%d]     [%d] stat=%s", requestID, idx, formatUIStat(softCap.GetUnitStat()))
		log.Printf("[reforgeOptimize:%d]         type=%s", requestID, softCap.GetCapType().String())
		log.Printf("[reforgeOptimize:%d]         breakpoints=%s", requestID, formatFloat64Slice(softCap.GetBreakpoints()))
		log.Printf("[reforgeOptimize:%d]         postCapEPs=%s", requestID, formatFloat64Slice(softCap.GetPostCap_EPs()))
	}
}

func baselineItemCount(request *proto.ReforgeOptimizeRequest) int {
	if request.GetRaid() == nil || len(request.GetRaid().GetParties()) == 0 || len(request.GetRaid().GetParties()[0].GetPlayers()) == 0 {
		return 0
	}
	return len(request.GetRaid().GetParties()[0].GetPlayers()[0].GetEquipment().GetItems())
}

func logProtoUnitStats(requestID uint64, label string, unitStats *proto.UnitStats) {
	parts := protoUnitStatParts(unitStats)
	if len(parts) == 0 {
		log.Printf("[reforgeOptimize:%d] %s=none", requestID, label)
		return
	}
	if strings.TrimSpace(label) != "" {
		log.Printf("[reforgeOptimize:%d] %s", requestID, label)
		label = "      "
	}
	for _, part := range parts {
		log.Printf("[reforgeOptimize:%d] %s%s", requestID, label, part)
	}
}

func logOptimizedGearSummary(requestID uint64, equipment *proto.EquipmentSpec) {
	gems := 0
	for _, item := range equipment.GetItems() {
		for _, gemID := range item.GetGems() {
			if gemID != 0 {
				gems++
			}
		}
	}
	log.Printf("[reforgeOptimize:%d] optimized gear contains gems=%d", requestID, gems)
}

func unitStatName(unitStat stats.UnitStat) string {
	if unitStat.IsStat() {
		return stats.Stat(unitStat.StatIdx()).StatName()
	}
	return proto.PseudoStat(unitStat.PseudoStatIdx()).String()
}

func formatItemSlots(slots []proto.ItemSlot) string {
	if len(slots) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(slots))
	for _, slot := range slots {
		parts = append(parts, slot.String())
	}
	return formatStringList(parts)
}

func formatStringList(parts []string) string {
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ",")
}

func protoUnitStatParts(unitStats *proto.UnitStats) []string {
	if unitStats == nil {
		return nil
	}
	parts := make([]string, 0)
	for statIdx, value := range unitStats.GetStats() {
		if value == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%.3f", stats.Stat(statIdx).StatName(), value))
	}
	for pseudoStatIdx, value := range unitStats.GetPseudoStats() {
		if value == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%.3f", proto.PseudoStat(pseudoStatIdx).String(), value))
	}
	return parts
}

func formatUIStat(uiStat *proto.UIStat) string {
	unitStat, ok := unitStatFromUIStat(uiStat)
	if !ok {
		return "unknown"
	}
	return unitStatName(unitStat)
}

func formatFloat64Slice(values []float64) string {
	if len(values) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%.3f", value))
	}
	return strings.Join(parts, ",")
}
