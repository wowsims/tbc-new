package protection

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/paladin"
)

/*
Hurls your shield at an enemy target, dealing (<6058-7405> + 0.8175 * <AP> + 0.315 * <SP>) Holy damage,

silencing and interrupting spellcasting for 3 sec, and then jumping to additional nearby enemies.

Affects 3 total targets.
*/
func (prot *ProtectionPaladin) registerAvengersShieldSpell() {
	maxTargets := int32(3)

	prot.AvengersShield = prot.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 31935},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: paladin.SpellMaskAvengersShield,

		MaxRange:     30,
		MissileSpeed: 35,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 7,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    prot.NewTimer(),
				Duration: time.Second * 15,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return prot.PseudoStats.CanBlock
		},

		DamageMultiplier: 1,
		CritMultiplier:   prot.DefaultCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			bonusDamage := 0.31499999762*spell.SpellPower() + 0.81749999523*spell.MeleeAttackPower()

			spell.CalcCleaveDamageWithVariance(sim, target, maxTargets, spell.OutcomeMagicHitAndCrit, func(sim *core.Simulation, _ *core.Spell) float64 {
				return prot.CalcAndRollDamageRange(sim, 5.89499998093, 0.20000000298) + bonusDamage
			})

			spell.DealBatchedAoeDamage(sim)
		},
	})
}
