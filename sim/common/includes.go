package common

// Just import other directories, so importing common from elsewhere is enough.
import (
	"github.com/wowsims/tbc/sim/common/mop"
)

func RegisterAllEffects() {
	mop.RegisterAllOnUseCds()
	mop.RegisterAllProcs()
	mop.RegisterAllEnchants()
}
