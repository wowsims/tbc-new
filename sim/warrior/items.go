package warrior

import (
	"github.com/wowsims/tbc/sim/core"
)

// Dungeon Set 3 - Tank
var ItemSetBoldArmor = core.NewItemSet(core.ItemSet{
	Name: "Bold Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask: SpellMaskShouts,
					Kind:      core.SpellMod_PowerCost_Flat,
					IntValue:  -2,
				}).
				ExposeToAPL(37512)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.ExposeToAPL(37513)
		},
	},
})

// T6 - DPS
var ItemSetOnslaughtBattlegear = core.NewItemSet(core.ItemSet{
	Name: "Onslaught Battlegear",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask: SpellMaskExecute,
					Kind:      core.SpellMod_PowerCost_Flat,
					IntValue:  -3,
				}).
				ExposeToAPL(38398)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask:  SpellMaskMortalStrike | SpellMaskBloodthirst,
					Kind:       core.SpellMod_DamageDone_Flat,
					FloatValue: 0.05,
				}).
				ExposeToAPL(38399)
		},
	},
})

// T6 - Tank
var ItemSetOnslaughtArmor = core.NewItemSet(core.ItemSet{
	Name: "Onslaught Armor",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.ExposeToAPL(38408)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				ClassMask:  SpellMaskShieldSlam,
				Kind:       core.SpellMod_DamageDone_Flat,
				FloatValue: 0.1,
			}).
				ExposeToAPL(38407)
		},
	},
})

func init() {}
