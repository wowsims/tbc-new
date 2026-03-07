package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (war *Warrior) registerWhirlwind() {
	actionID := core.ActionID{SpellID: 1680}

	whirlwindOH := war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(2),
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeOHSpecial,
		ClassSpellMask: SpellMaskWhirlwindOh,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1.25,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := war.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			spell.CalcCleaveDamage(sim, target, 4, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			spell.DealBatchedAoeDamage(sim)
		},
	})

	war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(1),
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: SpellMaskWhirlwind,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost: 25,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Second * 10,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1.25,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(BerserkerStance)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := war.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			results := spell.CalcCleaveDamage(sim, target, 4, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			war.CastNormalizedSweepingStrikesAttack(results, sim)
			spell.DealBatchedAoeDamage(sim)

			if war.HasOHWeapon() {
				whirlwindOH.Cast(sim, target)
			}
		},
	})
}
