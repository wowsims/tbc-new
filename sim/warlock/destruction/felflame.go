package destruction

import (
	"github.com/wowsims/tbc/sim/core"
)

func (destruction DestructionWarlock) registerFelflame() {
	destruction.RegisterFelflame(func(resultList core.SpellResultSlice, spell *core.Spell, sim *core.Simulation) {
		destruction.BurningEmbers.Gain(sim, core.TernaryFloat64(resultList[0].DidCrit(), 2, 1), spell.ActionID)
	})
}
