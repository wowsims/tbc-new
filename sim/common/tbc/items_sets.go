package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

// Dungeon Set 3 - DPS - Plate
var ItemSetDoomplateBattlegear = core.NewItemSet(core.ItemSet{
	Name: "Doomplate Battlegear",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachStatBuff(stats.MeleeHitRating, 35).
				ExposeToAPL(37610)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your attacks have a chance to grant you 160 attack power for 15 sec.
			aura := agent.GetCharacter().NewTemporaryStatsAura(
				"Heroic Resolution",
				core.ActionID{SpellID: 37612},
				stats.Stats{stats.AttackPower: 160, stats.RangedAttackPower: 160},
				time.Second*15,
			)

			setBonusAura.
				AttachProcTrigger(core.ProcTrigger{
					Name:       "Doomplate Battlegear - 4PC",
					ProcChance: 0.02,
					Callback:   core.CallbackOnSpellHitDealt,
					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						aura.Activate(sim)
					},
				}).
				ExposeToAPL(37611)
		},
	},
})

func init() {}
