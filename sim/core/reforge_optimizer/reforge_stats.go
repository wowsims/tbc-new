package reforgeoptimizer

import (
	"slices"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func protoToCoreUnitStats(protoStats *proto.UnitStats) core.UnitStats {
	if protoStats == nil {
		return core.NewUnitStats()
	}
	return core.UnitStats{
		Stats:       stats.FromUnitStatsProto(protoStats),
		PseudoStats: slices.Clone(protoStats.PseudoStats),
	}
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

// subtractUnitStats is used by the fixture parity test to diff expected against optimized stats.
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

// eachUnitStat invokes fn for every stat and pseudo-stat index in the vector.
func eachUnitStat(vec core.UnitStats, fn func(unitStat stats.UnitStat, value float64)) {
	for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
		fn(stats.UnitStatFromStat(stats.Stat(statIdx)), vec.Stats[statIdx])
	}
	for pseudoStatIdx := 0; pseudoStatIdx < int(stats.PseudoStatsLen); pseudoStatIdx++ {
		value := 0.0
		if pseudoStatIdx < len(vec.PseudoStats) {
			value = vec.PseudoStats[pseudoStatIdx]
		}
		fn(stats.UnitStatFromPseudoStat(proto.PseudoStat(pseudoStatIdx)), value)
	}
}

func childPseudoStats(parent stats.Stat) []proto.PseudoStat {
	switch parent {
	case stats.MeleeHitRating:
		return []proto.PseudoStat{proto.PseudoStat_PseudoStatMeleeHitPercent, proto.PseudoStat_PseudoStatRangedHitPercent}
	case stats.SpellHitRating:
		return []proto.PseudoStat{
			proto.PseudoStat_PseudoStatSpellHitPercent,
			proto.PseudoStat_PseudoStatSchoolHitPercentArcane,
			proto.PseudoStat_PseudoStatSchoolHitPercentFire,
			proto.PseudoStat_PseudoStatSchoolHitPercentFrost,
			proto.PseudoStat_PseudoStatSchoolHitPercentHoly,
			proto.PseudoStat_PseudoStatSchoolHitPercentNature,
			proto.PseudoStat_PseudoStatSchoolHitPercentShadow,
		}
	case stats.MeleeCritRating:
		return []proto.PseudoStat{proto.PseudoStat_PseudoStatMeleeCritPercent, proto.PseudoStat_PseudoStatRangedCritPercent}
	case stats.SpellCritRating:
		return []proto.PseudoStat{proto.PseudoStat_PseudoStatSpellCritPercent}
	case stats.MeleeHasteRating:
		return []proto.PseudoStat{proto.PseudoStat_PseudoStatMeleeHastePercent, proto.PseudoStat_PseudoStatRangedHastePercent}
	case stats.SpellHasteRating:
		return []proto.PseudoStat{proto.PseudoStat_PseudoStatSpellHastePercent}
	case stats.ResilienceRating, stats.DefenseRating:
		return []proto.PseudoStat{proto.PseudoStat_PseudoStatReducedCritTakenPercent}
	default:
		return nil
	}
}

func ratingPerPseudoStatPercent(pseudoStat proto.PseudoStat, parent stats.Stat) float64 {
	switch pseudoStat {
	case proto.PseudoStat_PseudoStatMeleeHitPercent:
		return core.PhysicalHitRatingPerHitPercent
	case proto.PseudoStat_PseudoStatRangedHitPercent:
		return core.PhysicalHitRatingPerHitPercent
	case proto.PseudoStat_PseudoStatSpellHitPercent:
		return core.SpellHitRatingPerHitPercent
	case proto.PseudoStat_PseudoStatSchoolHitPercentArcane, proto.PseudoStat_PseudoStatSchoolHitPercentFire, proto.PseudoStat_PseudoStatSchoolHitPercentFrost, proto.PseudoStat_PseudoStatSchoolHitPercentHoly, proto.PseudoStat_PseudoStatSchoolHitPercentNature, proto.PseudoStat_PseudoStatSchoolHitPercentShadow:
		return core.SpellHitRatingPerHitPercent
	case proto.PseudoStat_PseudoStatMeleeCritPercent:
		return core.PhysicalCritRatingPerCritPercent
	case proto.PseudoStat_PseudoStatRangedCritPercent:
		return core.PhysicalCritRatingPerCritPercent
	case proto.PseudoStat_PseudoStatSpellCritPercent:
		return core.SpellCritRatingPerCritPercent
	case proto.PseudoStat_PseudoStatMeleeHastePercent, proto.PseudoStat_PseudoStatRangedHastePercent:
		return core.PhysicalHasteRatingPerHastePercent
	case proto.PseudoStat_PseudoStatSpellHastePercent:
		return core.SpellHasteRatingPerHastePercent
	case proto.PseudoStat_PseudoStatReducedCritTakenPercent:
		if parent == stats.DefenseRating {
			return core.DefenseRatingPerDefenseLevel / core.MissDodgeParryBlockCritChancePerDefense
		}
		if parent == stats.ResilienceRating {
			return core.ResilienceRatingPerCritReductionChance
		}
		return 1
	default:
		return 1
	}
}

// computeGapToCap returns the remaining room from the current sheet value to cap. A gap of
// exactly 0 is returned as 1e-12 so the cap still registers as configured (a zero entry means
// "no cap").
func computeGapToCap(baseStats core.UnitStats, unitStat stats.UnitStat, cap float64) float64 {
	statDelta := cap - getUnitStat(baseStats, unitStat)
	if statDelta == 0 {
		return 1e-12
	}
	return statDelta
}

// computeStatCapsDelta returns the per-unit-stat gap-to-cap, but only for caps whose configured
// value is > 0 (a cap of 0 maps to 0 and means "no cap").
func computeStatCapsDelta(baseStats core.UnitStats, statCaps core.UnitStats) core.UnitStats {
	result := core.NewUnitStats()
	for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
		if cap := statCaps.Stats[statIdx]; cap > 0 {
			unitStat := stats.UnitStatFromStat(stats.Stat(statIdx))
			result = setUnitStat(result, unitStat, computeGapToCap(baseStats, unitStat, cap))
		}
	}
	for pseudoStatIdx := 0; pseudoStatIdx < int(stats.PseudoStatsLen); pseudoStatIdx++ {
		cap := 0.0
		if pseudoStatIdx < len(statCaps.PseudoStats) {
			cap = statCaps.PseudoStats[pseudoStatIdx]
		}
		if cap > 0 {
			unitStat := stats.UnitStatFromPseudoStat(proto.PseudoStat(pseudoStatIdx))
			result = setUnitStat(result, unitStat, computeGapToCap(baseStats, unitStat, cap))
		}
	}
	return result
}

