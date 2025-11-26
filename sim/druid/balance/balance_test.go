package balance

import (
	"testing"

	_ "github.com/wowsims/tbc/sim/common" // imported to get caster sets included. (we use spellfire here)
)

func init() {
	RegisterBalanceDruid()
}

func TestBalance(t *testing.T) {
}
