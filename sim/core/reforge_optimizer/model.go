package reforgeoptimizer

import (
	"fmt"
	"sort"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// model.go builds the LP model (variables + constraints) for the gem/socket-bonus optimizer.
// It is a port of the reference reforger's model layer: buildYalpsVariables, buildGemOptions,
// applyReforgeStat and buildYalpsConstraints.
//
// Coefficient space: the reference keeps ONE coefficient map per variable — the same map feeds
// both the objective score and the cap constraints. lpVariables carries two spaces (byName for
// caps, objByName for the objective) to allow the split later; until then every variable stores
// the SAME map in both, which reproduces the single-space behaviour exactly.

// gemData is a candidate gem for a socket color, with its precomputed coefficient map (which
// includes the "score" key produced by updateReforgeScores).
type gemData struct {
	gem      *proto.ReforgeGemOption
	isJC     bool
	isUnique bool
	coeffs   map[string]float64
}

// scoreCoeffKey is the objective key holding each variable's EP score.
const scoreCoeffKey = "score"

// computeCoeffScore sums weight*value over the coefficient keys that name a stat or pseudo-stat.
// Structural keys (socket keys, SocketBonusLink_*, UniqueGem_*, GemColor_*, and "score" itself)
// are not stat names and contribute nothing.
func computeCoeffScore(coeffs map[string]float64, weights core.UnitStats) float64 {
	score := 0.0
	for key, value := range coeffs {
		if unitStat, ok := unitStatFromCoeffKey(key); ok {
			score += getUnitStat(weights, unitStat) * value
		}
	}
	return score
}

// updateReforgeScores recomputes every variable's "score" coefficient against the given weights
// and returns the updated variable set. The weights are expected to already carry post-cap values
// for any stat constrained to its cap in an earlier pass.
func updateReforgeScores(variables *lpVariables, weights core.UnitStats) *lpVariables {
	updated := newLPVariables()
	variables.each(func(name string, coeffs map[string]float64) {
		updatedCoeffs := make(map[string]float64, len(coeffs)+1)
		for key, value := range coeffs {
			updatedCoeffs[key] = value
		}
		updatedCoeffs[scoreCoeffKey] = computeCoeffScore(coeffs, weights)
		updated.set(name, updatedCoeffs)
		// Single coefficient space (see file header): the objective reads the same map.
		updated.setObj(name, updatedCoeffs)
	})
	return updated
}

// scoreCoeffMap scores a standalone coefficient map, mirroring the reference's trick of running a
// single-entry variable map through updateReforgeScores.
func scoreCoeffMap(coeffs map[string]float64, weights core.UnitStats) map[string]float64 {
	scored := make(map[string]float64, len(coeffs)+1)
	for key, value := range coeffs {
		scored[key] = value
	}
	scored[scoreCoeffKey] = computeCoeffScore(coeffs, weights)
	return scored
}

// applyReforgeStat adds a stat amount to a coefficient map, applying the racial stat modifiers and
// resolving the stat onto its child pseudo-stats when the root stat itself carries no EP.
func applyReforgeStat(coeffs map[string]float64, stat stats.Stat, amount float64, preCapEPs core.UnitStats, race proto.Race) {
	if stat == stats.Spirit && race == proto.Race_RaceHuman {
		amount *= 1.1
	}
	if stat == stats.Intellect && race == proto.Race_RaceGnome {
		amount *= 1.05
	}

	// If the pre-cap EP for the root stat is non-zero, apply the root stat directly and don't look
	// for any children.
	if preCapEPs.Stats[stat] != 0 {
		coeffs[statCoeffKey(proto.Stat(stat))] += amount
		return
	}

	for _, child := range childPseudoStats(stat) {
		if getUnitStat(preCapEPs, stats.UnitStatFromPseudoStat(child)) == 0 {
			continue
		}
		// ratingPerPseudoStatPercent encodes the per-parent conversion (including the dual
		// Defense/Resilience parents of ReducedCritTakenPercent), so dividing by it reproduces the
		// reference's convertStatToChildPseudoStat exactly.
		if ratingPerPercent := ratingPerPseudoStatPercent(child, stat); ratingPerPercent != 0 {
			coeffs[pseudoStatCoeffKey(child)] += amount / ratingPerPercent
		}
	}
}

// applyReforgeStats applies every non-zero entry of a stats vector through applyReforgeStat.
func applyReforgeStats(coeffs map[string]float64, statValues stats.Stats, preCapEPs core.UnitStats, race proto.Race) {
	for statIdx, value := range statValues {
		if value != 0 {
			applyReforgeStat(coeffs, stats.Stat(statIdx), value, preCapEPs, race)
		}
	}
}

// applyPositiveReforgeStats applies only the strictly positive entries of a stats vector. This
// mirrors the reference's getBuffedStats(), which filters a stats vector down to its positive
// entries (it applies no buffs or stat dependencies despite the name) and is how socket-bonus
// stats are read.
func applyPositiveReforgeStats(coeffs map[string]float64, statValues stats.Stats, preCapEPs core.UnitStats, race proto.Race) {
	for statIdx, value := range statValues {
		if value > 0 {
			applyReforgeStat(coeffs, stats.Stat(statIdx), value, preCapEPs, race)
		}
	}
}

// ---------------------------------------------------------------------------
// Gem pool
// ---------------------------------------------------------------------------

// epRelevantStats returns the stats this spec is willing to gem for. The request carries the spec's
// configured list, which deliberately includes stats weighted at 0 (Stamina, spell penetration) so
// those gems stay selectable — that is why it cannot be derived from the EP weights.
//
// When the list is absent (an older client), fall back to deriving it: a stat counts when it, or
// any of its child pseudo-stats, carries a non-zero pre-cap EP. Either way attack power and ranged
// attack power imply each other, matching the reference.
func epRelevantStats(preCapEPs core.UnitStats, settings *proto.ReforgeSettings) map[stats.Stat]bool {
	result := make(map[stats.Stat]bool)
	if configured := settings.GetEpStats(); len(configured) > 0 {
		for _, stat := range configured {
			result[stats.Stat(stat)] = true
		}
	} else {
		for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
			stat := stats.Stat(statIdx)
			if preCapEPs.Stats[statIdx] != 0 {
				result[stat] = true
				continue
			}
			for _, child := range childPseudoStats(stat) {
				if getUnitStat(preCapEPs, stats.UnitStatFromPseudoStat(child)) != 0 {
					result[stat] = true
					break
				}
			}
		}
	}
	if result[stats.AttackPower] && !result[stats.RangedAttackPower] {
		result[stats.RangedAttackPower] = true
	} else if result[stats.RangedAttackPower] && !result[stats.AttackPower] {
		result[stats.AttackPower] = true
	}
	return result
}

