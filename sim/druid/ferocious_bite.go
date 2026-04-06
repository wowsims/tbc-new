package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (druid *Druid) registerFerociousBiteSpell() {
	// Idol of the Beast (25667) adds 14 damage per combo point.
	dmgPerCP := 169.0
	if druid.HasItemEquipped(25667, []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}) {
		dmgPerCP += 14
	}

	druid.FerociousBite = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 24248},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellFerociousBite,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost: 35,
			// Ferocious Bite consumes ALL available energy; refund is not applicable.
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
			// Energy above the base 35 is consumed for bonus damage at 4.1 per energy.
			excessEnergy := druid.CurrentEnergy() - 35
			if excessEnergy < 0 {
				excessEnergy = 0
			}

			// Spend all energy before dealing damage.
			if excessEnergy > 0 {
				druid.SpendEnergy(sim, excessEnergy, spell.EnergyMetrics())
			}

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
			// Expected value: midpoint of random roll, no excess energy assumed.
			baseDamage := 57 + dmgPerCP*cp + 33 + 0.05*cp*ap
			return spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMeleeWeaponSpecialHitAndCrit)
		},
	})
}

func (druid *Druid) CurrentFerociousBiteCost() float64 {
	return druid.FerociousBite.Cost.GetCurrentCost()
}
