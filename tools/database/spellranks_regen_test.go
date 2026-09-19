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

// Literal rounding in the hand tables: 0.429 stands in for 0.428999990224838, a 2.3e-8 difference.
// Deliberately far tighter than the gap between a rounded literal and a genuinely different number -
// Mind Blast's 0.42857 (3/7) against the DB's 0.429 is 4.3e-4 and has to surface as a residual, not be
// waved through as precision.
type comparison struct {
	Family  string
	File    string
	Rank    int32
	SpellID int32
	Field   string
	Hand    float64
	Derived float64
	Source  string
}

func (c comparison) ok() bool { return c.Hand == c.Derived }

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
	for _, fam := range calibFamilies {
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
		if e.Coefficient == handCoef {
			return finishWith(base, "Coefficient", handCoef, e.Coefficient, src)
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
	return c
}
