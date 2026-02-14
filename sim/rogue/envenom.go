package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (rogue *Rogue) registerEnvenom() {
	baseDamage := 180.0 + rogue.DeathmantleBonus
	apScalingPerComboPoint := 0.03

	rogue.Envenom = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 32645},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskMeleeMHSpecial, // not core.ProcMaskSpellDamage
		Flags:          core.SpellFlagMeleeMetrics | SpellFlagFinisher | core.SpellFlagAPL,
		MetricSplits:   6,
		ClassSpellMask: RogueSpellEnvenom,

		EnergyCost: core.EnergyCostOptions{
			Cost:          35,
			Refund:        0.4 * float64(rogue.Talents.QuickRecovery),
			RefundMetrics: rogue.EnergyRefundMetrics,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				spell.SetMetricsSplit(rogue.ComboPoints())
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return rogue.ComboPoints() > 0
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           rogue.DefaultMeleeCritMultiplier(),
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			comboPoints := rogue.ComboPoints()
			dp := rogue.DeadlyPoison.Dot(target)
			consumed := min(dp.GetStacks(), comboPoints)

			baseDamage := baseDamage*float64(consumed) +
				apScalingPerComboPoint*float64(consumed)*spell.MeleeAttackPower()

			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if result.Landed() {
				rogue.ApplyFinisher(sim, spell)
				if newStacks := dp.GetStacks() - comboPoints; newStacks > 0 {
					dp.SetStacks(sim, newStacks)
				} else {
					dp.Deactivate(sim)
				}
			} else {
				spell.IssueRefund(sim)
			}

			spell.DealDamage(sim, result)
		},
	})
}
