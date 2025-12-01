package demonology

import (
	"github.com/wowsims/tbc/sim/core"
)

func (demonology *DemonologyWarlock) registerHellfire() {
	hellfire := demonology.RegisterHellfire(func(resultList core.SpellResultSlice, spell *core.Spell, sim *core.Simulation) {
		if demonology.IsInMeta() {
			return
		}

		// 10 for primary, 3 for every other target
		fury := 10 + ((len(resultList))-1)*3
		demonology.GainDemonicFury(sim, float64(fury), spell.ActionID)
	})

	oldExtra := hellfire.ExtraCastCondition
	hellfire.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		if oldExtra != nil && !oldExtra(sim, target) {
			return false
		}

		return !demonology.IsInMeta()
	}

	demonology.Metamorphosis.RelatedSelfBuff.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		demonology.Hellfire.SelfHot().Deactivate(sim)
	})

}
