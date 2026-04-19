package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

// PVP - S1, S2, S3, S4 Sets
var ItemSetGladiatorsSanctuary = core.NewItemSet(core.ItemSet{
	ID:      584,
	Name:    "Gladiator's Sanctuary",
	Bonuses: pvpResilience2PBonus(32145),
})

var ItemSetGladiatorsRefuge = core.NewItemSet(core.ItemSet{
	ID:      685,
	Name:    "Gladiator's Refuge",
	Bonuses: pvpResilience2PBonus(40043),
})

var ItemSetGladiatorsWildhide = core.NewItemSet(core.ItemSet{
	ID:      585,
	Name:    "Gladiator's Wildhide",
	Bonuses: pvpResilience2PBonus(40042),
})

// PVP - PvP Rare Set
var ItemSetOathboundsKodohideBattlegear = core.NewItemSet(core.ItemSet{
	ID:      2027,
	Name:    "Oathbound's Kodohide Battlegear",
	Bonuses: pvpResilience2PBonus(46437),
})

var ItemSetOathboundsDragonhideBattlegear = core.NewItemSet(core.ItemSet{
	ID:      2025,
	Name:    "Oathbound's Dragonhide Battlegear",
	Bonuses: pvpResilience2PBonus(46435),
})

var ItemSetOathboundsWyrmhideBattlegear = core.NewItemSet(core.ItemSet{
	ID:      2026,
	Name:    "Oathbound's Wyrmhide Battlegear",
	Bonuses: pvpResilience2PBonus(46436),
})

func pvpResilience2PBonus(spellID int32) map[int32]core.ApplySetBonus {
	return map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachStatBuff(stats.ResilienceRating, 35).ExposeToAPL(spellID)
		},
	}
}

// Druid Dungeon Set
var ItemSetMoongladeRaiment = core.NewItemSet(core.ItemSet{
	ID:   637,
	Name: "Moonglade Raiment",
	Bonuses: map[int32]core.ApplySetBonus{
		// Your Rejuvenation spell now also grants 35 dodge rating.
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// TODO: Not implemented (Spell ID 37286)
		},
		// Reduces the mana cost of all shapeshifting by 25%.
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				ClassMask:  DruidSpellCatForm | DruidSpellBearForm,
				Kind:       core.SpellMod_PowerCost_Pct,
				FloatValue: -0.25,
			})
		},
	},
})

// Feral T4
var ItemSetMalorneHarness = core.NewItemSet(core.ItemSet{
	ID:   640,
	Name: "Malorne Harness",
	Bonuses: map[int32]core.ApplySetBonus{
		// Your melee attacks in Cat Form have a chance to generate 20 additional energy.
		// Your melee attacks in Bear Form and Dire Bear Form have a chance to generate 10 additional rage.
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			druid := agent.(DruidAgent).GetDruid()

			energyMetrics := druid.NewEnergyMetrics(core.ActionID{SpellID: 37311})
			rageMetrics := druid.NewRageMetrics(core.ActionID{SpellID: 37306})

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Callback:   core.CallbackOnSpellHitDealt,
				ProcMask:   core.ProcMaskMelee,
				Outcome:    core.OutcomeLanded,
				ProcChance: 0.04,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if druid.InForm(Cat) {
						druid.AddEnergy(sim, 20, energyMetrics)
					} else if druid.InForm(Bear) {
						druid.AddRage(sim, 10, rageMetrics)
					}
				},
			})
		},
		// Increases your strength by 30 in Cat Form.
		// Increases your armor by 1400 in Bear Form and Dire Bear Form.
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			druid := agent.(DruidAgent).GetDruid()
			if druid.CatFormAura != nil {
				druid.CatFormAura.AttachStatsBuff(stats.Stats{stats.Strength: 30})
			}
			if druid.BearFormAura != nil {
				druid.BearFormAura.AttachStatsBuff(stats.Stats{stats.Armor: 1400})
			}
		},
	},
})

// Balance T4
var ItemSetMalorneRegalia = core.NewItemSet(core.ItemSet{
	ID:   639,
	Name: "Malorne Regalia",
	Bonuses: map[int32]core.ApplySetBonus{
		// Your harmful spells have a chance to restore up to 120 mana.
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			druid := agent.(DruidAgent).GetDruid()
			manaMetrics := druid.NewManaMetrics(core.ActionID{SpellID: 37295 /* T4 2P Mana Restore */})

			setBonusAura.AttachProcTrigger(core.ProcTrigger{
				Callback:   core.CallbackOnCastComplete,
				ProcMask:   core.ProcMaskSpellDamage,
				ProcChance: 0.05,
				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					druid.AddMana(sim, 120, manaMetrics)
				},
			})

		},
		// Reduces the cooldown on your Innervate ability by 48 sec.
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				ClassMask: DruidSpellInnervate,
				Kind:      core.SpellMod_Cooldown_Flat,
				TimeValue: -48 * time.Second,
			})
		},
	},
})

