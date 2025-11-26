package demonology

import "github.com/wowsims/tbc/sim/core"

func (demo *DemonologyWarlock) registerDrainLife() {
	demo.RegisterDrainLife(func(_ core.SpellResultSlice, spell *core.Spell, sim *core.Simulation) {
		if demo.IsInMeta() {
			if demo.CanSpendDemonicFury(30) {
				demo.SpendDemonicFury(sim, 30, spell.ActionID)
			} else {
				demo.ChanneledDot.Deactivate(sim)
			}
		} else {
			demo.GainDemonicFury(sim, 10, spell.ActionID)
		}
	})
}
