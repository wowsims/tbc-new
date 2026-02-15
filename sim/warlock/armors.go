package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (warlock *Warlock) registerArmors() {

	felArmorBonus := 100.0
	felArmorHealingBonus := 1.2

	demonArmorBonus := 660.0
	demonArmorSRBonus := 18.0

	if warlock.Talents.DemonicAegis > 0 {
		bonusMultipler := 1.0 + (0.1 * float64(warlock.Talents.DemonicAegis))

		felArmorBonus *= bonusMultipler
		felArmorHealingBonus *= bonusMultipler

		demonArmorBonus *= bonusMultipler
		demonArmorSRBonus *= bonusMultipler
	}

	warlock.FelArmor = warlock.RegisterAura(core.Aura{
		Label:    "Fel Armor",
		ActionID: core.ActionID{SpellID: 28176},
		Duration: time.Minute * 30,
	}).AttachMultiplicativePseudoStatBuff(&warlock.PseudoStats.SelfHealingMultiplier, felArmorHealingBonus).AttachStatBuff(stats.SpellDamage, felArmorBonus)

	warlock.DemonArmor = warlock.RegisterAura(core.Aura{
		Label:    "Demon Armor",
		ActionID: core.ActionID{SpellID: 27260},
		Duration: time.Minute * 30,
	}).AttachStatBuff(stats.Armor, demonArmorBonus).AttachStatBuff(stats.ShadowResistance, demonArmorSRBonus)

	// Armor selection
	switch warlock.Options.Armor {

	case proto.WarlockOptions_FelArmor:
		core.MakePermanent(warlock.FelArmor)

	case proto.WarlockOptions_DemonArmor:
		core.MakePermanent(warlock.DemonArmor)
	}

}
