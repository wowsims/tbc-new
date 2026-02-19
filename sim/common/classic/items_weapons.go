package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// Thunderfury, Blessed Blade of the Windseeker
	core.NewItemEffect(19019, func(agent core.Agent) {
		character := agent.GetCharacter()

		procActionID := core.ActionID{SpellID: 21992}

		attackSpeedDebuffAura := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			aura := target.GetOrRegisterAura(core.Aura{
				Label:    "Cyclone",
				ActionID: core.ActionID{SpellID: 27648},
				Duration: time.Second * 12,
			})

			core.AtkSpeedReductionEffect(aura, 1.2)

			return aura
		})

		singleTargetSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    procActionID.WithTag(1),
			SpellSchool: core.SpellSchoolNature,
			ProcMask:    core.ProcMaskEmpty,

			DamageMultiplier: 1,
			CritMultiplier:   character.DefaultSpellCritMultiplier(),
			ThreatMultiplier: 0.5,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				result := spell.CalcAndDealDamage(sim, target, 300, spell.OutcomeMagicHitAndCrit)
				if result.Landed() {
					attackSpeedDebuffAura.Get(result.Target).Activate(sim)
				}
			},
		})

		resistanceDebuffAura := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return target.GetOrRegisterAura(core.Aura{
				Label:    "Thunderfury",
				ActionID: procActionID,
				Duration: time.Second * 12,
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					target.AddStatDynamic(sim, stats.NatureResistance, -25)
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					target.AddStatDynamic(sim, stats.NatureResistance, 25)
				},
			})
		})

		bounceSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    procActionID.WithTag(2),
			SpellSchool: core.SpellSchoolNature,
			ProcMask:    core.ProcMaskEmpty,

			ThreatMultiplier: 1,
			FlatThreatBonus:  63,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				results := spell.CalcCleaveDamage(sim, target, 5, 0, spell.OutcomeMagicHit)
				for _, result := range results {
					if result.Landed() {
						resistanceDebuffAura.Get(result.Target).Activate(sim)
					}
				}
				spell.DealBatchedAoeDamage(sim)
			},
		})

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Thunderfury",
			Callback: core.CallbackOnSpellHitDealt,
			Outcome:  core.OutcomeLanded,
			DPM:      character.NewDynamicLegacyProcForWeapon(19019, 6, 0),
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				singleTargetSpell.Cast(sim, result.Target)
				bounceSpell.Cast(sim, result.Target)
			},
		})

		character.ItemSwap.RegisterProc(19019, procAura)
	})

}
