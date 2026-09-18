package database

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
)

const RankLevel = 70

const (
	effSchoolDamage      = 2
	effHeal              = 10
	effEnergize          = 30
	effAuraPeriodic      = 3
	effAuraPeriodicLeech = 53
)

// Devouring Plague ticks as a leech rather than as plain periodic damage, so "is this a DoT" cannot be
// a single aura check.
func IsPeriodicAura(aura int32) bool {
	return aura == effAuraPeriodic || aura == effAuraPeriodicLeech
}

type RankEffect struct {
	Index        int32
	Effect       int32
	Aura         int32
	BasePoints   int32
	DieSides     int32
	PointsPerLvl float64
	Coefficient  float64
	OwnerSpellID int32
}

type RankSpell struct {
	SpellID    int32
	SpellLevel int32
	MaxLevel   int32
	ManaCost   sql.NullInt64
	Effects    []RankEffect
}

// The calibrated derivation rule, shared by the generator and the calibration gate so the two cannot
// drift apart. Scored 86/87 against the hand-written tables; the one residual is a hand-row bug.
//
// float32 throughout is load-bearing, not incidental: EffectRealPointsPerLevel is a float32 widened into
// the DB (3.79999995231628), and doing the multiply in float64 breaks 6 rows that float32 gets right.
//
// This reproduces the tooltip values, which is not the same claim as reproducing the server: the TBC
// server roll is plausibly trunc(base)+1 .. trunc(base)+dieSides, one lower on max for a fractional
// base. Kept as one named function so that can be changed deliberately rather than by accident.
func DeriveRankAmount(e RankEffect, spellLevel, maxLevel int32) (min float64, max float64) {
	cap := maxLevel
	if cap <= 0 {
		cap = RankLevel
	}
	lvl := int32(RankLevel)
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

func LoadRankSpell(db *sql.DB, spellID int32) (RankSpell, error) {
	s := RankSpell{SpellID: spellID}
	err := db.QueryRow(`
		SELECT l.SpellLevel, l.MaxLevel,
		       (SELECT ManaCost FROM SpellPower WHERE SpellID = l.SpellID ORDER BY OrderIndex LIMIT 1)
		FROM SpellLevels l WHERE l.SpellID = ?`, spellID).Scan(&s.SpellLevel, &s.MaxLevel, &s.ManaCost)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// Passive talents such as the warrior's Blood Craze have no SpellLevels row at all. No level
		// data means no level scaling, which is what setting the spell's own level to the cap gives.
		s.SpellLevel = RankLevel
	case err != nil:
		return s, fmt.Errorf("levels for spell %d: %w", spellID, err)
	}

	s.Effects, err = RankEffectsOf(db, spellID)
	return s, err
}

func RankEffectsOf(db *sql.DB, spellID int32) ([]RankEffect, error) {
	rows, err := db.Query(`
		SELECT EffectIndex, Effect, EffectAura, EffectBasePoints, EffectDieSides,
		       EffectRealPointsPerLevel, EffectBonusCoefficient
		FROM SpellEffect WHERE SpellID = ? ORDER BY EffectIndex`, spellID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RankEffect
	for rows.Next() {
		e := RankEffect{OwnerSpellID: spellID}
		if err := rows.Scan(&e.Index, &e.Effect, &e.Aura, &e.BasePoints, &e.DieSides, &e.PointsPerLvl, &e.Coefficient); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Holy Shock's registered spells carry only Effect=3 (dummy) and have no EffectTriggerSpell edge to the
// damage and heal spells that share their name and rank - the association exists nowhere but the name.
// Restricted to the family's class so an NPC copy of the name cannot be picked up.
func SiblingRankEffects(db *sql.DB, spellID int32, classBit int) ([]RankEffect, error) {
	rows, err := db.Query(`
		SELECT DISTINCT sla.Spell
		FROM SkillLineAbility sla
		JOIN SpellName n ON n.ID = sla.Spell
		JOIN Spell s ON s.ID = sla.Spell
		WHERE n.Name_lang = (SELECT Name_lang FROM SpellName WHERE ID = ?)
		  AND s.NameSubtext_lang = (SELECT NameSubtext_lang FROM Spell WHERE ID = ?)
		  AND (sla.ClassMask & ?) != 0
		  AND sla.Spell != ?
		ORDER BY sla.Spell`, spellID, spellID, classBit, spellID)
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

	var out []RankEffect
	for _, id := range ids {
		effs, err := RankEffectsOf(db, id)
		if err != nil {
			return nil, err
		}
		out = append(out, effs...)
	}
	return out, nil
}

func HasValueEffect(effects []RankEffect) bool {
	for _, e := range effects {
		if e.Effect == effSchoolDamage || e.Effect == effHeal || e.Effect == effEnergize || IsPeriodicAura(e.Aura) {
			return true
		}
	}
	return false
}

// Every effect that could carry the numbers a rank row wants, including the ones reached only through
// a same-name sibling.
func RankCandidates(db *sql.DB, spellID int32, classBit int) (RankSpell, []RankEffect, error) {
	spell, err := LoadRankSpell(db, spellID)
	if err != nil {
		return spell, nil, err
	}

	candidates := spell.Effects
	if !HasValueEffect(candidates) {
		sibs, err := SiblingRankEffects(db, spellID, classBit)
		if err != nil {
			return spell, nil, err
		}
		candidates = append(candidates, sibs...)
	}
	return spell, candidates, nil
}