// Feral T5
var ItemSetNordrassilHarness = core.NewItemSet(core.ItemSet{
	ID:   641,
	Name: "Nordrassil Harness",
	Bonuses: map[int32]core.ApplySetBonus{
		// Your Shred ability deals an additional 75 damage, and your Lacerate
		// ability does an additional 15 per application.
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			druid := agent.(DruidAgent).GetDruid()
			druid.ShredFlatBonus += 75
			druid.LacerateTickBonus += 15
		},
	},
})

// Balance T5
var ItemSetNordrassilRegalia = core.NewItemSet(core.ItemSet{
	ID:   643,
	Name: "Nordrassil Regalia",
	Bonuses: map[int32]core.ApplySetBonus{
		// When you shift out of Moonkin Form, your next Regrowth spell costs 450 less mana.
		2: func(agent core.Agent, setBonusAura *core.Aura) {
		},
		// Increases your Starfire damage against targets afflicted with Moonfire or Insect Swarm by 10%.
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			druid := agent.(DruidAgent).GetDruid()

			bonusStarfireDmgT5 := func(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
				if spell.Matches(DruidSpellStarfire) {
					return 1.1
				}

				return 1.0
			}

			t5DotBonusDummyAuras := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
				return target.GetOrRegisterAura(core.Aura{
					ActionID: core.ActionID{SpellID: 37327},
					Label:    "Item - Druid T5 Balance 2P Bonus",
					Duration: core.NeverExpires,
					OnGain: func(aura *core.Aura, sim *core.Simulation) {
						druid.AttackTables[aura.Unit.UnitIndex].DamageDoneByCasterMultiplier = bonusStarfireDmgT5
					},
					OnExpire: func(aura *core.Aura, sim *core.Simulation) {
						druid.AttackTables[aura.Unit.UnitIndex].DamageDoneByCasterMultiplier = nil
					},
				})
			})

			druid.OnSpellRegistered(func(spell *core.Spell) {
				if !spell.Matches(DruidSpellInsectSwarm | DruidSpellMoonfire) {
					return
				}

				for _, target := range druid.Env.Encounter.AllTargetUnits {
					dot := spell.Dot(target)
					if dot == nil {
						return
					}

					dot.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
						if setBonusAura.IsActive() {
							t5DotBonusDummyAuras.Get(aura.Unit).Activate(sim)
						}
					}).ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
						t5DotBonusDummyAuras.Get(aura.Unit).Deactivate(sim)
					})
				}
			})
		},
	},
})

// Feral T6
var ItemSetThunderheartHarness = core.NewItemSet(core.ItemSet{
	ID:   676,
	Name: "Thunderheart Harness",
	Bonuses: map[int32]core.ApplySetBonus{
		// Reduces the energy cost of your Mangle (Cat) by 5 and increases the
		// threat generated by your Mangle (Bear) by 15%.
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				ClassMask: DruidSpellMangleCat,
				Kind:      core.SpellMod_PowerCost_Flat,
				IntValue:  -5,
			})
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				ClassMask:  DruidSpellMangleBear,
				Kind:       core.SpellMod_ThreatMultiplier_Pct,
				FloatValue: 0.15,
			})
		},
		// Increases the damage dealt by your Rip, Swipe, and Ferocious Bite by 15%.
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				ClassMask:  DruidSpellRip | DruidSpellSwipe | DruidSpellFerociousBite,
				Kind:       core.SpellMod_DamageDone_Flat,
				FloatValue: 0.15,
			})
		},
	},
})

// Balance T6
var ItemSetThunderheartRegalia = core.NewItemSet(core.ItemSet{
	ID:   677,
	Name: "Thunderheart Regalia",
	Bonuses: map[int32]core.ApplySetBonus{
		// Increases the duration of your Moonfire ability by 3 sec.
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				ClassMask: DruidSpellMoonfire,
				Kind:      core.SpellMod_DotNumberOfTicks_Flat,
				IntValue:  1,
			})
		},
		// Increases the critical strike chance of your Starfire ability by 5%.
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				ClassMask:  DruidSpellStarfire,
				Kind:       core.SpellMod_BonusCrit_Percent,
				FloatValue: 5,
			})
		},
	},
})
