package warrior

import "github.com/wowsims/tbc/sim/core"

func (war *Warrior) registerSunderArmor() {
	war.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 25225},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: SpellMaskSunderArmor,
		MaxRange:       core.MaxMeleeRange,

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
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.CanApplySunderAura(target)
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  301.5,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)

			if result.Landed() {
				war.TryApplySunderArmorEffect(sim, target)
			} else {
				spell.IssueRefund(sim)
			}
		},

		RelatedAuraArrays: war.SunderArmorAuras.ToMap(),
	})
}

func (warrior *Warrior) CanApplySunderAura(target *core.Unit) bool {
	return warrior.SunderArmorAuras.Get(target).IsActive() || !warrior.SunderArmorAuras.Get(target).ExclusiveEffects[0].Category.AnyActive()
}

func (warrior *Warrior) TryApplySunderArmorEffect(sim *core.Simulation, target *core.Unit) {
	if warrior.CanApplySunderAura(target) {
		aura := warrior.SunderArmorAuras.Get(target)
		aura.Activate(sim)
		if aura.IsActive() {
			aura.AddStack(sim)
		}
	}
}
