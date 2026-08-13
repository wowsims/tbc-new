package reforgeoptimizer

import (
	"fmt"
	"maps"
	"sort"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// model.go builds the LP model (variables + constraints) for the gem/socket-bonus optimizer.
// It is a port of the reference reforger's model layer: buildYalpsVariables, buildGemOptions,
// applyReforgeStat and buildYalpsConstraints.
//
// Coefficient space: the reference keeps ONE coefficient map per variable — the same map feeds both
// the objective score and the cap constraints, so a stat only counts toward a cap when
// applyReforgeStat happened to resolve it onto that cap's pseudo-stat (which it does only for a root
// stat carrying no EP). This port splits the two spaces:
//   - objByName keeps the EP-calibrated applyReforgeStat output, so the objective stays exactly as
//     the reference calibrates it.
//   - byName holds the CAP coefficients, built by pushing the variable's RAW stat delta through the
//     full stat dependency graph (resolveCapCoeffs), plus the structural/constraint keys. Cap rows
//     and checkCaps therefore also count the dependencies applyReforgeStat never models at all —
//     Intellect -> SpellCrit%, Agility -> Crit%/Dodge% — instead of silently under-counting them.
//
// The racial multipliers (Human Spirit x1.1, Gnome Intellect x1.05) are registered in the stat
// dependency manager, so the cap space resolves the raw delta and must NOT pre-apply them; only the
// objective space applies them by hand (see applyReforgeStat).

// gemData is a candidate gem for a socket color, with its precomputed coefficient maps: coeffs is
// the objective space (including the "score" key produced by scoreCoeffMap) and capCoeffs the cap
// space. rawStats is kept so the force-socket-bonus case can re-resolve the cap space.
type gemData struct {
	gem       *proto.ReforgeGemOption
	isJC      bool
	isUnique  bool
	coeffs    map[string]float64
	rawStats  stats.Stats
	capCoeffs map[string]float64
}

var gemBuildSocketColors = []proto.GemColor{
	proto.GemColor_GemColorPrismatic,
	proto.GemColor_GemColorRed,
	proto.GemColor_GemColorBlue,
	proto.GemColor_GemColorYellow,
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
//
// The score is computed from the OBJECTIVE coefficients and written into the cap-space map, which is
// what buildLPObjective reads — so the objective stays exactly as calibrated while the constraint
// rows keep their stat-dependency-resolved coefficients.
func (o *reforgeOptimizer) updateReforgeScores(variables *lpVariables, weights core.UnitStats) *lpVariables {
	updated := newLPVariables()
	variables.each(func(name string, coeffs map[string]float64) {
		out := make(map[string]float64, len(coeffs)+1)
		for key, value := range coeffs {
			out[key] = value
		}
		objCoeffs := variables.getObj(name)
		out[scoreCoeffKey] = computeCoeffScore(objCoeffs, weights)
		updated.set(name, out)
		// Carry the objective coefficients forward so the cap-refinement recursion (which re-invokes
		// updateReforgeScores on this returned value) can re-score against the tightened weights
		// instead of collapsing every score to zero.
		updated.setObj(name, objCoeffs)
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
func (o *reforgeOptimizer) applyReforgeStat(coeffs map[string]float64, stat stats.Stat, amount float64, preCapEPs core.UnitStats) {
	race := o.player.GetRace()
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

// applyPositiveReforgeStats applies only the strictly positive entries of a stats vector. This
// mirrors the reference's getBuffedStats(), which filters a stats vector down to its positive
// entries (it applies no buffs or stat dependencies despite the name) and is how socket-bonus
// stats are read.
func (o *reforgeOptimizer) applyPositiveReforgeStats(coeffs map[string]float64, statValues stats.Stats, preCapEPs core.UnitStats) {
	for statIdx, value := range statValues {
		if value > 0 {
			o.applyReforgeStat(coeffs, stats.Stat(statIdx), value, preCapEPs)
		}
	}
}

// resolveCapCoeffs builds a variable's CAP-space coefficient map from a raw stat delta by resolving
// it through the full stat dependency graph and keying every nonzero resolved stat/pseudo-stat by its
// coefficient-key name. These feed the LP cap constraint rows and checkCaps, so every dependency
// (Intellect -> SpellCrit%, Agility -> PhysicalCrit%/Dodge%, the haste speed multiplier) counts
// toward the caps. The racial stat multipliers live in the dependency manager, so rawStats must be
// passed through unscaled.
func (o *reforgeOptimizer) resolveCapCoeffs(rawDelta stats.Stats) map[string]float64 {
	resolved := resolveStatDelta(o.statDeps, o.baseStats, rawUnitStatsFromStats(rawDelta))
	coeffs := map[string]float64{}
	eachUnitStat(resolved, func(unitStat stats.UnitStat, value float64) {
		if value != 0 {
			coeffs[coeffKeyForUnitStat(unitStat)] = value
		}
	})
	return coeffs
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
func (o *reforgeOptimizer) buildGemOptions(preCapEPs core.UnitStats, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) map[proto.GemColor][]gemData {
	gemsToInclude := make(map[proto.GemColor][]gemData)
	hasJC := playerHasProfession(o.player, proto.Profession_Jewelcrafting)
	epStats := epRelevantStats(preCapEPs, o.settings)

	for _, socketColor := range gemBuildSocketColors {
		var filtered []gemData

		for _, gem := range o.gemOptions {
			isJC := gem.GetRequiredProfession() == proto.Profession_Jewelcrafting
			statCount := 0
			for _, value := range gem.GetStats() {
				if value > 0 {
					statCount++
				}
			}

			if (o.settings.GetDisableUniqueGems() && gem.GetUnique() && !isJC) ||
				(isJC && !hasJC) ||
				!gemMatchesSocket(gem.GetColor(), socketColor) ||
				statCount == 0 ||
				gem.GetPhase() > o.settings.GetMaxGemPhase() ||
				gem.GetQuality() > o.settings.GetMaxGemQuality() {
				continue
			}

			allStatsValid := true
			coeffs := make(map[string]float64)
			for statIdx, statValue := range gem.GetStats() {
				if statValue == 0 {
					continue
				}
				stat := stats.Stat(statIdx)
				if !gemStatIsAllowed(stat, statCount, epStats, o.isTankSpec) {
					allStatsValid = false
					break
				}
				o.applyReforgeStat(coeffs, stat, statValue, preCapEPs)
			}
			if !allStatsValid {
				continue
			}

			// A gem's stats are fixed, so resolve its cap coefficients once here rather than per
			// socket variable.
			gemStats := stats.FromProtoArray(gem.GetStats())
			filtered = append(filtered, gemData{
				gem:       gem,
				isJC:      isJC,
				isUnique:  gem.GetUnique(),
				coeffs:    scoreCoeffMap(coeffs, preCapEPs),
				rawStats:  gemStats,
				capCoeffs: o.resolveCapCoeffs(gemStats),
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

// isColoredSocket reports whether a socket takes an ordinary (non-meta) gem.
func isColoredSocket(socketColor proto.GemColor) bool {
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
func (o *reforgeOptimizer) buildYalpsVariables(equipment core.Equipment, preCapEPs core.UnitStats, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) *lpVariables {
	variables := newLPVariables()
	gemsToInclude := o.buildGemOptions(preCapEPs, reforgeCaps, softCaps)
	frozen := frozenItemSlots(o.settings)

	// setVar stores a variable's two coefficient spaces: capCoeffs (stat-dependency-resolved stats
	// plus the structural/constraint keys) in byName, and objCoeffs (the EP-calibrated
	// applyReforgeStat output) in objByName. See the file header. applyReforgeStat emits no
	// non-stat keys here, so nothing needs carrying between the two spaces.
	setVar := func(key string, capCoeffs, objCoeffs map[string]float64) {
		variables.set(key, capCoeffs)
		variables.setObj(key, objCoeffs)
	}

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
			if getUnitStat(o.undershootCaps, stats.UnitStatFromStat(stat)) != 0 {
				continue
			}
			undershot := false
			for _, child := range childPseudoStats(stat) {
				if getUnitStat(o.undershootCaps, stats.UnitStatFromPseudoStat(child)) != 0 {
					undershot = true
					break
				}
			}
			if undershot {
				continue
			}
			o.applyReforgeStat(socketBonusAsCoeff, stat, value, preCapEPs)
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
				if !isColoredSocket(socketColor) {
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
					objCoeffs := make(map[string]float64, len(candidate.coeffs)+2)
					for key, value := range candidate.coeffs {
						objCoeffs[key] = value
					}

					rawStats := candidate.rawStats
					socketBonusAdded := false
					useSocketBonusLink := false
					if gemMatchesSocket(candidate.gem.GetColor(), socketColor) {
						if forceSocketBonus {
							o.applyPositiveReforgeStats(objCoeffs, distributedSocketBonus, preCapEPs)
							rawStats = rawStats.Add(distributedSocketBonus)
							socketBonusAdded = true
						} else {
							useSocketBonusLink = true
						}
					}

					// The gem's cap coefficients were resolved once in buildGemOptions; only the
					// force-socket-bonus case changes rawStats, so only it re-resolves.
					var capCoeffs map[string]float64
					if socketBonusAdded {
						capCoeffs = o.resolveCapCoeffs(rawStats)
					} else {
						capCoeffs = maps.Clone(candidate.capCoeffs)
					}
					capCoeffs[constraintKey] = 1
					addMetaGemColorCoefficients(capCoeffs, candidate.gem.GetColor(), compareGreater, compareLesser)
					if useSocketBonusLink {
						capCoeffs[socketBonusLinkKey(slot, socketIdx)] = -1
					}
					if candidate.isUnique {
						capCoeffs[uniqueGemKey(candidate.gem.GetId())] = 1
					}
					if candidate.isJC {
						capCoeffs[jewelcraftingGemKey] = 1
					}

					setVar(variableKey, capCoeffs, objCoeffs)
				}
			}
		}

		if !forceSocketBonus && socketBonusNormalization > 0 {
			objCoeffs := make(map[string]float64)
			o.applyPositiveReforgeStats(objCoeffs, item.SocketBonus, preCapEPs)
			capCoeffs := o.resolveCapCoeffs(item.SocketBonus)
			for socketIdx, socketColor := range socketColors {
				if isColoredSocket(socketColor) {
					capCoeffs[socketBonusLinkKey(slot, socketIdx)] = 1
				}
			}
			setVar(socketBonusVariableKey(slot), capCoeffs, objCoeffs)
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
func (o *reforgeOptimizer) buildYalpsConstraints(equipment core.Equipment) *lpConstraints {
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
		if o.frozenSlots[itemSlot] {
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
