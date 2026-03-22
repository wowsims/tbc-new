package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// Bulwark of Azzinoth
	core.NewItemEffect(32375, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := character.NewTemporaryStatsAura(
			"Unbreakable",
			core.ActionID{SpellID: 40407},
			stats.Stats{stats.Armor: 2000},
			time.Second*10,
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Illidan Tank Shield",
			ActionID:           core.ActionID{ItemID: 32375},
			ProcMask:           core.ProcMaskDirect,
			ProcChance:         0.02,
			ICD:                time.Minute * 1,
			RequireDamageDealt: true,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitTaken,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				aura.Activate(sim)
			},
		})

		character.ItemSwap.RegisterProc(32375, procAura)
	})

	// Eye of the Night
	core.NewItemEffect(24116, func(agent core.Agent) {
		character := agent.GetCharacter()
		core.EyeOfTheNightAura(character)
	})

	// Chain of the Twilight Owl
	core.NewItemEffect(24121, func(agent core.Agent) {
		character := agent.GetCharacter()
		core.ChainOfTheTwilightOwlAura(character)
	})

	// Braided Eternium Chain
	core.NewItemEffect(24114, func(agent core.Agent) {
		character := agent.GetCharacter()
		core.BraidedEterniumChainAura(character)
	})
}
