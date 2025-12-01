package mage

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (mage *Mage) registerIceLanceSpell() {
	actionID := core.ActionID{SpellID: 30455}
	// Values found at https://wago.tools/db2/SpellEffect?build=5.5.0.60802&filter%5BSpellID%5D=30455
	iceLanceScaling := 0.33500000834
	iceLanceCoefficient := 0.33500000834
	iceLanceVariance := 0.25

	getIceLanceSpellBaseConfig := func(config core.SpellConfig) core.SpellConfig {
		return core.SpellConfig{
			ActionID:       config.ActionID,
			SpellSchool:    core.SpellSchoolFrost,
			ProcMask:       core.ProcMaskSpellDamage,
			Flags:          config.Flags,
			ClassSpellMask: MageSpellIceLance,
			MissileSpeed:   38,

			ManaCost: config.ManaCost,
			Cast:     config.Cast,

			DamageMultiplier: config.DamageMultiplier,
			CritMultiplier:   mage.DefaultCritMultiplier(),
			BonusCoefficient: iceLanceCoefficient,
			ThreatMultiplier: 1,

			ApplyEffects: config.ApplyEffects,
		}
	}

	castIceLance := func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := mage.CalcAndRollDamageRange(sim, iceLanceScaling, iceLanceVariance)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	mage.RegisterSpell(getIceLanceSpellBaseConfig(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 1,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			icyVeinsDamageMultiplier := 1.0

			// Secondary Target hit
			spell.DamageMultiplier *= icyVeinsDamageMultiplier

			// Main Target hit
			castIceLance(sim, target, spell)

			if mage.FingersOfFrostAura.IsActive() {
				mage.FingersOfFrostAura.RemoveStack(sim)
			}

			spell.DamageMultiplier /= icyVeinsDamageMultiplier

			if mage.Spec == proto.Spec_SpecFrostMage {
				// Confirmed in game Icicles launch even if ice lance misses.
				for _, icicle := range mage.Icicles {
					mage.SpendIcicle(sim, target, icicle)
				}
				mage.Icicles = make([]float64, 0)
			}

		},
	}))
}
