package rogue

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var garroteRank = genRanks.Garrote.BySpellID(26884)

func (rogue *Rogue) registerGarrote() {
	tick := garroteRank.Periodic.(shared.SpellRankPeriodic)

	rogue.Garrote = rogue.GetOrRegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: garroteRank.SpellID},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
		ClassSpellMask: RogueSpellGarrote,

		EnergyCost: core.EnergyCostOptions{
			Cost:   garroteRank.Cost,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: garroteRank.GCD,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !rogue.PseudoStats.InFrontOfTarget && rogue.IsStealthed()
		},

		DamageMultiplierAdditive: 1,
		DamageMultiplier:         1,
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Garrote",
				Tag:   RogueBleedTag,
			},
			NumberOfTicks: tick.Ticks,
			TickLength:    tick.Period,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.SnapshotPhysical(target, tick.Tick+dot.Spell.MeleeAttackPower(target)*0.03)
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
