package database

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
)

const RankLevel = 70

// The client stores rage in tenths - Heroic Strike costs 150, not 15 - because the server tracks a
// 0-1000 bar where the UI shows 0-100. Mana, energy and focus are stated as the player sees them, so
// rage is the one power type a cost has to be divided through. Confirmed on every warrior and bear
// ability in this build: each one is exactly ten times the sim's hand-written cost.
const powerTypeRage = 1

func NormalizePowerCost(cost int32, powerType int32) int32 {
	if powerType == powerTypeRage {
		return cost / 10
	}
	return cost
}

const (
	effSchoolDamage      = 2
	effHeal              = 10
	effEnergize          = 30
	effAuraPeriodic      = 3
	effAuraPeriodicLeech = 53

	// A melee ability states its bonus as weapon damage rather than school damage: Sinister Strike's
	// +98 is E_NORMALIZED_WEAPON_DMG, not E_SCHOOL_DAMAGE. E_WEAPON_PERCENT_DAMAGE is deliberately
	// absent - it is a multiplier on the weapon swing, not an amount a rank can carry.
	effWeaponDamageNoSchool = 17
	effWeaponDamage         = 58
	effNormalizedWeaponDmg  = 121
)

func IsWeaponDamageEffect(effect int32) bool {
	return effect == effWeaponDamageNoSchool || effect == effWeaponDamage || effect == effNormalizedWeaponDmg
}

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
	APCoef       float64
	AuraPeriod   int32
	OwnerSpellID int32
}

type RankSpell struct {
	SpellID    int32
	SpellLevel int32
	MaxLevel   int32
	ManaCost   sql.NullInt64
	PowerType  int32
	DurationMs int32
	CastTimeMs int32
	GCDMs      int32
	CooldownMs int32
	MinRange   float64
	MaxRange   float64

	// How fast the projectile flies, in yards per second, which core turns into the delay between
	// the cast landing and the damage arriving. Zero for a spell that hits the instant it is cast.
	MissileSpeed float64
	Effects      []RankEffect
}

// SpellCastTimes resolves SpellMisc.CastingTimeIndex and is absent from this build's database - the
// table is not in generator-settings.json's extraction list, so cast times stay zero until it is added
// and `make db` re-run against a client.
func castTimesAvailable(db *sql.DB) bool {
	var n int
	if err := db.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'SpellCastTimes'`).Scan(&n); err != nil {
		return false
	}
	return n > 0
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
		       (SELECT ManaCost FROM SpellPower WHERE SpellID = l.SpellID ORDER BY OrderIndex LIMIT 1),
		       COALESCE((SELECT PowerType FROM SpellPower WHERE SpellID = l.SpellID ORDER BY OrderIndex LIMIT 1), 0)
		FROM SpellLevels l WHERE l.SpellID = ?`, spellID).Scan(&s.SpellLevel, &s.MaxLevel, &s.ManaCost, &s.PowerType)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// Passive talents such as the warrior's Blood Craze have no SpellLevels row at all. No level
		// data means no level scaling, which is what setting the spell's own level to the cap gives.
		s.SpellLevel = RankLevel
	case err != nil:
		return s, fmt.Errorf("levels for spell %d: %w", spellID, err)
	}

	// Duration and its period are what a DoT's NumberOfTicks and TickLength are derived from.
	_ = db.QueryRow(`
		SELECT COALESCE(d.Duration, 0)
		FROM SpellMisc m LEFT JOIN SpellDuration d ON d.ID = m.DurationIndex
		WHERE m.SpellID = ?`, spellID).Scan(&s.DurationMs)

	// The GCD, the cooldown and the range are all reachable from tables already extracted. Cooldown
	// takes whichever of the two recovery times is longer: RecoveryTime is the spell's own, while
	// CategoryRecoveryTime is the shared category one that Fire Blast and Cone of Cold actually use,
	// and Cast.CD in the sim models whichever applies.
	_ = db.QueryRow(`
		SELECT COALESCE(max(RecoveryTime, CategoryRecoveryTime), 0), COALESCE(StartRecoveryTime, 0)
		FROM SpellCooldowns WHERE SpellID = ?`, spellID).Scan(&s.CooldownMs, &s.GCDMs)

	// RangeMin is nonzero on only 212 spells in this build - the dead zone on a charge, and a handful
	// of ranged abilities - but where it exists core gates the cast on it exactly as it does MaxRange.
	_ = db.QueryRow(`
		SELECT COALESCE(r.RangeMin_1, 0), COALESCE(r.RangeMax_1, 0)
		FROM SpellMisc m JOIN SpellRange r ON r.ID = m.RangeIndex
		WHERE m.SpellID = ?`, spellID).Scan(&s.MinRange, &s.MaxRange)

	_ = db.QueryRow(`SELECT COALESCE(Speed, 0) FROM SpellMisc WHERE SpellID = ?`, spellID).Scan(&s.MissileSpeed)

	if castTimesAvailable(db) {
		_ = db.QueryRow(`
			SELECT COALESCE(ct.Base, 0)
			FROM SpellMisc m JOIN SpellCastTimes ct ON ct.ID = m.CastingTimeIndex
			WHERE m.SpellID = ?`, spellID).Scan(&s.CastTimeMs)
	}

	s.Effects, err = RankEffectsOf(db, spellID)
	return s, err
}

func RankEffectsOf(db *sql.DB, spellID int32) ([]RankEffect, error) {
	rows, err := db.Query(`
		SELECT EffectIndex, Effect, EffectAura, EffectBasePoints, EffectDieSides,
		       EffectRealPointsPerLevel, EffectBonusCoefficient, BonusCoefficientFromAP, EffectAuraPeriod
		FROM SpellEffect WHERE SpellID = ? ORDER BY EffectIndex`, spellID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RankEffect
	for rows.Next() {
		e := RankEffect{OwnerSpellID: spellID}
		if err := rows.Scan(&e.Index, &e.Effect, &e.Aura, &e.BasePoints, &e.DieSides, &e.PointsPerLvl, &e.Coefficient, &e.APCoef, &e.AuraPeriod); err != nil {
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
		if e.Effect == effSchoolDamage || e.Effect == effHeal || e.Effect == effEnergize ||
			IsWeaponDamageEffect(e.Effect) || IsPeriodicAura(e.Aura) {
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
