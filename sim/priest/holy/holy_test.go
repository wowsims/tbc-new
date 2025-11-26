package holy

import (
	"testing"

	_ "github.com/wowsims/tbc/sim/common" // imported to get caster sets included.
)

func init() {
	RegisterHolyPriest()
}

func TestSmite(t *testing.T) {
}
