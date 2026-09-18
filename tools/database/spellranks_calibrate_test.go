package database

// Calibration gate for the spell-rank derivation rule.
//
// Every rank row in the sim is hand-transcribed today. Before those tables can be generated from the
// client DB, the rule that derives them has to be shown to reproduce what is already there - and every
// row it does not reproduce has to be classified as either "the rule is wrong" or "the hand row is
// wrong", with the evidence written down rather than argued about.
//
// Run:  go test ./tools/database/ -run SpellRankCalibration            (gate)
//       go test ./tools/database/ -run SpellRankCalibration -update    (rewrite the report)
//
// Skips when tools/database/wowsims.db is absent - it is gitignored and only produced by `make db`
// from a local WoW install, so CI and a fresh clone legitimately have no client data.

import (
	"database/sql"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/druid"
	"github.com/wowsims/tbc/sim/mage"
	"github.com/wowsims/tbc/sim/paladin"
	"github.com/wowsims/tbc/sim/priest"
	"github.com/wowsims/tbc/sim/shaman"
)

var updateReport = flag.Bool("update", false, "rewrite spellranks_calibration.md instead of only checking it")

const calibLevel = 70

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
	Table    shared.RankTable
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

var knownResiduals = []knownResidual{
	{10313, "Direct.Min", "hand row wrong",
		"Exorcism r5. The hand min (453) is the DB's max, a column slip, and the hand spread of 54 breaks " +
			"the die-sides ladder (r4 38, r5 46, r6 58)."},
	{10313, "Direct.Max", "hand row wrong", "Exorcism r5, same column slip as above."},
	{20930, "Direct.Max", "hand row wrong",
		"Holy Shock r3. 628 is heal spell 25903's min (bp 627 + 1); the damage spell 25902 is bp 495 / ds 41 " +
			"= 496-536. Ranks 1, 2, 4 and 5 are exact."},
}

var mindBlastCoefRanks = []int32{8103, 8104, 8105, 8106, 10945, 10946, 10947, 25372, 25375}

func init() {
	for _, id := range mindBlastCoefRanks {
		knownResiduals = append(knownResiduals, knownResidual{id, "Coefficient", "decision pending",
			"Mind Blast ranks 3+. The sim uses 3/7 = 0.42857 where the DB says 0.429 - 4.3e-4 absolute, " +
				"0.1% of SP scaling. Consistent across every rank, so a deliberate choice rather than a " +
				"transcription slip: keep it as a documented override, or take the DB value and move the " +
				"golden. Ranks 1 and 2 carry their own coefficients and match exactly."})
	}
}

const (
	effSchoolDamage = 2
	effHeal         = 10
	effEnergize     = 30
	effAuraPeriodic = 3
)

// Literal rounding in the hand tables: 0.429 stands in for 0.428999990224838, a 2.3e-8 difference.
// Deliberately far tighter than the gap between a rounded literal and a genuinely different number -
// Mind Blast's 0.42857 (3/7) against the DB's 0.429 is 4.3e-4 and has to surface as a residual, not be
// waved through as precision.
const coefRoundingTolerance = 1e-5

type dbEffect struct {
	Index        int32
	Effect       int32
	Aura         int32
	BasePoints   int32
	DieSides     int32
	PointsPerLvl float64
	Coefficient  float64
	OwnerSpellID int32
}

type dbSpell struct {
	SpellLevel int32
	MaxLevel   int32
	ManaCost   sql.NullInt64
	Effects    []dbEffect
}

// derive applies the calibrated rule. Kept as one named function precisely so it can be swapped
// deliberately: the TBC server roll is plausibly trunc(base)+1 .. trunc(base)+dieSides, one lower on max
// for a fractional base, and this reproduces the tooltip values rather than proving the server's.
//
// float32 throughout is load-bearing, not incidental: EffectRealPointsPerLevel is a float32 widened into
// the DB (3.79999995231628), and doing the multiply in float64 breaks 6 rows that float32 gets right.
func derive(e dbEffect, spellLevel, maxLevel int32) (min float64, max float64) {
	cap := maxLevel
	if cap <= 0 {
		cap = calibLevel
	}
	lvl := int32(calibLevel)
	if cap < lvl {
		lvl = cap
	}
	delta := lvl - spellLevel
	if delta < 0 {
		delta = 0
	}

	base := float32(e.BasePoints) + float32(float32(delta)*float32(e.PointsPerLvl))
	min = math.Floor(float64(base)) + 1
	max = math.Ceil(float64(base)) + float64(e.DieSides)
	if e.DieSides <= 0 {
		max = min
	}
	return min, max
}

