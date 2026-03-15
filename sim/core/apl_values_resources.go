package core

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
)

type APLValueCurrentHealth struct {
	DefaultAPLValueImpl
	unit UnitReference
}

func (rot *APLRotation) newValueCurrentHealth(config *proto.APLValueCurrentHealth, uuid *proto.UUID) APLValue {
	unit := rot.GetSourceUnit(config.SourceUnit)
	resolvedUnit := unit.Get()
	if resolvedUnit == nil {
		return nil
	}
	if !resolvedUnit.HasHealthBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Health", resolvedUnit.Label)
		return nil
	}
	return &APLValueCurrentHealth{
		unit: unit,
	}
}
func (value *APLValueCurrentHealth) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentHealth) GetFloat(sim *Simulation) float64 {
	return value.unit.Get().CurrentHealth()
}
func (value *APLValueCurrentHealth) String() string {
	return "Current Health"
}

type APLValueCurrentHealthPercent struct {
	DefaultAPLValueImpl
	unit UnitReference
}

func (rot *APLRotation) newValueCurrentHealthPercent(config *proto.APLValueCurrentHealthPercent, uuid *proto.UUID) APLValue {
	unit := rot.GetSourceUnit(config.SourceUnit)
	resolvedUnit := unit.Get()

	if resolvedUnit == nil {
		return nil
	}
	if !resolvedUnit.HasHealthBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Health", resolvedUnit.Label)
		return nil
	}
	return &APLValueCurrentHealthPercent{
		unit: unit,
	}
}
func (value *APLValueCurrentHealthPercent) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentHealthPercent) GetFloat(sim *Simulation) float64 {
	return value.unit.Get().CurrentHealthPercent()
}
func (value *APLValueCurrentHealthPercent) String() string {
	return fmt.Sprintf("Current Health %%")
}

type APLValueMaxHealth struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueMaxHealth(_ *proto.APLValueMaxHealth, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasHealthBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Health", unit.Label)
		return nil
	}
	return &APLValueMaxHealth{
		unit: unit,
	}
}
func (value *APLValueMaxHealth) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueMaxHealth) GetFloat(sim *Simulation) float64 {
	return value.unit.MaxHealth()
}
func (value *APLValueMaxHealth) String() string {
	return "Max Health"
}

type APLValueCurrentMana struct {
	DefaultAPLValueImpl
	unit UnitReference
}

func (rot *APLRotation) newValueCurrentMana(config *proto.APLValueCurrentMana, uuid *proto.UUID) APLValue {
	unit := rot.GetSourceUnit(config.SourceUnit)
	resolvedUnit := unit.Get()

	if resolvedUnit == nil {
		return nil
	}
	if !resolvedUnit.HasManaBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Mana", resolvedUnit.Label)
		return nil
	}
	return &APLValueCurrentMana{
		unit: unit,
	}
}
func (value *APLValueCurrentMana) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentMana) GetFloat(sim *Simulation) float64 {
	return value.unit.Get().CurrentMana()
}
func (value *APLValueCurrentMana) String() string {
	return "Current Mana"
}

type APLValueCurrentManaPercent struct {
	DefaultAPLValueImpl
	unit UnitReference
}

func (rot *APLRotation) newValueCurrentManaPercent(config *proto.APLValueCurrentManaPercent, uuid *proto.UUID) APLValue {
	unit := rot.GetSourceUnit(config.SourceUnit)
	resolvedUnit := unit.Get()

	if resolvedUnit == nil {
		return nil
	}
	if !resolvedUnit.HasManaBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Mana", resolvedUnit.Label)
		return nil
	}
	return &APLValueCurrentManaPercent{
		unit: unit,
	}
}
func (value *APLValueCurrentManaPercent) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentManaPercent) GetFloat(sim *Simulation) float64 {
	return value.unit.Get().CurrentManaPercent()
}
func (value *APLValueCurrentManaPercent) String() string {
	return fmt.Sprintf("Current Mana %%")
}

type APLValueMaxMana struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueMaxMana(_ *proto.APLValueMaxMana, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasManaBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Mana", unit.Label)
		return nil
	}
	return &APLValueMaxMana{
		unit: unit,
	}
}
func (value *APLValueMaxMana) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueMaxMana) GetFloat(sim *Simulation) float64 {
	return value.unit.MaxMana()
}
func (value *APLValueMaxMana) String() string {
	return "Max Mana"
}

type APLValueCurrentRage struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueCurrentRage(_ *proto.APLValueCurrentRage, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasRageBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Rage", unit.Label)
		return nil
	}
	return &APLValueCurrentRage{
		unit: unit,
	}
}
func (value *APLValueCurrentRage) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentRage) GetFloat(sim *Simulation) float64 {
	return value.unit.CurrentRage()
}
func (value *APLValueCurrentRage) String() string {
	return "Current Rage"
}

