package core

import (
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
)

type APLValueAutoSwingTime struct {
	DefaultAPLValueImpl
	unit     *Unit
	autoType proto.APLValueAutoSwingTime_SwingType
}

func (rot *APLRotation) newValueAutoSwingTime(config *proto.APLValueAutoSwingTime, _ *proto.UUID) APLValue {
	return &APLValueAutoSwingTime{
		unit:     rot.unit,
		autoType: config.AutoType,
	}
}
func (value *APLValueAutoSwingTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueAutoSwingTime) GetDuration(sim *Simulation) time.Duration {
	switch value.autoType {
	case proto.APLValueAutoSwingTime_MainHand:
		return max(0, value.unit.AutoAttacks.MainhandSwingSpeed())
	case proto.APLValueAutoSwingTime_OffHand:
		return max(0, value.unit.AutoAttacks.OffhandSwingSpeed())
	case proto.APLValueAutoSwingTime_Ranged:
		return max(0, value.unit.AutoAttacks.RangedSwingSpeed())
	}
	// defaults to 0
	return 0
}
func (value *APLValueAutoSwingTime) String() string {
	return "Auto Swing Time"
}

type APLValueAutoTimeToNext struct {
	DefaultAPLValueImpl
	unit     *Unit
	autoType proto.APLValueAutoAttackType
}

func (rot *APLRotation) newValueAutoTimeToNext(config *proto.APLValueAutoTimeToNext, _ *proto.UUID) APLValue {
	return &APLValueAutoTimeToNext{
		unit:     rot.unit,
		autoType: config.AutoType,
	}
}
func (value *APLValueAutoTimeToNext) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueAutoTimeToNext) GetDuration(sim *Simulation) time.Duration {
	switch value.autoType {
	case proto.APLValueAutoAttackType_MeleeAuto:
		return max(0, value.unit.AutoAttacks.NextAttackAt()-sim.CurrentTime)
	case proto.APLValueAutoAttackType_MainHandAuto:
		return max(0, value.unit.AutoAttacks.MainhandSwingAt()-sim.CurrentTime)
	case proto.APLValueAutoAttackType_OffHandAuto:
		return max(0, value.unit.AutoAttacks.OffhandSwingAt()-sim.CurrentTime)
	case proto.APLValueAutoAttackType_RangedAuto:
		return max(0, value.unit.AutoAttacks.NextRangedAttackAt()-sim.CurrentTime)
	}
	// defaults to Any
	return max(0, value.unit.AutoAttacks.NextAnyAttackAt()-sim.CurrentTime)
}
func (value *APLValueAutoTimeToNext) String() string {
	return "Auto Time To Next"
}

type APLValueAutoTimeSinceLast struct {
	DefaultAPLValueImpl
	unit     *Unit
	autoType proto.APLValueAutoAttackType
}

func (rot *APLRotation) newValueAutoTimeSinceLast(config *proto.APLValueAutoTimeSinceLast, _ *proto.UUID) APLValue {
	return &APLValueAutoTimeSinceLast{
		unit:     rot.unit,
		autoType: config.AutoType,
	}
}
func (value *APLValueAutoTimeSinceLast) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueAutoTimeSinceLast) GetDuration(sim *Simulation) time.Duration {
	var swingValue time.Duration
	switch value.autoType {
	case proto.APLValueAutoAttackType_MeleeAuto:
		swingValue = sim.CurrentTime - value.unit.AutoAttacks.MainhandPreviousSwing()
		if value.unit.AutoAttacks.oh.enabled {
			swingValue = min(swingValue, sim.CurrentTime-value.unit.AutoAttacks.OffhandPreviousSwing())
		}
	case proto.APLValueAutoAttackType_MainHandAuto:
		swingValue = sim.CurrentTime - value.unit.AutoAttacks.MainhandPreviousSwing()
	case proto.APLValueAutoAttackType_OffHandAuto:
		swingValue = sim.CurrentTime - value.unit.AutoAttacks.OffhandPreviousSwing()
	case proto.APLValueAutoAttackType_RangedAuto:
		swingValue = sim.CurrentTime - value.unit.AutoAttacks.PreviousRangedAttack()
	// defaults to Any
	default:
		swingValue = sim.CurrentTime - value.unit.AutoAttacks.MainhandPreviousSwing()
		if value.unit.AutoAttacks.oh.enabled {
			swingValue = min(swingValue, sim.CurrentTime-value.unit.AutoAttacks.OffhandPreviousSwing())
		}
		if value.unit.AutoAttacks.ranged.enabled {
			swingValue = min(swingValue, sim.CurrentTime-value.unit.AutoAttacks.PreviousRangedAttack())
		}
	}

	return swingValue
}
func (value *APLValueAutoTimeSinceLast) String() string {
	return "Auto Time Since Last"
}
