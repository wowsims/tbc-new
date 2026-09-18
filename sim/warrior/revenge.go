package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

var revengeRank = genRanks.Revenge.BySpellID(30357)

func (war *Warrior) registerRevenge() {
	actionID := core.ActionID{SpellID: revengeRank.SpellID}

	aura := war.RegisterAura(core.Aura{
		Label:    "Revenge",
		Duration: 5 * time.Second,
		ActionID: actionID,
	})

	war.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Revenge - Trigger",
		TriggerImmediately: true,
		Outcome:            core.OutcomeBlock | core.OutcomeDodge | core.OutcomeParry,
		Callback:           core.CallbackOnSpellHitDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura.Activate(sim)
		},
	})

	war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: SpellMaskRevenge,
		MaxRange:       core.MaxMeleeRange,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: revengeRank.GCD,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: revengeRank.Cooldown,
			},
		},

		RageCost: core.RageCostOptions{
			Cost:   revengeRank.Cost,
			Refund: 0.8,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		FlatThreatBonus:  200,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(DefensiveStance) && aura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := revengeRank.Direct.Damage(sim)
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			aura.Deactivate(sim)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},

		RelatedSelfBuff: aura,
	})
}
