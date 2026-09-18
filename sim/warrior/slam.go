package warrior

import (
	"github.com/wowsims/tbc/sim/core"
)

var slamRank = genRanks.Slam.BySpellID(25242)

func (war *Warrior) registerSlam() {

	actionID := core.ActionID{SpellID: slamRank.SpellID}

	war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: SpellMaskSlam,
		MaxRange:       core.MaxMeleeRange,

		RageCost: core.RageCostOptions{
			Cost:   slamRank.Cost,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      slamRank.GCD,
				CastTime: slamRank.CastTime,
			},
			IgnoreHaste: true,
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				if cast.CastTime > 0 {
					war.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+cast.CastTime)
				}
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			slamVal, _ := slamRank.Direct.Range()
			baseDamage := slamVal + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}
