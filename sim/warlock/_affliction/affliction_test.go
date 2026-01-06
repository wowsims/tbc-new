package affliction

import (
	"testing"

	_ "unsafe"

	"github.com/wowsims/tbc/sim/common"
)

func init() {
	RegisterAfflictionWarlock()
	common.RegisterAllEffects()
}

func TestAffliction(t *testing.T) {
}
