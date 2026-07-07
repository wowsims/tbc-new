package reforgeoptimizer

import (
	"slices"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// hasteRatingSpeedMultiplierPairs maps each (haste rating stat, haste% pseudo-stat) to its
// speed multiplier pseudo-stat for the analytical haste delta calculation.
// Δhaste% = speedMult × ΔHasteRating / HasteRatingPerHastePercent
var hasteRatingSpeedMultiplierPairs = [3]struct {
	hasteRatingStat  stats.Stat
	hastePS          proto.PseudoStat
	speedMultPS      proto.PseudoStat
	hasteRatingConst float64
}{
	{stats.MeleeHasteRating, proto.PseudoStat_PseudoStatMeleeHastePercent, proto.PseudoStat_PseudoStatMeleeSpeedMultiplier, core.PhysicalHasteRatingPerHastePercent},
	{stats.MeleeHasteRating, proto.PseudoStat_PseudoStatRangedHastePercent, proto.PseudoStat_PseudoStatRangedSpeedMultiplier, core.PhysicalHasteRatingPerHastePercent},
	{stats.SpellHasteRating, proto.PseudoStat_PseudoStatSpellHastePercent, proto.PseudoStat_PseudoStatCastSpeedMultiplier, core.SpellHasteRatingPerHastePercent},
}

// resolveStatDelta applies the character's stat dependency graph to delta, resolving
// conversions such as HitRating→Hit%, CritRating→Crit%, Agility→PhysicalCritPercent.
// It also mirrors the resolved Stats values back to their corresponding PseudoStats
// so that LP constraint evaluation (which reads PseudoStats for hit/crit/haste caps)
// sees the correct contribution.
//
// Haste% is multiplicative with a speed multiplier that is not captured by the dep
// manager. We read it from baseStats.PseudoStats (populated by GetPseudoStatsProto):
//
//	Δhaste% = speedMult × ΔHasteRating / HasteRatingPerHastePercent
func resolveStatDelta(sdm *stats.StatDependencyManager, baseStats core.UnitStats, delta core.UnitStats) core.UnitStats {
	if isEmptyUnitStats(delta) {
		return delta
	}
	delta.Stats = sdm.ApplyStatDependencies(delta.Stats)

	// Mirror dual-stored stats from Stats (updated by SDM — e.g. HitRating→Hit%,
	// CritRating→Crit%, Agility→PhysicalCritPercent) back to their PseudoStat indices.
	delta = setUnitStat(delta, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeHitPercent), delta.Stats[stats.PhysicalHitPercent])
	delta = setUnitStat(delta, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSpellHitPercent), delta.Stats[stats.SpellHitPercent])
	delta = setUnitStat(delta, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent), delta.Stats[stats.PhysicalHitPercent]+delta.Stats[stats.RangedHitPercent])
	spellHitDelta := delta.Stats[stats.SpellHitPercent]
	for _, schoolHitPS := range []proto.PseudoStat{
		proto.PseudoStat_PseudoStatSchoolHitPercentArcane,
		proto.PseudoStat_PseudoStatSchoolHitPercentFire,
		proto.PseudoStat_PseudoStatSchoolHitPercentFrost,
		proto.PseudoStat_PseudoStatSchoolHitPercentHoly,
		proto.PseudoStat_PseudoStatSchoolHitPercentNature,
		proto.PseudoStat_PseudoStatSchoolHitPercentShadow,
	} {
		delta = setUnitStat(delta, stats.UnitStatFromPseudoStat(schoolHitPS), spellHitDelta)
	}
	delta = setUnitStat(delta, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeCritPercent), delta.Stats[stats.PhysicalCritPercent])
	delta = setUnitStat(delta, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedCritPercent), delta.Stats[stats.PhysicalCritPercent]+delta.Stats[stats.RangedCritPercent])
	delta = setUnitStat(delta, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSpellCritPercent), delta.Stats[stats.SpellCritPercent])
	delta = setUnitStat(delta, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatBlockPercent), delta.Stats[stats.BlockPercent])

	// Haste% pseudo-stats: read speed multipliers from baseStats.PseudoStats, which
	// GetPseudoStatsProto populates as MeleeSpeedMultiplier×AttackSpeedMultiplier etc.
	for _, p := range hasteRatingSpeedMultiplierPairs {
		if hasteRatingDelta := delta.Stats[p.hasteRatingStat]; hasteRatingDelta != 0 {
			speedMult := getUnitStat(baseStats, stats.UnitStatFromPseudoStat(p.speedMultPS))
			delta = setUnitStat(delta, stats.UnitStatFromPseudoStat(p.hastePS), speedMult*hasteRatingDelta/p.hasteRatingConst)
		}
	}

	return delta
}

func protoToCoreUnitStats(protoStats *proto.UnitStats) core.UnitStats {
	if protoStats == nil {
		return core.NewUnitStats()
	}
	return core.UnitStats{
		Stats:       stats.FromUnitStatsProto(protoStats),
		PseudoStats: slices.Clone(protoStats.PseudoStats),
	}
}

func addUnitStats(unitStats core.UnitStats, other core.UnitStats) core.UnitStats {
	result := unitStats
	result.Stats = unitStats.Stats.Add(other.Stats)
	maxLen := max(len(unitStats.PseudoStats), len(other.PseudoStats))
	result.PseudoStats = make([]float64, maxLen)
	copy(result.PseudoStats, unitStats.PseudoStats)
	for idx, value := range other.PseudoStats {
		result.PseudoStats[idx] += value
	}
	return result
}

func subtractUnitStats(unitStats core.UnitStats, other core.UnitStats) core.UnitStats {
	result := unitStats
	result.Stats = unitStats.Stats.Subtract(other.Stats)
	maxLen := max(len(unitStats.PseudoStats), len(other.PseudoStats))
	result.PseudoStats = make([]float64, maxLen)
	copy(result.PseudoStats, unitStats.PseudoStats)
	for idx, value := range other.PseudoStats {
		result.PseudoStats[idx] -= value
	}
	return result
}

func dotUnitStats(unitStats core.UnitStats, weights core.UnitStats) float64 {
	score := 0.0
	for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
		score += unitStats.Stats[statIdx] * weights.Stats[statIdx]
	}
	for idx, value := range unitStats.PseudoStats {
		if idx < len(weights.PseudoStats) {
			score += value * weights.PseudoStats[idx]
		}
	}
	return score
}

func getUnitStat(unitStats core.UnitStats, unitStat stats.UnitStat) float64 {
	if unitStat.IsStat() {
		return unitStats.Stats[unitStat.StatIdx()]
	}
	pseudoStatIdx := int(unitStat.PseudoStatIdx())
	if pseudoStatIdx >= len(unitStats.PseudoStats) {
		return 0
	}
	return unitStats.PseudoStats[pseudoStatIdx]
}

func setUnitStat(unitStats core.UnitStats, unitStat stats.UnitStat, value float64) core.UnitStats {
	if unitStat.IsStat() {
		unitStats.Stats[unitStat.StatIdx()] = value
		return unitStats
	}
	pseudoStatIdx := int(unitStat.PseudoStatIdx())
	for len(unitStats.PseudoStats) <= pseudoStatIdx {
		unitStats.PseudoStats = append(unitStats.PseudoStats, 0)
	}
	unitStats.PseudoStats[pseudoStatIdx] = value
	return unitStats
}

func isEmptyUnitStats(unitStats core.UnitStats) bool {
	for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
		if unitStats.Stats[statIdx] != 0 {
			return false
		}
	}
	for _, value := range unitStats.PseudoStats {
		if value != 0 {
			return false
		}
	}
	return true
}
