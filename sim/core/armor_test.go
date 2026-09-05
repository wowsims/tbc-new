package core

import (
	"testing"

	"github.com/wowsims/tbc/sim/core/stats"
)

func TestArmorDamageReductionCap(t *testing.T) {
	// Boss attacker: armorConstant = 73*467.5 - 22167.5 = 11960, so 75% reduction is
	// reached at 3*11960 = 35880 armor.
	attacker := Unit{Type: EnemyUnit, Level: 73}
	tolerance := 0.0001

	modifierForArmor := func(armor float64) float64 {
		defender := Unit{
			Type:         PlayerUnit,
			Level:        70,
			initialStats: stats.Stats{stats.Armor: armor},
			PseudoStats:  stats.NewPseudoStats(),
		}
		defender.stats = defender.initialStats
		return NewAttackTable(&attacker, &defender).GetArmorDamageModifier(nil)
	}

	if modifier := modifierForArmor(23920); !WithinToleranceFloat64(1.0/3.0, modifier, tolerance) {
		t.Fatalf("Expected %f damage taken below the cap, got %f", 1.0/3.0, modifier)
	}
	if modifier := modifierForArmor(35880); !WithinToleranceFloat64(0.25, modifier, tolerance) {
		t.Fatalf("Expected %f damage taken at the cap, got %f", 0.25, modifier)
	}
	if modifier := modifierForArmor(50000); !WithinToleranceFloat64(0.25, modifier, tolerance) {
		t.Fatalf("Expected %f damage taken above the cap, got %f", 0.25, modifier)
	}
}
