package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

// Incarnate Regalia (Tier 4 - Shadow Priest)
var ItemSetIncarnateRegalia = core.NewItemSet(core.ItemSet{
	ID:   664,
	Name: "Incarnate Regalia",
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Shadowfiend now has 75 more stamina and lasts 3 sec. longer.
			priest := agent.(PriestAgent).GetPriest()

			// Extend shadowfiend duration by 3 seconds
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:      core.SpellMod_BuffDuration_Flat,
				TimeValue: time.Second * 3,
				ClassMask: PriestSpellShadowFiend,
			})

			// Add 75 stamina to shadowfiend
			setBonusAura.
				ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
					if priest.ShadowfiendPet != nil {
						priest.ShadowfiendPet.AddStatDynamic(sim, stats.Stamina, 75.0)
					}
				}).
				ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
					if priest.ShadowfiendPet != nil {
						priest.ShadowfiendPet.AddStatDynamic(sim, stats.Stamina, -75.0)
					}
				})
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your Mind Flay and Smite spells deal 5% more damage.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				Kind:       core.SpellMod_DamageDone_Flat,
				FloatValue: 0.05,
				ClassMask:  PriestSpellMindFlay | PriestSpellSmite,
			})
		},
	},
})
