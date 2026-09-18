package rogue

import (
	"github.com/wowsims/tbc/sim/core"
)

var sinisterStrikeRank = genRanks.SinisterStrike.BySpellID(26862)

func (rogue *Rogue) registerSinisterStrikeSpell() {
	baseDamage, _ := sinisterStrikeRank.Direct.Range()

	rogue.SinisterStrike = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: sinisterStrikeRank.SpellID},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
		ClassSpellMask: RogueSpellSinisterStrike,

		EnergyCost: core.EnergyCostOptions{
			Cost:   sinisterStrikeRank.Cost,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: sinisterStrikeRank.GCD,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		ThreatMultiplier:         1,

		BonusCoefficient: sinisterStrikeRank.Direct.BonusCoefficient(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := baseDamage +
				spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				rogue.AddComboPoints(sim, 1, spell.ComboPointMetrics())
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
