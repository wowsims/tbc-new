package core

import (
	"testing"
	"time"
)

func newTestEnergyBar() (*Simulation, *energyBar) {
	sim := &Simulation{
		CurrentTime: time.Second,
	}
	eb := &energyBar{
		maxEnergy:             100,
		currentEnergy:         30,
		nextEnergyTick:        time.Millisecond * 1500,
		EnergyTickDuration:    time.Second * 2,
		EnergyPerTick:         20,
		energyRegenMultiplier: 1,
	}
	return sim, eb
}

func TestTimeToNextEnergyTick(t *testing.T) {
	sim, eb := newTestEnergyBar()

	if got := eb.TimeToNextEnergyTick(sim); got != time.Millisecond*500 {
		t.Fatalf("Unexpected time to next tick %s", got)
	}
}

func TestTimeToTargetEnergy(t *testing.T) {
	sim, eb := newTestEnergyBar()

	testCases := []struct {
		targetEnergy float64
		expected     time.Duration
	}{
		{targetEnergy: 20, expected: 0},                        // Already above target.
		{targetEnergy: 30, expected: 0},                        // Already at target.
		{targetEnergy: 40, expected: time.Millisecond * 500},   // Partial tick still rounds up to a full tick.
		{targetEnergy: 50, expected: time.Millisecond * 500},   // Exactly one tick.
		{targetEnergy: 51, expected: time.Millisecond * 2500},  // Just past one tick, so two are needed.
		{targetEnergy: 100, expected: time.Millisecond * 6500}, // Four ticks to reach max Energy.
		{targetEnergy: 101, expected: NeverExpires},            // Unreachable.
	}

	for _, testCase := range testCases {
		if got := eb.TimeToTargetEnergy(sim, testCase.targetEnergy); got != testCase.expected {
			t.Fatalf("Unexpected time to %0.0f energy: expected %s but was %s", testCase.targetEnergy, testCase.expected, got)
		}
	}
}

func TestTimeToTargetEnergyWithRegenMultiplier(t *testing.T) {
	sim, eb := newTestEnergyBar()
	eb.energyRegenMultiplier = 1.3 // 26 Energy per tick.

	if got := eb.TimeToTargetEnergy(sim, 51); got != time.Millisecond*500 {
		t.Fatalf("Unexpected time to 51 energy %s", got)
	}
	if got := eb.TimeToTargetEnergy(sim, 57); got != time.Millisecond*2500 {
		t.Fatalf("Unexpected time to 57 energy %s", got)
	}
}

func TestTimeToTargetEnergyWithoutRegen(t *testing.T) {
	sim, eb := newTestEnergyBar()
	eb.hasNoRegen = true

	if got := eb.TimeToTargetEnergy(sim, 50); got != NeverExpires {
		t.Fatalf("Unexpected time to 50 energy %s", got)
	}
}