// gemStatIsAllowed reports whether a gem stat keeps the gem eligible. A stat outside the spec's EP
// list is tolerated when it is Stamina on a multi-stat gem (or on a tank spec) — so a gem whose
// second stat is Stamina is kept rather than discarded — or when it is healing power on a spec
// that values spell damage.
func gemStatIsAllowed(stat stats.Stat, statCount int, epStats map[stats.Stat]bool, isTank bool) bool {
	if epStats[stat] {
		return true
	}
	if stat == stats.Stamina && (isTank || statCount > 1) {
		return true
	}
	if stat == stats.HealingPower && epStats[stats.SpellDamage] {
		return true
	}
	return false
}

// buildGemOptions builds the per-socket-color candidate gem lists, sorted by descending pre-cap EP
// and pruned once an uncapped normal gem (and an uncapped Jewelcrafting gem) has been found.
func buildGemOptions(
	gemOptions []*proto.ReforgeGemOption,
	player *proto.Player,
	preCapEPs core.UnitStats,
	reforgeCaps core.UnitStats,
	softCaps []*reforgeSoftCap,
	settings *proto.ReforgeSettings,
	isTank bool,
) map[proto.GemColor][]gemData {
	gemsToInclude := make(map[proto.GemColor][]gemData)
	hasJC := playerHasProfession(player, proto.Profession_Jewelcrafting)
	epStats := epRelevantStats(preCapEPs, settings)
	race := player.GetRace()

	for _, socketColor := range []proto.GemColor{
		proto.GemColor_GemColorPrismatic,
		proto.GemColor_GemColorRed,
		proto.GemColor_GemColorBlue,
		proto.GemColor_GemColorYellow,
	} {
		var filtered []gemData

		for _, gem := range gemOptions {
			isJC := gem.GetRequiredProfession() == proto.Profession_Jewelcrafting
			statCount := 0
			for _, value := range gem.GetStats() {
				if value > 0 {
					statCount++
				}
			}

			if (settings.GetDisableUniqueGems() && gem.GetUnique() && !isJC) ||
				(isJC && !hasJC) ||
				!gemMatchesSocket(gem.GetColor(), socketColor) ||
				statCount == 0 ||
				gem.GetPhase() > settings.GetMaxGemPhase() ||
				gem.GetQuality() > settings.GetMaxGemQuality() {
				continue
			}

			allStatsValid := true
			coeffs := make(map[string]float64)
			for statIdx, statValue := range gem.GetStats() {
				if statValue == 0 {
					continue
				}
				stat := stats.Stat(statIdx)
				if !gemStatIsAllowed(stat, statCount, epStats, isTank) {
					allStatsValid = false
					break
				}
				applyReforgeStat(coeffs, stat, statValue, preCapEPs, race)
			}
			if !allStatsValid {
				continue
			}

			filtered = append(filtered, gemData{
				gem:      gem,
				isJC:     isJC,
				isUnique: gem.GetUnique(),
				coeffs:   scoreCoeffMap(coeffs, preCapEPs),
			})
		}

		// Sort from highest to lowest pre-cap EP (stable, matching the reference's sort).
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].coeffs[scoreCoeffKey] > filtered[j].coeffs[scoreCoeffKey]
		})

		var included []gemData
		foundUncappedJCGem := false
		foundUncappedNormalGem := false
		for _, candidate := range filtered {
			cappedStatKeys := getCappedStatKeys(candidate.coeffs, reforgeCaps, softCaps)

			if (!candidate.isJC || !foundUncappedJCGem) && (len(cappedStatKeys) == 0 || !foundUncappedNormalGem) {
				included = append(included, candidate)
			}

			if len(cappedStatKeys) == 0 {
				if candidate.isJC {
					foundUncappedJCGem = true
				} else {
					foundUncappedNormalGem = true
				}
			}
		}

		gemsToInclude[socketColor] = included
	}

	return gemsToInclude
}

