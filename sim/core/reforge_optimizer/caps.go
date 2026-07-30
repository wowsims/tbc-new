package reforgeoptimizer

import (
	"slices"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// reforgeSoftCap represents a soft cap or threshold expressed in reforge-relative (gap-to-cap)
// breakpoints. breakpoints/postCapEPs are mutated in place across solver passes as breakpoints
// are consumed.
type reforgeSoftCap struct {
	unitStat    stats.UnitStat
	breakpoints []float64
	postCapEPs  []float64
	capType     proto.StatCapType
}

// buildDebuffUnitStats returns the pseudo-stat contributions from raid debuffs that the
// UI adds to the character-sheet display. These debuffs (e.g. Improved Faerie Fire, Improved
// Seal of the Crusader) lower the target's effective miss/crit chance rather than raising
// the player's stats, so they are absent from FinalStats. Soft-cap breakpoints configured
// by the user are based on the UI display values (which include the debuff contribution),
// so we add these offsets to the base stats before computing the gap to each cap.
func buildDebuffUnitStats(raid *proto.Raid) core.UnitStats {
	debuffs := raid.GetDebuffs()
	result := core.NewUnitStats()
	if debuffs.GetFaerieFire() == proto.TristateEffect_TristateEffectImproved {
		result = setUnitStat(result, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeHitPercent), 3)
		result = setUnitStat(result, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent), 3)
	}
	if debuffs.GetImprovedSealOfTheCrusader() != proto.TristateEffect_TristateEffectMissing {
		result = setUnitStat(result, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeCritPercent), 3)
		result = setUnitStat(result, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedCritPercent), 3)
		result = setUnitStat(result, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSpellCritPercent), 3)
	}
	return result
}

// ---------------------------------------------------------------------------
// LP-path cap computation (mirrors the reference solver). These run alongside the legacy
// buildReforgeHardCaps/buildReforgeSoftCaps/validateReforgeWeights until the MIP path is removed.
// ---------------------------------------------------------------------------

// computeReforgeSoftCaps converts each configured soft cap's absolute breakpoints into
// gap-to-cap deltas. For TypeThreshold caps the breakpoints are reversed (largest gap evaluated
// first) and every post-cap EP is set to the FIRST configured value, which is interpreted as the
// residual stat value just after passing a threshold discontinuity.
func computeReforgeSoftCaps(baseStats core.UnitStats, configs []*proto.StatCapConfig) []*reforgeSoftCap {
	result := make([]*reforgeSoftCap, 0, len(configs))
	for _, config := range configs {
		unitStat, ok := unitStatFromUIStat(config.GetUnitStat())
		if !ok {
			continue
		}

		weights := slices.Clone(config.GetPostCap_EPs())
		relativeBreakpoints := make([]float64, 0, len(config.GetBreakpoints()))
		for _, breakpoint := range config.GetBreakpoints() {
			relativeBreakpoints = append(relativeBreakpoints, computeGapToCap(baseStats, unitStat, breakpoint))
		}

		if config.GetCapType() == proto.StatCapType_TypeThreshold {
			slices.Reverse(relativeBreakpoints)
			first := 0.0
			if len(weights) > 0 {
				first = weights[0]
			}
			weights = make([]float64, len(relativeBreakpoints))
			for i := range weights {
				weights[i] = first
			}
		}

		result = append(result, &reforgeSoftCap{
			unitStat:    unitStat,
			breakpoints: relativeBreakpoints,
			capType:     config.GetCapType(),
			postCapEPs:  weights,
		})
	}
	return result
}

// isSchoolHitChildOfSpellHit reports whether a child pseudo-stat is one of the per-school spell-hit
// percentages. A non-zero EP on one of these does NOT suppress the SpellHitRating conversion: a
// spec can weight both the pure rating and its school percentages, and the two are summed.
func isSchoolHitChildOfSpellHit(parent stats.Stat, child proto.PseudoStat) bool {
	if parent != stats.SpellHitRating {
		return false
	}
	switch child {
	case proto.PseudoStat_PseudoStatSchoolHitPercentArcane,
		proto.PseudoStat_PseudoStatSchoolHitPercentFire,
		proto.PseudoStat_PseudoStatSchoolHitPercentFrost,
		proto.PseudoStat_PseudoStatSchoolHitPercentHoly,
		proto.PseudoStat_PseudoStatSchoolHitPercentNature,
		proto.PseudoStat_PseudoStatSchoolHitPercentShadow:
		return true
	default:
		return false
	}
}

