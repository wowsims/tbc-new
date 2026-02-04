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
}
