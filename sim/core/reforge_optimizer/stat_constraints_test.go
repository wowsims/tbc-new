//go:build with_db

package reforgeoptimizer

import (
	"math"
	"slices"
	"testing"

	"github.com/wowsims/tbc/sim"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/simsignals"
	"github.com/wowsims/tbc/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

func statConstraint(stat proto.Stat, op proto.BulkStatConstraintOp, value float64) *proto.BulkStatConstraint {
	return &proto.BulkStatConstraint{UnitStat: &proto.BulkStatConstraint_Stat{Stat: stat}, Op: op, Value: value}
}

func optimizedFinalStats(t *testing.T, request *proto.ReforgeOptimizeRequest) (core.UnitStats, float64) {
	t.Helper()
	result := Optimize(request)
	if result.GetError() != nil {
		t.Fatalf("Optimize failed: %s", result.GetError().GetMessage())
	}
	return protoToCoreUnitStats(result.GetOptimizedPlayerStats().GetFinalStats()), result.GetScore()
}

// Stat constraints are rows of the model: the solver picks gems that keep the constrained
// stat on the right side of the threshold, reports infeasibility when no gem choice can,
// and decides constraints on stats the gems cannot move on the base stats.
func TestStatConstraintsInModel(t *testing.T) {
	sim.RegisterAll()
	request := loadPreset(t, "gem-pool-wide.test.json")
	request.Debug = false

	optimizer, err := newReforgeOptimizer(request, simsignals.CreateSignals())
	if err != nil {
		t.Fatalf("newReforgeOptimizer: %v", err)
	}
	base := optimizer.capBaseStats
	unconstrained, unconstrainedScore := optimizedFinalStats(t, request)

	// The stat the unconstrained gems raise the most is the one to constrain.
	var target stats.Stat
	var targetDelta float64
	for statIdx := range unconstrained.Stats {
		if delta := unconstrained.Stats[statIdx] - base.Stats[statIdx]; delta > targetDelta {
			target, targetDelta = stats.Stat(statIdx), delta
		}
	}
	if targetDelta <= 0 {
		t.Fatal("unconstrained gems raise no stat; fixture unsuitable")
	}
	targetStat := proto.Stat(target)
	t.Logf("constraining %s: base %.1f, unconstrained %.1f", target.StatName(), base.Stats[target], unconstrained.Stats[target])

	// An upper bound below what the free solve reached: satisfiable with other gems, at a cost.
	bound := base.Stats[target] + targetDelta/2
	request.StatConstraints = []*proto.BulkStatConstraint{statConstraint(targetStat, proto.BulkStatConstraintOp_BulkStatConstraintOpLessThanOrEqual, bound)}
	constrained, constrainedScore := optimizedFinalStats(t, request)
	if constrained.Stats[target] > bound+1e-6 {
		t.Fatalf("%s = %.1f exceeds the constraint bound %.1f", target.StatName(), constrained.Stats[target], bound)
	}
	if constrainedScore > unconstrainedScore+1e-9 {
		t.Fatalf("constrained score %.3f beats the unconstrained optimum %.3f", constrainedScore, unconstrainedScore)
	}

	// A lower bound above anything the gems can reach: infeasible, flagged as such.
	request.StatConstraints = []*proto.BulkStatConstraint{statConstraint(targetStat, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, unconstrained.Stats[target]+10000)}
	result := Optimize(request)
	if !result.GetInfeasibleStatConstraints() || result.GetError() == nil {
		t.Fatalf("unreachable constraint should be reported infeasible, got %+v", result)
	}

	// A lower bound the free solve already clears costs nothing: the optimum scores the same. The
	// gems may differ between equally scored solutions, because gems that move a constrained stat
	// are kept in the pool, which can change the solver's tie-breaking.
	request.StatConstraints = []*proto.BulkStatConstraint{statConstraint(targetStat, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan, base.Stats[target])}
	loose, looseScore := optimizedFinalStats(t, request)
	if math.Abs(looseScore-unconstrainedScore) > 1e-9 || loose.Stats[target] <= base.Stats[target] {
		t.Fatalf("a constraint the optimum satisfies must cost nothing: %.1f/%.3f vs %.1f/%.3f", loose.Stats[target], looseScore, unconstrained.Stats[target], unconstrainedScore)
	}

	// Frost resistance is carried by no gem here, so it is decided on the base stats.
	frostRes := base.Stats[stats.FrostResistance]
	request.StatConstraints = []*proto.BulkStatConstraint{statConstraint(proto.Stat_StatFrostResistance, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, frostRes)}
	if result := Optimize(request); result.GetError() != nil {
		t.Fatalf("a base-satisfied constraint on an unmovable stat must solve: %s", result.GetError().GetMessage())
	}
	request.StatConstraints = []*proto.BulkStatConstraint{statConstraint(proto.Stat_StatFrostResistance, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThan, frostRes)}
	if result := Optimize(request); !result.GetInfeasibleStatConstraints() {
		t.Fatalf("a base-failed constraint on an unmovable stat must be infeasible, got %+v", result)
	}
}

func pseudoStatConstraint(pseudoStat proto.PseudoStat, op proto.BulkStatConstraintOp, value float64) *proto.BulkStatConstraint {
	return &proto.BulkStatConstraint{UnitStat: &proto.BulkStatConstraint_PseudoStat{PseudoStat: pseudoStat}, Op: op, Value: value}
}

var tankPercentStats = []proto.PseudoStat{
	proto.PseudoStat_PseudoStatReducedCritTakenPercent,
	proto.PseudoStat_PseudoStatDodgePercent,
	proto.PseudoStat_PseudoStatParryPercent,
}

// The model's stat accounting has to agree with the character sheet for crit reduction, dodge and
// parry, the tank stats the sheet derives from ratings rather than from stat dependencies. Checked
// against the sim itself by adding the ratings as bonus stats: exact for dodge, parry, agility and
// resilience, and within one defense point for defense, which the sheet floors.
func TestResolveStatDeltaTankStatsMatchSheet(t *testing.T) {
	sim.RegisterAll()
	request := loadPreset(t, "tank-caps.test.json") // Protection Warrior: can parry, agility gives dodge.
	optimizer, err := newReforgeOptimizer(request, simsignals.CreateSignals())
	if err != nil {
		t.Fatalf("newReforgeOptimizer: %v", err)
	}
	sheet := func(bonus stats.Stats) core.UnitStats {
		raid := googleProto.Clone(optimizer.baseRaidProto).(*proto.Raid)
		raid.Parties[0].Players[0].BonusStats = &proto.UnitStats{Stats: bonus[:], PseudoStats: make([]float64, stats.PseudoStatsLen)}
		result := computeReforgeStats(&proto.ComputeStatsRequest{Raid: raid})
		if result.ErrorResult != "" {
			t.Fatalf("ComputeStats: %s", result.ErrorResult)
		}
		return protoToCoreUnitStats(result.RaidStats.Parties[0].Players[0].FinalStats)
	}
	base := sheet(stats.Stats{})

	for _, tc := range []struct {
		name      string
		ratings   map[stats.Stat]float64
		tolerance float64
	}{
		{"dodge, parry, agility and resilience", map[stats.Stat]float64{stats.DodgeRating: 40, stats.ParryRating: 25, stats.Agility: 30, stats.ResilienceRating: 20}, 1e-9},
		{"defense", map[stats.Stat]float64{stats.DefenseRating: 20}, core.MissDodgeParryBlockCritChancePerDefense + 1e-9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var bonus stats.Stats
			for stat, value := range tc.ratings {
				bonus[stat] = value
			}
			withBonus := sheet(bonus)
			modelled := resolveStatDelta(optimizer.statDeps, optimizer.baseStats, rawUnitStatsFromStats(bonus))
			for _, pseudoStat := range tankPercentStats {
				unitStat := stats.UnitStatFromPseudoStat(pseudoStat)
				sheetDelta := getUnitStat(withBonus, unitStat) - getUnitStat(base, unitStat)
				if sheetDelta == 0 {
					t.Fatalf("%s: the ratings should move the sheet's value; the case is unsuitable", pseudoStat)
				}
				if got := getUnitStat(modelled, unitStat); math.Abs(got-sheetDelta) > tc.tolerance {
					t.Fatalf("%s: model credits %.4f, the sheet moves %.4f", pseudoStat, got, sheetDelta)
				}
			}
		})
	}
}

