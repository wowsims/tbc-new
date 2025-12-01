package survival

import (
	"testing"

	"github.com/wowsims/tbc/sim/common" // imported to get item effects included.
)

func init() {
	RegisterSurvivalHunter()
	common.RegisterAllEffects()
}

func TestSurvival(t *testing.T) {
}
