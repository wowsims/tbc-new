package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

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
					spell.Dot(target).ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
						if setBonusAura.IsActive() {
							t5DotBonusDummyAuras.Get(aura.Unit).Activate(sim)
						}
					})

					spell.Dot(target).ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
						t5DotBonusDummyAuras.Get(aura.Unit).Deactivate(sim)
					})
				}
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
				FloatValue: 0.05,
			})
		},
	},
})

func init() {
	// Idol of the Avenger
	core.NewItemEffect(31025, func(agent core.Agent) {
		// Increases the damage dealt by Wrath by 25.
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Wrath Damage",
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask:  DruidSpellWrath,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 25.0,
		}))

		character.ItemSwap.RegisterProc(32330, aura)
	})

	// Ivory Idol of the Moongoddess
	core.NewItemEffect(27518, func(agent core.Agent) {
		// Increases the damage of your Starfire spell by up to 55.
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Starfire Damage",
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask:  DruidSpellStarfire,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 55.0,
		}))

		character.ItemSwap.RegisterProc(32330, aura)
	})

	core.NewItemEffect(32486, func(agent core.Agent) {
		// Ashtongue Talisman of Equilibrium
		// Mangle has a 40% chance to grant 140 Strength for 8 sec,
		// Starfire has a 25% chance to grant up to 150 spell damage for 8 sec, and
		// Rejuvenation has a 25% chance to grant up to 210 healing for 8 sec.
		druid := agent.(DruidAgent).GetDruid()
		ashtongueAuraMangle := druid.NewTemporaryStatsAura("Ashtongue Talisman of Equilibrium (Mangle)", core.ActionID{SpellID: 32486}, stats.Stats{stats.Strength: 140}, time.Second*8)
		ashtongueAuraStarfire := druid.NewTemporaryStatsAura("Ashtongue Talisman of Equilibrium (Starfire)", core.ActionID{SpellID: 32486}, stats.Stats{stats.SpellDamage: 150}, time.Second*8)

		aura := core.MakePermanent(druid.RegisterAura(core.Aura{
			Label: "Ashtongue Talisman of Equilibrium",
		}).AttachProcTrigger(core.ProcTrigger{
			ClassSpellMask: DruidSpellMangle,
			Callback:       core.CallbackOnSpellHitDealt,
			ProcChance:     0.4,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				ashtongueAuraMangle.Activate(sim)
			},
		}).AttachProcTrigger(core.ProcTrigger{
			ClassSpellMask: DruidSpellStarfire,
			Callback:       core.CallbackOnSpellHitDealt,
			ProcChance:     0.25,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				ashtongueAuraStarfire.Activate(sim)
			},
		}))

		druid.ItemSwap.RegisterProc(32486, aura)
		druid.ItemSwap.RegisterActive(32486)
	})
}