// Constraints on crit reduction, dodge and parry are met by choosing gems. They used to be judged
// on the gem-stripped base alone, which made every raised threshold infeasible.
func TestStatConstraintsOnTankStats(t *testing.T) {
	sim.RegisterAll()
	unconstrained, _ := optimizedFinalStats(t, loadPreset(t, "tank-caps.test.json"))

	for _, pseudoStat := range tankPercentStats {
		unitStat := stats.UnitStatFromPseudoStat(pseudoStat)
		t.Run(pseudoStat.String(), func(t *testing.T) {
			// Above what the unconstrained gems reach, so meeting it takes different gems.
			target := getUnitStat(unconstrained, unitStat) + 0.5
			request := loadPreset(t, "tank-caps.test.json")
			request.StatConstraints = []*proto.BulkStatConstraint{pseudoStatConstraint(pseudoStat, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, target)}
			finalStats, _ := optimizedFinalStats(t, request)
			if got := getUnitStat(finalStats, unitStat); got < target {
				t.Fatalf("the gems reach %.3f, below the constraint's %.3f", got, target)
			}

			// Far beyond anything the sockets can add: still infeasible.
			request.StatConstraints = []*proto.BulkStatConstraint{pseudoStatConstraint(pseudoStat, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, target+100)}
			if result := Optimize(request); !result.GetInfeasibleStatConstraints() {
				t.Fatalf("an unreachable threshold must be infeasible, got %+v", result.GetError())
			}
		})
	}

	// An upper bound: the gems stay under it.
	dodge := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatDodgePercent)
	ceiling := getUnitStat(unconstrained, dodge) - 0.1
	request := loadPreset(t, "tank-caps.test.json")
	request.StatConstraints = []*proto.BulkStatConstraint{pseudoStatConstraint(proto.PseudoStat_PseudoStatDodgePercent, proto.BulkStatConstraintOp_BulkStatConstraintOpLessThanOrEqual, ceiling)}
	finalStats, _ := optimizedFinalStats(t, request)
	if got := getUnitStat(finalStats, dodge); got > ceiling {
		t.Fatalf("dodge %.3f exceeds the constraint's %.3f", got, ceiling)
	}
}

