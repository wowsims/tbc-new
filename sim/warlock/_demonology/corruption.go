package demonology

import "github.com/wowsims/tbc/sim/core"

func (demonology *DemonologyWarlock) registerCorruption() {
	corruption := demonology.RegisterCorruption(nil, func(resultList core.SpellResultSlice, spell *core.Spell, sim *core.Simulation) {
		if resultList[0].Landed() {
			demonology.GainDemonicFury(sim, 4, spell.ActionID)
		}
	})

	// replaced by doom in meta
	corruption.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return !demonology.IsInMeta()
	}
}
