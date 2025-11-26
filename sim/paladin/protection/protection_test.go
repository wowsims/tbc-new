package protection

import (
	"testing"

	"github.com/wowsims/tbc/sim/common" // imported to get item effects included.
)

func init() {
	RegisterProtectionPaladin()
	common.RegisterAllEffects()
}

func TestProtection(t *testing.T) {
}
