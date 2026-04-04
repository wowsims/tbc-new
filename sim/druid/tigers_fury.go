package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (druid *Druid) registerTigersFurySpell() {
	const weaponDamageBonus = 40.0

	druid.TigersFuryAura = druid.RegisterAura(core.Aura{
		Label:    "Tiger's Fury",
		ActionID: core.ActionID{SpellID: 9846},
		Duration: time.Second * 6,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			druid.AutoAttacks.MH().BaseDamageMin += weaponDamageBonus
			druid.AutoAttacks.MH().BaseDamageMax += weaponDamageBonus
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.AutoAttacks.MH().BaseDamageMin -= weaponDamageBonus
			druid.AutoAttacks.MH().BaseDamageMax -= weaponDamageBonus
		},
	})

	druid.TigersFury = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 9846},
		ClassSpellMask: DruidSpellTigersFury,
		Flags:          core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost: 30,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: 0,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Second,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.TigersFuryAura.Activate(sim)
		},
	})
}
