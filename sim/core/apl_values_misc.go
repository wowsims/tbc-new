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
	alignCdEnd  bool // Tries to align the CD at the end of combat
}

func (rot *APLRotation) newValueMultipleCdUsages(config *proto.APLValueMultipleCdUsages, _ *proto.UUID) APLValue {
	baseSpell := rot.GetAPLSpell(config.BaseSpellId)

	targetSpell := rot.GetAPLSpell(config.TargetSpellId)
	if targetSpell == nil {
		return nil
	}

	if baseSpell == nil || baseSpell.RelatedSelfBuff == nil {
		if baseSpell == nil {
			rot.ValidationMessage(proto.LogLevel_Warning, "%s is not known. Only using offset to delay CD usage.", ProtoToActionID(config.BaseSpellId))
		} else {
			rot.ValidationMessage(proto.LogLevel_Warning, "%s does not have a related self buff. Only using offset to delay CD usage.", ProtoToActionID(config.BaseSpellId))
		}

	}

	if targetSpell.RelatedSelfBuff == nil {
		rot.ValidationMessage(proto.LogLevel_Warning, "%s does not have a related self buff. Will always consider it ready.", ProtoToActionID(config.TargetSpellId))
		return nil
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
		alignCdEnd:  config.AlignCdEnd,
	}
}
func (value *APLValueMultipleCdUsages) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueMultipleCdUsages) GetBool(sim *Simulation) bool {
	totalTime := sim.Duration
	remainingTime := sim.GetRemainingDuration()
	castOffset := value.offset.GetDuration(sim)

	targetSpell := value.targetSpell
	targetSpellCD := targetSpell.CD.Duration
	targetSpellDuration := value.targetSpell.RelatedSelfBuff.Duration
	targetSpellEffectiveTime := targetSpellCD + targetSpellDuration

	maxUses := 1 + int(float64(totalTime)/float64(targetSpellCD))
	canGetFullDurationAtEncounterEnd := remainingTime <= targetSpellDuration
	canGetSecondUptimeWithinRemainingTime := remainingTime >= targetSpellEffectiveTime

	baseSpell := value.baseSpell
	if baseSpell == nil || baseSpell.RelatedSelfBuff == nil {
		if value.alignCdEnd {
			return canGetFullDurationAtEncounterEnd || canGetSecondUptimeWithinRemainingTime
		}
		return true
	}

	baseSpellAura := baseSpell.RelatedSelfBuff
	baseSpellIsReady := baseSpell.CD.IsReady(sim)

	baseSpellIsActiveAndShouldCast := baseSpellAura.IsActive() && sim.CurrentTime >= castOffset
	willLoseUses := baseSpellIsReady && (time.Duration(maxUses-1)*targetSpellCD)+targetSpellDuration+castOffset > totalTime
	baseSpellWasCastAndCanCastAgain := !baseSpellIsReady && canGetSecondUptimeWithinRemainingTime
	baseSpellIsActiveAtEncounterEnd := castOffset+baseSpellAura.Duration >= totalTime
	baseAuraExceedsTargetAura := baseSpellAura.Duration >= remainingTime-targetSpellDuration

	if value.alignCdEnd {
		if maxUses == 1 {
			return canGetFullDurationAtEncounterEnd || (baseSpellIsActiveAtEncounterEnd && baseSpellIsActiveAndShouldCast && !baseAuraExceedsTargetAura) || (!baseSpellIsActiveAtEncounterEnd && baseSpellIsActiveAndShouldCast)
		}
		return canGetFullDurationAtEncounterEnd || baseSpellIsActiveAndShouldCast || willLoseUses || baseSpellWasCastAndCanCastAgain
	}

	return baseSpellIsActiveAndShouldCast || baseSpellWasCastAndCanCastAgain || willLoseUses || !baseSpellIsReady && targetSpell.IsReady(sim)
}

func (value *APLValueMultipleCdUsages) String() string {
	return fmt.Sprintf("Can use CD multiple times(%s)", value.targetSpell.ActionID)
}
