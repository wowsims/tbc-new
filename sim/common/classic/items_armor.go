package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func init() {
	// Skullflame Shield
	core.NewItemEffect(1168, func(agent core.Agent) {
		character := agent.GetCharacter()

		drainLifeActionID := core.ActionID{SpellID: 18817}
		healthMetrics := character.NewHealthMetrics(drainLifeActionID)
		drainLifeSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    drainLifeActionID,
			SpellSchool: core.SpellSchoolShadow,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			BonusCoefficient: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				result := spell.CalcAndDealDamage(sim, target, 35, spell.OutcomeAlwaysHit)
				character.GainHealth(sim, result.Damage, healthMetrics)
			},
		})

		rollFlamestrikeDamage := func(sim *core.Simulation, _ *core.Spell) float64 {
			return sim.Roll(75, 125)
		}

		flamestrikeSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 18818},
			SpellSchool: core.SpellSchoolFire,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHit, rollFlamestrikeDamage)
			},
		})

		drainLifeTriggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:       "Drain Life Trigger",
			ProcMask:   core.ProcMaskMelee,
			ProcChance: 0.03,
			Outcome:    core.OutcomeLanded,
			Callback:   core.CallbackOnSpellHitTaken,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				drainLifeSpell.Cast(sim, spell.Unit)
			},
		})

		flameStrikeTriggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:       "Flamestrike Trigger",
			ProcMask:   core.ProcMaskMelee,
			ProcChance: 0.01,
			Outcome:    core.OutcomeLanded,
			Callback:   core.CallbackOnSpellHitTaken,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				flamestrikeSpell.Cast(sim, result.Target)
			},
		})

		character.ItemSwap.RegisterProc(1168, drainLifeTriggerAura)
		character.ItemSwap.RegisterProc(1168, flameStrikeTriggerAura)
	})

	// Force Reactive Disk
	core.NewItemEffect(18168, func(agent core.Agent) {
		character := agent.GetCharacter()

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: 18168},
			SpellSchool: core.SpellSchoolNature,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

			DamageMultiplier: 1,
			CritMultiplier:   character.DefaultSpellCritMultiplier(),
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				spell.CalcAndDealAoeDamage(sim, 25, spell.OutcomeMagicHitAndCrit)
			},
		})

		aura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Force Reactive Disk",
			ProcMask: core.ProcMaskMelee,
			ICD:      time.Second,
			Outcome:  core.OutcomeBlock,
			Callback: core.CallbackOnSpellHitTaken,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				spell.Cast(sim, result.Target)
			},
		})

		character.ItemSwap.RegisterProc(18168, aura)
	})
}
