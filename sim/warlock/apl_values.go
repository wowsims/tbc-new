package warlock

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (warlock *Warlock) NewAPLValue(rot *core.APLRotation, config *proto.APLValue) core.APLValue {
	switch config.Value.(type) {
	case *proto.APLValue_WarlockAssignedCurseIsActive:
		return warlock.newValueWarlockAssignedCurseIsActive(rot, config.GetWarlockAssignedCurseIsActive())
	default:
		return nil
	}
}

type APLValueWarlockAssignedCurseIsActive struct {
	core.DefaultAPLValueImpl
	warlock *Warlock
	target  core.UnitReference
}

func (x *APLValueWarlockAssignedCurseIsActive) GetInnerActions() []*core.APLAction { return nil }
func (x *APLValueWarlockAssignedCurseIsActive) GetAPLValues() []core.APLValue      { return nil }
func (x *APLValueWarlockAssignedCurseIsActive) Finalize(*core.APLRotation)         {}
func (x *APLValueWarlockAssignedCurseIsActive) GetNextAction(*core.Simulation) *core.APLAction {
	return nil
}
func (x *APLValueWarlockAssignedCurseIsActive) GetSpellFromAction(sim *core.Simulation) *core.Spell {
	return x.warlock.GetAssignedCurse()
}

func (warlock *Warlock) newValueWarlockAssignedCurseIsActive(rot *core.APLRotation, config *proto.APLValueWarlockAssignedCurseIsActive) core.APLValue {
	target := rot.GetTargetUnit(config.TargetUnit)

	if target.Get() == nil {
		return nil
	}
	return &APLValueWarlockAssignedCurseIsActive{
		warlock: warlock,
		target:  target,
	}
}

func (x *APLValueWarlockAssignedCurseIsActive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (x *APLValueWarlockAssignedCurseIsActive) GetBool(sim *core.Simulation) bool {
	assignedCurse := x.GetSpellFromAction(sim)
	aura := x.target.Get().GetAuraByID(assignedCurse.ActionID)

	return aura.IsActive()
}

func (x *APLValueWarlockAssignedCurseIsActive) String() string {
	return fmt.Sprintf("Is Assigned Curse Active (%s)", x.warlock.GetAssignedCurse().ActionID)
}

func (warlock *Warlock) NewAPLAction(rot *core.APLRotation, config *proto.APLAction) core.APLActionImpl {
	switch config.Action.(type) {
	case *proto.APLAction_CastWarlockAssignedCurse:
		return warlock.newActionWarlockAssignedCurseAction(rot, config.GetCastWarlockAssignedCurse())
	default:
		return nil
	}
}

type APLActionCastWarlockAssignedCurse struct {
	warlock    *Warlock
	lastAction time.Duration
	target     core.UnitReference
}

func (x *APLActionCastWarlockAssignedCurse) GetInnerActions() []*core.APLAction { return nil }
func (x *APLActionCastWarlockAssignedCurse) GetAPLValues() []core.APLValue      { return nil }
func (x *APLActionCastWarlockAssignedCurse) Finalize(*core.APLRotation)         {}
func (x *APLActionCastWarlockAssignedCurse) PostFinalize(*core.APLRotation)     {}
func (x *APLActionCastWarlockAssignedCurse) GetNextAction(*core.Simulation) *core.APLAction {
	return nil
}
func (x *APLActionCastWarlockAssignedCurse) ReResolveVariableRefs(*core.APLRotation, map[string]*proto.APLValue) {
}

func (x *APLActionCastWarlockAssignedCurse) GetSpellFromAction(sim *core.Simulation) *core.Spell {
	return x.warlock.GetAssignedCurse()
}

func (warlock *Warlock) newActionWarlockAssignedCurseAction(rot *core.APLRotation, config *proto.APLActionCastWarlockAssignedCurse) core.APLActionImpl {
	target := rot.GetTargetUnit(config.Target)
	if target.Get() == nil {
		return nil
	}
	return &APLActionCastWarlockAssignedCurse{
		warlock: warlock,
		target:  target,
	}
}

func (x *APLActionCastWarlockAssignedCurse) Execute(sim *core.Simulation) {
	x.GetSpellFromAction(sim).Cast(sim, x.target.Get())
}

func (x *APLActionCastWarlockAssignedCurse) IsReady(sim *core.Simulation) bool {
	return x.GetSpellFromAction(sim).CanCast(sim, x.warlock.CurrentTarget)
}

func (x *APLActionCastWarlockAssignedCurse) Reset(*core.Simulation) {
	x.lastAction = -core.NeverExpires
}

func (x *APLActionCastWarlockAssignedCurse) String() string {
	return fmt.Sprintf("Cast Assigned Curse(%s)", x.warlock.GetAssignedCurse().ActionID)
}

func (warlock *Warlock) GetAssignedCurse() *core.Spell {
	switch warlock.Options.CurseOptions {
	case proto.WarlockOptions_Agony:
		return warlock.CurseOfAgony

	case proto.WarlockOptions_Doom:
		return warlock.CurseOfDoom

	case proto.WarlockOptions_Elements:
		return warlock.CurseOfElements

	case proto.WarlockOptions_Recklessness:
		return warlock.CurseOfRecklessness
	}

	return nil
}
