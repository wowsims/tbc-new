package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	core.NewItemEffect(27683, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 6
		spellHasteRating := 320.0

		quagmirransEyeAura := character.NewTemporaryStatsAura(
			"Spell Haste Trinket",
			core.ActionID{SpellID: 33297},
			stats.Stats{stats.SpellHasteRating: spellHasteRating},
			duration,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Quagmirran's Eye",
			ActionID:           core.ActionID{ItemID: 27683},
			ProcChance:         .1,
			ICD:                time.Second * 45,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			RequireDamageDealt: true,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				quagmirransEyeAura.Activate(sim)
			},
		})

		character.ItemSwap.RegisterProc(27683, procAura)
	})

	core.NewItemEffect(28785, func(agent core.Agent) {
		character := agent.GetCharacter()

		lightningBolt := character.RegisterSpell(core.SpellConfig{
			ActionID:     core.ActionID{SpellID: 123},
			SpellSchool:  core.SpellSchoolNature,
			ProcMask:     core.ProcMaskEmpty,
			Flags:        core.SpellFlagPassiveSpell,
			MissileSpeed: 28, // this is a guess atm

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.WaitTravelTime(sim, func(s *core.Simulation) {
					baseDamage := sim.Roll(694, 806)
					spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeAlwaysHit)
				})

			},
		})

		lightningCapacitorAura := character.RegisterAura(core.Aura{
			ActionID:  core.ActionID{SpellID: 37657},
			Duration:  core.NeverExpires,
			MaxStacks: 3,
			OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
				if newStacks == 3 {
					aura.Deactivate(sim)
					lightningBolt.Cast(sim, character.CurrentTarget)
				}
			},
		})

		//procTrigger
		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "The Lightning Capacitor",
			ActionID: core.ActionID{ItemID: 28785},
			ICD:      time.Millisecond * 2500,
			Outcome:  core.OutcomeCrit,
			Callback: core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				lightningCapacitorAura.Activate(sim)
				lightningCapacitorAura.AddStack(sim)
			},
		})

		character.ItemSwap.RegisterProc(28785, procAura)
	})
}
