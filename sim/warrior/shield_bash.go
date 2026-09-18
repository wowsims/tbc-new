package warrior

import (
	"github.com/wowsims/tbc/sim/core"
)

var shieldBashRank = genRanks.ShieldBash.BySpellID(29704)

func (war *Warrior) registerShieldBash() {
	actionID := core.ActionID{SpellID: shieldBashRank.SpellID}

	war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		ClassSpellMask: SpellMaskShieldBash,
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		MaxRange:       core.MaxMeleeRange,

		RageCost: core.RageCostOptions{
			Cost:   shieldBashRank.Cost,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: shieldBashRank.GCD,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: shieldBashRank.Cooldown,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1.5,
		FlatThreatBonus:  192,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.PseudoStats.CanBlock && war.StanceMatches(DefensiveStance|BattleStance)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := shieldBashRank.Direct.Damage(sim)
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}
