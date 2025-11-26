package sim

import (
	"github.com/wowsims/tbc/sim/common"
	"github.com/wowsims/tbc/sim/druid/balance"
	"github.com/wowsims/tbc/sim/druid/feral"
	"github.com/wowsims/tbc/sim/druid/guardian"
	restoDruid "github.com/wowsims/tbc/sim/druid/restoration"
	_ "github.com/wowsims/tbc/sim/encounters"
	"github.com/wowsims/tbc/sim/hunter/beast_mastery"
	"github.com/wowsims/tbc/sim/hunter/marksmanship"
	"github.com/wowsims/tbc/sim/hunter/survival"
	"github.com/wowsims/tbc/sim/mage/arcane"
	"github.com/wowsims/tbc/sim/mage/fire"
	frostMage "github.com/wowsims/tbc/sim/mage/frost"
	holyPaladin "github.com/wowsims/tbc/sim/paladin/holy"
	protPaladin "github.com/wowsims/tbc/sim/paladin/protection"
	"github.com/wowsims/tbc/sim/paladin/retribution"
	"github.com/wowsims/tbc/sim/priest/discipline"
	holyPriest "github.com/wowsims/tbc/sim/priest/holy"
	"github.com/wowsims/tbc/sim/priest/shadow"
	"github.com/wowsims/tbc/sim/rogue/assassination"
	"github.com/wowsims/tbc/sim/rogue/combat"
	"github.com/wowsims/tbc/sim/rogue/subtlety"
	"github.com/wowsims/tbc/sim/shaman/elemental"
	"github.com/wowsims/tbc/sim/shaman/enhancement"
	restoShaman "github.com/wowsims/tbc/sim/shaman/restoration"
	"github.com/wowsims/tbc/sim/warlock/affliction"
	"github.com/wowsims/tbc/sim/warlock/demonology"
	"github.com/wowsims/tbc/sim/warlock/destruction"
	"github.com/wowsims/tbc/sim/warrior/arms"
	"github.com/wowsims/tbc/sim/warrior/fury"
	protWarrior "github.com/wowsims/tbc/sim/warrior/protection"
)

var registered = false

func RegisterAll() {
	if registered {
		return
	}
	registered = true

	balance.RegisterBalanceDruid()
	feral.RegisterFeralDruid()
	guardian.RegisterGuardianDruid()
	restoDruid.RegisterRestorationDruid()

	beast_mastery.RegisterBeastMasteryHunter()
	marksmanship.RegisterMarksmanshipHunter()
	survival.RegisterSurvivalHunter()

	arcane.RegisterArcaneMage()
	fire.RegisterFireMage()
	frostMage.RegisterFrostMage()

	holyPaladin.RegisterHolyPaladin()
	protPaladin.RegisterProtectionPaladin()
	retribution.RegisterRetributionPaladin()

	discipline.RegisterDisciplinePriest()
	holyPriest.RegisterHolyPriest()
	shadow.RegisterShadowPriest()

	assassination.RegisterAssassinationRogue()
	combat.RegisterCombatRogue()
	subtlety.RegisterSubtletyRogue()

	elemental.RegisterElementalShaman()
	enhancement.RegisterEnhancementShaman()
	restoShaman.RegisterRestorationShaman()

	affliction.RegisterAfflictionWarlock()
	demonology.RegisterDemonologyWarlock()
	destruction.RegisterDestructionWarlock()

	arms.RegisterArmsWarrior()
	fury.RegisterFuryWarrior()
	protWarrior.RegisterProtectionWarrior()

	common.RegisterAllEffects()
}
