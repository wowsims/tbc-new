package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

// Incarnate Regalia (Tier 4 - Shadow Priest)
var ItemSetIncarnateRegalia = core.NewItemSet(core.ItemSet{
	ID:   664,
	Name: "Incarnate Regalia",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Shadowfiend now has 75 more stamina and lasts 3 sec. longer.
			priest := agent.(PriestAgent).GetPriest()

			// Extend shadowfiend duration by 3 seconds
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_BuffDuration_Flat,
				TimeValue: time.Second * 3,
				ClassMask: PriestSpellShadowFiend,
			})

			// Add 75 stamina to shadowfiend
			setBonusAura.
				ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
					if priest.ShadowfiendPet != nil {
						priest.ShadowfiendPet.AddStatDynamic(sim, stats.Stamina, 75.0)
					}
				}).
				ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
					if priest.ShadowfiendPet != nil {
						priest.ShadowfiendPet.AddStatDynamic(sim, stats.Stamina, -75.0)
					}
				})
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Mind Flay and Smite spells deal 5% more damage.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Flat,
				FloatValue: 0.05,
				ClassMask:  PriestSpellMindFlay | PriestSpellSmite,
			})
		},
	},
})

// Avatar Regalia (Tier 5 - Shadow Priest)
var ItemSetAvatarRegalia = core.NewItemSet(core.ItemSet{
	ID:   666,
	Name: "Avatar Regalia",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Each time you cast an offensive spell, there is a chance your next spell will cost 150 less mana.
			// 6% proc rate. Discount is consumed on the next spell cast.
			character := agent.GetCharacter()

			discountAura := character.RegisterAura(core.Aura{
				Label:    "Avatar Regalia 2pc Discount",
				ActionID: core.ActionID{SpellID: 37601},
				Duration: time.Second * 15,
				OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
					if spell.ClassSpellMask == 0 {
						return
					}
					aura.Deactivate(sim)
				},
			}).AttachSpellMod(core.SpellModConfig{
				Kind:     core.SpellMod_PowerCost_Flat,
				IntValue: -150,
			})

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:       "Avatar Regalia 2pc",
				ProcChance: 0.06,
				ProcMask:   core.ProcMaskSpellDamage,
				Callback:   core.CallbackOnCastComplete,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					discountAura.Activate(sim)
				},
			})
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Each time your Shadow Word: Pain deals damage, it has a chance to grant your
			// next spell cast within 15 sec up to 100 damage and healing.
			// 40% proc rate. Buff is consumed on next spell cast.
			character := agent.GetCharacter()

			buffAura := character.RegisterAura(core.Aura{
				Label:    "Avatar Regalia 4pc",
				ActionID: core.ActionID{SpellID: 37604},
				Duration: time.Second * 15,
				OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
					if spell.ClassSpellMask == 0 {
						return
					}
					aura.Deactivate(sim)
				},
			}).AttachStatsBuff(stats.Stats{stats.SpellDamage: 100, stats.HealingPower: 100})

			// Trigger the buff from SWP periodic ticks
			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:           "Avatar Regalia 4pc Trigger",
				ProcChance:     0.40,
				ClassSpellMask: PriestSpellShadowWordPain,
				Callback:       core.CallbackOnPeriodicDamageDealt,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					buffAura.Activate(sim)
				},
			})
		},
	},
})

// Absolution Regalia (Tier 6 - Shadow Priest)
var ItemSetAbsolutionRegalia = core.NewItemSet(core.ItemSet{
	ID:   674,
	Name: "Absolution Regalia",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases the duration of your Shadow Word: Pain ability by 3 sec.
			// SWP ticks every 3 sec, so +3 sec = +1 tick.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_DotNumberOfTicks_Flat,
				IntValue:  1,
				ClassMask: PriestSpellShadowWordPain,
			})
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases the damage from your Mind Blast ability by 10%.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Flat,
				FloatValue: 0.10,
				ClassMask:  PriestSpellMindBlast,
			})
		},
	},
})