// ---------------------------------------------------------------------------
// Variables
// ---------------------------------------------------------------------------

// jewelcraftingGemKey counts Jewelcrafting-only gems, capped at maxJewelcraftingGems. This limit
// is TBC-specific (the reference reforger has no such constraint).
const jewelcraftingGemKey = "JewelcraftingGem"
const maxJewelcraftingGems = 2

// isGemmableSocketColor reports whether a socket takes an ordinary (non-meta) gem.
func isGemmableSocketColor(socketColor proto.GemColor) bool {
	switch socketColor {
	case proto.GemColor_GemColorRed, proto.GemColor_GemColorBlue, proto.GemColor_GemColorYellow, proto.GemColor_GemColorPrismatic:
		return true
	default:
		return false
	}
}

// addMetaGemColorCoefficients records how a gem contributes to the equipped meta gem's colour
// requirements: one count per socket colour it satisfies, plus the greater/lesser comparison term.
func addMetaGemColorCoefficients(coeffs map[string]float64, gemColor proto.GemColor, compareGreater, compareLesser proto.GemColor) {
	red, yellow, blue := metaGemActivationColorContribution(gemColor)
	colors := make([]proto.GemColor, 0, 3)
	if red != 0 {
		colors = append(colors, proto.GemColor_GemColorRed)
	}
	if yellow != 0 {
		colors = append(colors, proto.GemColor_GemColorYellow)
	}
	if blue != 0 {
		colors = append(colors, proto.GemColor_GemColorBlue)
	}

	for _, color := range colors {
		coeffs[gemColorKey(color)]++
	}

	if compareGreater == proto.GemColor_GemColorUnknown || compareLesser == proto.GemColor_GemColorUnknown {
		return
	}
	compareValue := 0.0
	for _, color := range colors {
		if color == compareGreater {
			compareValue++
		}
		if color == compareLesser {
			compareValue--
		}
	}
	if compareValue != 0 {
		coeffs[gemColorCompareKey(compareGreater, compareLesser)] += compareValue
	}
}

