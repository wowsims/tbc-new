package frost

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/mage"
)

func (frost *FrostMage) registerHotfixes() {
	// 2025-09-22 - Frostbolt/Frostfire bolt damage increased by 15%
	frost.AddStaticMod(core.SpellModConfig{
		ClassMask:  mage.MageSpellFrostbolt | mage.MageSpellFrostfireBolt | mage.MageSpellIceLance,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.15,
	})
}
