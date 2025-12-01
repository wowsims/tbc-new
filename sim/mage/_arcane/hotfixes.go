package arcane

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/mage"
)

func (arcane *ArcaneMage) registerHotfixes() {

	// 2025-09-22 - Arcane Blast damage increase lowered from 29% to 15%
	arcane.AddStaticMod(core.SpellModConfig{
		ClassMask:  mage.MageSpellArcaneBlast,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.15,
	})
	// 2025-07-01 - Arcane Blast mana cost lowered by 10% to 1.5% of base mana (was 1.666%) -  https://eu.forums.blizzard.com/en/wow/t/mists-of-pandaria-classic-development-notes-updated-20-june/571162/13
	arcane.AddStaticMod(core.SpellModConfig{
		ClassMask:  mage.MageSpellArcaneBlast,
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -0.1,
	})

	// 2025-07-01 - Arcane Barrage damage increase lowered to 19% (was 30%)
	arcane.AddStaticMod(core.SpellModConfig{
		ClassMask:  mage.MageSpellArcaneBarrage,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.19,
	})

	// 2025-07-01 - Arcane Missiles damage increase lowered to 15% (was 28%).
	arcane.AddStaticMod(core.SpellModConfig{
		ClassMask:  mage.MageSpellArcaneMissilesTick,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.15,
	})
}
