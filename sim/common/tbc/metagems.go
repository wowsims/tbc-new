package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// Keep these in order by item ID

	// Destructive Skyfire Diamond
	// +14 Spell Crit Rating and 1% Spell Reflect
	// core.NewItemEffect(25890, func(agent core.Agent) {})

	// Mystical Skyfire Diamond
	// Chance to Increase Spell Cast Speed
	core.NewItemEffect(25893, func(agent core.Agent) {
		character := agent.GetCharacter()
		procAura := character.NewTemporaryStatsAura("Focus", core.ActionID{SpellID: 18803}, stats.Stats{stats.SpellHasteRating: 320}, time.Second*4)

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:       "Mystical Skyfire Diamond",
			ProcChance: 0.15,
			ICD:        time.Second * 35,
			Callback:   core.CallbackOnCastComplete,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				procAura.Activate(sim)
			},
		})
	})

	// Swift Skyfire Diamond
	// +24 Attack Power and Minor Run Speed Increase
	// core.NewItemEffect(25894, func(agent core.Agent) {})

	// Enigmatic Skyfire Diamond
	// +12 Critical Strike Rating & 5% Snare and Root Resist
	// core.NewItemEffect(25895, func(agent core.Agent) {})

	// Powerful Earthstorm Diamond,
	// +18 Stamina & 5% Stun Resist
	// core.NewItemEffect(25896, func(agent core.Agent) {})

	// Bracing Earthstorm Diamond
	// +26 Healing +9 Spell Damage and 2% Reduced Threat
	core.NewItemEffect(25897, func(agent core.Agent) {
		character := agent.GetCharacter()
		character.AddStats(stats.Stats{
			stats.HealingPower: 26,
			stats.SpellDamage:  9,
		})
		character.PseudoStats.ThreatMultiplier *= 0.98
	})

	// Tenacious Earthstorm Diamond
	// +12 Defense Rating & Chance to Restore Health on hit
	// core.NewItemEffect(25898, func(agent core.Agent) {})

	// Brutal Earthstorm Diamond
	// +3 Melee Damage & Chance to Stun Target
	core.NewItemEffect(25899, func(agent core.Agent) {
		character := agent.GetCharacter()
		character.AddStat(stats.PhysicalDamage, 3)
	})

	// Insightful Earthstorm Diamond
	// +12 Intellect & Chance to restore mana on spellcast
	// core.NewItemEffect(25901, func(agent core.Agent) {})

	// Relentless Earthstorm Diamond
	// +12 Agility & 3% Increased Critical Damage
	core.NewItemEffect(32409, core.ApplyMetaGemCriticalDamageEffect)

	// Thundering Skyfire Diamond
	// Chance to Increase Melee/Ranged Attack Speed
	core.NewItemEffect(32410, func(agent core.Agent) {
		character := agent.GetCharacter()
		procAura := character.NewTemporaryStatsAura("Skyfire Swiftness", core.ActionID{SpellID: 39959}, stats.Stats{stats.MeleeHasteRating: 240}, time.Second*6)

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Thundering Skyfire Diamond",
			DPM:      character.NewLegacyPPMManager(1.5, core.ProcMaskWhiteHit),
			ICD:      time.Second * 40,
			Outcome:  core.OutcomeLanded,
			Callback: core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				procAura.Activate(sim)
			},
		})

	})

	// Chaotic Skyfire Diamond
	// +12 Spell Critical & 3% Increased Critical Damage
	core.NewItemEffect(34220, core.ApplyMetaGemCriticalDamageEffect)

	// Swift Starfire Diamond
	// +12 Spell Damage and Minor Run Speed Increase
	// core.NewItemEffect(28557, func(agent core.Agent) {})

	// Swift Windfire Diamond
	// +20 Attack Power and Minor Run Speed Increase
	// core.NewItemEffect(28556, func(agent core.Agent) {})

	// Potent Unstable Diamond
	// +24 Attack Power and 5% Stun Resistance
	// core.NewItemEffect(32640, func(agent core.Agent) {})

	// Imbued Unstable Diamond
	// +14 Spell Damage & 5% Stun Resistance
	// core.NewItemEffect(32641, func(agent core.Agent) {})

	// Eternal Earthstorm Diamond
	// +12 Defense Rating & +10% Shield Block Value
	core.NewItemEffect(35501, func(agent core.Agent) {
		character := agent.GetCharacter()
		character.PseudoStats.BlockValueMultiplier *= 1.1
	})

	// Ember Skyfire Diamond
	// +14 Spell Damage & +2% Intellect
	core.NewItemEffect(35503, func(agent core.Agent) {
		character := agent.GetCharacter()
		character.MultiplyStat(stats.Intellect, 1.02)
	})
}
