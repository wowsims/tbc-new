package sim

import (
	"github.com/wowsims/tbc/sim/common"
	"github.com/wowsims/tbc/sim/druid/balance"
	"github.com/wowsims/tbc/sim/druid/feralbear"
	"github.com/wowsims/tbc/sim/druid/feralcat"
	restoDruid "github.com/wowsims/tbc/sim/druid/restoration"
	_ "github.com/wowsims/tbc/sim/encounters"
	"github.com/wowsims/tbc/sim/hunter"
	"github.com/wowsims/tbc/sim/mage"
	holyPaladin "github.com/wowsims/tbc/sim/paladin/holy"
	protPaladin "github.com/wowsims/tbc/sim/paladin/protection"
	"github.com/wowsims/tbc/sim/paladin/retribution"
	"github.com/wowsims/tbc/sim/priest/discipline"
	holyPriest "github.com/wowsims/tbc/sim/priest/holy"
	"github.com/wowsims/tbc/sim/priest/shadow"
	"github.com/wowsims/tbc/sim/rogue"
	"github.com/wowsims/tbc/sim/shaman/elemental"
	"github.com/wowsims/tbc/sim/shaman/enhancement"
	restoShaman "github.com/wowsims/tbc/sim/shaman/restoration"
	"github.com/wowsims/tbc/sim/warlock"
	dpswarrior "github.com/wowsims/tbc/sim/warrior/dps"
	protWarrior "github.com/wowsims/tbc/sim/warrior/protection"
)

var registered = false

func RegisterAll() {
	if registered {
		return
	}
	registered = true

	balance.RegisterBalanceDruid()
	feralcat.RegisterFeralCatDruid()
	feralbear.RegisterFeralBearDruid()
	restoDruid.RegisterRestorationDruid()

	hunter.RegisterHunter()

	mage.RegisterMage()

	holyPaladin.RegisterHolyPaladin()
	protPaladin.RegisterProtectionPaladin()
	retribution.RegisterRetributionPaladin()

	discipline.RegisterDisciplinePriest()
	holyPriest.RegisterHolyPriest()
	shadow.RegisterShadowPriest()

	rogue.RegisterRogue()

	elemental.RegisterElementalShaman()
	enhancement.RegisterEnhancementShaman()
	restoShaman.RegisterRestorationShaman()

	warlock.RegisterWarlock()

	dpswarrior.RegisterDpsWarrior()
	protWarrior.RegisterProtectionWarrior()

	common.RegisterAllEffects()
}