type APLValueMaxRage struct {
	DefaultAPLValueImpl
	maxRage float64
}

func (rot *APLRotation) newValueMaxRage(_ *proto.APLValueMaxRage, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasRageBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Error, "%s does not use Rage", unit.Label)
		return nil
	}
	return &APLValueMaxRage{
		maxRage: unit.MaximumRage(),
	}
}
func (value *APLValueMaxRage) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueMaxRage) GetFloat(sim *Simulation) float64 {
	return value.maxRage
}
func (value *APLValueMaxRage) String() string {
	return fmt.Sprintf("Max Rage(%f)", value.maxRage)
}

type APLValueCurrentEnergy struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueCurrentEnergy(_ *proto.APLValueCurrentEnergy, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Energy", unit.Label)
		return nil
	}
	return &APLValueCurrentEnergy{
		unit: unit,
	}
}
func (value *APLValueCurrentEnergy) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentEnergy) GetFloat(sim *Simulation) float64 {
	return value.unit.CurrentEnergy()
}
func (value *APLValueCurrentEnergy) String() string {
	return "Current Energy"
}

type APLValueMaxEnergy struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueMaxEnergy(_ *proto.APLValueMaxEnergy, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Error, "%s does not use Energy", unit.Label)
		return nil
	}
	return &APLValueMaxEnergy{
		unit: unit,
	}
}
func (value *APLValueMaxEnergy) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueMaxEnergy) GetFloat(sim *Simulation) float64 {
	return value.unit.MaximumEnergy()
}
func (value *APLValueMaxEnergy) String() string {
	return "Max Energy"
}

type APLValueEnergyRegenPerSecond struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueEnergyRegenPerSecond(_ *proto.APLValueEnergyRegenPerSecond, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Energy", unit.Label)
		return nil
	}
	return &APLValueEnergyRegenPerSecond{
		unit: unit,
	}
}
func (value *APLValueEnergyRegenPerSecond) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueEnergyRegenPerSecond) GetFloat(sim *Simulation) float64 {
	return value.unit.EnergyRegenPerSecond()
}
func (value *APLValueEnergyRegenPerSecond) String() string {
	return "Energy Regen Per Second"
}

type APLValueEnergyTimeToTarget struct {
	DefaultAPLValueImpl
	unit         *Unit
	targetEnergy APLValue
}

func (rot *APLRotation) newValueEnergyTimeToTarget(config *proto.APLValueEnergyTimeToTarget, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Energy", unit.Label)
		return nil
	}

	targetEnergy := rot.coerceTo(rot.newAPLValue(config.TargetEnergy), proto.APLValueType_ValueTypeFloat)
	if targetEnergy == nil {
		return nil
	}

	return &APLValueEnergyTimeToTarget{
		unit:         unit,
		targetEnergy: targetEnergy,
	}
}
func (value *APLValueEnergyTimeToTarget) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueEnergyTimeToTarget) GetDuration(sim *Simulation) time.Duration {
	return value.unit.TimeToTargetEnergy(value.targetEnergy.GetFloat(sim))
}
func (value *APLValueEnergyTimeToTarget) String() string {
	return "Estimated Time To Target Energy"
}

type APLValueCurrentComboPoints struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueCurrentComboPoints(_ *proto.APLValueCurrentComboPoints, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Combo Points", unit.Label)
		return nil
	}
	return &APLValueCurrentComboPoints{
		unit: unit,
	}
}
func (value *APLValueCurrentComboPoints) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueCurrentComboPoints) GetInt(sim *Simulation) int32 {
	return value.unit.ComboPoints()
}
func (value *APLValueCurrentComboPoints) String() string {
	return "Current Combo Points"
}

type APLValueMaxComboPoints struct {
	DefaultAPLValueImpl
	maxComboPoints int32
}

func (rot *APLRotation) newValueMaxComboPoints(_ *proto.APLValueMaxComboPoints, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Error, "%s does not use Combo Points", unit.Label)
		return nil
	}
	return &APLValueMaxComboPoints{
		maxComboPoints: unit.MaxComboPoints(),
	}
}
func (value *APLValueMaxComboPoints) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueMaxComboPoints) GetInt(sim *Simulation) int32 {
	return value.maxComboPoints
}
func (value *APLValueMaxComboPoints) String() string {
	return fmt.Sprintf("Max Combo Points(%d)", value.maxComboPoints)
}

type APLValueCurrentGenericResource struct {
	DefaultAPLValueImpl
	unit *Unit
}
