package fire

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
	"github.com/wowsims/tbc/sim/mage"
)

func (fire *FireMage) registerCriticalMass() {

	getCritPercent := func() float64 {
		return fire.GetStat(stats.SpellCritPercent) * fire.criticalMassMultiplier
	}

	criticalMassCritBuffMod := fire.AddDynamicMod(core.SpellModConfig{
		FloatValue: getCritPercent(),
		ClassMask:  mage.MageSpellFireball | mage.MageSpellFrostfireBolt | mage.MageSpellScorch | mage.MageSpellPyroblast | mage.MageSpellPyroblastDot,
		Kind:       core.SpellMod_BonusCrit_Percent,
	})

	core.MakePermanent(fire.RegisterAura(core.Aura{
		Label: "Critical Mass",
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			criticalMassCritBuffMod.Activate()
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			criticalMassCritBuffMod.Deactivate()
		},
	}))

	fire.AddOnTemporaryStatsChange(func(sim *core.Simulation, buffAura *core.Aura, statsChangeWithoutDeps stats.Stats) {
		criticalMassCritBuffMod.UpdateFloatValue(getCritPercent())
	})

	fire.RegisterResetEffect(func(sim *core.Simulation) {
		criticalMassCritBuffMod.UpdateFloatValue(getCritPercent())
	})

}
