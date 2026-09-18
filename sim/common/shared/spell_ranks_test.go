package shared

import "testing"

// Attack power coefficients are hand-supplied, so the guards around them are the only thing standing
// between a typo and a spell that silently scales off nothing.
func apCoefTestTable() SpellRankTable {
	return SpellRankTable{
		{Rank: 1, SpellID: 100, Direct: SpellRankRange{Min: 10, Max: 20, Coef: 0.1}},
		{Rank: 2, SpellID: 200, Direct: SpellRankRange{Min: 30, Max: 40, Coef: 0.2}},
	}
}

func TestPerRankAPCoefficients(t *testing.T) {
	src := apCoefTestTable()
	out := WithSpellRankAPCoefs(src, map[int32]float64{1: 0.05, 2: 0.15})

	if got := SpellRankAPCoef(out[0].Direct); got != 0.05 {
		t.Errorf("rank 1 AP coef = %v, want 0.05", got)
	}
	if got := SpellRankAPCoef(out[1].Direct); got != 0.15 {
		t.Errorf("rank 2 AP coef = %v, want 0.15", got)
	}
	if got := SpellRankAPCoef(src[0].Direct); got != 0 {
		t.Errorf("source table was mutated: %v", got)
	}
	if got := SpellRankCoef(out[1].Direct); got != 0.2 {
		t.Errorf("SP coef lost: %v", got)
	}
}

func TestAPCoefficientMissingRankPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a panic for a rank with no coefficient")
		}
	}()
	WithSpellRankAPCoefs(apCoefTestTable(), map[int32]float64{1: 0.05})
}

func TestAPCoefficientUnknownRankPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a panic for a coefficient naming a rank that does not exist")
		}
	}()
	WithSpellRankAPCoefs(apCoefTestTable(), map[int32]float64{1: 0.05, 2: 0.15, 7: 0.25})
}