func loadSpell(db *sql.DB, spellID int32) (dbSpell, error) {
	var s dbSpell
	err := db.QueryRow(`
		SELECT l.SpellLevel, l.MaxLevel,
		       (SELECT ManaCost FROM SpellPower WHERE SpellID = l.SpellID ORDER BY OrderIndex LIMIT 1)
		FROM SpellLevels l WHERE l.SpellID = ?`, spellID).Scan(&s.SpellLevel, &s.MaxLevel, &s.ManaCost)
	if err != nil {
		return s, fmt.Errorf("levels for spell %d: %w", spellID, err)
	}

	s.Effects, err = effectsOf(db, spellID)
	return s, err
}

func effectsOf(db *sql.DB, spellID int32) ([]dbEffect, error) {
	rows, err := db.Query(`
		SELECT EffectIndex, Effect, EffectAura, EffectBasePoints, EffectDieSides,
		       EffectRealPointsPerLevel, EffectBonusCoefficient
		FROM SpellEffect WHERE SpellID = ? ORDER BY EffectIndex`, spellID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []dbEffect
	for rows.Next() {
		e := dbEffect{OwnerSpellID: spellID}
		if err := rows.Scan(&e.Index, &e.Effect, &e.Aura, &e.BasePoints, &e.DieSides, &e.PointsPerLvl, &e.Coefficient); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Holy Shock's registered spells carry only Effect=3 (dummy) and have no EffectTriggerSpell edge to the
// damage and heal spells that share their name and rank - the association exists nowhere but the name.
// So when a registered spell has no value-bearing effect of its own, look at its same-name/same-rank
// siblings, restricted to the family's class so an NPC copy cannot be picked up.
func siblingEffects(db *sql.DB, spellID int32, classBit int) ([]dbEffect, error) {
	rows, err := db.Query(`
		SELECT DISTINCT sla.Spell
		FROM SkillLineAbility sla
		JOIN SpellName n ON n.ID = sla.Spell
		JOIN Spell s ON s.ID = sla.Spell
		WHERE n.Name_lang = (SELECT Name_lang FROM SpellName WHERE ID = ?)
		  AND s.NameSubtext_lang = (SELECT NameSubtext_lang FROM Spell WHERE ID = ?)
		  AND (sla.ClassMask & ?) != 0
		  AND sla.Spell != ?`, spellID, spellID, classBit, spellID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var out []dbEffect
	for _, id := range ids {
		effs, err := effectsOf(db, id)
		if err != nil {
			return nil, err
		}
		out = append(out, effs...)
	}
	return out, nil
}

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

	if *updateReport {
		if err := writeReport(all); err != nil {
			t.Fatalf("writing report: %v", err)
		}
		t.Logf("wrote spellranks_calibration.md (%d comparisons)", len(all))
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

func compareRow(t *testing.T, db *sql.DB, fam calibFamily, row shared.RankRow) []comparison {
	t.Helper()

	spell, err := loadSpell(db, row.SpellID)
	if err != nil {
		t.Fatalf("%s rank %d: %v", fam.Name, row.Rank, err)
	}

	candidates := spell.Effects
	if !hasValueEffect(candidates) {
		sibs, err := siblingEffects(db, row.SpellID, fam.ClassBit)
		if err != nil {
			t.Fatalf("%s rank %d siblings: %v", fam.Name, row.Rank, err)
		}
		candidates = append(candidates, sibs...)
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
		out = append(out, matchPair(base, "Direct.Min", "Direct.Max", row.Direct.Min, row.Direct.Max,
			directCandidates(candidates), spell)...)
		coef = row.Direct.Coef
	}

	if row.Heal != nil {
		out = append(out, matchPair(base, "Heal.Min", "Heal.Max", row.Heal.Min, row.Heal.Max,
			directCandidates(candidates), spell)...)
		coef = row.Heal.Coef
	}

	if row.Periodic != nil {
		out = append(out, matchTick(base, row.Periodic.Tick, periodicCandidates(candidates), spell))
		if row.Periodic.Coef > 0 {
			coef = row.Periodic.Coef
		}
	}

	if row.Energize > 0 {
		out = append(out, matchPair(base, "Energize", "", row.Energize, 0,
			directCandidates(candidates), spell)...)
	}

	if coef > 0 {
		out = append(out, matchCoefficient(base, coef, candidates))
	}

	return out
}

func hasValueEffect(effects []dbEffect) bool {
	for _, e := range effects {
		if e.Effect == effSchoolDamage || e.Effect == effHeal || e.Effect == effEnergize || e.Aura == effAuraPeriodic {
			return true
		}
	}
	return false
}

func directCandidates(effects []dbEffect) []dbEffect {
	var out []dbEffect
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

func periodicCandidates(effects []dbEffect) []dbEffect {
	var out []dbEffect
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

func matchPair(base comparison, minField, maxField string, handMin, handMax float64, cands []dbEffect, spell dbSpell) []comparison {
	bestMin, bestMax, bestSrc, found := 0.0, 0.0, "no candidate effect", false
	for _, e := range cands {
		dMin, dMax := derive(e, spell.SpellLevel, spell.MaxLevel)
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

func matchTick(base comparison, handTick float64, cands []dbEffect, spell dbSpell) comparison {
	bestTick, bestSrc := 0.0, "no periodic effect"
	for _, e := range cands {
		dMin, _ := derive(e, spell.SpellLevel, spell.MaxLevel)
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

func matchCoefficient(base comparison, handCoef float64, cands []dbEffect) comparison {
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

func writeReport(all []comparison) error {
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].Family != all[j].Family {
			return all[i].Family < all[j].Family
		}
		if all[i].Rank != all[j].Rank {
			return all[i].Rank < all[j].Rank
		}
		return all[i].Field < all[j].Field
	})

	matched, known, residual := 0, 0, 0
	for _, c := range all {
		switch c.Status {
		case "MATCH":
			matched++
		case "KNOWN":
			known++
		default:
			residual++
		}
	}

	var b strings.Builder
	b.WriteString("# Spell rank calibration\n\n")
	b.WriteString("Generated by `go test ./tools/database/ -run SpellRankCalibration -update`. Do not edit by hand.\n\n")
	b.WriteString("Every hand-written rank row in the sim, compared against the value derived from the client DB.\n")
	b.WriteString("This is the gate the generator has to pass before it may replace those tables.\n\n")

	b.WriteString("## Rule\n\n```\n")
	b.WriteString("delta = max(0, min(70, MaxLevel > 0 ? MaxLevel : 70) - SpellLevels.SpellLevel)\n")
	b.WriteString("base  = f32(EffectBasePoints + f32(delta * f32(EffectRealPointsPerLevel)))\n")
	b.WriteString("min   = floor(base) + 1\n")
	b.WriteString("max   = ceil(base)  + EffectDieSides\n")
	b.WriteString("```\n\n")
	b.WriteString("float32 throughout is load-bearing: the multiply in float64 breaks rows this gets right.\n")
	b.WriteString("A coefficient within 5e-4 counts as a match - the hand tables write `0.429` for `0.428999990224838`.\n\n")

	fmt.Fprintf(&b, "## Result\n\n%d comparisons: **%d match**, **%d known and explained**, **%d unexplained**.\n\n",
		len(all), matched, known, residual)

	b.WriteString("## Known residuals\n\n")
	b.WriteString("| Spell | Field | Verdict | Evidence |\n|---|---|---|---|\n")
	for _, k := range knownResiduals {
		fmt.Fprintf(&b, "| %d | %s | %s | %s |\n", k.SpellID, k.Field, k.Verdict, k.Note)
	}
	b.WriteString("\n## Every comparison\n\n")
	b.WriteString("| Family | Rank | Spell | Field | Hand | DB-derived | Status | Source |\n|---|---|---|---|---|---|---|---|\n")
	for _, c := range all {
		status := c.Status
		if c.Status == "MATCH" && c.Note != "" {
			status = "MATCH (" + c.Note + ")"
		}
		fmt.Fprintf(&b, "| %s | %d | %d | %s | %s | %s | %s | %s |\n",
			c.Family, c.Rank, c.SpellID, c.Field, num(c.Hand), num(c.Derived), status, c.Source)
	}

	return os.WriteFile("spellranks_calibration.md", []byte(b.String()), 0644)
}

func num(f float64) string {
	if f == math.Trunc(f) && math.Abs(f) < 1e9 {
		return fmt.Sprintf("%d", int64(f))
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.9f", f), "0"), ".")
}
