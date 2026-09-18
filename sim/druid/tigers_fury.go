package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var tigersFuryRank = genRanks.TigersFury.BySpellID(9846)

func (druid *Druid) registerTigersFurySpell() {
	weaponDamageBonus := shared.SpellRankMin(tigersFuryRank.Direct)

	druid.TigersFuryAura = druid.RegisterAura(core.Aura{
		Label:    "Tiger's Fury",
		ActionID: core.ActionID{SpellID: tigersFuryRank.SpellID},
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
		ActionID:       core.ActionID{SpellID: tigersFuryRank.SpellID},
		ClassSpellMask: DruidSpellTigersFury,
		Flags:          core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost: tigersFuryRank.Cost,
		},
		Cast: core.CastConfig{
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: tigersFuryRank.Cooldown,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.TigersFuryAura.Activate(sim)
		},
	})
}
