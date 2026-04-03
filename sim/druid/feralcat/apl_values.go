package feralcat

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (cat *FeralDruid) NewAPLAction(rot *core.APLRotation, config *proto.APLAction) core.APLActionImpl {
	switch config.Action.(type) {
	case *proto.APLAction_CatOptimalRotationAction:
		return cat.newActionCatOptimalRotationAction(rot, config.GetCatOptimalRotationAction())
	default:
		return nil
	}
}

type APLActionCatOptimalRotationAction struct {
	cat        *FeralDruid
	lastAction time.Duration
}

func (impl *APLActionCatOptimalRotationAction) GetInnerActions() []*core.APLAction { return nil }
func (impl *APLActionCatOptimalRotationAction) GetAPLValues() []core.APLValue      { return nil }
func (impl *APLActionCatOptimalRotationAction) Finalize(*core.APLRotation)         {}
func (impl *APLActionCatOptimalRotationAction) PostFinalize(*core.APLRotation)     {}
func (impl *APLActionCatOptimalRotationAction) GetNextAction(*core.Simulation) *core.APLAction {
	return nil
}
func (impl *APLActionCatOptimalRotationAction) ReResolveVariableRefs(*core.APLRotation, map[string]*proto.APLValue) {
}

func (cat *FeralDruid) newActionCatOptimalRotationAction(_ *core.APLRotation, config *proto.APLActionCatOptimalRotationAction) core.APLActionImpl {
	rotationOptions := &proto.FeralCatDruid_Rotation{
		FinishingMove:      config.FinishingMove,
		Biteweave:          config.Biteweave,
		Ripweave:           config.Ripweave,
		RipMinComboPoints:  config.RipMinComboPoints,
		BiteMinComboPoints: config.BiteMinComboPoints,
		MangleTrick:        config.MangleTrick,
		RakeTrick:          config.RakeTrick,
		MaintainFaerieFire: config.MaintainFaerieFire,
	}

	cat.setupRotation(rotationOptions)

	return &APLActionCatOptimalRotationAction{
		cat: cat,
	}
}

func (action *APLActionCatOptimalRotationAction) IsReady(sim *core.Simulation) bool {
	return sim.CurrentTime > action.lastAction
}

func (action *APLActionCatOptimalRotationAction) Execute(sim *core.Simulation) {
	action.lastAction = sim.CurrentTime
	action.cat.doRotation(sim)
}

func (action *APLActionCatOptimalRotationAction) Reset(*core.Simulation) {
	action.lastAction = -100 * time.Second
}

func (action *APLActionCatOptimalRotationAction) String() string {
	return "Execute Optimal Cat Action()"
}
