package warrior

import (
	"github.com/wowsims/tbc/sim/core"
)

var executeRank = genRanks.Execute.BySpellID(25236)

func (war *Warrior) registerExecute() {

	var rageMetrics *core.ResourceMetrics

	spell := war.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: executeRank.SpellID},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: SpellMaskExecute,
		MaxRange:       core.MaxMeleeRange,

		RageCost: core.RageCostOptions{
			Cost:   executeRank.Cost,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: executeRank.GCD,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1.25,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(BerserkerStance|BattleStance) && sim.IsExecutePhase20()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			extraRage := spell.Unit.CurrentRage()
			maxRage := war.MaximumRage()
			if extraRage > maxRage-spell.Cost.GetCurrentCost() {
				extraRage = maxRage - spell.Cost.GetCurrentCost()
			}
			war.SpendRage(sim, extraRage, rageMetrics)
			rageMetrics.Events--

			baseDamage := 925 + 21*extraRage
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})

	rageMetrics = spell.Cost.ResourceCostImpl.(*core.RageCost).ResourceMetrics

}