func unitStatFromUIStat(uiStat *proto.UIStat) (stats.UnitStat, bool) {
	if uiStat == nil {
		return 0, false
	}
	switch unitStat := uiStat.UnitStat.(type) {
	case *proto.UIStat_Stat:
		return stats.UnitStatFromStat(stats.Stat(unitStat.Stat)), true
	case *proto.UIStat_PseudoStat:
		return stats.UnitStatFromPseudoStat(unitStat.PseudoStat), true
	default:
		return 0, false
	}
}

// ---------------------------------------------------------------------------
// LP coefficient-key scheme
// ---------------------------------------------------------------------------
//
// Coefficients are keyed by the proto enum NAME (e.g. "StatMeleeHitRating",
// "PseudoStatSpellHitPercent", "ItemSlotHead") so checkCaps can recover the stat from a key.
// Slot keys ("ItemSlot...") and special keys (SocketBonusLink_*, JewelcraftingGem, unique-gem
// IDs, score) are not stat names and parse to (_, false).

func statCoeffKey(stat proto.Stat) string {
	return proto.Stat_name[int32(stat)]
}

func pseudoStatCoeffKey(pseudoStat proto.PseudoStat) string {
	return proto.PseudoStat_name[int32(pseudoStat)]
}

func slotCoeffKey(slot proto.ItemSlot) string {
	return proto.ItemSlot_name[int32(slot)]
}

// unitStatFromCoeffKey recovers the stat from a coefficient key, or (_, false) for non-stat
// keys. PseudoStat and Stat names occupy disjoint namespaces so lookup order is irrelevant.
func unitStatFromCoeffKey(key string) (stats.UnitStat, bool) {
	if value, ok := proto.PseudoStat_value[key]; ok {
		return stats.UnitStatFromPseudoStat(proto.PseudoStat(value)), true
	}
	if value, ok := proto.Stat_value[key]; ok {
		return stats.UnitStatFromStat(stats.Stat(value)), true
	}
	return 0, false
}

// coeffKeyForUnitStat returns the coefficient/constraint key for a unit stat: its proto enum
// name.
func coeffKeyForUnitStat(unitStat stats.UnitStat) string {
	if unitStat.IsStat() {
		return statCoeffKey(proto.Stat(unitStat.StatIdx()))
	}
	return pseudoStatCoeffKey(proto.PseudoStat(unitStat.PseudoStatIdx()))
}