// scaleStats returns a copy of statValues scaled by factor.
func scaleStats(statValues stats.Stats, factor float64) stats.Stats {
	scaled := statValues
	for idx := range scaled {
		scaled[idx] *= factor
	}
	return scaled
}

// buildYalpsVariables builds one binary variable per (socket, candidate gem) plus an optional
// all-or-nothing socket-bonus variable per item.
func buildYalpsVariables(
	equipment core.Equipment,
	gemOptions []*proto.ReforgeGemOption,
	player *proto.Player,
	preCapEPs core.UnitStats,
	reforgeCaps core.UnitStats,
	softCaps []*reforgeSoftCap,
	undershootCaps core.UnitStats,
	settings *proto.ReforgeSettings,
	isTank bool,
) *lpVariables {
	variables := newLPVariables()
	gemsToInclude := buildGemOptions(gemOptions, player, preCapEPs, reforgeCaps, softCaps, settings, isTank)
	race := player.GetRace()
	frozen := frozenItemSlots(settings)

	compareGreater := proto.GemColor_GemColorUnknown
	compareLesser := proto.GemColor_GemColorUnknown
	if metaConstraint, ok := equippedMetaGemConstraint(equipment); ok {
		compareGreater = metaConstraint.compareColorGreater
		compareLesser = metaConstraint.compareColorLesser
	}

	for slotIdx, item := range equipment {
		slot := proto.ItemSlot(slotIdx)
		socketColors := currentSocketColors(item)
		if item.ID == 0 || len(socketColors) == 0 || frozen[slot] {
			continue
		}

		socketBonusNormalization := len(socketColors)
		if socketBonusNormalization == 0 {
			socketBonusNormalization = 1
		}
		// A meta socket cannot carry the socket bonus, so it does not dilute the per-socket share.
		if socketBonusNormalization > 1 && socketColors[0] == proto.GemColor_GemColorMeta {
			socketBonusNormalization--
		}
		distributedSocketBonus := scaleStats(item.SocketBonus, 1.0/float64(socketBonusNormalization))

		// Decide up front whether matching the socket bonus is obviously correct, which lets the
		// model drop the off-colour gem candidates for this item.
		forceSocketBonus := false
		socketBonusAsCoeff := make(map[string]float64)
		for statIdx, value := range distributedSocketBonus {
			if value <= 0 {
				continue
			}
			stat := stats.Stat(statIdx)
			if getUnitStat(undershootCaps, stats.UnitStatFromStat(stat)) != 0 {
				continue
			}
			undershot := false
			for _, child := range childPseudoStats(stat) {
				if getUnitStat(undershootCaps, stats.UnitStatFromPseudoStat(child)) != 0 {
					undershot = true
					break
				}
			}
			if undershot {
				continue
			}
			applyReforgeStat(socketBonusAsCoeff, stat, value, preCapEPs, race)
		}

		if len(socketBonusAsCoeff) > 0 {
			bonusHelpsUncappedStat := includesStatWithCap(socketBonusAsCoeff, reforgeCaps, softCaps) &&
				!includesCappedStat(socketBonusAsCoeff, reforgeCaps, softCaps)

			if bonusHelpsUncappedStat && socketBonusNormalization > 1 {
				forceSocketBonus = true
			}

			matched := make(map[string]float64)
			unmatched := make(map[string]float64)
			for _, socketColor := range socketColors {
				if !isGemmableSocketColor(socketColor) {
					continue
				}
				if best := gemsToInclude[socketColor]; len(best) > 0 {
					for key, value := range best[0].coeffs {
						matched[key] += value
					}
				}
				for key, value := range socketBonusAsCoeff {
					matched[key] += value
				}
				if bestUnmatched := gemsToInclude[proto.GemColor_GemColorPrismatic]; len(bestUnmatched) > 0 {
					for key, value := range bestUnmatched[0].coeffs {
						unmatched[key] += value
					}
				}
			}

			if computeCoeffScore(matched, preCapEPs) > computeCoeffScore(unmatched, preCapEPs) &&
				(socketBonusNormalization > 1 || bonusHelpsUncappedStat) {
				forceSocketBonus = true
			}
		}

		for socketIdx, socketColor := range socketColors {
			var gemColorKeys []proto.GemColor
			switch socketColor {
			case proto.GemColor_GemColorPrismatic:
				gemColorKeys = []proto.GemColor{socketColor}
			case proto.GemColor_GemColorRed, proto.GemColor_GemColorBlue, proto.GemColor_GemColorYellow:
				gemColorKeys = []proto.GemColor{socketColor}
				if !forceSocketBonus {
					gemColorKeys = append(gemColorKeys, proto.GemColor_GemColorPrismatic)
				}
			default:
				continue
			}

			constraintKey := socketConstraintKey(slot, socketIdx)
			for _, gemColorKey := range gemColorKeys {
				for _, candidate := range gemsToInclude[gemColorKey] {
					variableKey := fmt.Sprintf("%s_%d", constraintKey, candidate.gem.GetId())
					coeffs := make(map[string]float64, len(candidate.coeffs)+4)
					for key, value := range candidate.coeffs {
						coeffs[key] = value
					}
					coeffs[constraintKey] = 1
					addMetaGemColorCoefficients(coeffs, candidate.gem.GetColor(), compareGreater, compareLesser)

					if gemMatchesSocket(candidate.gem.GetColor(), socketColor) {
						if forceSocketBonus {
							applyPositiveReforgeStats(coeffs, distributedSocketBonus, preCapEPs, race)
						} else {
							coeffs[socketBonusLinkKey(slot, socketIdx)] = -1
						}
					}

					if candidate.isUnique {
						coeffs[uniqueGemKey(candidate.gem.GetId())] = 1
					}
					if candidate.isJC {
						coeffs[jewelcraftingGemKey] = 1
					}

					variables.set(variableKey, coeffs)
					// Single coefficient space (see file header).
					variables.setObj(variableKey, coeffs)
				}
			}
		}

		if !forceSocketBonus && socketBonusNormalization > 0 {
			socketBonusCoeffs := make(map[string]float64)
			applyPositiveReforgeStats(socketBonusCoeffs, item.SocketBonus, preCapEPs, race)
			for socketIdx, socketColor := range socketColors {
				if isGemmableSocketColor(socketColor) {
					socketBonusCoeffs[socketBonusLinkKey(slot, socketIdx)] = 1
				}
			}
			socketBonusKey := socketBonusVariableKey(slot)
			variables.set(socketBonusKey, socketBonusCoeffs)
			variables.setObj(socketBonusKey, socketBonusCoeffs)
		}
	}

	return variables
}

