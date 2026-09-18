package shared

import "fmt"

type Amount struct {
	Min  float64
	Max  float64
	Coef float64
}

type Periodic struct {
	Tick float64
	Coef float64
}

// One rank of a spell, as the client database describes it.
//
// The value lives in the field named for the effect that carries it rather than in one shared
// MinDamage: the same hand-written field currently means direct damage for Exorcism, a periodic tick
// for Consecration, a heal for Holy Light and a mana restore for Lay on Hands, and nothing at the call
// site said which.
type RankRow struct {
	Rank     int32
	SpellID  int32
	Cost     int32
	CostPct  float64
	Direct   *Amount
	Heal     *Amount
	Periodic *Periodic
	Energize float64

	FlatThreatBonus float64
}

func (r RankRow) GetRank() int32 { return r.Rank }

func (r RankRow) GetRankLabel() string { return fmt.Sprintf("Rank %d", r.Rank) }

type Ranked interface {
	GetRank() int32
}

// A spell's ranks, keyed by rank number but held as a slice.
//
// Not a map: iteration order would vary per run, which would reorder the Spellbook and so change which
// spell GetSpell returns wherever two share an ActionID - as the six Seal of Command procs do.
//
// Declaration order is preserved rather than sorted, because it is registration order: Flamestrike
// declares rank 7 before rank 6, and re-sorting it would quietly change the Spellbook.
type RankedTable[T Ranked] []T

// Panics on a rank the table does not have. A downrank chosen by typo has to fail loudly rather than
// silently register nothing.
func (t RankedTable[T]) ByRank(rank int32) T {
	for _, row := range t {
		if row.GetRank() == rank {
			return row
		}
	}
	panic(fmt.Sprintf("no rank %d in table of %d ranks", rank, len(t)))
}

// The highest rank, which is not always the last element - Flamestrike is declared 7 then 6.
func (t RankedTable[T]) Max() T {
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
func (t RankedTable[T]) Ranks(ranks ...int32) RankedTable[T] {
	out := make(RankedTable[T], 0, len(ranks))
	for _, rank := range ranks {
		out = append(out, t.ByRank(rank))
	}
	return out
}

func (t RankedTable[T]) RegisterAll(factory func(T)) {
	for _, row := range t {
		factory(row)
	}
}

type RankTable = RankedTable[RankRow]
