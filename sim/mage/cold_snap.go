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
