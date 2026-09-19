package database

// Regeneration check for the generated rank tables.
//
// It re-derives every row of the families below straight from the client database and asserts the
// committed table says the same thing. It began as a calibration gate, proving the derivation rule
// against 141 hand-transcribed rows; those rows are gone, so both sides now come from one database
// and the residual bookkeeping that classified a disagreement went with them.
//
// What it still catches: a generated file edited by hand, one left stale after the client data moved,
// and any change to DeriveRankAmount that moves a number.
//
// Run:  go test --tags=with_db ./tools/database/ -run GeneratedRankTables
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

type rankFamily struct {
	Name     string
	ClassBit int
	Table    shared.SpellRankTable
}

// The two shaman tables were inline anonymous literals until they were hoisted to package vars so this
// gate could read them.
var rankFamilies = []rankFamily{
	{"Consecration", classPaladin, paladin.ConsecrationRankMap},
	{"Hammer of Wrath", classPaladin, paladin.HammerOfWrathRankMap},
	{"Holy Wrath", classPaladin, paladin.HolyWrathRankMap},
	{"Exorcism", classPaladin, paladin.ExorcismRankMap},
	{"Holy Light", classPaladin, paladin.HolyLightRankMap},
	{"Flash of Light", classPaladin, paladin.FlashOfLightRankMap},
	{"Lay on Hands", classPaladin, paladin.LayOnHandsRankMap},
	{"Holy Shield", classPaladin, paladin.HolyShieldRankMap},
	{"Holy Shock", classPaladin, paladin.HolyShockRankMap},
	{"Avenger's Shield", classPaladin, paladin.AvengersShieldRankMap},

	{"Mind Blast", classPriest, priest.MindBlastRankMap},
	{"Mind Flay", classPriest, priest.MindFlayRankMap},
	{"Shadow Word: Pain", classPriest, priest.ShadowWordPainRankMap},
	{"Shadow Word: Death", classPriest, priest.ShadowWordDeathRankMap},
	{"Smite", classPriest, priest.SmiteRankMap},
	{"Devouring Plague", classPriest, priest.DevouringPlagueRankMap},
	{"Holy Nova", classPriest, priest.HolyNovaRankMap},
	{"Starshards", classPriest, priest.StarshardsRankMap},
	{"Vampiric Touch", classPriest, priest.VampiricTouchRankMap},

	{"Lightning Bolt", classShaman, shaman.LightningBoltRankMap},
	{"Chain Lightning", classShaman, shaman.ChainLightningRankMap},

	{"Flamestrike", classMage, mage.FlameStrikeRankMap},
	{"Starfire", classDruid, druid.StarfireRankMap},
}

// One value in the committed table against the same value re-derived from the database.
type comparison struct {
	Family    string
	Rank      int32
	SpellID   int32
	Field     string
	Generated float64
	Derived   float64
	Source    string
}

func (c comparison) ok() bool { return c.Generated == c.Derived }

func TestGeneratedRankTablesMatchTheDatabase(t *testing.T) {
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
	for _, fam := range rankFamilies {
		for _, row := range fam.Table {
			all = append(all, compareRow(t, db, fam, row)...)
		}
	}

	var mismatched []comparison
	for _, c := range all {
		if !c.ok() {
			mismatched = append(mismatched, c)
		}
	}
	t.Logf("%d comparisons, %d mismatched", len(all), len(mismatched))

	for _, c := range mismatched {
		t.Errorf("%s rank %d (spell %d) %s: table says %v, the database derives %v (%s)\n"+
			"    regenerate with `go run ./tools/database/gen_spellranks`, or fix DeriveRankAmount",
			c.Family, c.Rank, c.SpellID, c.Field, c.Generated, c.Derived, c.Source)
	}
}

