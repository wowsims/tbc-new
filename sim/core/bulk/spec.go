package bulk

import (
	"fmt"

	"github.com/wowsims/tbc/sim/core/proto"
)

func getPlayerSpec(player *proto.Player) (proto.Spec, error) {
	if player == nil {
		return proto.Spec_SpecUnknown, fmt.Errorf("unsupported player spec for backend bulk candidate generation")
	}

	switch {
	case player.GetBalanceDruid() != nil:
		return proto.Spec_SpecBalanceDruid, nil
	case player.GetFeralCatDruid() != nil:
		return proto.Spec_SpecFeralCatDruid, nil
	case player.GetFeralBearDruid() != nil:
		return proto.Spec_SpecFeralBearDruid, nil
	case player.GetRestorationDruid() != nil:
		return proto.Spec_SpecRestorationDruid, nil
	case player.GetHunter() != nil:
		return proto.Spec_SpecHunter, nil
	case player.GetMage() != nil:
		return proto.Spec_SpecMage, nil
	case player.GetHolyPaladin() != nil:
		return proto.Spec_SpecHolyPaladin, nil
	case player.GetProtectionPaladin() != nil:
		return proto.Spec_SpecProtectionPaladin, nil
	case player.GetRetributionPaladin() != nil:
		return proto.Spec_SpecRetributionPaladin, nil
	case player.GetPriest() != nil:
		return proto.Spec_SpecPriest, nil
	case player.GetRogue() != nil:
		return proto.Spec_SpecRogue, nil
	case player.GetElementalShaman() != nil:
		return proto.Spec_SpecElementalShaman, nil
	case player.GetEnhancementShaman() != nil:
		return proto.Spec_SpecEnhancementShaman, nil
	case player.GetRestorationShaman() != nil:
		return proto.Spec_SpecRestorationShaman, nil
	case player.GetWarlock() != nil:
		return proto.Spec_SpecWarlock, nil
	case player.GetDpsWarrior() != nil:
		return proto.Spec_SpecDpsWarrior, nil
	case player.GetProtectionWarrior() != nil:
		return proto.Spec_SpecProtectionWarrior, nil
	default:
		return proto.Spec_SpecUnknown, fmt.Errorf("unsupported player spec for backend bulk candidate generation")
	}
}
