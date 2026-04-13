package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (druid *Druid) registerFaerieFireSpell() {
	auras := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.FaerieFireAura(target, float64(druid.Talents.ImprovedFaerieFire))
	})

	druid.FaerieFire = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ClassSpellMask: DruidSpellFaerieFire,
		ActionID:       core.ActionID{SpellID: 26993},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: 145,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  132,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)

			if result.Landed() {
				auras.Get(target).Activate(sim)
			}
		},

		RelatedAuraArrays: auras.ToMap(),
	})
}

func (druid *Druid) registerFaerieFireFeralSpell() {
	if !druid.Talents.FaerieFireFeral {
		return
	}

	druid.FaerieFireAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.FaerieFireAura(target, float64(druid.Talents.ImprovedFaerieFire))
	})

	druid.FaerieFireFeral = druid.RegisterSpell(Cat|Bear, core.SpellConfig{
		ClassSpellMask: DruidSpellFaerieFireFeral,
		ActionID:       core.ActionID{SpellID: 27011},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDMin,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Second * 6,
			},
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  132,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)

			if result.Landed() {
				druid.FaerieFireAuras.Get(target).Activate(sim)
			}
		},

		RelatedAuraArrays: druid.FaerieFireAuras.ToMap(),
	})
}
