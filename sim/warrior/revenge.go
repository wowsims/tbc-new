package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (war *Warrior) registerRevenge() {
	actionID := core.ActionID{SpellID: 30357}

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
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: SpellMaskRevenge,
		MaxRange:       core.MaxMeleeRange,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Second * 5,
			},
		},

		RageCost: core.RageCostOptions{
			Cost:   5,
			Refund: 0.8,
		},

		DamageMultiplier: 1,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,
		FlatThreatBonus:  200,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(DefensiveStance) && aura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(414, 506)
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			aura.Deactivate(sim)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},

		RelatedSelfBuff: aura,
	})
}
