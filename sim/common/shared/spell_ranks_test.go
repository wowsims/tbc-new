package shared

import (
	"math"
	"testing"
)

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
			{Index: 0, Effect: E_APPLY_AURA, Aura: A_ADD_PCT_MODIFIER, Misc: 8, Value: 50},
			{Index: 1, Effect: E_APPLY_AURA, Aura: A_ADD_FLAT_MODIFIER, Misc: 12, Value: -6},
		},
		Direct: SpellRankFlat{Value: 50, Coef: 1},
	}
}

func TestEffectPicksByAura(t *testing.T) {
	rank := effectTestRank()
	if got := rank.Effect(A_ADD_PCT_MODIFIER, 8).Value; got != 50 {
		t.Errorf("threat effect: want 50, got %v", got)
	}
	if got := rank.Effect(A_ADD_FLAT_MODIFIER, 12).Value; got != -6 {
		t.Errorf("damage-taken effect: want -6, got %v", got)
	}
}

func TestEffectMissingPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a panic for an aura the rank does not carry")
		}
	}()
	effectTestRank().Effect(A_MOD_DAMAGE_PERCENT_DONE, 0)
}

// 186 ranked spells in this build carry two effects with the same aura and misc value. Returning the
// first is how a caller silently reads the wrong one.
func TestEffectAmbiguousPanics(t *testing.T) {
	rank := SpellRank{
		Rank: 1, SpellID: 1,
		Effects: []SpellRankEffect{
			{Index: 0, Effect: E_APPLY_AURA, Aura: A_MOD_DAMAGE_PERCENT_DONE, Misc: 0, Value: 10},
			{Index: 1, Effect: E_APPLY_AURA, Aura: A_MOD_DAMAGE_PERCENT_DONE, Misc: 0, Value: 20},
		},
	}
	defer func() {
		if recover() == nil {
			t.Error("expected a panic for two effects sharing an aura and misc value")
		}
	}()
	rank.Effect(A_MOD_DAMAGE_PERCENT_DONE, 0)
}

func talentLadder() SpellRankTable {
	return SpellRankTable{
		{Rank: 1, SpellID: 20468, Effects: []SpellRankEffect{
			{Index: 0, Effect: E_APPLY_AURA, Aura: A_ADD_PCT_MODIFIER, Misc: SPELLMOD_ALL_EFFECTS, Value: 16},
			{Index: 1, Effect: E_APPLY_AURA, Aura: A_ADD_FLAT_MODIFIER, Misc: SPELLMOD_EFFECT2, Value: -2}}},
		{Rank: 2, SpellID: 20469, Effects: []SpellRankEffect{
			{Index: 0, Effect: E_APPLY_AURA, Aura: A_ADD_PCT_MODIFIER, Misc: SPELLMOD_ALL_EFFECTS, Value: 33},
			{Index: 1, Effect: E_APPLY_AURA, Aura: A_ADD_FLAT_MODIFIER, Misc: SPELLMOD_EFFECT2, Value: -4}}},
		{Rank: 3, SpellID: 20470, Effects: []SpellRankEffect{
			{Index: 0, Effect: E_APPLY_AURA, Aura: A_ADD_PCT_MODIFIER, Misc: SPELLMOD_ALL_EFFECTS, Value: 50},
			{Index: 1, Effect: E_APPLY_AURA, Aura: A_ADD_FLAT_MODIFIER, Misc: SPELLMOD_EFFECT2, Value: -6}}},
	}
}

// An untaken talent is rank 0, which ByRank would panic on.
func TestLadderUntakenTalentIsZero(t *testing.T) {
	table := talentLadder()
	if got := table.Effect(A_ADD_PCT_MODIFIER, SPELLMOD_ALL_EFFECTS).ValueAt(0); got != 0 {
		t.Errorf("rank 0 value: want 0, got %v", got)
	}
	if got := table.Effect(A_ADD_FLAT_MODIFIER, SPELLMOD_EFFECT2).MultiplierAt(0); got != 1 {
		t.Errorf("rank 0 multiplier: want 1, got %v", got)
	}
}

// The ladder is 16/33/50, not 16/32/48, which is what a per-point literal would give.
func TestLadderIsNotPerPointTimesRank(t *testing.T) {
	threat := talentLadder().Effect(A_ADD_PCT_MODIFIER, SPELLMOD_ALL_EFFECTS)
	for rank, want := range map[int32]float64{1: 1.16, 2: 1.33, 3: 1.50} {
		if got := threat.MultiplierAt(rank); math.Abs(got-want) > 1e-9 {
			t.Errorf("rank %d: want %v, got %v", rank, want, got)
		}
	}
}

// The client states the reduction negative, so the caller never writes the minus.
func TestLadderMultiplierTakesItsSignFromTheData(t *testing.T) {
	if got := talentLadder().Effect(A_ADD_FLAT_MODIFIER, SPELLMOD_EFFECT2).MultiplierAt(3); math.Abs(got-0.94) > 1e-9 {
		t.Errorf("want 0.94, got %v", got)
	}
}

func TestLadderUnnamedEffectOnMultiEffectTalentPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a panic for an unnamed read of a two-effect talent")
		}
	}()
	talentLadder().ValueAt(3)
}
