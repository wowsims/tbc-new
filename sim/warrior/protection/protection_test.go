package protection

import (
	"testing"

	"github.com/wowsims/tbc/sim/common" // imported to get item effects included.
)

func init() {
	RegisterProtectionWarrior()
	common.RegisterAllEffects()
}

func TestProtectionWarrior(t *testing.T) {
}
