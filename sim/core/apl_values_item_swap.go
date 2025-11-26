package core

import (
	"fmt"

	"github.com/wowsims/tbc/sim/core/proto"
)

type APLValueActiveItemSwapSet struct {
	DefaultAPLValueImpl
	character *Character
	swapSet   proto.APLActionItemSwap_SwapSet
}

func (rot *APLRotation) newValueActiveItemSwapSet(config *proto.APLValueActiveItemSwapSet, _ *proto.UUID) APLValue {
	if config.SwapSet == proto.APLActionItemSwap_Unknown {
		rot.ValidationMessage(proto.LogLevel_Warning, "Unknown item swap set")
		return nil
	}

	character := rot.unit.Env.Raid.GetPlayerFromUnit(rot.unit).GetCharacter()
	if !character.ItemSwap.IsEnabled() {
		if config.SwapSet != proto.APLActionItemSwap_Main {
			rot.ValidationMessage(proto.LogLevel_Warning, "No swap set configured in Settings.")
		}
		return nil
	}

	return &APLValueActiveItemSwapSet{
		character: character,
		swapSet:   config.SwapSet,
	}
}
func (value *APLValueActiveItemSwapSet) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueActiveItemSwapSet) GetBool(sim *Simulation) bool {
	return value.character.ItemSwap.swapSet == value.swapSet
}
func (value *APLValueActiveItemSwapSet) String() string {
	return fmt.Sprintf("Active Item Swap Set(%s)", value.swapSet)
}
