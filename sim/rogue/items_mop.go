package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

var PVPSet = core.NewItemSet(core.ItemSet{
	Name: "Gladiator's Vestments",
	ID:   1113,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Nothing relevant
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			character := agent.GetCharacter()
			metric := character.NewEnergyMetrics(core.ActionID{SpellID: 21975})

			setBonusAura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
				character.UpdateMaxEnergy(sim, 30, metric)
			})
			setBonusAura.ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
				character.UpdateMaxEnergy(sim, -30, metric)
			})
			setBonusAura.ExposeToAPL(21975)
		},
	},
})

var Tier14 = core.NewItemSet(core.ItemSet{
	Name:                    "Battlegear of the Thousandfold Blades",
	DisabledInChallengeMode: true,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases the damage done by your Venomous Wounds ability by 20%,
			// increases the damage done by your Sinister Strike ability by 15%,
			// and increases the damage done by your Backstab ability by 10%.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Pct,
				ClassMask:  RogueSpellVenomousWounds,
				FloatValue: 0.2,
			})
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Pct,
				ClassMask:  RogueSpellSinisterStrike,
				FloatValue: 0.15,
			})
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Pct,
				ClassMask:  RogueSpellBackstab,
				FloatValue: 0.1,
			})
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases the duration of your Shadow Blades ability by Combat: 6 /  Assassination,  Subtlety: 12 sec.
			rogue := agent.(RogueAgent).GetRogue()
			addTime := time.Second * time.Duration(core.Ternary(rogue.Spec == proto.Spec_SpecCombatRogue, 6, 12))
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_BuffDuration_Flat,
				ClassMask: RogueSpellShadowBlades,
				TimeValue: addTime,
			})
		},
	},
})

var Tier15 = core.NewItemSet(core.ItemSet{
	Name:                    "Nine-Tail Battlegear",
	DisabledInChallengeMode: true,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases the duration of your finishing moves as if you had used an additional combo point, up to a maximum of 6 combo points.
			rogue := agent.(RogueAgent).GetRogue()

			rogue.Has2PT15 = true
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Shadow Blades also reduces the cost of all your abilities by 15%.
			// Additionally, reduces the GCD of all rogue abilities by 300ms
			rogue := agent.(RogueAgent).GetRogue()
			energyMod := rogue.AddDynamicMod(core.SpellModConfig{
				Kind:       core.SpellMod_PowerCost_Pct,
				ClassMask:  RogueSpellsAll,
				FloatValue: -0.15,
			})
			gcdMod := rogue.AddDynamicMod(core.SpellModConfig{
				Kind:      core.SpellMod_GlobalCooldown_Flat,
				ClassMask: RogueSpellActives,
				TimeValue: time.Millisecond * -300,
			})
			aura := rogue.RegisterAura(core.Aura{
				Label:    "Shadow Blades Energy Cost Reduction",
				ActionID: core.ActionID{SpellID: 138151},
				Duration: time.Second * 12,
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					energyMod.Activate()
					gcdMod.Activate()
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					energyMod.Deactivate()
					gcdMod.Deactivate()
				},
			})

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:           "Rogue T15 4P Bonus",
				Callback:       core.CallbackOnCastComplete,
				ClassSpellMask: RogueSpellShadowBlades,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					aura.Activate(sim)
				},
			})
		},
	},
})

var Tier16 = core.NewItemSet(core.ItemSet{
	Name:                    "Barbed Assassin Battlegear",
	DisabledInChallengeMode: true,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// When you generate a combo point from Revealing Strike's effect, Honor Among Thieves, or Seal Fate
			// your next combo point generating ability has its energy cost reduced by {Subtlety: 2, Assassination: 6, Combat: 15].
			// Stacks up to 5 times.
			rogue := agent.(RogueAgent).GetRogue()

			energyReduction := 0
			switch rogue.Spec {
			case proto.Spec_SpecSubtletyRogue:
				energyReduction = -2
			case proto.Spec_SpecAssassinationRogue:
				energyReduction = -6
			default:
				energyReduction = -15
			}

			energyMod := rogue.AddDynamicMod(core.SpellModConfig{
				Kind:      core.SpellMod_PowerCost_Flat,
				ClassMask: RogueSpellGenerator,
				IntValue:  0, // Set dynamically
			})

			// This aura gets activated by the applicable spell scripts
			rogue.T16EnergyAura = rogue.RegisterAura(core.Aura{
				Label:     "Silent Blades",
				ActionID:  core.ActionID{SpellID: 145193},
				Duration:  time.Second * 30,
				MaxStacks: 5,
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					energyMod.UpdateIntValue(aura.GetStacks() * int32(energyReduction))
					energyMod.Activate()
				},
				OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
					energyMod.UpdateIntValue(aura.GetStacks() * int32(energyReduction))
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					energyMod.Deactivate()
				},
				OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if result.Landed() && spell.Flags.Matches(SpellFlagBuilder) && spell.DefaultCast.Cost > 0 {
						// Free action casts (such as Dispatch w/ Blindside) will not consume the aura
						aura.Deactivate(sim)
					}
				},
			})
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Killing Spree deals 10% more damage every time it strikes a target.
			// Abilities against a target with Vendetta on it increase your mastery by 250 for 5 sec, stacking up to 20 times.
			// Every time you Backstab, you have a 4% chance to replace your Backstab with Ambush that can be used regardless of Stealth.
			rogue := agent.(RogueAgent).GetRogue()

			if rogue.Spec == proto.Spec_SpecSubtletyRogue {
				aura := rogue.RegisterAura(core.Aura{
					Label:    "Sleight of Hand",
					ActionID: core.ActionID{SpellID: 145211},
					Duration: time.Second * 10,
					OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						if spell == rogue.Ambush {
							aura.Deactivate(sim)
						}
					},
				})

				setBonusAura.AttachProcTrigger(core.ProcTrigger{
					Name:           "Rogue T16 4P Bonus",
					Callback:       core.CallbackOnApplyEffects,
					ClassSpellMask: RogueSpellBackstab,
					Outcome:        core.OutcomeLanded,
					ProcChance:     0.04,
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						aura.Activate(sim)
					},
				})
			} else if rogue.Spec == proto.Spec_SpecCombatRogue {
				rogue.T16SpecMod = rogue.AddDynamicMod(core.SpellModConfig{
					Kind:       core.SpellMod_DamageDone_Pct,
					ClassMask:  RogueSpellKillingSpreeHit,
					FloatValue: 0.1, // Set dynamically in Killing Spree
				})
			} else if rogue.Spec == proto.Spec_SpecAssassinationRogue {
				aura := rogue.RegisterAura(core.Aura{
					Label:     "Toxicologist",
					ActionID:  core.ActionID{SpellID: 145249},
					Duration:  time.Second * 5,
					MaxStacks: 20,
					OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
						change := newStacks - oldStacks
						aura.Unit.AddStatDynamic(sim, stats.MasteryRating, float64(250*change))
					},
				})

				setBonusAura.AttachProcTrigger(core.ProcTrigger{
					Name:           "Rogue T16 4P Bonus",
					Callback:       core.CallbackOnCastComplete,
					ClassSpellMask: RogueSpellVendetta,
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						aura.Activate(sim)
					},
				})

				setBonusAura.AttachProcTrigger(core.ProcTrigger{
					Name:           "Toxicologist Trigger",
					Callback:       core.CallbackOnSpellHitDealt,
					ClassSpellMask: RogueSpellActives,
					ExtraCondition: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
						return rogue.Vendetta.RelatedAuraArrays.AnyActive(result.Target)
					},
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						aura.Activate(sim)
						aura.AddStack(sim)
					},
				})
			}
		},
	},
})

