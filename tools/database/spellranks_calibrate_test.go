package database

// Calibration gate for the spell-rank derivation rule.
//
// The rule that derives a rank's numbers has to keep reproducing what the sim registers, and every row
// it does not reproduce has to be classified as either "the rule is wrong" or "the hand row is wrong".
// knownResiduals below is that record: an unexplained residual fails the build.
//
// Run:  go test ./tools/database/ -run SpellRankCalibration
//
// Skips when tools/database/wowsims.db is absent - it is gitignored and only produced by `make db`
// from a local WoW install, so CI and a fresh clone legitimately have no client data.

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/druid"
	"github.com/wowsims/tbc/sim/mage"
	"github.com/wowsims/tbc/sim/paladin"
	"github.com/wowsims/tbc/sim/priest"
	"github.com/wowsims/tbc/sim/shaman"
)

const (
	classPaladin = 2
	classPriest  = 16
	classShaman  = 64
	classMage    = 128
	classDruid   = 1024
)

type calibFamily struct {
	Name     string
	File     string
	ClassBit int
	Table    shared.SpellRankTable
}

// The two shaman tables were inline anonymous literals until they were hoisted to package vars so this
// gate could read them.
var calibFamilies = []calibFamily{
	{"Consecration", "sim/paladin/consecration.go", classPaladin, paladin.ConsecrationRankMap},
	{"Hammer of Wrath", "sim/paladin/hammer_of_wrath.go", classPaladin, paladin.HammerOfWrathRankMap},
	{"Holy Wrath", "sim/paladin/holy_wrath.go", classPaladin, paladin.HolyWrathRankMap},
	{"Exorcism", "sim/paladin/exorcism.go", classPaladin, paladin.ExorcismRankMap},
	{"Holy Light", "sim/paladin/healing.go", classPaladin, paladin.HolyLightRankMap},
	{"Flash of Light", "sim/paladin/healing.go", classPaladin, paladin.FlashOfLightRankMap},
	{"Lay on Hands", "sim/paladin/healing.go", classPaladin, paladin.LayOnHandsRankMap},
	{"Holy Shield", "sim/paladin/holy_shield.go", classPaladin, paladin.HolyShieldRankMap},
	{"Holy Shock", "sim/paladin/holy_shock.go", classPaladin, paladin.HolyShockRankMap},
	{"Avenger's Shield", "sim/paladin/avengers_shield.go", classPaladin, paladin.AvengersShieldRankMap},

	{"Mind Blast", "sim/priest/mind_blast.go", classPriest, priest.MindBlastRankMap},
	{"Mind Flay", "sim/priest/mind_flay.go", classPriest, priest.MindFlayRankMap},
	{"Shadow Word: Pain", "sim/priest/shadow_word_pain.go", classPriest, priest.ShadowWordPainRankMap},
	{"Shadow Word: Death", "sim/priest/shadow_word_death.go", classPriest, priest.ShadowWordDeathRankMap},
	{"Smite", "sim/priest/smite.go", classPriest, priest.SmiteRankMap},
	{"Devouring Plague", "sim/priest/devouring_plague.go", classPriest, priest.DevouringPlagueRankMap},
	{"Holy Nova", "sim/priest/holy_nova.go", classPriest, priest.HolyNovaRankMap},
	{"Starshards", "sim/priest/starshards.go", classPriest, priest.StarshardsRankMap},
	{"Vampiric Touch", "sim/priest/vampiric_touch.go", classPriest, priest.VampiricTouchRankMap},

	{"Lightning Bolt", "sim/shaman/lightning_bolt.go", classShaman, shaman.LightningBoltRankMap},
	{"Chain Lightning", "sim/shaman/chain_lightning.go", classShaman, shaman.ChainLightningRankMap},

	{"Flamestrike", "sim/mage/flamestrike.go", classMage, mage.FlameStrikeRankMap},
	{"Starfire", "sim/druid/starfire.go", classDruid, druid.StarfireRankMap},
}

type knownResidual struct {
	SpellID int32
	Field   string
	Verdict string
	Note    string
}

// Empty, and that is the point: every disagreement the gate ever found - Exorcism rank 5, Holy Shock
// rank 3, Mind Blast's and Mind Flay's coefficients - was resolved in favour of the client DB, so the
// tables now agree with it by construction. What the gate still catches is a generated file edited by
// hand, or one left stale after the client data moved. A residual that cannot be explained here fails
// the build.
var knownResiduals []knownResidual