// checkWeights routes each pure-rating stat's EP into the percent pseudo-stat that actually carries
// a cap, so the solver values the stat in the same units the caps are expressed in.
//
// When a child already carries EP the parent rating is simply zeroed (it would double count) — the
// per-school spell-hit children are exempt, since a spec may weight both. Conversions ACCUMULATE
// onto any existing child EP because several parents can share one child (both DefenseRating and
// ResilienceRating feed ReducedCritTakenPercent).
func checkWeights(weights core.UnitStats, reforgeCaps core.UnitStats, reforgeSoftCaps []*reforgeSoftCap) core.UnitStats {
	validated := weights
	for _, parent := range []stats.Stat{
		stats.MeleeHitRating,
		stats.SpellHitRating,
		stats.MeleeCritRating,
		stats.SpellCritRating,
		stats.MeleeHasteRating,
		stats.SpellHasteRating,
		stats.DefenseRating,
		stats.ResilienceRating,
	} {
		children := childPseudoStats(parent)
		if len(children) == 0 {
			continue
		}

		childHasWeight := false
		for _, child := range children {
			if isSchoolHitChildOfSpellHit(parent, child) {
				continue
			}
			if getUnitStat(validated, stats.UnitStatFromPseudoStat(child)) != 0 {
				childHasWeight = true
				break
			}
		}
		if childHasWeight {
			validated.Stats[parent] = 0
			continue
		}

		for _, child := range children {
			if !pseudoStatHasCap(child, reforgeCaps, reforgeSoftCaps) {
				continue
			}
			if ratingPerPercent := ratingPerPseudoStatPercent(child, parent); ratingPerPercent != 0 {
				childUnitStat := stats.UnitStatFromPseudoStat(child)
				rescaled := validated.Stats[parent] * ratingPerPercent
				validated = setUnitStat(validated, childUnitStat, getUnitStat(validated, childUnitStat)+rescaled)
				validated.Stats[parent] = 0
			}
			break
		}
	}
	return validated
}

// ---------------------------------------------------------------------------
// Cap-detection helpers. reforgeCaps is a gap-to-cap vector (the remaining rating room before
// each hard cap).
// ---------------------------------------------------------------------------

func pseudoStatHasCap(pseudoStat proto.PseudoStat, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	unitStat := stats.UnitStatFromPseudoStat(pseudoStat)
	if getUnitStat(reforgeCaps, unitStat) != 0 {
		return true
	}
	for _, softCap := range softCaps {
		if softCap.unitStat == unitStat {
			return true
		}
	}
	return false
}

func statHasCap(stat stats.Stat, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	unitStat := stats.UnitStatFromStat(stat)
	if getUnitStat(reforgeCaps, unitStat) != 0 {
		return true
	}
	for _, softCap := range softCaps {
		if softCap.unitStat == unitStat {
			return true
		}
	}
	return false
}

func pseudoStatIsCapped(pseudoStat proto.PseudoStat, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	unitStat := stats.UnitStatFromPseudoStat(pseudoStat)
	return getUnitStat(reforgeCaps, unitStat) < 0
}

func statIsCapped(stat stats.Stat, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	unitStat := stats.UnitStatFromStat(stat)
	return getUnitStat(reforgeCaps, unitStat) < 0
}

// includesStatWithCap reports whether any key in the coefficient map names a stat that has a
// configured cap.
func includesStatWithCap(coeffs map[string]float64, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	for key := range coeffs {
		if unitStat, ok := unitStatFromCoeffKey(key); ok {
			if unitStat.IsPseudoStat() {
				if pseudoStatHasCap(proto.PseudoStat(unitStat.PseudoStatIdx()), reforgeCaps, softCaps) {
					return true
				}
			} else if statHasCap(stats.Stat(unitStat.StatIdx()), reforgeCaps, softCaps) {
				return true
			}
		}
	}
	return false
}

// includesCappedStat reports whether any key in the coefficient map names a stat that is already
// capped.
func includesCappedStat(coeffs map[string]float64, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	for key := range coeffs {
		if unitStat, ok := unitStatFromCoeffKey(key); ok {
			if unitStat.IsPseudoStat() {
				if pseudoStatIsCapped(proto.PseudoStat(unitStat.PseudoStatIdx()), reforgeCaps, softCaps) {
					return true
				}
			} else if statIsCapped(stats.Stat(unitStat.StatIdx()), reforgeCaps, softCaps) {
				return true
			}
		}
	}
	return false
}

// getCappedStatKeys returns the coefficient keys whose stat has a configured cap (hard or soft).
func getCappedStatKeys(coeffs map[string]float64, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) []string {
	var keys []string
	for key := range coeffs {
		unitStat, ok := unitStatFromCoeffKey(key)
		if !ok {
			continue
		}
		if unitStat.IsPseudoStat() {
			if pseudoStatHasCap(proto.PseudoStat(unitStat.PseudoStatIdx()), reforgeCaps, softCaps) {
				keys = append(keys, key)
			}
		} else if statHasCap(stats.Stat(unitStat.StatIdx()), reforgeCaps, softCaps) {
			keys = append(keys, key)
		}
	}
	return keys
}
