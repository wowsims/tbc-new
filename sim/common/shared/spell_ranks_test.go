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

// Improved Righteous Fury's shape: one talent, two effects, and the role fields can only hold one.
func effectTestRank() SpellRank {
	return SpellRank{
		Rank: 3, SpellID: 20470,
		Effects: []SpellRankEffect{
			{Index: 0, Effect: 6, Aura: 108, Misc: 8, Value: 50},
			{Index: 1, Effect: 6, Aura: 107, Misc: 12, Value: -6},
		},
		Direct: SpellRankFlat{Value: 50, Coef: 1},
	}
}

func TestEffectPicksByAura(t *testing.T) {
	rank := effectTestRank()
	if got := rank.Effect(108, 8).Value; got != 50 {
		t.Errorf("threat effect: want 50, got %v", got)
	}
	if got := rank.Effect(107, 12).Value; got != -6 {
		t.Errorf("damage-taken effect: want -6, got %v", got)
	}
}

func TestEffectMissingPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a panic for an aura the rank does not carry")
		}
	}()
	effectTestRank().Effect(79, 0)
}

// 186 ranked spells in this build carry two effects with the same aura and misc value. Returning the
// first is how a caller silently reads the wrong one.
func TestEffectAmbiguousPanics(t *testing.T) {
	rank := SpellRank{
		Rank: 1, SpellID: 1,
		Effects: []SpellRankEffect{
			{Index: 0, Effect: 6, Aura: 79, Misc: 0, Value: 10},
			{Index: 1, Effect: 6, Aura: 79, Misc: 0, Value: 20},
		},
	}
	defer func() {
		if recover() == nil {
			t.Error("expected a panic for two effects sharing an aura and misc value")
		}
	}()
	rank.Effect(79, 0)
}
