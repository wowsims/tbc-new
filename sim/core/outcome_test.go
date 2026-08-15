package core

import (
	"testing"

	"github.com/wowsims/tbc/sim/core/stats"
)

func TestCritChancesSeparateResilienceFromDefense(t *testing.T) {
	tests := []struct {
		name                 string
		rawChance            float64
		reducedCritTaken     float64
		resilienceReduction  float64
		wantActual, wantSupp float64
	}{
		{
			name:                "resilience immunity",
			rawChance:           0.05,
			reducedCritTaken:    0.05,
			resilienceReduction: 0.05,
			wantActual:          0,
			wantSupp:            0.05,
		},
		{
			name:                "defense immunity",
			rawChance:           0.05,
			reducedCritTaken:    0.05,
			resilienceReduction: 0,
			wantActual:          0,
			wantSupp:            0,
		},
		{
			name:                "mixed reduction",
			rawChance:           0.05,
			reducedCritTaken:    0.06,
			resilienceReduction: 0.03,
			wantActual:          0,
			wantSupp:            0.03,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			target := &Unit{PseudoStats: stats.NewPseudoStats()}
			target.PseudoStats.ReducedCritTakenPercent = test.reducedCritTaken
			target.AddStats(stats.Stats{stats.ResilienceRating: test.resilienceReduction * ResilienceRatingPerCritReductionChance * 100})

			actual := getCritChances(test.rawChance, target)
			if actual.actual != test.wantActual || actual.suppressed != test.wantSupp {
				t.Fatalf("got actual=%v, suppressed=%v; want actual=%v, suppressed=%v", actual.actual, actual.suppressed, test.wantActual, test.wantSupp)
			}
		})
	}
}

func TestSuppressedCritProc(t *testing.T) {
	unit := &Unit{}
	aura := &Aura{Unit: unit}
	spell := &Spell{ProcMask: ProcMaskMelee}
	result := &SpellResult{Outcome: OutcomeSuppressedCrit}
	procced := false

	aura.AttachProcTriggerCallback(unit, ProcTrigger{
		Callback:           CallbackOnSpellHitTaken,
		ProcMask:           ProcMaskMelee,
		Outcome:            OutcomeCrit,
		TriggerImmediately: true,
		Handler: func(_ *Simulation, _ *Spell, _ *SpellResult) {
			procced = true
		},
	})
	aura.OnSpellHitTaken(aura, nil, spell, result)

	if !procced {
		t.Fatal("resilience-suppressed crit did not trigger a crit proc")
	}
}

func TestSuppressedCritIsNotActualCrit(t *testing.T) {
	result := &SpellResult{Outcome: OutcomeSuppressedCrit}
	if result.DidCrit() {
		t.Fatal("suppressed crit reported as actual crit")
	}
	if !result.DidSuppressedCrit() {
		t.Fatal("suppressed crit was not identified")
	}
	if !result.Landed() {
		t.Fatal("suppressed crit was not treated as landed")
	}
}
