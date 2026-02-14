package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

// Forbearance
// https://www.wowhead.com/tbc/spell=25771
//
// Cannot be made invulnerable by Divine Shield, Divine Protection, Blessing of Protection or be affected by Avenging Wrath.
func (paladin *Paladin) registerForbearance() {
	paladin.Forbearance = paladin.RegisterAura(core.Aura{
		Label:    "Forbearance",
		ActionID: core.ActionID{SpellID: 25771},
		Duration: time.Minute,
	})
}