package feralbear

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (bear *GuardianDruid) NewAPLAction(rot *core.APLRotation, config *proto.APLAction) core.APLActionImpl {
	switch config.Action.(type) {
	case *proto.APLAction_BearOptimalRotationAction:
		return bear.newActionBearOptimalRotationAction(rot, config.GetBearOptimalRotationAction())
	default:
		return nil
	}
}

type APLActionBearOptimalRotationAction struct {
	bear       *GuardianDruid
	lastAction time.Duration
}

func (impl *APLActionBearOptimalRotationAction) GetInnerActions() []*core.APLAction { return nil }
func (impl *APLActionBearOptimalRotationAction) GetAPLValues() []core.APLValue      { return nil }
func (impl *APLActionBearOptimalRotationAction) Finalize(*core.APLRotation)         {}
func (impl *APLActionBearOptimalRotationAction) PostFinalize(*core.APLRotation)     {}
func (impl *APLActionBearOptimalRotationAction) GetNextAction(*core.Simulation) *core.APLAction {
	return nil
}
func (impl *APLActionBearOptimalRotationAction) ReResolveVariableRefs(*core.APLRotation, map[string]*proto.APLValue) {
}

func (bear *GuardianDruid) newActionBearOptimalRotationAction(_ *core.APLRotation, config *proto.APLActionBearOptimalRotationAction) core.APLActionImpl {
	rotationOptions := &proto.FeralBearDruid_Rotation{
		MaintainFaerieFire:       config.MaintainFaerieFire,
		MaintainDemoralizingRoar: config.MaintainDemoralizingRoar,
		MaulRageThreshold:        config.MaulRageThreshold,
		SwipeUsage:               config.SwipeUsage,
		SwipeApThreshold:         config.SwipeApThreshold,
	}

	bear.setupRotation(rotationOptions)

	return &APLActionBearOptimalRotationAction{
		bear: bear,
	}
}

func (action *APLActionBearOptimalRotationAction) IsReady(sim *core.Simulation) bool {
	return sim.CurrentTime > action.lastAction
}

func (action *APLActionBearOptimalRotationAction) Execute(sim *core.Simulation) {
	action.lastAction = sim.CurrentTime
	action.bear.doRotation(sim)
}

func (action *APLActionBearOptimalRotationAction) Reset(*core.Simulation) {
	action.lastAction = -100 * time.Second
}

func (action *APLActionBearOptimalRotationAction) String() string {
	return "Execute Optimal Bear Action()"
}
