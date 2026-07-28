package druid

import (
	"github.com/wowsims/tbc/sim/core"
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
	druid.DemoralizingRoarAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.DemoralizingRoarAura(target, druid.Talents.FeralAggression)
	})
}
