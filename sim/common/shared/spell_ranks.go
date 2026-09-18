package shared

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
)

// What a rank is worth, discriminated by shape rather than by a struct that carries every field and
// leaves the caller to guess which are meaningful. A talent's flat number has no Min and Max to
// misread, and only a periodic value has a tick schedule.
//
// The interface carries only what every shape can answer. Period and Ticks are reached by asserting
// to SpellRankPeriodic, rather than through a method the other two would have to answer with zeroes -
// which would put the guesswork back, one level up.
type SpellRankValue interface {
	// Returns the damage range of the spell.
	// Min/Max are the same if the spell only has a single value.
	Range() (float64, float64)

	// Returns the SP scaling coefficient
	BonusCoefficient() float64

	// Attack Power scaling coefficient
	// Defined manually, DBC does not hold this value - see WithSpellRankAPCoef.
	APBonusCoefficient() float64

	// Returns a static damage value or rolls between the min/max of the range.
	Damage(sim *core.Simulation) float64

	isSpellRankValue()
}

// A single number: a mana restore, a talent's value, damage the client does not roll.
type SpellRankFlat struct {
	Value  float64
	Coef   float64
	APCoef float64
}

// Damage or Healing "Direct/Impact"
type SpellRankRange struct {
	Min    float64
	Max    float64
	Coef   float64
	APCoef float64
}

// DoT
type SpellRankPeriodic struct {
	Tick   float64
	Coef   float64
	APCoef float64
	Period time.Duration
	Ticks  int32
}

func (v SpellRankFlat) Range() (float64, float64)     { return v.Value, v.Value }
func (v SpellRankRange) Range() (float64, float64)    { return v.Min, v.Max }
func (v SpellRankPeriodic) Range() (float64, float64) { return v.Tick, v.Tick }

func (v SpellRankFlat) Damage(_ *core.Simulation) float64     { return v.Value }
func (v SpellRankRange) Damage(sim *core.Simulation) float64  { return sim.Roll(v.Min, v.Max) }
func (v SpellRankPeriodic) Damage(_ *core.Simulation) float64 { return v.Tick }

func (v SpellRankFlat) BonusCoefficient() float64     { return v.Coef }
func (v SpellRankRange) BonusCoefficient() float64    { return v.Coef }
func (v SpellRankPeriodic) BonusCoefficient() float64 { return v.Coef }

func (v SpellRankFlat) APBonusCoefficient() float64     { return v.APCoef }
func (v SpellRankRange) APBonusCoefficient() float64    { return v.APCoef }
func (v SpellRankPeriodic) APBonusCoefficient() float64 { return v.APCoef }

// Convenience for the common reads, so a factory that knows its spell's shape is not forced through
// a two-value return.
//
// A nil value reads as zero rather than panicking: a rank can legitimately carry nothing in a role
// its ladder otherwise uses, as Lay on Hands rank 1 does by restoring no mana where ranks 2-4 do.
func SpellRankCoef(v SpellRankValue) float64 {
	if v == nil {
		return 0
	}
	return v.BonusCoefficient()
}

func SpellRankAPCoef(v SpellRankValue) float64 {
	if v == nil {
		return 0
	}
	return v.APBonusCoefficient()
}

func SpellRankMin(v SpellRankValue) float64 {
	if v == nil {
		return 0
	}
	min, _ := v.Range()
	return min
}

func SpellRankMax(v SpellRankValue) float64 {
	if v == nil {
		return 0
	}
	_, max := v.Range()
	return max
}

func (v SpellRankFlat) isSpellRankValue()     {}
func (v SpellRankRange) isSpellRankValue()    {}
func (v SpellRankPeriodic) isSpellRankValue() {}

// One rank of a spell, as the DBC describes it.
type SpellRank struct {
	Rank    int32
	SpellID int32
	Cost    int32
	CostPct float64

	// Cast times differ per rank - Fireball is 1.5s at rank 1 and 3.5s at rank 13 - so a downrank
	// cannot be registered faithfully without them.
	CastTime time.Duration
	GCD      time.Duration
	Cooldown time.Duration
	MaxRange float64
	Direct   SpellRankValue
	Heal     SpellRankValue
	Periodic SpellRankValue
	Energize SpellRankValue

	FlatThreatBonus float64
}

