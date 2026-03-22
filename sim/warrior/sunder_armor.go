package warrior

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

func (war *Warrior) registerSunderArmor() {
	actionId := core.ActionID{SpellID: 25225}

	war.SunderArmorAuras = war.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.SunderArmorAura(target)
	})

	getSunderArmorConfig := func(config core.SpellConfig, outcome shared.OutcomeType) core.SpellConfig {
		return core.SpellConfig{
			ActionID:       config.ActionID,
			SpellSchool:    core.SpellSchoolPhysical,
			ProcMask:       core.ProcMaskMeleeMHSpecial,
			Flags:          config.Flags,
			ClassSpellMask: SpellMaskSunderArmor,
			MaxRange:       core.MaxMeleeRange,

			RageCost: config.RageCost,
			Cast:     config.Cast,
			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return war.CanApplySunderAura(target)
			},

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			FlatThreatBonus:  301.5,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				result := spell.CalcOutcome(sim, target, shared.GetOutcome(spell, outcome))

				if result.Landed() {
					aura := war.SunderArmorAuras.Get(target)
					aura.Activate(sim)
					aura.AddStack(sim)
				} else if spell.Cost != nil {
					spell.IssueRefund(sim)
				}

				spell.DealOutcome(sim, result)
			},

			RelatedAuraArrays: war.SunderArmorAuras.ToMap(),
		}
	}

	war.RegisterSpell(getSunderArmorConfig(core.SpellConfig{
		ActionID: actionId,
		Flags:    core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost:   15,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},
	}, shared.OutcomeMeleeNoCrit))

	war.SunderArmorDevastate = war.RegisterSpell(getSunderArmorConfig(core.SpellConfig{
		ActionID: actionId.WithTag(1),
	}, shared.OutcomeAlwaysHit))
}

func (warrior *Warrior) CanApplySunderAura(target *core.Unit) bool {
	return warrior.SunderArmorAuras.Get(target).IsActive() || !warrior.SunderArmorAuras.Get(target).ExclusiveEffects[0].Category.AnyActive()
}
