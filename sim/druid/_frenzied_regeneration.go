package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (druid *Druid) registerFrenziedRegenerationSpell() {
	actionID := core.ActionID{SpellID: 22842}
	rageMetrics := druid.NewRageMetrics(actionID)

	druid.FrenziedRegeneration = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:         actionID,
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskSpellHealing,
		Flags:            core.SpellFlagAPL,
		DamageMultiplier: 1,
		CritMultiplier:   druid.DefaultCritMultiplier(),
		ThreatMultiplier: 1,
		ClassSpellMask:   DruidSpellFrenziedRegeneration,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Millisecond * 1500,
			},
		},

		RageCost: core.RageCostOptions{
			Cost: 0,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			const maxRageCost = 60.0
			rageDumped := min(druid.CurrentRage(), maxRageCost)
			healthGained := max((druid.GetStat(stats.AttackPower)-2*druid.GetStat(stats.Agility))*2.2, druid.GetStat(stats.Stamina)*2.5) * rageDumped / maxRageCost
			spell.CalcAndDealHealing(sim, spell.Unit, healthGained, spell.OutcomeHealing)
			druid.SpendRage(sim, rageDumped, rageMetrics)
		},
	})
}
