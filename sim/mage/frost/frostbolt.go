package frost

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/mage"
)

const frostboltVariance = 0.24   // Per https://wago.tools/db2/SpellEffect?build=5.5.0.60802&filter%5BSpellID%5D=exact%253A116 Field: "Variance"
const frostboltScale = 1.5       // Per https://wago.tools/db2/SpellEffect?build=5.5.0.60802&filter%5BSpellID%5D=exact%253A116 Field: "Coefficient"
const frostboltCoefficient = 1.5 // Per https://wago.tools/db2/SpellEffect?build=5.5.0.60802&filter%5BSpellID%5D=exact%253A116 Field: "BonusCoefficient"

func (frost *FrostMage) frostBoltConfig(config core.SpellConfig) core.SpellConfig {
	return core.SpellConfig{
		ActionID:       config.ActionID,
		SpellSchool:    core.SpellSchoolFrost,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          config.Flags,
		ClassSpellMask: mage.MageSpellFrostbolt,
		MissileSpeed:   28,

		ManaCost: config.ManaCost,
		Cast:     config.Cast,

		DamageMultiplier: config.DamageMultiplier,
		CritMultiplier:   frost.DefaultCritMultiplier(),
		BonusCoefficient: frostboltCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: config.ApplyEffects,
	}
}

func (frost *FrostMage) registerFrostboltSpell() {
	actionID := core.ActionID{SpellID: 116}
	hasGlyph := frost.HasMajorGlyph(proto.MageMajorGlyph_GlyphOfIcyVeins)
	var icyVeinsFrostBolt *core.Spell

	frost.RegisterSpell(frost.frostBoltConfig(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 4,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Second * 2,
			},
		},

		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			hasSplitBolts := frost.IcyVeinsAura.IsActive() && hasGlyph
			damageMultiplier := core.TernaryFloat64(hasSplitBolts, 0.4, 1.0)

			spell.DamageMultiplier *= damageMultiplier
			baseDamage := frost.CalcAndRollDamageRange(sim, frostboltScale, frostboltVariance)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.DamageMultiplier /= damageMultiplier

			if result.Landed() {
				frost.ProcFingersOfFrost(sim, spell)
			}

			if hasSplitBolts {
				icyVeinsFrostBolt.Cast(sim, target)
			}

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
				if result.Landed() {
					frost.GainIcicle(sim, target, result.Damage)
				}
			})
		},
	}))

	// Glyph of Icy Veins - Frostbolt
	icyVeinsFrostBolt = frost.RegisterSpell(frost.frostBoltConfig(core.SpellConfig{
		ActionID:       actionID.WithTag(1), // Real SpellID: 131079
		ClassSpellMask: mage.MageSpellFrostbolt,
		Flags:          core.SpellFlagPassiveSpell,

		DamageMultiplier: 0.4,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			results := make([]*core.SpellResult, 2)

			for idx := range results {
				baseDamage := frost.CalcAndRollDamageRange(sim, frostboltScale, frostboltVariance)
				results[idx] = spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
				if results[idx].Landed() {
					frost.ProcFingersOfFrost(sim, spell)
				}
			}

			for _, result := range results {
				spell.WaitTravelTime(sim, func(sim *core.Simulation) {
					spell.DealDamage(sim, result)
					if result.Landed() {
						frost.GainIcicle(sim, target, result.Damage)
					}
				})
			}
		},
	}))
}
