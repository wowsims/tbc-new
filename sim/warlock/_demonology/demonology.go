package demonology

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/warlock"
)

func RegisterDemonologyWarlock() {
	core.RegisterAgentFactory(
		proto.Player_DemonologyWarlock{},
		proto.Spec_SpecDemonologyWarlock,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewDemonologyWarlock(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_DemonologyWarlock)
			if !ok {
				panic("Invalid spec value for Demonology Warlock!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewDemonologyWarlock(character *core.Character, options *proto.Player) *DemonologyWarlock {
	demoOptions := options.GetDemonologyWarlock().Options

	demonology := &DemonologyWarlock{
		Warlock: warlock.NewWarlock(character, options, demoOptions.ClassOptions),
	}

	demonology.Felguard = demonology.registerFelguard()
	demonology.registerWildImp(15)
	demonology.registerGrimoireOfService()
	return demonology
}

type DemonologyWarlock struct {
	*warlock.Warlock

	DemonicFury   core.SecondaryResourceBar
	Metamorphosis *core.Spell
	HandOfGuldan  *core.Spell
	ChaosWave     *core.Spell

	MoltenCore *core.Aura

	Felguard               *warlock.WarlockPet
	WildImps               []*WildImpPet
	HandOfGuldanImpactTime time.Duration
	ImpSwarm               *core.Spell
}

func (demonology *DemonologyWarlock) GetWarlock() *warlock.Warlock {
	return demonology.Warlock
}

const DefaultDemonicFury = 200

func (demonology *DemonologyWarlock) Initialize() {
	demonology.Warlock.Initialize()

	demonology.DemonicFury = demonology.RegisterNewDefaultSecondaryResourceBar(core.SecondaryResourceConfig{
		Type:    proto.SecondaryResourceType_SecondaryResourceTypeDemonicFury,
		Max:     1000, // Multiplied by 10 to avoid having to refactor to float
		Default: DefaultDemonicFury,
	})

	demonology.registerMetamorphosis()
	demonology.registerMasterDemonologist()
	demonology.registerShadowBolt()
	demonology.registerFelFlame()
	demonology.registerCorruption()
	demonology.registerDrainLife()
	demonology.registerHandOfGuldan()
	demonology.registerHellfire()
	demonology.registerSoulfire()
	demonology.registerMoltenCore()
	demonology.registerCarrionSwarm()
	demonology.registerChaosWave()
	demonology.registerDoom()
	demonology.registerImmolationAura()
	demonology.registerTouchOfChaos()
	demonology.registerVoidRay()
	demonology.registerDarksoulKnowledge()

	demonology.registerHotfixes()
}

func (demonology *DemonologyWarlock) ApplyTalents() {
	demonology.Warlock.ApplyTalents()

	// Demo specific versions
	demonology.registerGrimoireOfSupremacy()
	demonology.registerGrimoireOfSacrifice()
}

func (demonology *DemonologyWarlock) Reset(sim *core.Simulation) {
	demonology.Warlock.Reset(sim)

	demonology.HandOfGuldanImpactTime = 0
}

func (demonology *DemonologyWarlock) OnEncounterStart(sim *core.Simulation) {
	demonology.DemonicFury.ResetBarTo(sim, DefaultDemonicFury)
	demonology.Warlock.OnEncounterStart(sim)
}

func NewDemonicFuryCost(cost int) *warlock.SecondaryResourceCost {
	return &warlock.SecondaryResourceCost{
		SecondaryCost: cost,
		Name:          "Demonic Fury",
	}
}

func (demo *DemonologyWarlock) IsInMeta() bool {
	return demo.Metamorphosis.RelatedSelfBuff.IsActive()
}

func (demo *DemonologyWarlock) CanSpendDemonicFury(amount float64) bool {
	if demo.T15_2pc.IsActive() {
		amount *= 0.7
	}

	return demo.DemonicFury.CanSpend(amount)
}

func (demo *DemonologyWarlock) SpendUpToDemonicFury(sim *core.Simulation, limit float64, actionID core.ActionID) {
	if demo.T15_2pc.IsActive() {
		limit *= 0.7
	}

	demo.DemonicFury.SpendUpTo(sim, limit, actionID)
}

func (demo *DemonologyWarlock) SpendDemonicFury(sim *core.Simulation, amount float64, actionID core.ActionID) {
	if demo.T15_2pc.IsActive() {
		amount *= 0.7
	}

	demo.DemonicFury.Spend(sim, amount, actionID)
}

func (demo *DemonologyWarlock) GainDemonicFury(sim *core.Simulation, amount float64, actionID core.ActionID) {
	if demo.T15_4pc.IsActive() {
		amount *= 1.1
	}

	demo.DemonicFury.Gain(sim, amount, actionID)
}
