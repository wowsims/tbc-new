package shaman

import (
	"slices"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

///////////////////////////////////////////////////////////////////////////
//							Elemental
///////////////////////////////////////////////////////////////////////////

var ItemSetTidefuryRaiment = core.NewItemSet(core.ItemSet{
	ID:   630,
	Name: "Tidefury Raiment",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Chain Lightning Spell now only loses 17% of its damage per jump.
			// Implemented in chain_lightning.go
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Water Shield ability grants an additional 56 mana each time it triggers and an additional 3 mana per 5 sec.
			// Implemented in shields.go
		},
	},
})

var ItemSetCycloneRegalia = core.NewItemSet(core.ItemSet{
	ID:   632,
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
				Kind:     core.SpellMod_PowerCost_Flat,
				IntValue: -270,
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
	ID:   635,
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
				ClassSpellMask: SpellMaskLightningBolt, // Does not have Can Proc from Procs so presumably does not proc from overloads
				ProcChance:     0.25,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					character.AddMana(sim, 120, manaMetrics)
				},
			})
		},
	},
})

var ItemSetSkyshatterRegalia = core.NewItemSet(core.ItemSet{
	ID:   684,
	Name: "Skyshatter Regalia",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Whenever you have an air totem, an earth totem, a fire totem, and a water totem active at the same time,
			// you gain 15 mana per 5 sec, 35 spell critical strike rating, and up to 45 spell damage.
			shaman := agent.(ShamanAgent).GetShaman()
			aura := shaman.RegisterAura(core.Aura{
				Label:    "Totemic Mastery",
				ActionID: core.ActionID{SpellID: 38437},
				Duration: core.NeverExpires,
			}).AttachStatsBuff(stats.Stats{
				stats.MP5:             15,
				stats.SpellCritRating: 35,
				stats.SpellDamage:     45,
			})

			var periodicAction *core.PendingAction

			core.MakePermanent(shaman.RegisterAura(core.Aura{
				Label: "Totemic Mastery Periodic Check",
				OnGain: func(_ *core.Aura, sim *core.Simulation) {
					periodicAction = core.StartPeriodicAction(sim, core.PeriodicActionOptions{
						Period:          time.Second * 3,
						TickImmediately: true,
						OnAction: func(sim *core.Simulation) {
							if slices.Min(shaman.TotemExpirations[:]) < sim.CurrentTime {
								aura.Deactivate(sim)
							} else if !aura.IsActive() {
								aura.Activate(sim)
							}
						},
						CleanUp: func(sim *core.Simulation) {
							aura.Deactivate(sim)
						},
					})
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					periodicAction.Cancel(sim)
				},
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					periodicAction = nil
				},
			}))
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases the damage dealt by your Lightning Bolt ability by 5%.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Flat,
				FloatValue: 0.05,
				ClassMask:  SpellMaskLightningBolt | SpellMaskLightningBoltOverload,
			}).ExposeToAPL(38436)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//							Enhancement
///////////////////////////////////////////////////////////////////////////

var ItemSetCycloneHarness = core.NewItemSet(core.ItemSet{
	ID:   633,
	Name: "Cyclone Harness",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Strength of Earth Totem ability grants an additional 12 strength.
			// Implemented in totems.go
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Stormstrike ability does an additional 30 damage per weapon.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_BaseDamage_Flat,
				FloatValue: 30.0,
				ClassMask:  SpellMaskStormstrikeDamage,
			}).ExposeToAPL(37224)
		},
	},
})

var ItemSetCataclysmHarness = core.NewItemSet(core.ItemSet{
	ID:   636,
	Name: "Cataclysm Harness",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your melee attacks have a chance to reduce the cast time of your next Lesser Healing Wave by 1.5 sec.
			// Not implementing
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// You gain 5% additional haste from your Flurry ability.
			// Implemented in talents_enhancement.go
		},
	},
})

var ItemSetSkyshatterHarness = core.NewItemSet(core.ItemSet{
	ID:   682,
	Name: "Skyshatter Harness",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Earth Shock, Flame Shock, and Frost Shock abilities cost 10% less mana.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_PowerCost_Pct,
				FloatValue: -10,
				ClassMask:  SpellMaskShock,
			}).ExposeToAPL(38429)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Whenever you use Stormstrike, you gain 70 attack power for 12 sec.
			character := agent.GetCharacter()
			bonusPower := character.NewTemporaryStatsAura("Stormstrike AP Buff", core.ActionID{SpellID: 38432}, stats.Stats{stats.AttackPower: 70}, time.Second*12)

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:           "Skyshatter Harness 4pc",
				Callback:       core.CallbackOnCastComplete,
				ClassSpellMask: SpellMaskStormstrikeCast,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					bonusPower.Activate(sim)
				},
			})
		},
	},
})

func init() {
}
