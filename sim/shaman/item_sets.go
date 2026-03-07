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

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
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
		},
	},
})

var ItemSetCataclysmRegalia = core.NewItemSet(core.ItemSet{
	Name: "Cataclysm Regalia",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Each time you cast an offensive spell, there is a chance your next Lesser Healing Wave will cost 380 less mana.
			// Not implementing
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Lightning Bolt critical strikes have a chance to grant you 120 mana.
			character := agent.GetCharacter()
			manaMetrics := character.NewManaMetrics(core.ActionID{SpellID: 37238})

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:           "Lightning Bolt Discount",
				ActionID:       core.ActionID{ItemID: 37237},
				Callback:       core.CallbackOnSpellHitDealt,
				Outcome:        core.OutcomeCrit,
				ClassSpellMask: SpellMaskLightningBolt,
				ProcChance:     0.25,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					character.AddMana(sim, 120, manaMetrics)
				},
			})
		},
	},
})

func init() {
}
