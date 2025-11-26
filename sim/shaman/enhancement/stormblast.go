package enhancement

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/shaman"
)

func (enh *EnhancementShaman) newStormblastHitSpell(isMh bool) *core.Spell {
	config := enh.newStormstrikeHitSpellConfig(115356, isMh)
	config.SpellSchool = core.SpellSchoolNature
	config.ClassSpellMask = shaman.SpellMaskStormblastDamage
	return enh.RegisterSpell(config)
}

func (enh *EnhancementShaman) registerStormblastSpell() {
	mhHit := enh.newStormblastHitSpell(true)
	ohHit := enh.newStormblastHitSpell(false)

	config := enh.newStormstrikeSpellConfig(115356, &enh.StormStrikeDebuffAuras, mhHit, ohHit)
	config.SpellSchool = core.SpellSchoolNature
	config.ManaCost.BaseCostPercent = 9.372
	config.ClassSpellMask = shaman.SpellMaskStormblastCast
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return enh.AscendanceAura.IsActive()
	}

	enh.Stormblast = enh.RegisterSpell(config)
}
