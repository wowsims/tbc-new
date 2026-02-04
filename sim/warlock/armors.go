package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (warlock *Warlock) registerArmors() {
	warlock.FelArmor = warlock.RegisterAura(core.Aura{
		Label:    "Fel Armor",
		ActionID: core.ActionID{SpellID: 28176},
		Duration: time.Minute * 30,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SelfHealingMultiplier *= 1.20
			aura.Unit.AddStatDynamic(sim, stats.SpellDamage, 100)
		},

		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SelfHealingMultiplier /= 1.20
			aura.Unit.AddStatDynamic(sim, stats.SpellDamage, -100)
		},
	})

	warlock.DemonArmor = warlock.RegisterAura(core.Aura{
		Label:    "Demon Armor",
		ActionID: core.ActionID{SpellID: 27260},
		Duration: time.Minute * 30,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.Armor, 660)
			aura.Unit.AddStatDynamic(sim, stats.ShadowResistance, 18)
			//18 hp5
		},

		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.AddStatDynamic(sim, stats.Armor, -660)
			aura.Unit.AddStatDynamic(sim, stats.ShadowResistance, -18)
			//-18 hp5
		},
	})

}
