package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

var overpowerRank = genRanks.Overpower.BySpellID(11585)

func (war *Warrior) registerOverpower() {
	actionID := core.ActionID{SpellID: overpowerRank.SpellID}

	aura := war.RegisterAura(core.Aura{
		ActionID: actionID,
		Label:    "Overpower Aura",
		Duration: time.Second * 5,
	})

	war.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Overpower - Trigger",
		TriggerImmediately: true,
		Outcome:            core.OutcomeDodge,
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
		ClassSpellMask: SpellMaskOverpower,
		MaxRange:       core.MaxMeleeRange,

		RageCost: core.RageCostOptions{
			Cost:   overpowerRank.Cost,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: overpowerRank.GCD,
			},
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: overpowerRank.Cooldown,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 0.75,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(BattleStance)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			overpowerVal, _ := overpowerRank.Direct.Range()
			baseDamage := overpowerVal + spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialNoBlockDodgeParry)
			aura.Deactivate(sim)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}
