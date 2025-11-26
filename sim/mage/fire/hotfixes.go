package fire

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/mage"
)

func (fire *FireMage) registerHotfixes() {
	// 2025-07-01 - Critical Mass Critical Strike bonus increased to 1.5x (was 1.3x).
	fire.criticalMassMultiplier += 0.2

	// 2025-07-01 - Pyroblast's direct damage increase raised to 30% (was 11%).
	fire.AddStaticMod(core.SpellModConfig{
		ClassMask:  mage.MageSpellPyroblast,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.3,
	})

	// 2025-07-01 - Combustion Ignite scaling increased to 50% (was 20%).
	fire.combustionDotDamageMultiplier += 0.3
}
