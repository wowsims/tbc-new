package warrior

import (
	"github.com/wowsims/tbc/sim/core"
)

var hamstringRank = genRanks.Hamstring.BySpellID(25212)

func (war *Warrior) registerHamstring() {
	war.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: hamstringRank.SpellID},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHamstring,
		MaxRange:       core.MaxMeleeRange,

		RageCost: core.RageCostOptions{
			Cost:   hamstringRank.Cost,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: hamstringRank.GCD,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1.25,
		FlatThreatBonus:  167.5,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(BerserkerStance)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage, _ := hamstringRank.Direct.Range()
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}
