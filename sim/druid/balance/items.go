package balance

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
	"github.com/wowsims/tbc/sim/druid"
)

// T14 Balance
var ItemSetRegaliaOfTheEternalBloosom = core.NewItemSet(core.ItemSet{
	Name:                    "Regalia of the Eternal Blossom",
	DisabledInChallengeMode: true,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(_ core.Agent, setBonusAura *core.Aura) {
			// Your Starfall deals 20% additional damage.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Pct,
				ClassMask:  druid.DruidSpellStarfall,
				FloatValue: 0.2,
			})
		},
		4: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increases the duration of your Moonfire and Sunfire spells by 2 sec.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_DotNumberOfTicks_Flat,
				ClassMask: druid.DruidSpellMoonfireDoT | druid.DruidSpellSunfireDoT,
				IntValue:  1,
			})
		},
	},
})

// T15 Balance
var ItemSetRegaliaOfTheHauntedForest = core.NewItemSet(core.ItemSet{
	Name:                    "Regalia of the Haunted Forest",
	DisabledInChallengeMode: true,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increases the critical strike chance of Starsurge by 10%.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_BonusCrit_Percent,
				ClassMask:  druid.DruidSpellStarsurge,
				FloatValue: 10,
			})
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Nature's Grace now also grants 1000 critical strike and 1000 mastery for its duration.
			druid := agent.(*BalanceDruid)
			aura := druid.NewTemporaryStatsAura(
				"Druid T15 Balance 4P Bonus",
				core.ActionID{SpellID: 138350},
				stats.Stats{stats.CritRating: 1000, stats.MasteryRating: 1000},
				time.Second*15,
			)

			druid.Env.RegisterPreFinalizeEffect(func() {
				druid.NaturesGrace.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
					if setBonusAura.IsActive() {
						aura.Activate(sim)
					}
				})
			})

			setBonusAura.ExposeToAPL(138350)
		},
	},
})