func (r SpellRank) GetRank() int32 { return r.Rank }

func (r SpellRank) GetSpellID() int32 { return r.SpellID }

func (r SpellRank) GetRankLabel() string { return fmt.Sprintf("Rank %d", r.Rank) }

type SpellRanked interface {
	GetRank() int32
	GetSpellID() int32
}

// A spell's ranks, keyed by rank number but held as a slice.
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

func (t SpellRankTableOf[T]) BySpellID(spellID int32) T {
	for _, row := range t {
		if row.GetSpellID() == spellID {
			return row
		}
	}
	panic(fmt.Sprintf("no spell %d in table of %d ranks", spellID, len(t)))
}

// The highest rank present in the DATA, which is not always the last element -
// Flamestrike is declared rank 7 then 6 - and is not necessarily a rank the game grants. See BySpellID.
func (t SpellRankTableOf[T]) HighestRank() T {
	if len(t) == 0 {
		panic("HighestRank() on an empty rank table")
	}

	best := t[0]
	for _, row := range t[1:] {
		if row.GetRank() > best.GetRank() {
			best = row
		}
	}
	return best
}

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
// The single-value forms give every rank the same coefficient. The plural forms take one per rank, the
// way the spell power coefficient already varies per rank because each row carries its own.
//
// All of them return a copy, so the generated table keeps whatever the database said, and all panic if
// that table already carries a coefficient: a value appearing upstream should be noticed rather than
// silently shadowed by the hand-written one.
func WithSpellRankAPCoef(table SpellRankTable, coef float64) SpellRankTable {
	return applyAPCoef(table, func(SpellRank) float64 { return coef }, apDirect)
}

func WithSpellRankPeriodicAPCoef(table SpellRankTable, coef float64) SpellRankTable {
	return applyAPCoef(table, func(SpellRank) float64 { return coef }, apPeriodic)
}

// Every rank in the table has to be named. A ladder that gains a rank then fails loudly instead of
// scaling the new one off nothing, which is the failure a map would otherwise hide.
func WithSpellRankAPCoefs(table SpellRankTable, coefs map[int32]float64) SpellRankTable {
	requireEveryRank(table, coefs)
	return applyAPCoef(table, func(r SpellRank) float64 { return coefs[r.Rank] }, apDirect)
}

func WithSpellRankPeriodicAPCoefs(table SpellRankTable, coefs map[int32]float64) SpellRankTable {
	requireEveryRank(table, coefs)
	return applyAPCoef(table, func(r SpellRank) float64 { return coefs[r.Rank] }, apPeriodic)
}

type apRole int

const (
	apDirect apRole = iota
	apPeriodic
)

func requireEveryRank(table SpellRankTable, coefs map[int32]float64) {
	for _, row := range table {
		if _, ok := coefs[row.Rank]; !ok {
			panic(fmt.Sprintf("spell %d rank %d has no AP coefficient in a per-rank map of %d",
				row.SpellID, row.Rank, len(coefs)))
		}
	}
	for rank := range coefs {
		found := false
		for _, row := range table {
			if row.Rank == rank {
				found = true
				break
			}
		}
		if !found {
			panic(fmt.Sprintf("AP coefficient given for rank %d, which the table does not have", rank))
		}
	}
}

func applyAPCoef(table SpellRankTable, coefOf func(SpellRank) float64, role apRole) SpellRankTable {
	out := make(SpellRankTable, len(table))
	for i, row := range table {
		out[i] = row
		value := row.Direct
		if role == apPeriodic {
			value = row.Periodic
		}

		updated := withAPCoef(value, coefOf(row), row.SpellID, row.Rank)
		if role == apPeriodic {
			out[i].Periodic = updated
		} else {
			out[i].Direct = updated
		}
	}
	return out
}

func withAPCoef(value SpellRankValue, coef float64, spellID, rank int32) SpellRankValue {
	if value == nil {
		panic(fmt.Sprintf("spell %d rank %d has no value to give an AP coefficient", spellID, rank))
	}
	if existing := value.APBonusCoefficient(); existing != 0 {
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
