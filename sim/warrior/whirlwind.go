package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (war *Warrior) registerWhirlwind() {
	actionID := core.ActionID{SpellID: 1680}

	whirlwindOH := war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(2),
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeOHSpecial,
		ClassSpellMask: SpellMaskWhirlwindOh,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,

		ThreatMultiplier: 1.25,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := war.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			spell.CalcCleaveDamage(sim, target, 4, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			spell.DealBatchedAoeDamage(sim)
		},
	})

	war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(1),
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: SpellMaskWhirlwind,

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

		ThreatMultiplier: 1.25,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(BerserkerStance)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := war.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			spell.CalcCleaveDamage(sim, target, 4, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			spell.DealBatchedAoeDamage(sim)

			if whirlwindOH != nil && war.OffHand() != nil && war.OffHand().WeaponType != proto.WeaponType_WeaponTypeUnknown {
				whirlwindOH.Cast(sim, target)
			}
		},
	})
}