// Literal rounding in the hand tables: 0.429 stands in for 0.428999990224838, a 2.3e-8 difference.
// Deliberately far tighter than the gap between a rounded literal and a genuinely different number -
// Mind Blast's 0.42857 (3/7) against the DB's 0.429 is 4.3e-4 and has to surface as a residual, not be
// waved through as precision.
const coefRoundingTolerance = 1e-5

type comparison struct {
	Family  string
	File    string
	Rank    int32
	SpellID int32
	Field   string
	Hand    float64
	Derived float64
	Source  string
	Status  string
	Note    string
}

func (c comparison) ok() bool { return c.Status == "MATCH" || c.Status == "KNOWN" }

func classify(spellID int32, field string) (string, string, bool) {
	for _, k := range knownResiduals {
		if k.SpellID == spellID && k.Field == field {
			return "KNOWN", k.Verdict + ": " + k.Note, true
		}
	}
	return "", "", false
}

func TestSpellRankCalibration(t *testing.T) {
	DatabasePath = "wowsims.db"
	if _, err := os.Stat(DatabasePath); err != nil {
		t.Skipf("no client database at %s - run `make db` from a local WoW install to enable this gate", DatabasePath)
	}

	helper, err := NewDBHelper()
	if err != nil {
		t.Fatalf("opening %s: %v", DatabasePath, err)
	}
	defer helper.Close()
	db := helper.db

	var all []comparison
	for _, fam := range calibFamilies {
		for _, row := range fam.Table {
			all = append(all, compareRow(t, db, fam, row)...)
		}
	}

	var residuals []comparison
	for _, c := range all {
		if !c.ok() {
			residuals = append(residuals, c)
		}
	}

	matched := 0
	known := 0
	for _, c := range all {
		switch c.Status {
		case "MATCH":
			matched++
		case "KNOWN":
			known++
		}
	}
	t.Logf("%d comparisons: %d match, %d known-and-explained, %d unexplained", len(all), matched, known, len(residuals))

	for _, c := range residuals {
		t.Errorf("unexplained residual %s rank %d (spell %d) %s: hand %v, DB-derived %v (%s)\n"+
			"    classify it in knownResiduals with evidence, or fix the rule",
			c.Family, c.Rank, c.SpellID, c.Field, c.Hand, c.Derived, c.Source)
	}
}

func compareRow(t *testing.T, db *sql.DB, fam calibFamily, row shared.SpellRank) []comparison {
	t.Helper()

	spell, candidates, err := RankCandidates(db, row.SpellID, fam.ClassBit)
	if err != nil {
		t.Fatalf("%s rank %d: %v", fam.Name, row.Rank, err)
	}

	base := comparison{Family: fam.Name, File: fam.File, Rank: row.Rank, SpellID: row.SpellID}
	var out []comparison

	if row.Cost > 0 || spell.ManaCost.Valid {
		c := base
		c.Field = "Cost"
		c.Hand = float64(row.Cost)
		c.Source = "SpellPower.ManaCost"
		if spell.ManaCost.Valid {
			c.Derived = float64(spell.ManaCost.Int64)
		}
		out = append(out, finish(c))
	}

	// Reporting which effect reproduced each value is the point, not a nicety: it is the evidence the
	// generator's role rules get built from, and it is how a value that turns out to belong to a
	// sibling spell (Holy Shock's damage, Lay on Hands' energize) shows itself.
	coef := 0.0
	if row.Direct != nil {
		out = append(out, matchPair(base, "Direct.Min", "Direct.Max", shared.SpellRankMin(row.Direct), shared.SpellRankMax(row.Direct),
			directCandidates(candidates), spell)...)
		coef = shared.SpellRankCoef(row.Direct)
	}

	if row.Heal != nil {
		out = append(out, matchPair(base, "Heal.Min", "Heal.Max", shared.SpellRankMin(row.Heal), shared.SpellRankMax(row.Heal),
			directCandidates(candidates), spell)...)
		coef = shared.SpellRankCoef(row.Heal)
	}

	if row.Periodic != nil {
		out = append(out, matchTick(base, shared.SpellRankMin(row.Periodic), periodicCandidates(candidates), spell))
		if shared.SpellRankCoef(row.Periodic) > 0 {
			coef = shared.SpellRankCoef(row.Periodic)
		}
	}

	if row.Energize != nil {
		out = append(out, matchPair(base, "Energize", "", shared.SpellRankMin(row.Energize), 0,
			directCandidates(candidates), spell)...)
	}

	if coef > 0 {
		out = append(out, matchCoefficient(base, coef, candidates))
	}

	return out
}

