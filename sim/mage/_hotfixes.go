package mage

import "github.com/wowsims/tbc/sim/core"

func (mage *Mage) registerHotfixes() {
	// 2013-09-23 Ice Lance's damage has been increased by 20%
	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellIceLance,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.2,
	})
}