func compareRow(t *testing.T, db *sql.DB, fam rankFamily, row shared.SpellRank) []comparison {
	t.Helper()

	spell, candidates, err := RankCandidates(db, row.SpellID, fam.ClassBit)
	if err != nil {
		t.Fatalf("%s rank %d: %v", fam.Name, row.Rank, err)
	}

	base := comparison{Family: fam.Name, Rank: row.Rank, SpellID: row.SpellID}
	var out []comparison

	if row.Cost > 0 || spell.ManaCost.Valid {
		c := base
		c.Field = "Cost"
		c.Generated = float64(row.Cost)
		c.Source = "SpellPower.ManaCost"
		if spell.ManaCost.Valid {
			c.Derived = float64(NormalizePowerCost(int32(spell.ManaCost.Int64), spell.PowerType))
		}
		out = append(out, c)
	}

	// Reporting which effect reproduced each value is the point, not a nicety: it is the evidence the
	// generator's role rules get built from, and it is how a value that turns out to belong to a
	// sibling spell (Holy Shock's damage, Lay on Hands' energize) shows itself.
	coef := 0.0
	if row.Direct != nil {
		out = append(out, matchPair(base, "Direct.Min", "Direct.Max", shared.SpellRankMin(row.Direct), shared.SpellRankMax(row.Direct),
			directCandidates(candidates), spell)...)
		coef = row.Direct.BonusCoefficient()
	}

	if row.Heal != nil {
		out = append(out, matchPair(base, "Heal.Min", "Heal.Max", shared.SpellRankMin(row.Heal), shared.SpellRankMax(row.Heal),
			directCandidates(candidates), spell)...)
		coef = row.Heal.BonusCoefficient()
	}

	if row.Periodic != nil {
		out = append(out, matchTick(base, shared.SpellRankMin(row.Periodic), periodicCandidates(candidates), spell))
		if row.Periodic.BonusCoefficient() > 0 {
			coef = row.Periodic.BonusCoefficient()
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

// Holy Shield's per-block damage lives on an aura effect (EffectAura 43) rather than on a damage
// effect, and the generator files it under Direct all the same, so any aura effect is a candidate too.
func directCandidates(effects []RankEffect) []RankEffect {
	var out []RankEffect
	for _, e := range effects {
		if e.Effect == effSchoolDamage || e.Effect == effHeal || e.Effect == effEnergize || e.Aura != 0 {
			out = append(out, e)
		}
	}
	return out
}

func periodicCandidates(effects []RankEffect) []RankEffect {
	var out []RankEffect
	for _, e := range effects {
		if IsPeriodicAura(e.Aura) {
			out = append(out, e)
		}
	}
	if len(out) == 0 {
		out = effects
	}
	return out
}

func matchPair(base comparison, minField, maxField string, genMin, genMax float64, cands []RankEffect, spell RankSpell) []comparison {
	bestMin, bestMax, bestSrc, found := 0.0, 0.0, "no candidate effect", false
	for _, e := range cands {
		dMin, dMax := DeriveRankAmount(e, spell.SpellLevel, spell.MaxLevel)
		src := fmt.Sprintf("spell %d effect %d (Effect=%d, Aura=%d)", e.OwnerSpellID, e.Index, e.Effect, e.Aura)
		if dMin == genMin && (genMax == 0 || dMax == genMax) {
			out := []comparison{finishWith(base, minField, genMin, dMin, src)}
			if genMax > 0 {
				out = append(out, finishWith(base, maxField, genMax, dMax, src))
			}
			return out
		}
		// Closest candidate, so a residual reports a near-miss rather than "no match".
		if !found || math.Abs(dMin-genMin) < math.Abs(bestMin-genMin) {
			bestMin, bestMax, bestSrc, found = dMin, dMax, src, true
		}
	}

	out := []comparison{finishWith(base, minField, genMin, bestMin, bestSrc)}
	if genMax > 0 {
		out = append(out, finishWith(base, maxField, genMax, bestMax, bestSrc))
	}
	return out
}

func matchTick(base comparison, genTick float64, cands []RankEffect, spell RankSpell) comparison {
	bestTick, bestSrc := 0.0, "no periodic effect"
	for _, e := range cands {
		dMin, _ := DeriveRankAmount(e, spell.SpellLevel, spell.MaxLevel)
		src := fmt.Sprintf("spell %d effect %d (periodic)", e.OwnerSpellID, e.Index)
		if dMin == genTick {
			return finishWith(base, "DotTickDamage", genTick, dMin, src)
		}
		if bestSrc == "no periodic effect" || math.Abs(dMin-genTick) < math.Abs(bestTick-genTick) {
			bestTick, bestSrc = dMin, src
		}
	}
	return finishWith(base, "DotTickDamage", genTick, bestTick, bestSrc)
}

func matchCoefficient(base comparison, genCoef float64, cands []RankEffect) comparison {
	best, bestSrc := 0.0, "no candidate effect"
	for _, e := range cands {
		if e.Coefficient <= 0 {
			continue
		}
		src := fmt.Sprintf("spell %d effect %d", e.OwnerSpellID, e.Index)
		if e.Coefficient == genCoef {
			return finishWith(base, "Coefficient", genCoef, e.Coefficient, src)
		}
		if bestSrc == "no candidate effect" || math.Abs(e.Coefficient-genCoef) < math.Abs(best-genCoef) {
			best, bestSrc = e.Coefficient, src
		}
	}
	return finishWith(base, "Coefficient", genCoef, best, bestSrc)
}

func finishWith(base comparison, field string, generated, derived float64, source string) comparison {
	c := base
	c.Field = field
	c.Generated = generated
	c.Derived = derived
	c.Source = source
	return c
}
