package shared

import (
	"fmt"
	"time"
)

// What a rank is worth, discriminated by shape rather than by a struct that carries every field and
// leaves the caller to guess which are meaningful. A talent's flat number has no Min and Max to
// misread, and only a periodic value has a tick schedule.
type SpellRankValue interface {
	// The amount, as a range. A flat value and a periodic tick report the same number twice, so a
	// caller that just wants to roll does not have to know which shape it was handed.
	Amount() (float64, float64)

	// Spell power, then attack power.
	Coefficients() (float64, float64)

	// Tick length and count, both zero unless the value is periodic.
	Schedule() (time.Duration, int32)

	isSpellRankValue()
}

// A single number: a mana restore, a talent's value, damage the client does not roll.
type SpellRankFlat struct {
	Value  float64
	Coef   float64
	APCoef float64
}

// Damage or healing rolled between two ends.
type SpellRankRange struct {
	Min    float64
	Max    float64
	Coef   float64
	APCoef float64
}

// A tick, and the schedule it lands on. Ticks is Duration/Period as the client states them, which is
// how every hand-written NumberOfTicks in the sim was arrived at.
type SpellRankPeriodic struct {
	Tick   float64
	Coef   float64
	APCoef float64
	Period time.Duration
	Ticks  int32
}

func (v SpellRankFlat) Amount() (float64, float64)     { return v.Value, v.Value }
func (v SpellRankRange) Amount() (float64, float64)    { return v.Min, v.Max }
func (v SpellRankPeriodic) Amount() (float64, float64) { return v.Tick, v.Tick }

func (v SpellRankFlat) Coefficients() (float64, float64)     { return v.Coef, v.APCoef }
func (v SpellRankRange) Coefficients() (float64, float64)    { return v.Coef, v.APCoef }
func (v SpellRankPeriodic) Coefficients() (float64, float64) { return v.Coef, v.APCoef }

func (v SpellRankFlat) Schedule() (time.Duration, int32)     { return 0, 0 }
func (v SpellRankRange) Schedule() (time.Duration, int32)    { return 0, 0 }
func (v SpellRankPeriodic) Schedule() (time.Duration, int32) { return v.Period, v.Ticks }

// Convenience for the common reads, so a factory that knows its spell's shape is not forced through
// a two-value return.
//
// A nil value reads as zero rather than panicking: a rank can legitimately carry nothing in a role
// its ladder otherwise uses, as Lay on Hands rank 1 does by restoring no mana where ranks 2-4 do.
func SpellRankCoef(v SpellRankValue) float64 {
	if v == nil {
		return 0
	}
	c, _ := v.Coefficients()
	return c
}

func SpellRankAPCoef(v SpellRankValue) float64 {
	if v == nil {
		return 0
	}
	_, ap := v.Coefficients()
	return ap
}

func SpellRankMin(v SpellRankValue) float64 {
	if v == nil {
		return 0
	}
	min, _ := v.Amount()
	return min
}

func SpellRankMax(v SpellRankValue) float64 {
	if v == nil {
		return 0
	}
	_, max := v.Amount()
	return max
}

func (v SpellRankFlat) isSpellRankValue()     {}
func (v SpellRankRange) isSpellRankValue()    {}
func (v SpellRankPeriodic) isSpellRankValue() {}

// One rank of a spell, as the client database describes it.
//
// The value lives in the field named for the effect that carries it rather than in one shared
// MinDamage: the same hand-written field currently means direct damage for Exorcism, a periodic tick
// for Consecration, a heal for Holy Light and a mana restore for Lay on Hands, and nothing at the call
// site said which.
type SpellRank struct {
	Rank    int32
	SpellID int32
	Cost    int32
	CostPct float64

	// Zero until SpellCastTimes is extracted; the table is absent from this build's database, so the
	// generator has nothing to read. Cast times genuinely differ per rank - Fireball's ranks point at
	// casting-time indices 16, 14 and 22 - so a downrank cannot be registered faithfully without it.
	CastTime time.Duration
	Direct   SpellRankValue
	Heal     SpellRankValue
	Periodic SpellRankValue
	Energize SpellRankValue

	FlatThreatBonus float64
}