func directCandidates(effects []RankEffect) []RankEffect {
	var out []RankEffect
	for _, e := range effects {
		switch e.Effect {
		case effSchoolDamage, effHeal, effEnergize:
			out = append(out, e)
		default:
			// Holy Shield's block damage lives on an aura effect (EffectAura 43) rather than a damage
			// effect, and Consecration's per-tick damage is the periodic one. Both are still "the number
			// the hand table wrote into MinDamage".
			if e.Aura != 0 {
				out = append(out, e)
			}
		}
	}
	return out
}

func periodicCandidates(effects []RankEffect) []RankEffect {
	var out []RankEffect
	for _, e := range effects {
		if e.Aura == effAuraPeriodic {
			out = append(out, e)
		}
	}
	if len(out) == 0 {
		out = effects
	}
	return out
}

func matchPair(base comparison, minField, maxField string, handMin, handMax float64, cands []RankEffect, spell RankSpell) []comparison {
	bestMin, bestMax, bestSrc, found := 0.0, 0.0, "no candidate effect", false
	for _, e := range cands {
		dMin, dMax := DeriveRankAmount(e, spell.SpellLevel, spell.MaxLevel)
		src := fmt.Sprintf("spell %d effect %d (Effect=%d, Aura=%d)", e.OwnerSpellID, e.Index, e.Effect, e.Aura)
		if dMin == handMin && (handMax == 0 || dMax == handMax) {
			out := []comparison{finishWith(base, minField, handMin, dMin, src)}
			if handMax > 0 {
				out = append(out, finishWith(base, maxField, handMax, dMax, src))
			}
			return out
		}
		// Closest candidate, so a residual reports a near-miss rather than "no match".
		if !found || math.Abs(dMin-handMin) < math.Abs(bestMin-handMin) {
			bestMin, bestMax, bestSrc, found = dMin, dMax, src, true
		}
	}

	out := []comparison{finishWith(base, minField, handMin, bestMin, bestSrc)}
	if handMax > 0 {
		out = append(out, finishWith(base, maxField, handMax, bestMax, bestSrc))
	}
	return out
}

func matchTick(base comparison, handTick float64, cands []RankEffect, spell RankSpell) comparison {
	bestTick, bestSrc := 0.0, "no periodic effect"
	for _, e := range cands {
		dMin, _ := DeriveRankAmount(e, spell.SpellLevel, spell.MaxLevel)
		src := fmt.Sprintf("spell %d effect %d (periodic)", e.OwnerSpellID, e.Index)
		if dMin == handTick {
			return finishWith(base, "DotTickDamage", handTick, dMin, src)
		}
		if bestSrc == "no periodic effect" || math.Abs(dMin-handTick) < math.Abs(bestTick-handTick) {
			bestTick, bestSrc = dMin, src
		}
	}
	return finishWith(base, "DotTickDamage", handTick, bestTick, bestSrc)
}

func matchCoefficient(base comparison, handCoef float64, cands []RankEffect) comparison {
	best, bestSrc := 0.0, "no candidate effect"
	for _, e := range cands {
		if e.Coefficient <= 0 {
			continue
		}
		src := fmt.Sprintf("spell %d effect %d", e.OwnerSpellID, e.Index)
		if math.Abs(e.Coefficient-handCoef) <= coefRoundingTolerance {
			return finishWith(base, "Coefficient", handCoef, e.Coefficient, src+" (literal rounding)")
		}
		if bestSrc == "no candidate effect" || math.Abs(e.Coefficient-handCoef) < math.Abs(best-handCoef) {
			best, bestSrc = e.Coefficient, src
		}
	}
	return finishWith(base, "Coefficient", handCoef, best, bestSrc)
}

func finishWith(base comparison, field string, hand, derived float64, source string) comparison {
	c := base
	c.Field = field
	c.Hand = hand
	c.Derived = derived
	c.Source = source
	return finish(c)
}

func finish(c comparison) comparison {
	if c.Hand == c.Derived {
		c.Status = "MATCH"
		return c
	}
	if c.Field == "Coefficient" && math.Abs(c.Hand-c.Derived) <= coefRoundingTolerance {
		c.Status = "MATCH"
		c.Note = "literal rounding"
		return c
	}
	if status, note, ok := classify(c.SpellID, c.Field); ok {
		c.Status = status
		c.Note = note
		return c
	}
	c.Status = "RESIDUAL"
	return c
}
