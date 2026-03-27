package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerManaGems() {

	var manaGain float64
	actionID := core.ActionID{ItemID: 22044}
	manaMetrics := mage.NewManaMetrics(actionID)

	minManaGain := 2340.0
	maxManaGain := 2460.0

	manaGemAura := core.MakePermanent(mage.GetOrRegisterAura(core.Aura{
		Label:     "Mana Gem Charges",
		ActionID:  actionID,
		MaxStacks: 3,
	})).ApplyOnReset(func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, 3)
	})

	mage.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL | core.SpellFlagHelpful,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    mage.GetConjuredCD(),
				Duration: time.Minute * 2,
			},
		},

		// Don't use if we don't have any gems remaining!
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return manaGemAura.GetStacks() > 0
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			manaGemAura.RemoveStack(sim)
			manaGain = sim.Roll(minManaGain, maxManaGain)
			mage.AddMana(sim, manaGain, manaMetrics)
		},
	})
}