func (r SpellRank) GetRank() int32 { return r.Rank }

func (r SpellRank) GetRankLabel() string { return fmt.Sprintf("Rank %d", r.Rank) }

type SpellRanked interface {
	GetRank() int32
}

// A spell's ranks, keyed by rank number but held as a slice.
//
// Not a map: iteration order would vary per run, which would reorder the Spellbook and so change which
// spell GetSpell returns wherever two share an ActionID - as the six Seal of Command procs do.
//
// Declaration order is preserved rather than sorted, because it is registration order: Flamestrike
// declares rank 7 before rank 6, and re-sorting it would quietly change the Spellbook.
type SpellRankTableOf[T SpellRanked] []T

// Panics on a rank the table does not have. A downrank chosen by typo has to fail loudly rather than
// silently register nothing.
func (t SpellRankTableOf[T]) ByRank(rank int32) T {
	for _, row := range t {
		if row.GetRank() == rank {
			return row
		}
	}
	panic(fmt.Sprintf("no rank %d in table of %d ranks", rank, len(t)))
}

// The highest rank, which is not always the last element - Flamestrike is declared 7 then 6.
func (t SpellRankTableOf[T]) Max() T {
	if len(t) == 0 {
		panic("Max() on an empty rank table")
	}

	best := t[0]
	for _, row := range t[1:] {
		if row.GetRank() > best.GetRank() {
			best = row
		}
	}
	return best
}

// The named ranks, in the order named - which becomes registration order, so it is the caller's to
// choose rather than something to normalise.
func (t SpellRankTableOf[T]) Ranks(ranks ...int32) SpellRankTableOf[T] {
	out := make(SpellRankTableOf[T], 0, len(ranks))
	for _, rank := range ranks {
		out = append(out, t.ByRank(rank))
	}
	return out
}

func (t SpellRankTableOf[T]) RegisterAll(factory func(T)) {
	for _, row := range t {
		factory(row)
	}
}

type SpellRankTable = SpellRankTableOf[SpellRank]

// Attack power scaling has to be supplied by hand. This build's client data carries a nonzero
// BonusCoefficientFromAP on exactly one effect out of 38357 - for everything else the coefficient lives
// in server script, which is why melee spells in this sim still hardcode theirs (sim/druid/rip.go:52
// reads 990 + 0.18*ap).
//
// Returns a copy, so the generated table keeps whatever the database said, and panics if that table
// already carries a coefficient: a value appearing upstream should be noticed rather than silently
// shadowed by the hand-written one.
func WithSpellRankAPCoef(table SpellRankTable, coef float64) SpellRankTable {
	out := make(SpellRankTable, len(table))
	for i, row := range table {
		out[i] = row
		out[i].Direct = withAPCoef(row.Direct, coef, row.SpellID, row.Rank)
	}
	return out
}

func WithSpellRankPeriodicAPCoef(table SpellRankTable, coef float64) SpellRankTable {
	out := make(SpellRankTable, len(table))
	for i, row := range table {
		out[i] = row
		out[i].Periodic = withAPCoef(row.Periodic, coef, row.SpellID, row.Rank)
	}
	return out
}

func withAPCoef(value SpellRankValue, coef float64, spellID, rank int32) SpellRankValue {
	if value == nil {
		panic(fmt.Sprintf("spell %d rank %d has no value to give an AP coefficient", spellID, rank))
	}
	if _, existing := value.Coefficients(); existing != 0 {
		panic(fmt.Sprintf("spell %d rank %d already has AP coefficient %v from the client DB", spellID, rank, existing))
	}

	switch v := value.(type) {
	case SpellRankFlat:
		v.APCoef = coef
		return v
	case SpellRankRange:
		v.APCoef = coef
		return v
	case SpellRankPeriodic:
		v.APCoef = coef
		return v
	}
	panic(fmt.Sprintf("spell %d rank %d has an unknown value shape %T", spellID, rank, value))
}
