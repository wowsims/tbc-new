package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (druid *Druid) registerDemoralizingRoarSpell() {
	druid.registerDemoralizingRoarAura()

	druid.DemoralizingRoar = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26998},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskEmpty,
		ClassSpellMask: DruidSpellDemoralizingRoar,
		Flags:          core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost: 10,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  62 * 2,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range druid.Env.Encounter.AllTargetUnits {
				result := spell.CalcOutcome(sim, aoeTarget, spell.OutcomeMeleeSpecialHit)
				if result.Landed() {
					druid.DemoralizingRoarAuras.Get(aoeTarget).Activate(sim)
				}
			}
		},
	})
}

func (druid *Druid) registerDemoralizingRoarAura() {
	// Rank 6 (TBC max): reduces melee AP by 248.
	// FeralAggression talent: +5% per rank (up to 5 ranks = +25%).
	apReduction := 248.0 * (1 + 0.05*float64(druid.Talents.FeralAggression))

	druid.DemoralizingRoarAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    "Demoralizing Roar-" + druid.Label,
			ActionID: core.ActionID{SpellID: 26998},
			Duration: time.Second * 30,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.AddStatsDynamic(sim, stats.Stats{stats.AttackPower: -apReduction})
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.AddStatsDynamic(sim, stats.Stats{stats.AttackPower: apReduction})
			},
		})
	})
}
