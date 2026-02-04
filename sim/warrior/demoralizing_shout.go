package warrior

import (
	"github.com/wowsims/tbc/sim/core"
)

func (war *Warrior) registerDemoralizingShout() {
	war.DemoralizingShoutAuras = war.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.DemoralizingShoutAura(target, war.Talents.BoomingVoice, war.Talents.ImprovedDemoralizingShout)
	})

	war.DemoralizingShout = war.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 25203},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL,

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
		FlatThreatBonus:  56,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
				result := spell.CalcOutcome(sim, aoeTarget, spell.OutcomeMagicHit)
				if result.Landed() {
					war.DemoralizingShoutAuras.Get(aoeTarget).Activate(sim)
				}
			}
		},

		RelatedAuraArrays: war.DemoralizingShoutAuras.ToMap(),
	})
}