// A model that is infeasible for a reason unrelated to the stat constraints must be reported as
// that reason, not as unmet constraints: the batch drops a candidate flagged as infeasible stat
// constraints, where any other optimizer failure falls back to the candidate's own gear.
func TestInfeasibleModelIsNotBlamedOnStatConstraints(t *testing.T) {
	sim.RegisterAll()
	baseRequest := loadPreset(t, "gem-pool-wide.test.json")
	optimizer, err := newReforgeOptimizer(baseRequest, simsignals.CreateSignals())
	if err != nil {
		t.Fatalf("newReforgeOptimizer: %v", err)
	}
	base := optimizer.capBaseStats

	// An upper-bound Stamina cap below what the gear has without gems: no gem choice can lower
	// Stamina, so the solver finds the caps impossible, constraints or not.
	impossibleCaps := func() *proto.ReforgeOptimizeRequest {
		request := loadPreset(t, "gem-pool-wide.test.json")
		caps := make([]float64, stats.ProtoStatsLen)
		caps[stats.Stamina] = base.Stats[stats.Stamina] - 10
		request.Settings.StatCaps = &proto.UnitStats{Stats: caps, PseudoStats: make([]float64, stats.PseudoStatsLen)}
		request.UndershootCaps = &proto.UnitStats{Stats: slices.Clone(caps), PseudoStats: make([]float64, stats.PseudoStatsLen)}
		return request
	}
	unconstrained := Optimize(impossibleCaps())
	if unconstrained.GetError() == nil || unconstrained.GetInfeasibleStatConstraints() {
		t.Fatalf("the impossible cap alone should fail as a cap error, got %+v", unconstrained)
	}

	for _, tc := range []struct {
		name       string
		constraint *proto.BulkStatConstraint
	}{
		// Decided on the base stats: it adds no row to the model at all.
		{"constraint without a row", statConstraint(proto.Stat_StatFireResistance, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, base.Stats[stats.FireResistance])},
		// A row, but one the gems meet trivially.
		{"constraint with a satisfiable row", statConstraint(proto.Stat_StatSpellDamage, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, base.Stats[stats.SpellDamage])},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := impossibleCaps()
			request.StatConstraints = []*proto.BulkStatConstraint{tc.constraint}
			result := Optimize(request)
			if result.GetInfeasibleStatConstraints() {
				t.Fatalf("the cap made the model infeasible, but it was blamed on the stat constraint: %s", result.GetError().GetMessage())
			}
			if result.GetError().GetMessage() != unconstrained.GetError().GetMessage() {
				t.Fatalf("got error %q, want the cap error %q", result.GetError().GetMessage(), unconstrained.GetError().GetMessage())
			}
		})
	}

	// And the reverse still holds: when the constraint is what cannot be met, it is blamed.
	request := loadPreset(t, "gem-pool-wide.test.json")
	request.StatConstraints = []*proto.BulkStatConstraint{statConstraint(proto.Stat_StatSpellDamage, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, base.Stats[stats.SpellDamage]+100000)}
	if result := Optimize(request); !result.GetInfeasibleStatConstraints() {
		t.Fatalf("an unreachable constraint must be reported as infeasible stat constraints, got %+v", result.GetError())
	}
}

