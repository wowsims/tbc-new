package rogue

import (
	"github.com/wowsims/tbc/sim/core"
)

var ambushRank = genRanks.Ambush.BySpellID(27441)

func (rogue *Rogue) registerAmbushSpell() {
	baseDamage, _ := ambushRank.Direct.Range()
	weaponDamage := 2.75

	rogue.Ambush = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: ambushRank.SpellID},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
		ClassSpellMask: RogueSpellAmbush,
		MaxRange:       core.MaxMeleeRange,

		EnergyCost: core.EnergyCostOptions{
			Cost:   ambushRank.Cost,
			Refund: 0,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: ambushRank.GCD,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !rogue.PseudoStats.InFrontOfTarget && rogue.IsStealthed()
		},

		DamageMultiplier:         weaponDamage,
		DamageMultiplierAdditive: 1,
		ThreatMultiplier:         1,

		BonusCoefficient: ambushRank.Direct.BonusCoefficient(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := baseDamage +
				spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialNoBlockDodgeParry)

			if result.Landed() {
				rogue.AddComboPoints(sim, 1, spell.ComboPointMetrics())
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
