package warlock

import (
	"github.com/wowsims/tbc/sim/core"
)

// Dungeon Set 3
var ItemSetOblivionRaiment = core.NewItemSet(core.ItemSet{
	ID:   644,
	Name: "Oblivion Raiment",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Grants your pet 45 mana per 5 sec.
			// Pet Mana Regen - 37375
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Seed of Corruption deals 180 additional damage when it detonates.
			// Improved Seed of Corruption - 37376
		},
	},
})

// T4
var ItemSetVoidheartRaiment = core.NewItemSet(core.ItemSet{
	ID:   645,
	Name: "Voidheart Raiment",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your shadow damage spells have a chance to grant you 135 bonus shadow damage for 15 sec.
			// Shadowflame - 37377
			// Your fire damage spells have a chance to grant you 135 bonus fire damage for 15 sec.
			// Hellfire - 39437
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Increases the duration of your Corruption and Immolate abilities by 3 sec.
			// Improved Corruption and Immolate - 37380
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
		},
	},
})

func init() {
	core.NewItemEffect(19337, func(agent core.Agent) {
		// The Black Book
	})

	core.NewItemEffect(30449, func(agent core.Agent) {
		// Void Star Talisman
	})

	core.NewItemEffect(32493, func(agent core.Agent) {
		// Ashtongue Talisman of Shadows
	})
}
