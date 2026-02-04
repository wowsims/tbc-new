package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {

	core.NewItemEffect(21670, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 30
		arpAura := core.MakeStackingAura(
			character,
			core.StackingStatAura{
				Aura: core.Aura{
					Label:     "Insight of the Qiraji",
					ActionID:  core.ActionID{SpellID: 26481},
					Duration:  duration,
					MaxStacks: 6,
				},
				BonusPerStack: stats.Stats{stats.ArmorPenetration: 200},
			},
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Badge of the Swarmguard",
			ActionID:           core.ActionID{ItemID: 26480},
			DPM:                character.NewLegacyPPMManager(10, core.ProcMaskMeleeOrRanged),
			Duration:           duration,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			RequireDamageDealt: true,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if arpAura.IsActive() && arpAura.GetStacks() == arpAura.MaxStacks {
					return
				}
				arpAura.Activate(sim)
				arpAura.AddStack(sim)
			},
		})

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID: core.ActionID{ItemID: 21670},
			Flags:    core.SpellFlagNoOnCastComplete,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 3,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				procAura.Activate(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell:    spell,
			Type:     core.CooldownTypeDPS,
			BuffAura: arpAura,
		})

		character.ItemSwap.RegisterProc(21670, procAura)
		character.ItemSwap.RegisterActive(21670)
	})
}
