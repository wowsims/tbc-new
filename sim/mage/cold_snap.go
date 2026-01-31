package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerColdSnapSpell() {
	if !mage.Talents.ColdSnap {
		return
	}

	mage.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 11958},
		Flags:    core.SpellFlagNoOnCastComplete,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: time.Second * 480,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			// Don't use if there are no cooldowns to reset.
			// I don't know if I should leave this up to APL or not
			return (mage.IcyVeins != nil && !mage.IcyVeins.IsReady(sim)) ||
				(mage.SummonWaterElemental != nil && !mage.SummonWaterElemental.IsReady(sim))
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			if mage.IcyVeins != nil {
				mage.IcyVeins.CD.Reset()
			}
			if mage.SummonWaterElemental != nil {
				mage.SummonWaterElemental.CD.Reset()
			}
		},
	})
}
