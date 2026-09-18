package druid

import (
	"github.com/wowsims/tbc/sim/core"
)

var faerieFireRank = genRanks.FaerieFire.BySpellID(26993)
var faerieFireFeralRank = genRanks.FaerieFireFeral.BySpellID(27011)

func (druid *Druid) registerFaerieFireSpell() {
	auras := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.FaerieFireAura(target, float64(druid.Talents.ImprovedFaerieFire))
	})

	druid.FaerieFire = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ClassSpellMask: DruidSpellFaerieFire,
		ActionID:       core.ActionID{SpellID: faerieFireRank.SpellID},
		SpellSchool:    core.SpellSchoolNature,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: faerieFireRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: faerieFireRank.GCD,
			},
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  132,
		MaxRange:         faerieFireRank.MaxRange,

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
		ActionID:       core.ActionID{SpellID: faerieFireFeralRank.SpellID},
		SpellSchool:    core.SpellSchoolNature,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: faerieFireFeralRank.GCD,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: faerieFireFeralRank.Cooldown,
			},
		},

		ThreatMultiplier: 1,
		FlatThreatBonus:  132,
		MaxRange:         faerieFireFeralRank.MaxRange,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)

			if result.Landed() {
				druid.FaerieFireAuras.Get(target).Activate(sim)
			}
		},

		RelatedAuraArrays: druid.FaerieFireAuras.ToMap(),
	})
}
