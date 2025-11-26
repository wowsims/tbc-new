package combat

import (
	"testing"

	_ "github.com/wowsims/tbc/sim/common" // imported to get item effects included.
)

func init() {
	RegisterCombatRogue()
}

func TestCombat(t *testing.T) {
}
