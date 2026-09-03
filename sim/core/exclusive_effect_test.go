package core

import (
	"testing"
	"time"
)

func newExclusiveTestTarget() *Unit {
	return &Unit{
		Type:        EnemyUnit,
		Index:       0,
		Level:       73,
		auraTracker: newAuraTracker(),
	}
}

// Registers an aura with an exclusive effect in "TestCategory" whose callbacks
// track the currently applied value via *applied, without touching unit stats
// (dynamic stat changes require a full Environment).
func makeTestExclusiveAura(target *Unit, label string, duration time.Duration, priority float64, applied *float64) *Aura {
	aura := target.GetOrRegisterAura(Aura{
		Label:    label,
		ActionID: ActionID{SpellID: 100000 + int32(len(target.auras))},
		Duration: duration,
	})
	aura.NewExclusiveEffect("TestCategory", true, ExclusiveEffect{
		Priority: priority,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			*applied += ee.Priority
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			*applied -= ee.Priority
		},
	})
	return aura
}

func TestSingleAuraExclusiveDurationNoOverwrite(t *testing.T) {
	sim := &Simulation{}
	target := newExclusiveTestTarget()
	var applied float64

	permanent := MakePermanent(makeTestExclusiveAura(target, "Permanent Breath", time.Second*15, 1, &applied))
	short := makeTestExclusiveAura(target, "Short Breath", time.Second*15, 1, &applied)

	// The permanent aura should never be overwritten by an equal-priority
	// aura with a shorter duration.
	permanent.Activate(sim)

	sim.CurrentTime = 1 * time.Second

	short.Activate(sim)

	if !(permanent.IsActive() && !short.IsActive()) {
		t.Fatalf("lower duration exclusive aura overwrote previous!")
	}
	if applied != 1 {
		t.Fatalf("expected applied value 1, got %f", applied)
	}
}

func TestSingleAuraExclusiveDurationOverwrite(t *testing.T) {
	sim := &Simulation{}
	target := newExclusiveTestTarget()
	var applied float64

	short := makeTestExclusiveAura(target, "Short Breath", time.Second*15, 1, &applied)
	long := makeTestExclusiveAura(target, "Long Breath", time.Second*30, 1, &applied)

	short.Activate(sim)

	sim.CurrentTime = 1 * time.Second

	// Equal priority but longer duration than the active aura's remaining
	// duration should overwrite it.
	long.Activate(sim)

	if !(long.IsActive() && !short.IsActive()) {
		t.Fatalf("longer duration exclusive aura failed to overwrite")
	}
	if applied != 1 {
		t.Fatalf("expected applied value 1 after overwrite, got %f", applied)
	}
}

func TestSingleAuraExclusiveHigherPriorityOverwrites(t *testing.T) {
	sim := &Simulation{}
	target := newExclusiveTestTarget()
	var applied float64

	weak := makeTestExclusiveAura(target, "Weak Debuff", time.Second*30, 1, &applied)
	strong := makeTestExclusiveAura(target, "Strong Debuff", time.Second*15, 2, &applied)

	weak.Activate(sim)
	strong.Activate(sim)

	if !(strong.IsActive() && !weak.IsActive()) {
		t.Fatalf("higher priority exclusive aura failed to overwrite")
	}
	if applied != 2 {
		t.Fatalf("expected applied value 2, got %f", applied)
	}

	strong.Deactivate(sim)
	if applied != 0 {
		t.Fatalf("expected applied value 0 after deactivation, got %f", applied)
	}
}

func TestNewExclusiveEffectDedupsPerAura(t *testing.T) {
	target := newExclusiveTestTarget()
	var applied float64

	aura := makeTestExclusiveAura(target, "Dedup Debuff", time.Second*10, 5, &applied)
	first := aura.ExclusiveEffects[0]

	second := aura.NewExclusiveEffect("TestCategory", true, ExclusiveEffect{Priority: 99})

	if second != first {
		t.Fatalf("duplicate (category, aura) registration created a new effect")
	}
	if len(aura.ExclusiveEffects) != 1 {
		t.Fatalf("expected 1 exclusive effect, got %d", len(aura.ExclusiveEffects))
	}
	// The dedup keeps the first registrant's priority; callers that need a
	// higher value must bump it manually (see FaerieFireAura).
	if first.Priority != 5 {
		t.Fatalf("expected priority 5, got %f", first.Priority)
	}
}

// Regression for PR #425: the raid debuff config registers Faerie Fire before
// player Initialize(), so a talented druid's improved value must survive the
// (category, aura) dedup regardless of registration order.
func TestFaerieFireImprovedSurvivesDedup(t *testing.T) {
	target := newExclusiveTestTarget()
	configAura := FaerieFireAura(target, 0)
	druidAura := FaerieFireAura(target, 3)

	if configAura != druidAura {
		t.Fatalf("expected both Faerie Fire registrations to share one aura")
	}
	if prio := configAura.ExclusiveEffects[0].Priority; prio != 3 {
		t.Fatalf("expected improved Faerie Fire priority 3, got %f", prio)
	}

	reversed := newExclusiveTestTarget()
	FaerieFireAura(reversed, 3)
	aura := FaerieFireAura(reversed, 0)
	if prio := aura.ExclusiveEffects[0].Priority; prio != 3 {
		t.Fatalf("unimproved registration lowered priority to %f", prio)
	}
}

// Regression: the config and druid Demoralizing Roar auras used to have
// different labels and no exclusive category, double-dipping the AP reduction.
func TestDemoralizingRoarSharedAura(t *testing.T) {
	target := newExclusiveTestTarget()
	configAura := DemoralizingRoarAura(target, 0)
	druidAura := DemoralizingRoarAura(target, 5)

	if configAura != druidAura {
		t.Fatalf("expected both Demoralizing Roar registrations to share one aura")
	}
	expected := 248.0 * (1 + 0.08*5)
	if prio := configAura.ExclusiveEffects[0].Priority; prio != expected {
		t.Fatalf("expected Demoralizing Roar priority %f, got %f", expected, prio)
	}
}

// Demoralizing Roar and Demoralizing Shout are mutually exclusive — the
// stronger of the two wins via shared category priority.
func TestDemoralizingRoarAndShoutShareCategory(t *testing.T) {
	target := newExclusiveTestTarget()
	roar := DemoralizingRoarAura(target, 5)
	shout := DemoralizingShoutAura(target, 0, 5)

	if roar.ExclusiveEffects[0].Category != shout.ExclusiveEffects[0].Category {
		t.Fatalf("expected Demoralizing Roar and Shout to share an exclusive category")
	}
}

// Regression: the config's Demoralizing Shout registered first and clobbered a
// talented warrior's stronger values via the shared-label dedup.
func TestDemoralizingShoutTalentsSurviveDedup(t *testing.T) {
	target := newExclusiveTestTarget()
	configAura := DemoralizingShoutAura(target, 0, 0)
	warriorAura := DemoralizingShoutAura(target, 5, 5)

	if configAura != warriorAura {
		t.Fatalf("expected both Demoralizing Shout registrations to share one aura")
	}
	expectedPrio := 300.0 * (1 + 0.1*5)
	if prio := configAura.ExclusiveEffects[0].Priority; prio != expectedPrio {
		t.Fatalf("expected Demoralizing Shout priority %f, got %f", expectedPrio, prio)
	}
	expectedDuration := time.Duration(float64(time.Second*30) * (1 + 0.1*5))
	if configAura.Duration != expectedDuration {
		t.Fatalf("expected Demoralizing Shout duration %v, got %v", expectedDuration, configAura.Duration)
	}
}
