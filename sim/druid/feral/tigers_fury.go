package feral

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
	"github.com/wowsims/tbc/sim/druid"
)

func (cat *FeralDruid) registerTigersFurySpell() {
	actionID := core.ActionID{SpellID: 5217}
	energyMetrics := cat.NewEnergyMetrics(actionID)

	const instantEnergy = 60.0

	cat.TigersFuryAura = cat.RegisterAura(core.Aura{
		Label:    "Tiger's Fury",
		ActionID: actionID,
		Duration: 6 * time.Second,

		OnGain: func(aura *core.Aura, _ *core.Simulation) {
			aura.Unit.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= 1.15
		},

		OnExpire: func(aura *core.Aura, _ *core.Simulation) {
			aura.Unit.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] /= 1.15
		},
	})

	cat.CatFormAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		cat.TigersFuryAura.Deactivate(sim)
	})

	cat.TigersFury = cat.RegisterSpell(druid.Cat, core.SpellConfig{
		ActionID:       actionID,
		Flags:          core.SpellFlagAPL | core.SpellFlagReadinessTrinket,
		ClassSpellMask: druid.DruidSpellTigersFury,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    cat.NewTimer(),
				Duration: time.Second * 30,
			},
		},

		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return !cat.BerserkCatAura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			cat.AddEnergy(sim, instantEnergy, energyMetrics)
			cat.TigersFuryAura.Activate(sim)
		},
	})
}
