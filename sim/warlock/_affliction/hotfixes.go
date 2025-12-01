package affliction

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/warlock"
)

func (affliction *AfflictionWarlock) registerHotfixes() {

	// 2025-07-31 - Agony’s damage over time increased by 5%.
	affliction.AddStaticMod(core.SpellModConfig{
		ClassMask:  warlock.WarlockSpellAgony,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.05,
	})

	// 2025-09-22 - Corruption’s damage over time decreased from 33% to 20%.
	affliction.AddStaticMod(core.SpellModConfig{
		ClassMask:  warlock.WarlockSpellCorruption,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.20,
	})

	// 2025-07-31 -  Malefic Damage increased by 50%
	affliction.AddStaticMod(core.SpellModConfig{
		ClassMask:  warlock.WarlockSpellMaleficGrasp | warlock.WarlockSpellDrainSoul,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.5,
	})

	// 2025-07-31 - The damage your Malefic Grasp causes your other DoTs to deal increased to 50% (was 30%).
	affliction.MaleficGraspMaleficEffectMultiplier += 0.2
	// 2025-07-31 - The damage your Drain Soul causes your other DoTs to deal increased to 100% (was 60%).
	affliction.DrainSoulMaleficEffectMultiplier += 0.4

}
