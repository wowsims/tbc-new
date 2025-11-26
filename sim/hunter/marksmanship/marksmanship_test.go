package marksmanship

import (
	"testing"

	"github.com/wowsims/tbc/sim/common" // imported to get item effects included.
)

func init() {
	RegisterMarksmanshipHunter()
	common.RegisterAllEffects()
}

func TestMarksmanship(t *testing.T) {
}
