package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// Dungeon Set 3
var ItemSetOblivionRaiment = core.NewItemSet(core.ItemSet{
	ID:   644,
	Name: "Oblivion Raiment",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Grants your pet 45 mana per 5 sec.
			// Pet Mana Regen - 37375
			petAgents := agent.GetCharacter().PetAgents
			if petAgents != nil {
				agent.GetCharacter().PetAgents[0].GetCharacter().AddStat(stats.MP5, 45.0)
			}

		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Seed of Corruption deals 180 additional damage when it detonates.
			// Improved Seed of Corruption - 37376
			char := agent.GetCharacter()
			if char.Class != proto.Class_ClassWarlock {
				return
			}

			char.AddDynamicMod(core.SpellModConfig{
				Kind:       core.SpellMod_BonusSpellDamage_Flat,
				FloatValue: 180.0,
				ClassMask:  WarlockSpellSeedOfCorruptionExplosion,
			})
		},
	},
})

// T4
var ItemSetVoidheartRaiment = core.NewItemSet(core.ItemSet{
	ID:   645,
	Name: "Voidheart Raiment",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			warlock := agent.(WarlockAgent).GetWarlock()
			// Your shadow damage spells have a chance to grant you 135 bonus shadow damage for 15 sec.
			// Shadowflame - 37377
			shadowBonus := warlock.NewTemporaryStatsAura("Shadowflame", core.ActionID{SpellID: 37377}, stats.Stats{stats.ShadowDamage: 135}, time.Second*15)

			// Your fire damage spells have a chance to grant you 135 bonus fire damage for 15 sec.
			// Hellfire - 39437
			fireBonus := warlock.NewTemporaryStatsAura("Shadowflame Hellfire", core.ActionID{SpellID: 39437}, stats.Stats{stats.FireDamage: 135}, time.Second*15)

			warlock.RegisterAura(core.Aura{
				Label:    "Voidheart Raiment 2pc",
				Duration: time.Second * 15,
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					aura.Activate(sim)
				},
				OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {

					if !sim.Proc(0.05, "voidheart4pc") {
						return
					}
					if spell.SpellSchool.Matches(core.SpellSchoolShadow) {
						shadowBonus.Activate(sim)
					}
					if spell.SpellSchool.Matches(core.SpellSchoolFire) {
						fireBonus.Activate(sim)
					}
				},
			})
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases the duration of your Corruption and Immolate abilities by 3 sec.
			// Improved Corruption and Immolate - 37380
			warlock := agent.(WarlockAgent).GetWarlock()

			warlock.AddDynamicMod(core.SpellModConfig{
				Kind:      core.SpellMod_DotNumberOfTicks_Flat,
				IntValue:  1,
				ClassMask: WarlockSpellCorruption | WarlockSpellImmolate,
			})
		},
	},
})

// T5
var ItemSetCorruptorRaiment = core.NewItemSet(core.ItemSet{
	ID:   646,
	Name: "Corruptor Raiment",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Causes your pet to be healed for 15% of the damage you deal.
			// Pet Healing - 37381
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Shadowbolt spell hits increase the damage of Corruption by 10% and your Incinerate spell hits increase the damage of Immolate by 10%.
			// Improved Corruption and Immolate - 37384
			warlock := agent.(WarlockAgent).GetWarlock()
			warlock.RegisterAura(core.Aura{
				Label:    "Corruptor Raiment 4pc",
				Duration: core.NeverExpires,
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					aura.Activate(sim)
				},
				OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {

					if !(spell.ClassSpellMask != WarlockSpellShadowBolt && spell.ClassSpellMask != WarlockSpellIncinerate) || result.Outcome != core.OutcomeHit {
						return
					}
					if spell.ClassSpellMask == WarlockSpellShadowBolt {
						warlock.AddDynamicMod(core.SpellModConfig{
							Kind:       core.SpellMod_DotDamageDone_Pct,
							FloatValue: 0.10,
							ClassMask:  WarlockSpellShadowBolt,
						}).Activate()
					}
					if spell.ClassSpellMask == WarlockSpellIncinerate {
						warlock.AddDynamicMod(core.SpellModConfig{
							Kind:       core.SpellMod_DotDamageDone_Pct,
							FloatValue: 0.10,
							ClassMask:  WarlockSpellImmolate,
						}).Activate()
					}
				},
			})
		},
	},
})

// T6
var ItemSetMaleficRaiment = core.NewItemSet(core.ItemSet{
	ID:   670,
	Name: "Malefic Raiment",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Each time one of your Corruption or Immolate spells deals periodic damage, you heal 70 health.
			// Dot Heals - 38394
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases damage done by shadowbolt and incinerate by 6%.
			// Improved Shadow Bolt and Incinerate - 38393
			warlock := agent.(WarlockAgent).GetWarlock()

			warlock.AddStaticMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Pct,
				FloatValue: 0.06,
				ClassMask:  WarlockSpellShadowBolt | WarlockSpellIncinerate,
			})
		},
	},
})

func init() {
	core.NewItemEffect(19337, func(agent core.Agent) {
		// The Black Book
		warlockPet := agent.(WarlockAgent).GetWarlock().ActivePet
		warlockPet.NewTemporaryStatsAura("Blessing of The Black Book", core.ActionID{SpellID: 23720}, stats.Stats{stats.SpellDamage: 200, stats.AttackPower: 325, stats.Armor: 1600}, time.Second*30)
	})

	core.NewItemEffect(30449, func(agent core.Agent) {
		// Void Star Talisman
		warlockPet := agent.(WarlockAgent).GetWarlock().ActivePet
		warlockPet.AddStats(stats.Stats{
			stats.SpellDamage:      48,
			stats.ArcaneResistance: 130,
			stats.FireResistance:   130,
			stats.FrostResistance:  130,
			stats.NatureResistance: 130,
			stats.ShadowResistance: 130,
		})
	})

	core.NewItemEffect(32493, func(agent core.Agent) {
		// Ashtongue Talisman of Shadows
		warlock := agent.(WarlockAgent).GetWarlock()
		warlock.RegisterAura(core.Aura{
			Label:    "Corruptor Raiment 4pc",
			Duration: core.NeverExpires,
			OnReset: func(aura *core.Aura, sim *core.Simulation) {
				aura.Activate(sim)
			},
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {

				if !(spell.ClassSpellMask != WarlockSpellCorruption) || result.Outcome != core.OutcomeHit {
					return
				}

				if !sim.Proc(0.20, "Ashtongue Talisman of Shadows") {
					return
				}

				warlock.NewTemporaryStatsAura("Ashtongue Talisman of Shadows Proc", core.ActionID{SpellID: 40478}, stats.Stats{stats.SpellDamage: 220}, time.Second*5)
			},
		})
	})
}
