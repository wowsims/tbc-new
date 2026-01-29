package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerArcaneCharges() {
	abCostMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellArcaneBlast,
		FloatValue: 1.75,
		Kind:       core.SpellMod_PowerCost_Pct,
	})

	abCastMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask: MageSpellArcaneBlast,
		TimeValue: time.Millisecond * -334,
		Kind:      core.SpellMod_CastTime_Flat,
	})

	mage.ArcaneChargesAura = core.BlockPrepull(mage.GetOrRegisterAura(core.Aura{
		Label:     "Arcane Charges Aura",
		ActionID:  core.ActionID{SpellID: 36032},
		Duration:  time.Second * 8,
		MaxStacks: 3,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			abCastMod.Activate()
			abCostMod.Activate()
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			abCastMod.Deactivate()
			abCostMod.Deactivate()
		},
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			stacks := float64(newStacks)
			abCastMod.UpdateTimeValue(time.Millisecond * -334 * time.Duration(newStacks))
			abCostMod.UpdateFloatValue(1.75 * stacks)
		},
	}))
}
