package paladin

import (
	"github.com/wowsims/tbc/sim/core"
)

// T4 Ret - Justicar Battlegear
// (2) Set: Increases the damage bonus of your Judgement of the Crusader by 15%.
// (4) Set: Increases the damage dealt by your Judgement of Command by 10%.
var ItemSetJusticarBattlegear = core.NewItemSet(core.ItemSet{
	ID:   626,
	Name: "Justicar Battlegear",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.ExposeToAPL(37186)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Flat,
				FloatValue: 0.10,
				ClassMask:  SpellMaskJudgementOfCommand,
			}).ExposeToAPL(37187)
		},
	},
})

// T5 Ret - Crystalforge Battlegear
// (2) Set: Reduces the cost of your Judgements by 35.
// (4) Set: Each time you cast a Judgement, there is a chance it will heal all nearby party members for 244 to 256.
var ItemSetCrystalforgeBattlegear = core.NewItemSet(core.ItemSet{
	ID:   629,
	Name: "Crystalforge Battlegear",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_PowerCost_Flat,
				ClassMask: SpellMaskJudgement,
				IntValue:  -35,
			}).ExposeToAPL(37194)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			paladin := agent.(PaladinAgent).GetPaladin()

			divineLightHeal := paladin.RegisterSpell(core.SpellConfig{
				ActionID:    core.ActionID{SpellID: 37196},
				SpellSchool: core.SpellSchoolHoly,
				ProcMask:    core.ProcMaskSpellHealing,
				Flags:       core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

				DamageMultiplier: 1,
				ThreatMultiplier: 1,

				ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					for _, agent := range paladin.Party.Players {
						spell.CalcAndDealHealing(sim, &agent.GetCharacter().Unit, sim.Roll(244, 256), spell.OutcomeAlwaysHit)
					}
				},
			})

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:           "Judgement Group Heal",
				Callback:       core.CallbackOnCastComplete,
				ClassSpellMask: SpellMaskAllJudgements,
				ProcChance:     0.06,

				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					divineLightHeal.Cast(sim, &paladin.Unit)
				},
			}).ExposeToAPL(37195)
		},
	},
})

// T6 Ret - Lightbringer Battlegear
// (2) Set: Your melee attacks have a chance to grant you 50 mana.
// (4) Set: Increases the damage dealt by your Hammer of Wrath ability by 10%.
var ItemSetLightbringerBattlegear = core.NewItemSet(core.ItemSet{
	ID:   680,
	Name: "Lightbringer Battlegear",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			paladin := agent.(PaladinAgent).GetPaladin()
			manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: 38428})

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Name:       "Mana Regen Proc",
				Callback:   core.CallbackOnSpellHitDealt,
				ProcMask:   core.ProcMaskMelee,
				ProcChance: 0.20,

				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					paladin.AddMana(sim, 50, manaMetrics)
				},
			}).ExposeToAPL(38427)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Flat,
				ClassMask:  SpellMaskHammerOfWrath,
				FloatValue: 0.10,
			}).ExposeToAPL(38424)
		},
	},
})
