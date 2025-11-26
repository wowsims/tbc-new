package beast_mastery

import (
	"testing"

	"github.com/wowsims/tbc/sim/common" // imported to get item effects included.
)

func init() {
	RegisterBeastMasteryHunter()
	common.RegisterAllEffects()
}

func TestBeastMastery(t *testing.T) {
}
