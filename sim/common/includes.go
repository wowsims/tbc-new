package common

// Just import other directories, so importing common from elsewhere is enough.
import (
	"github.com/wowsims/tbc/sim/common/tbc"
)

func RegisterAllEffects() {
	tbc.RegisterAllOnUseCds()
	tbc.RegisterAllProcs()
	tbc.RegisterAllEnchants()
}
