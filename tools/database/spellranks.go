package database

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sync"
)

const RankLevel = 70

// Rage is stored in tenths: Heroic Strike costs 150, not 15. Mana, energy and focus are not.
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
	MiscValue    int32
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

// SpellCastTimes resolves SpellMisc.CastingTimeIndex and is absent from a database extracted before it
// was added to generator-settings.json. Only the generator insists on it - the regeneration check
// compares amounts and coefficients, neither of which needs a cast time.
// Answered once per database rather than once per row: LoadRankSpell runs 3327 times and the schema
// cannot change underneath it.
var castTimesOnce sync.Once
var castTimesPresent bool

func castTimesAvailable(db *sql.DB) bool {
	castTimesOnce.Do(func() {
		var n int
		if err := db.QueryRow(
			`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'SpellCastTimes'`).Scan(&n); err == nil {
			castTimesPresent = n > 0
		}
	})
	return castTimesPresent
}

func RequireSpellCastTimes(db *sql.DB) error {
	if castTimesAvailable(db) {
		return nil
	}
	return errors.New("the client database has no SpellCastTimes table, so every generated cast time would " +
		"be zero - add \"SpellCastTimes\" to tools/database/generator-settings.json and re-run `make db`")
}

// The calibrated rule, shared with the regeneration check. Scored 86/87 against the hand tables.
//
// float32 is load-bearing: EffectRealPointsPerLevel is a float32 widened into the DB
// (3.79999995231628), and multiplying in float64 breaks 6 rows. Reproduces the tooltip, which is not
// the same as reproducing the server roll.
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
	if err := scanOptional(db, `
		SELECT COALESCE(d.Duration, 0)
		FROM SpellMisc m LEFT JOIN SpellDuration d ON d.ID = m.DurationIndex
		WHERE m.SpellID = ?`, spellID, &s.DurationMs); err != nil {
		return s, fmt.Errorf("duration for spell %d: %w", spellID, err)
	}

	// Cooldown takes the longer of the two: Fire Blast and Cone of Cold use the shared
	// CategoryRecoveryTime, everything else its own RecoveryTime.
	if err := scanOptional(db, `
		SELECT COALESCE(max(RecoveryTime, CategoryRecoveryTime), 0), COALESCE(StartRecoveryTime, 0)
		FROM SpellCooldowns WHERE SpellID = ?`, spellID, &s.CooldownMs, &s.GCDMs); err != nil {
		return s, fmt.Errorf("cooldown for spell %d: %w", spellID, err)
	}

	// RangeMin is nonzero on only 212 spells in this build - the dead zone on a charge, and a handful
	// of ranged abilities - but where it exists core gates the cast on it exactly as it does MaxRange.
	if err := scanOptional(db, `
		SELECT COALESCE(r.RangeMin_0, 0), COALESCE(r.RangeMax_0, 0)
		FROM SpellMisc m JOIN SpellRange r ON r.ID = m.RangeIndex
		WHERE m.SpellID = ?`, spellID, &s.MinRange, &s.MaxRange); err != nil {
		return s, fmt.Errorf("range for spell %d: %w", spellID, err)
	}

	if err := scanOptional(db,
		`SELECT COALESCE(Speed, 0) FROM SpellMisc WHERE SpellID = ?`, spellID, &s.MissileSpeed); err != nil {
		return s, fmt.Errorf("missile speed for spell %d: %w", spellID, err)
	}

	if castTimesAvailable(db) {
		if err := scanOptional(db, `
			SELECT COALESCE(ct.Base, 0)
			FROM SpellMisc m JOIN SpellCastTimes ct ON ct.ID = m.CastingTimeIndex
			WHERE m.SpellID = ?`, spellID, &s.CastTimeMs); err != nil {
			return s, fmt.Errorf("cast time for spell %d: %w", spellID, err)
		}
	}

	s.Effects, err = RankEffectsOf(db, spellID)
	return s, err
}

// A spell with no row in one of the optional tables is ordinary - most spells have no cooldown - and
// leaves the destination at its zero, which is what core reads as "ungated". Any other error means the
// schema moved, and silently zeroing a cast time or a range on that is the failure this loader exists
// to avoid.
func scanOptional(db *sql.DB, query string, spellID int32, dest ...any) error {
	err := db.QueryRow(query, spellID).Scan(dest...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func RankEffectsOf(db *sql.DB, spellID int32) ([]RankEffect, error) {
	rows, err := db.Query(`
		SELECT EffectIndex, Effect, EffectAura, EffectBasePoints, EffectDieSides,
		       EffectRealPointsPerLevel, EffectBonusCoefficient, BonusCoefficientFromAP, EffectAuraPeriod,
		       COALESCE(EffectMiscValue_0, 0)
		FROM SpellEffect WHERE SpellID = ? ORDER BY EffectIndex`, spellID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RankEffect
	for rows.Next() {
		e := RankEffect{OwnerSpellID: spellID}
		if err := rows.Scan(&e.Index, &e.Effect, &e.Aura, &e.BasePoints, &e.DieSides, &e.PointsPerLvl, &e.Coefficient, &e.APCoef, &e.AuraPeriod, &e.MiscValue); err != nil {
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
