package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

var ItemSetCycloneRegalia = core.NewItemSet(core.ItemSet{
	Name: "Cyclone Regalia",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Wrath of Air Totem ability grants an additional 20 spell damage.
			// Implemented in totems.go
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your offensive spell critical strikes have a chance to reduce the base mana cost of your next spell by 270.
			character := agent.GetCharacter()

			aura := character.RegisterAura(core.Aura{
				Label:    "Energized (Cyclone Regalia)",
				ActionID: core.ActionID{SpellID: 37214},
				Duration: time.Second * 15,
				OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
					// Has the Only Proc From Class Abilities flag
					if spell.ClassSpellMask == 0 {
						return
					}

					aura.Deactivate(sim)
				},
			}).AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_PowerCost_Flat,
				FloatValue: -270,
			})

			procAura := character.MakeProcTriggerAura(core.ProcTrigger{
				Name:       "Cyclone Regalia",
				ActionID:   core.ActionID{ItemID: 37214},
				Callback:   core.CallbackOnSpellHitDealt,
				Outcome:    core.OutcomeCrit,
				ProcMask:   core.ProcMaskSpellDamage,
				ProcChance: 0.11,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					aura.Activate(sim)
				},
			})

			character.ItemSwap.RegisterProc(33506, procAura)
		},
	},
})

func init() {
}
