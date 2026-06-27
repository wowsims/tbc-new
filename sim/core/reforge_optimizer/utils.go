package reforgeoptimizer

import (
	"github.com/wowsims/tbc/sim/core/proto"
)

func playerIsTankSpec(player *proto.Player) bool {
	switch player.GetSpec().(type) {
	case *proto.Player_FeralBearDruid,
		*proto.Player_ProtectionPaladin,
		*proto.Player_ProtectionWarrior:
		return true
	default:
		return false
	}
}

func playerHasProfession(player *proto.Player, profession proto.Profession) bool {
	return player.GetProfession1() == profession || player.GetProfession2() == profession
}
