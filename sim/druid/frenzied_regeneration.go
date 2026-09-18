package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var frenziedRegenerationRank = genRanks.FrenziedRegeneration.BySpellID(26999)
var frenziedRegenerationTick = frenziedRegenerationRank.Periodic.(shared.SpellRankPeriodic)

func (druid *Druid) registerFrenziedRegenerationSpell() {
	actionID := core.ActionID{SpellID: frenziedRegenerationRank.SpellID}
	rageMetrics := druid.NewRageMetrics(actionID)

	druid.FrenziedRegenerationAura = druid.RegisterAura(core.Aura{
		Label:    "Frenzied Regeneration",
		ActionID: actionID,
		Duration: 10 * time.Second,
	})

	// Deactivate when leaving Bear Form.
	druid.BearFormAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.FrenziedRegenerationAura.Deactivate(sim)
	})

	druid.FrenziedRegeneration = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:         actionID,
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskEmpty,
		ClassSpellMask:   DruidSpellFrenziedRegeneration,
		Flags:            core.SpellFlagAPL,
		DamageMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: frenziedRegenerationRank.GCD,
			},
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: frenziedRegenerationRank.Cooldown,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			druid.FrenziedRegenerationAura.Activate(sim)
			// Converts up to 10 rage per second into 25 health per rage, for 10 sec.
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   frenziedRegenerationTick.Period,
				NumTicks: int(frenziedRegenerationTick.Ticks),
				Priority: core.ActionPriorityDOT,
				OnAction: func(sim *core.Simulation) {
					rage := min(druid.CurrentRage(), 10)
					if rage > 0 {
						druid.SpendRage(sim, rage, rageMetrics)
						spell.CalcAndDealPeriodicHealing(sim, &druid.Unit, rage*frenziedRegenerationTick.Tick, spell.OutcomeHealing)
					}
				},
			})
		},
	})
}
