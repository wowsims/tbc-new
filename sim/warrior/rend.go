package warrior

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var rendRank = genRanks.Rend.BySpellID(25208)

func (war *Warrior) registerRend() {
	tick := rendRank.Periodic.(shared.SpellRankPeriodic)

	war.Rend = war.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rendRank.SpellID},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ClassSpellMask: SpellMaskRend,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost:   rendRank.Cost,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: rendRank.GCD,
			},
			IgnoreHaste: true,
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(BattleStance | DefensiveStance)
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Rend",
			},
			NumberOfTicks: tick.NumberOfTicks,
			TickLength:    tick.TickLength,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.SnapshotBaseDamage = tick.Tick + war.AutoAttacks.MH().CalculateAverageWeaponDamage(dot.Spell.MeleeAttackPower(target))*0.00743
				dot.SnapshotAttackerMultiplier = dot.Spell.AttackerDamageMultiplier(dot.Spell.Unit.AttackTables[target.UnitIndex], true)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
