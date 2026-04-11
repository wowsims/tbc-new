package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// https://www.wowhead.com/tbc/item=27484/libram-of-avengement
	// Your Judgement spells grant +53 melee and spell crit rating for 5s.
	core.NewItemEffect(27484, func(agent core.Agent) {
		paladin := agent.(PaladinAgent).GetPaladin()

		buffAura := paladin.NewTemporaryStatsAura(
			"Justice",
			core.ActionID{SpellID: 34260},
			stats.Stats{stats.MeleeCritRating: 53, stats.SpellCritRating: 53},
			time.Second*5,
		)

		aura := core.MakePermanent(paladin.RegisterAura(core.Aura{
			Label:    "Libram of Avengement",
			ActionID: core.ActionID{SpellID: 34258},
		}).AttachProcTrigger(core.ProcTrigger{
			Callback:       core.CallbackOnSpellHitDealt,
			ClassSpellMask: SpellMaskJudgementOfCommand | SpellMaskJudgementOfRighteousness | SpellMaskJudgementOfBlood | SpellMaskJudgementOfVengeance,

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				buffAura.Activate(sim)
			},
		}))

		paladin.ItemSwap.RegisterProc(27484, aura)
	})

	// https://www.wowhead.com/tbc/item=22401/libram-of-hope
	// Reduces the base mana cost of your Seal spells by 20.
	core.NewItemEffect(22401, func(agent core.Agent) {
		paladin := agent.(PaladinAgent).GetPaladin()

		aura := core.MakePermanent(paladin.RegisterAura(core.Aura{
			Label:    "Libram of Hope",
			ActionID: core.ActionID{SpellID: 27848},
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask: SpellMaskAllSeals,
			Kind:      core.SpellMod_PowerCost_Flat,
			IntValue:  -20,
		}))

		paladin.ItemSwap.RegisterProc(22401, aura)
	})

	LibramMap{
		// https://www.wowhead.com/tbc/item=23203/libram-of-fervor
		// Increases the melee attack power bonus of your Seal of the Crusader by 48 and the Holy damage increase of your Judgement of the Crusader by 33.
		{ItemID: 23203, AuraID: 28852, StatValue: 48, Label: "Libram of Fervor"},

		// Increases the melee attack power bonus of your Seal of the Crusader by 68 and the Holy damage increase of your Judgement of the Crusader by 47.
		// https://www.wowhead.com/tbc/item=27949/libram-of-zeal
		{ItemID: 27949, AuraID: 33557, StatValue: 68, Label: "Libram of Zeal"},
		// https://www.wowhead.com/tbc/item=27983/libram-of-zeal
		{ItemID: 27983, AuraID: 33557, StatValue: 68, Label: "Libram of Zeal"},
	}.RegisterAll(func(config LibramConfig) {
		core.NewItemEffect(config.ItemID, func(agent core.Agent) {
			paladin := agent.(PaladinAgent).GetPaladin()

			aura := paladin.GetAura(config.Label)
			if aura == nil {
				aura = paladin.NewTemporaryStatsAura(
					config.Label,
					core.ActionID{SpellID: config.AuraID},
					stats.Stats{stats.AttackPower: config.StatValue},
					core.NeverExpires,
				).Aura
				// Bonus judge damage implemented in sim/paladin/seals.go
			}

			paladin.OnSpellRegistered(func(spell *core.Spell) {
				if !spell.Matches(SpellMaskSealOfTheCrusader) {
					return
				}

				spell.RelatedSelfBuff.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
					if paladin.Ranged().ID == config.ItemID {
						aura.Activate(sim)
					}
				}).ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
					aura.Deactivate(sim)
				})
			})

			paladin.ItemSwap.RegisterProc(config.ItemID, aura)
		})
	})

	// https://www.wowhead.com/tbc/item=27917/libram-of-the-eternal-rest
	// Increases the damage of your Consecration spell by up to 47.
	core.NewItemEffect(27917, func(agent core.Agent) {
		paladin := agent.(PaladinAgent).GetPaladin()

		aura := core.MakePermanent(paladin.RegisterAura(core.Aura{
			Label:    "Libram of the Eternal Rest",
			ActionID: core.ActionID{SpellID: 34252},
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask:  SpellMaskConsecration,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 47.0,
		}))

		paladin.ItemSwap.RegisterProc(27917, aura)
	})

	// https://www.wowhead.com/tbc/item=28065/libram-of-wracking
	// Increases the damage done by your Exorcism and Holy Wrath spells by up to 120.
	core.NewItemEffect(28065, func(agent core.Agent) {
		paladin := agent.(PaladinAgent).GetPaladin()

		aura := core.MakePermanent(paladin.RegisterAura(core.Aura{
			Label:    "Libram of Wracking",
			ActionID: core.ActionID{SpellID: 33695},
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask:  SpellMaskExorcism | SpellMaskHolyWrath,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 120.0,
		}))

		paladin.ItemSwap.RegisterProc(28065, aura)
	})

	// https://www.wowhead.com/tbc/item=29388/libram-of-repentance
	// Increases your block rating by 42 while Holy Shield is active.
	core.NewItemEffect(29388, func(agent core.Agent) {
		paladin := agent.(PaladinAgent).GetPaladin()

		if !paladin.Talents.HolyShield {
			return
		}

		aura := paladin.NewTemporaryStatsAura(
			"Libram of Repentance",
			core.ActionID{SpellID: 37742},
			stats.Stats{stats.BlockRating: 42},
			core.NeverExpires,
		)

		paladin.OnSpellRegistered(func(spell *core.Spell) {
			if !spell.Matches(SpellMaskHolyShield) {
				return
			}

			spell.RelatedSelfBuff.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
				if paladin.Ranged().ID == 29388 {
					aura.Activate(sim)
				}
			}).ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
				aura.Deactivate(sim)
			})
		})
	})

	// https://www.wowhead.com/tbc/item=31033/libram-of-righteous-power
	// Increases the damage dealt by Crusader Strike by 36.
	core.NewItemEffect(31033, func(agent core.Agent) {
		paladin := agent.(PaladinAgent).GetPaladin()

		aura := core.MakePermanent(paladin.RegisterAura(core.Aura{
			Label:    "Libram of Righteous Power",
			ActionID: core.ActionID{SpellID: 37763},
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask:  SpellMaskCrusaderStrike,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 36.0,
		}))

		paladin.ItemSwap.RegisterProc(31033, aura)
	})

	// https://www.wowhead.com/tbc/item=33503/libram-of-divine-judgement
	// Your Judgement of Command ability has a chance to grant 200 attack power for 10s.
	core.NewItemEffect(33503, func(agent core.Agent) {
		paladin := agent.(PaladinAgent).GetPaladin()

		buffAura := paladin.NewTemporaryStatsAura(
			"Crusader's Command",
			core.ActionID{SpellID: 43747},
			stats.Stats{stats.AttackPower: 200},
			time.Second*10,
		)

		aura := core.MakePermanent(paladin.RegisterAura(core.Aura{
			Label:    "Libram of Divine Judgement",
			ActionID: core.ActionID{SpellID: 43745},
		}).AttachProcTrigger(core.ProcTrigger{
			Callback:       core.CallbackOnSpellHitDealt,
			ClassSpellMask: SpellMaskJudgementOfCommand,
			ProcChance:     0.4,

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				buffAura.Activate(sim)
			},
		}))

		paladin.ItemSwap.RegisterProc(33503, aura)
	})

	// https://www.wowhead.com/tbc/item=33504/libram-of-divine-purpose
	// Increases the damage done by your Seal of Righteousness and Judgement of Righteousness abilities by up to 94.
	core.NewItemEffect(33504, func(agent core.Agent) {
		paladin := agent.(PaladinAgent).GetPaladin()

		aura := core.MakePermanent(paladin.RegisterAura(core.Aura{
			Label:    "Libram of Divine Purpose",
			ActionID: core.ActionID{SpellID: 43743},
		}).AttachSpellMod(core.SpellModConfig{
			ClassMask:  SpellMaskSealOfRighteousness | SpellMaskJudgementOfRighteousness,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 94.0,
		}))

		paladin.ItemSwap.RegisterProc(33504, aura)
	})

	LibramMap{
		// Judgement grants resilience for 6s.
		// https://www.wowhead.com/tbc/item=33936/gladiators-libram-of-fortitude
		{ItemID: 33936, AuraID: 43839, TriggerID: 43850, StatValue: 26, Label: "Gladiator's Libram of Fortitude", SpellMask: SpellMaskJudgement},
		// https://www.wowhead.com/tbc/item=33937/merciless-gladiators-libram-of-fortitude
		{ItemID: 33937, AuraID: 43848, TriggerID: 43851, StatValue: 31, Label: "Merciless Gladiator's Libram of Fortitude", SpellMask: SpellMaskJudgement},
		// https://www.wowhead.com/tbc/item=33938/vengeful-gladiators-libram-of-fortitude
		{ItemID: 33938, AuraID: 43849, TriggerID: 43852, StatValue: 34, Label: "Vengeful Gladiator's Libram of Fortitude", SpellMask: SpellMaskJudgement},
		// https://www.wowhead.com/tbc/item=35039/brutal-gladiators-libram-of-fortitude
		{ItemID: 35039, AuraID: 46089, TriggerID: 46091, StatValue: 39, Label: "Brutal Gladiator's Libram of Fortitude", SpellMask: SpellMaskJudgement},

		// Holy Shield grants resilience for 6s.
		// https://www.wowhead.com/tbc/item=33948/gladiators-libram-of-vengeance
		{ItemID: 33948, AuraID: 43839, TriggerID: 43854, StatValue: 26, Label: "Gladiator's Libram of Vengeance", SpellMask: SpellMaskHolyShield},
		// https://www.wowhead.com/tbc/item=33949/merciless-gladiators-libram-of-vengeance
		{ItemID: 33949, AuraID: 43848, TriggerID: 43855, StatValue: 31, Label: "Merciless Gladiator's Libram of Vengeance", SpellMask: SpellMaskHolyShield},
		// https://www.wowhead.com/tbc/item=33950/vengeful-gladiators-libram-of-vengeance
		{ItemID: 33950, AuraID: 43849, TriggerID: 43856, StatValue: 34, Label: "Vengeful Gladiator's Libram of Vengeance", SpellMask: SpellMaskHolyShield},
		// https://www.wowhead.com/tbc/item=35041/brutal-gladiators-libram-of-vengeance
		{ItemID: 35041, AuraID: 46089, TriggerID: 46095, StatValue: 39, Label: "Brutal Gladiator's Libram of Vengeance", SpellMask: SpellMaskHolyShield},
	}.RegisterAll(func(config LibramConfig) {
		core.NewItemEffect(config.ItemID, func(agent core.Agent) {
			paladin := agent.(PaladinAgent).GetPaladin()

			buffAura := paladin.NewTemporaryStatsAura(
				config.Label,
				core.ActionID{SpellID: config.AuraID},
				stats.Stats{stats.ResilienceRating: config.StatValue},
				time.Second*6,
			)

			aura := core.MakePermanent(paladin.RegisterAura(core.Aura{
				Label:    config.Label + " Trigger",
				ActionID: core.ActionID{SpellID: config.TriggerID},
			}).AttachProcTrigger(core.ProcTrigger{
				Callback:       core.CallbackOnCastComplete,
				ClassSpellMask: config.SpellMask,

				Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					buffAura.Activate(sim)
				},
			}))

			paladin.ItemSwap.RegisterProc(config.ItemID, aura)
		})
	})
}

type LibramConfig struct {
	ItemID    int32
	AuraID    int32
	TriggerID int32
	StatValue float64
	Label     string
	SpellMask int64
}

type LibramMap []LibramConfig
type LibramFactory func(config LibramConfig)

func (librams LibramMap) RegisterAll(factory LibramFactory) {
	for _, libramConfig := range librams {
		factory(libramConfig)
	}
}