// ---------------------------------------------------------------------------
// Constraints
// ---------------------------------------------------------------------------

// socketConstraintKey is the per-socket one-hot key, also carried as a coefficient by every gem
// variable for that socket.
func socketConstraintKey(slot proto.ItemSlot, socketIdx int) string {
	return fmt.Sprintf("%d_%d", int(slot), socketIdx)
}

func socketBonusLinkKey(slot proto.ItemSlot, socketIdx int) string {
	return fmt.Sprintf("SocketBonusLink_%d_%d", int(slot), socketIdx)
}

func socketBonusVariableKey(slot proto.ItemSlot) string {
	return fmt.Sprintf("SocketBonus_%d", int(slot))
}

func uniqueGemKey(gemID int32) string {
	return fmt.Sprintf("UniqueGem_%d", gemID)
}

func gemColorKey(color proto.GemColor) string {
	return fmt.Sprintf("GemColor_%d", int(color))
}

func gemColorCompareKey(greater, lesser proto.GemColor) string {
	return fmt.Sprintf("GemColorCompare_%d_%d", int(greater), int(lesser))
}

// buildYalpsConstraints builds the meta-gem activation rows plus the per-slot and per-socket
// one-hot rows. The UniqueGem_* (<=1) and SocketBonusLink_* (<=0) rows are added by the caller
// once the variables are known, mirroring the reference.
func buildYalpsConstraints(equipment core.Equipment, frozenSlots map[proto.ItemSlot]bool) *lpConstraints {
	constraints := newLPConstraints()

	if metaConstraint, ok := equippedMetaGemConstraint(equipment); ok {
		fixedCounts := metaGemColorCounts(equipment)

		if metaConstraint.compareColorGreater != proto.GemColor_GemColorUnknown && metaConstraint.compareColorLesser != proto.GemColor_GemColorUnknown {
			fixedGreater := metaGemCountForColor(fixedCounts, metaConstraint.compareColorGreater)
			fixedLesser := metaGemCountForColor(fixedCounts, metaConstraint.compareColorLesser)
			if remainingCompare := 1 - (fixedGreater - fixedLesser); remainingCompare > 0 {
				constraints.set(gemColorCompareKey(metaConstraint.compareColorGreater, metaConstraint.compareColorLesser), greaterEq(float64(remainingCompare)))
			}
		}

		for _, colorMin := range []struct {
			color proto.GemColor
			min   int
		}{
			{proto.GemColor_GemColorBlue, metaConstraint.minBlue},
			{proto.GemColor_GemColorRed, metaConstraint.minRed},
			{proto.GemColor_GemColorYellow, metaConstraint.minYellow},
		} {
			if remaining := colorMin.min - metaGemCountForColor(fixedCounts, colorMin.color); remaining > 0 {
				constraints.set(gemColorKey(colorMin.color), greaterEq(float64(remaining)))
			}
		}
	}

	for slot, item := range equipment {
		itemSlot := proto.ItemSlot(slot)
		constraints.set(slotCoeffKey(itemSlot), lessEq(1))
		if frozenSlots[itemSlot] {
			continue
		}
		for socketIdx := range currentSocketColors(item) {
			constraints.set(socketConstraintKey(itemSlot, socketIdx), lessEq(1))
		}
	}

	return constraints
}

// addStructuralConstraints adds the UniqueGem_* (<=1) and SocketBonusLink_* (<=0) rows implied by
// the built variables, matching the reference's post-pass over the variable coefficients.
func addStructuralConstraints(variables *lpVariables, constraints *lpConstraints) {
	variables.each(func(_ string, coeffs map[string]float64) {
		for key := range coeffs {
			if len(key) > len("UniqueGem_") && key[:len("UniqueGem_")] == "UniqueGem_" && !constraints.has(key) {
				constraints.set(key, lessEq(1))
			}
			if len(key) > len("SocketBonusLink_") && key[:len("SocketBonusLink_")] == "SocketBonusLink_" && !constraints.has(key) {
				constraints.set(key, lessEq(0))
			}
			// TBC-specific: at most two Jewelcrafting-only gems may be equipped.
			if key == jewelcraftingGemKey && !constraints.has(key) {
				constraints.set(key, lessEq(maxJewelcraftingGems))
			}
		}
	})
}
