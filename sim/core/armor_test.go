package core

import (
	"testing"

	"github.com/wowsims/tbc/sim/core/stats"
)

func TestFixedArmorReduction(t *testing.T) {
	baseArmor := 7684.0
	target := Unit{
		Type:         EnemyUnit,
		Index:        0,
		Level:        CharacterLevel + 3,
		auraTracker:  newAuraTracker(),
		initialStats: stats.Stats{stats.Armor: baseArmor},
		PseudoStats:  stats.NewPseudoStats(),
		Metrics:      NewUnitMetrics(),
	}
	target.stats = target.initialStats

	expectedArmor := baseArmor
	if target.Armor() != expectedArmor {
		t.Fatalf("Armor value for target should be %f but found %f", 7684.0, target.Armor())
	}
	tolerance := 0.001
	target.stats[stats.Armor] -= 610.0

	// Apply a fixed armor reduction of 610 to simulate Faerie Fire.
	expectedArmor = baseArmor - 610
	if !WithinToleranceFloat64(expectedArmor, target.Armor(), tolerance) {
		t.Fatalf("Armor value for target should be %f but found %f", expectedArmor, target.Armor())
	}
}

func TestDamageReductionFromArmor(t *testing.T) {
	baseArmor := 7684.0
	target := Unit{
		Type:         EnemyUnit,
		Index:        0,
		Level:        CharacterLevel + 3,
		auraTracker:  newAuraTracker(),
		initialStats: stats.Stats{stats.Armor: baseArmor},
		PseudoStats:  stats.NewPseudoStats(),
		Metrics:      NewUnitMetrics(),
	}
	attacker := Unit{
		Type:  PlayerUnit,
		Level: CharacterLevel,
	}
	target.stats = target.initialStats
	expectedDamageReduction := 0.421237
	attackTable := NewAttackTable(&attacker, &target)

	tolerance := 0.0001
	if !WithinToleranceFloat64(1-expectedDamageReduction, attackTable.getArmorDamageModifier(), tolerance) {
		t.Fatalf("Expected no armor modifiers to result in %f damage reduction got %f", expectedDamageReduction, 1-attackTable.getArmorDamageModifier())
	}
}
