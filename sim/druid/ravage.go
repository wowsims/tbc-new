package druid

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var ravageRank = genRanks.Ravage.BySpellID(27005)

func (druid *Druid) registerRavageSpell() {
	// 385% weapon damage, which the client states as E_WEAPON_PERCENT_DAMAGE = 384 on spell 27005 -
	// the same shape as Shred's 224 / 2.25. The flat addend is the rank's own value, scaled by that
	// multiplier the way the tooltip shows it.
	const weaponMultiplier = 3.85
	const highHpCritPercentBonus = 50.0

	druid.Ravage = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID:         core.ActionID{SpellID: ravageRank.SpellID},
		SpellSchool:      core.SpellSchoolPhysical,
		DefenseType:      core.DefenseTypeMelee,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		ClassSpellMask:   DruidSpellRavage,
		Flags:            core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		DamageMultiplier: weaponMultiplier,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,
		MaxRange:         core.MaxMeleeRange,

		EnergyCost: core.EnergyCostOptions{
			Cost:   ravageRank.Cost,
			Refund: 0.8,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: ravageRank.GCD,
			},

			IgnoreHaste: true,
		},

		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return druid.ProwlAura.IsActive() && !druid.PseudoStats.InFrontOfTarget && !druid.CannotShredTarget
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if sim.IsExecutePhase90() {
				spell.BonusCritPercent += highHpCritPercentBonus
			}

			baseDamage := ravageRank.Direct.Damage(sim) + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				druid.AddComboPoints(sim, 1, spell.ComboPointMetrics())
			} else {
				spell.IssueRefund(sim)
			}

			if sim.IsExecutePhase90() {
				spell.BonusCritPercent -= highHpCritPercentBonus
			}
		},

		ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			if sim.IsExecutePhase90() {
				spell.BonusCritPercent += highHpCritPercentBonus
			}

			baseDamage := shared.SpellRankMin(ravageRank.Direct) + spell.Unit.AutoAttacks.MH().CalculateAverageWeaponDamage(spell.MeleeAttackPower(target))
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMeleeWeaponSpecialHitAndCrit)

			if sim.IsExecutePhase90() {
				spell.BonusCritPercent -= highHpCritPercentBonus
			}

			return result
		},
	})
}
