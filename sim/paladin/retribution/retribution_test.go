package retribution

import (
	"testing"

	"github.com/wowsims/tbc/sim/common" // imported to get item effects included.
)

func init() {
	RegisterRetributionPaladin()
	common.RegisterAllEffects()
}

func TestRetribution(t *testing.T) {
}
