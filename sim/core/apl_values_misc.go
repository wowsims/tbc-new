package core

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
)

type APLValueMultipleCdUsages struct {
	DefaultAPLValueImpl
	unit *Unit

	baseSpell   *Spell
	targetSpell *Spell
	offset      APLValue
}

func (rot *APLRotation) newValueMultipleCdUsages(config *proto.APLValueMultipleCdUsages, _ *proto.UUID) APLValue {
	baseSpell := rot.GetAPLSpell(config.BaseSpellId)
	if baseSpell == nil {
		return nil
	}

	targetSpell := rot.GetAPLSpell(config.TargetSpellId)
	if targetSpell == nil {
		return nil
	}

	if baseSpell.RelatedSelfBuff == nil {
		rot.ValidationMessage(proto.LogLevel_Warning, "%s does not have a related self buff", ProtoToActionID(config.BaseSpellId))
	}

	if targetSpell.RelatedSelfBuff == nil {
		rot.ValidationMessage(proto.LogLevel_Warning, "%s does not have a related self buff", ProtoToActionID(config.TargetSpellId))
	}

	offset := rot.coerceTo(rot.newAPLValue(config.Offset), proto.APLValueType_ValueTypeDuration)
	if offset == nil {
		offset = rot.newValueConst(&proto.APLValueConst{Val: "0ms"}, &proto.UUID{Value: ""})
	}

	return &APLValueMultipleCdUsages{
		unit:        rot.unit,
		baseSpell:   baseSpell,
		targetSpell: targetSpell,
		offset:      offset,
	}
}
func (value *APLValueMultipleCdUsages) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueMultipleCdUsages) GetBool(sim *Simulation) bool {
	totalTime := sim.Duration
	remainingTime := sim.GetRemainingDuration()
	castOffset := value.offset.GetDuration(sim)
	baseSpellIsReady := value.baseSpell.CD.IsReady(sim)

	baseSpellAura := value.baseSpell.RelatedSelfBuff

	targetSpell := value.targetSpell
	targetSpellCD := targetSpell.CD.Duration
	targetSpellDuration := value.targetSpell.RelatedSelfBuff.Duration
	targetSpellEffectiveTime := targetSpellCD + targetSpellDuration

	maxUses := 1 + int(float64(totalTime)/float64(targetSpellCD))

	remainingTimeIsSmallerThanBuff := remainingTime <= targetSpellDuration
	baseSpellIsAciveAndShouldCast := baseSpellAura.IsActive() && sim.CurrentTime >= castOffset
	willLoseUses := baseSpellIsReady && (time.Duration(maxUses-1)*targetSpellCD)+targetSpellDuration+castOffset >= totalTime
	baseSpellWasCastAndCanCastAgain := !baseSpellIsReady && remainingTime >= targetSpellEffectiveTime

	return remainingTimeIsSmallerThanBuff || baseSpellIsAciveAndShouldCast || willLoseUses || baseSpellWasCastAndCanCastAgain
}

func (value *APLValueMultipleCdUsages) String() string {
	return fmt.Sprintf("Can use CD multiple times(%s)", value.targetSpell.ActionID)
}
