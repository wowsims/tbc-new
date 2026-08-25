package core

import (
	"fmt"

	"github.com/wowsims/tbc/sim/core/proto"
)

type APLValueSelectedPotion struct {
	DefaultAPLValueImpl
	character *Character
	potionId  int32
}

func (rot *APLRotation) newValueSelectedPotion(config *proto.APLValueSelectedPotion, uuid *proto.UUID) APLValue {
	if config.PotionId == nil || config.PotionId.GetItemId() == 0 {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "No potion selected")
		return nil
	}
	unit := rot.GetSourceUnit(nil).Get()
	character := unit.Env.GetAgentFromUnit(unit).GetCharacter()
	return &APLValueSelectedPotion{
		character: character,
		potionId:  config.PotionId.GetItemId(),
	}
}
func (value *APLValueSelectedPotion) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSelectedPotion) GetBool(sim *Simulation) bool {
	return value.character.Consumables.PotId == value.potionId
}
func (value *APLValueSelectedPotion) String() string {
	return fmt.Sprintf("Selected Potion (%d)", value.potionId)
}

type APLValueSelectedConjured struct {
	DefaultAPLValueImpl
	character  *Character
	conjuredId int32
}

func (rot *APLRotation) newValueSelectedConjured(config *proto.APLValueSelectedConjured, uuid *proto.UUID) APLValue {
	if config.ConjuredId == nil || config.ConjuredId.GetItemId() == 0 {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "No conjured item selected")
		return nil
	}
	unit := rot.GetSourceUnit(nil).Get()
	character := unit.Env.GetAgentFromUnit(unit).GetCharacter()
	return &APLValueSelectedConjured{
		character:  character,
		conjuredId: config.ConjuredId.GetItemId(),
	}
}
func (value *APLValueSelectedConjured) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSelectedConjured) GetBool(sim *Simulation) bool {
	return value.character.Consumables.ConjuredId == value.conjuredId
}
func (value *APLValueSelectedConjured) String() string {
	return fmt.Sprintf("Selected Conjured (%d)", value.conjuredId)
}