// A stat constraint and a cap on the same stat are independent rows: the cap still pins the stat
// and zeroes its value once a solution passes it, and the constraint still holds. Cap refinement
// used to skip a stat that already had a constraint row, so hit kept its full value past the cap.
func TestStatConstraintAndHardCapOnSameStat(t *testing.T) {
	sim.RegisterAll()
	spellHit := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatSpellHitPercent)
	optimizer, err := newReforgeOptimizer(loadPreset(t, "gem-pool-wide.test.json"), simsignals.CreateSignals())
	if err != nil {
		t.Fatalf("newReforgeOptimizer: %v", err)
	}
	base := getUnitStat(optimizer.capBaseStats, spellHit)
	// Final stats leave the sheet's debuffs out; caps and constraints include them.
	sheetOffset := getUnitStat(optimizer.capBaseStats, spellHit) - getUnitStat(optimizer.baseStats, spellHit)
	sheet := func(finalStats core.UnitStats) float64 { return getUnitStat(finalStats, spellHit) + sheetOffset }

	// The free gems add about 1.6% spell hit; cap it well below that so refinement has to fire.
	capValue := base + 0.5
	withCap := func(undershoot bool) *proto.ReforgeOptimizeRequest {
		request := loadPreset(t, "gem-pool-wide.test.json")
		caps := &proto.UnitStats{Stats: make([]float64, stats.ProtoStatsLen), PseudoStats: make([]float64, stats.PseudoStatsLen)}
		caps.PseudoStats[proto.PseudoStat_PseudoStatSpellHitPercent] = capValue
		request.Settings.StatCaps = caps
		if undershoot {
			request.UndershootCaps = googleProto.Clone(caps).(*proto.UnitStats)
		}
		return request
	}
	capOnly, capOnlyScore := optimizedFinalStats(t, withCap(false))
	if sheet(capOnly) >= base+1.5 {
		t.Fatalf("the cap alone should stop the gems short of the free solve, got %.2f", sheet(capOnly))
	}

	// A constraint the cap already implies: the same solution as the cap alone.
	request := withCap(false)
	request.StatConstraints = []*proto.BulkStatConstraint{pseudoStatConstraint(proto.PseudoStat_PseudoStatSpellHitPercent, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, base+0.2)}
	constrained, constrainedScore := optimizedFinalStats(t, request)
	if math.Abs(constrainedScore-capOnlyScore) > 1e-9 || sheet(constrained) > sheet(capOnly)+1e-9 {
		t.Fatalf("with the constraint the cap no longer holds: hit %.3f score %.3f, cap alone hit %.3f score %.3f", sheet(constrained), constrainedScore, sheet(capOnly), capOnlyScore)
	}

	// An upper-bound (undershoot) cap: the ceiling still applies with a constraint present.
	request = withCap(true)
	request.StatConstraints = []*proto.BulkStatConstraint{pseudoStatConstraint(proto.PseudoStat_PseudoStatSpellHitPercent, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, base+0.2)}
	undershot, _ := optimizedFinalStats(t, request)
	if sheet(undershot) > capValue+1e-9 {
		t.Fatalf("hit %.3f overshoots the upper-bound cap %.3f", sheet(undershot), capValue)
	}
}

// Soft-cap refinement adds its own row for a stat; it used to replace the constraint's row, so an
// upper bound on a soft-capped stat was lost once the first breakpoint was passed.
func TestStatConstraintAndSoftCapOnSameStat(t *testing.T) {
	sim.RegisterAll()
	critReduction := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatReducedCritTakenPercent)
	// Protection Paladin with a soft cap on crit reduction at 5.6%, which the gear already exceeds.
	unconstrained, _ := optimizedFinalStats(t, loadPreset(t, "tank-soft-caps.test.json"))
	free := getUnitStat(unconstrained, critReduction)

	// A floor above the free solve: the gems must reach it.
	floor := free + 0.2
	request := loadPreset(t, "tank-soft-caps.test.json")
	request.StatConstraints = []*proto.BulkStatConstraint{pseudoStatConstraint(proto.PseudoStat_PseudoStatReducedCritTakenPercent, proto.BulkStatConstraintOp_BulkStatConstraintOpGreaterThanOrEqual, floor)}
	raised, _ := optimizedFinalStats(t, request)
	if got := getUnitStat(raised, critReduction); got < floor {
		t.Fatalf("crit reduction %.3f is below the constraint's %.3f", got, floor)
	}

	// A ceiling below the free solve: the gems must stay under it.
	ceiling := free - 0.04
	request = loadPreset(t, "tank-soft-caps.test.json")
	request.StatConstraints = []*proto.BulkStatConstraint{pseudoStatConstraint(proto.PseudoStat_PseudoStatReducedCritTakenPercent, proto.BulkStatConstraintOp_BulkStatConstraintOpLessThanOrEqual, ceiling)}
	lowered, _ := optimizedFinalStats(t, request)
	if got := getUnitStat(lowered, critReduction); got > ceiling {
		t.Fatalf("crit reduction %.3f exceeds the constraint's %.3f", got, ceiling)
	}
}