// https://www.wowhead.com/mop-classic/item-set=1087/fangs-of-the-father#comments:id=1706397
// Assuming mob level 93 for the reduced chance modifier
func getFangsProcRate(character *core.Character) float64 {
	switch character.Spec {
	case proto.Spec_SpecSubtletyRogue:
		return 0.28223 * 0.5
	case proto.Spec_SpecAssassinationRogue:
		return 0.23139 * 0.5
	default:
		return 0.09438 * 0.5
	}
}

// Golad + Tiriosh
var FangsOfTheFather = core.NewItemSet(core.ItemSet{
	Name:  "Fangs of the Father",
	Slots: core.AllWeaponSlots(),
	Bonuses: map[int32]core.ApplySetBonus{
		// Your melee attacks have a chance to grant Shadows of the Destroyer, increasing your Agility by 17, stacking up to 50 times.
		// Each application past 30 grants an increasing chance to trigger Fury of the Destroyer.
		// When triggered, this consumes all applications of Shadows of the Destroyer, immediately granting 5 combo points and cause your finishing moves to generate 5 combo points.
		// Lasts 6 sec.

		// Tooltip is deceptive. The stacks of Shadows of the Destroyer only clear when the 5 Combo Point effect ends
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			character := agent.GetCharacter()
			cpMetrics := character.NewComboPointMetrics(core.ActionID{SpellID: 109950})

			agiAura := core.MakeStackingAura(character, core.StackingStatAura{
				Aura: core.Aura{
					Label:     "Shadows of the Destroyer",
					ActionID:  core.ActionID{SpellID: 109941},
					Duration:  time.Second * 30,
					MaxStacks: 50,
				},
				BonusPerStack: stats.Stats{stats.Agility: 17},
			})

			wingsProc := character.GetOrRegisterAura(core.Aura{
				Label:    "Fury of the Destroyer",
				ActionID: core.ActionID{SpellID: 109949},
				Duration: time.Second * 6,
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					aura.Unit.AddComboPoints(sim, 5, cpMetrics)
				},
				OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
					if spell.Flags.Matches(SpellFlagFinisher) {
						aura.Unit.AddComboPoints(sim, 5, cpMetrics)
					}
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					agiAura.SetStacks(sim, 0)
					agiAura.Deactivate(sim)
				},
			})

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:       "Rogue Legendary Daggers Stage 3",
				Callback:   core.CallbackOnSpellHitDealt,
				ProcMask:   core.ProcMaskMelee,
				Outcome:    core.OutcomeLanded,
				ProcChance: getFangsProcRate(character),
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					// Adding a stack and activating the combo point effect is mutually exclusive.
					// Agility bonus is lost when combo point effect ends
					stacks := float64(agiAura.GetStacks())
					if stacks > 30 && !wingsProc.IsActive() {
						if stacks == 50 || sim.Proc(1.0/(50-stacks), "Fangs of the Father") {
							wingsProc.Activate(sim)
						} else {
							agiAura.Activate(sim)
							agiAura.AddStack(sim)
						}
					} else {
						agiAura.Activate(sim)
						agiAura.AddStack(sim)
					}
				},
			})
		},
	},
})

// 45% SS/RvS Modifier for Legendary MH Dagger
func makeWeightedBladesModifier(itemID int32) {
	core.NewItemEffect(itemID, func(agent core.Agent, _ proto.ItemLevelState) {
		agent.GetCharacter().AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Pct,
			FloatValue: 0.45,
			ClassMask:  RogueSpellWeightedBlades,
		})
	})
}

func init() {
	makeWeightedBladesModifier(77949)
}
