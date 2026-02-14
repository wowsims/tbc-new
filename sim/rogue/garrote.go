package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (rogue *Rogue) registerGarrote() {
	baseDamage := 135.0

	rogue.Garrote = rogue.GetOrRegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 703},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
		ClassSpellMask: RogueSpellGarrote,

		EnergyCost: core.EnergyCostOptions{
			Cost:   50,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !rogue.PseudoStats.InFrontOfTarget && rogue.IsStealthed()
		},

		DamageMultiplierAdditive: 1,
		DamageMultiplier:         1,
		CritMultiplier:           rogue.DefaultMeleeCritMultiplier(),
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Garrote",
				Tag:   RogueBleedTag,
			},
			NumberOfTicks: 6,
			TickLength:    time.Second * 3,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.SnapshotPhysical(target, baseDamage+dot.Spell.MeleeAttackPower()*0.03)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialNoBlockDodgeParryNoCrit)
			if result.Landed() {
				rogue.AddComboPoints(sim, 1, spell.ComboPointMetrics())
				spell.Dot(target).Apply(sim)
			} else {
				spell.IssueRefund(sim)
			}
			spell.DealOutcome(sim, result)
		},
	})
}
