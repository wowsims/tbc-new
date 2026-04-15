package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (druid *Druid) registerFerociousBiteSpell() {
	const baseCost = 35

	var energyMetrics *core.ResourceMetrics

	druid.FerociousBite = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 24248},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellFerociousBite,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost: baseCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return druid.ComboPoints() > 0
		},

		DamageMultiplier: 1,
		CritMultiplier:   druid.FeralCritMultiplier(),
		ThreatMultiplier: 1,
		MaxRange:         core.MaxMeleeRange,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			cp := float64(druid.ComboPoints())
			ap := spell.MeleeAttackPower(target)
			excessEnergy := druid.CurrentEnergy()
			if excessEnergy > 0 {
				druid.SpendEnergy(sim, excessEnergy, energyMetrics)
				energyMetrics.Events--
			}

			dmgPerCP := 169.0 + druid.IdolFerociousBiteBonus
			baseDamage := 57 + dmgPerCP*cp + 4.1*excessEnergy + 0.05*cp*ap
			baseDamage += sim.RandomFloat("Ferocious Bite") * 66

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				druid.SpendComboPoints(sim, spell.ComboPointMetrics())
			}
		},

		ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			cp := float64(druid.ComboPoints())
			ap := spell.MeleeAttackPower(target)
			dmgPerCP := 169.0 + druid.IdolFerociousBiteBonus
			baseDamage := 57 + dmgPerCP*cp + 33 + 0.05*cp*ap
			return spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMeleeWeaponSpecialHitAndCrit)
		},
	})

	energyMetrics = druid.FerociousBite.Cost.ResourceCostImpl.(*core.EnergyCost).ResourceMetrics
}

func (druid *Druid) CurrentFerociousBiteCost() float64 {
	return druid.FerociousBite.Cost.GetCurrentCost()
}
