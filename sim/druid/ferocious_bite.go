package druid

import (
	"github.com/wowsims/tbc/sim/core"
)

var ferociousBiteRank = genRanks.FerociousBite.BySpellID(24248)
var ferociousBiteMin, ferociousBiteMax = ferociousBiteRank.Direct.Range()

func (druid *Druid) registerFerociousBiteSpell() {
	var energyMetrics *core.ResourceMetrics

	druid.FerociousBite = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: ferociousBiteRank.SpellID},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellFerociousBite,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost: ferociousBiteRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: ferociousBiteRank.GCD,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return druid.ComboPoints() > 0
		},

		DamageMultiplier: 1,
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
			baseDamage := ferociousBiteMin + dmgPerCP*cp + 4.1*excessEnergy + 0.05*cp*ap
			baseDamage += sim.RandomFloat("Ferocious Bite") * (ferociousBiteMax - ferociousBiteMin)

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				druid.SpendComboPoints(sim, spell.ComboPointMetrics())
			}
		},

		ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			cp := float64(druid.ComboPoints())
			ap := spell.MeleeAttackPower(target)
			dmgPerCP := 169.0 + druid.IdolFerociousBiteBonus
			baseDamage := ferociousBiteMin + dmgPerCP*cp + (ferociousBiteMax-ferociousBiteMin)/2 + 0.05*cp*ap
			return spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMeleeWeaponSpecialHitAndCrit)
		},
	})

	energyMetrics = druid.FerociousBite.Cost.ResourceCostImpl.(*core.EnergyCost).ResourceMetrics
}

func (druid *Druid) CurrentFerociousBiteCost() float64 {
	return druid.FerociousBite.Cost.GetCurrentCost()
}
