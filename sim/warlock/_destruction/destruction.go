package destruction

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/warlock"
)

func RegisterDestructionWarlock() {
	core.RegisterAgentFactory(
		proto.Player_DestructionWarlock{},
		proto.Spec_SpecDestructionWarlock,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewDestructionWarlock(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_DestructionWarlock)
			if !ok {
				panic("Invalid spec value for Destruction Warlock!")
			}
			player.Spec = playerSpec
		},
	)
}

const SpellFlagDestructionHavoc = core.SpellFlagAgentReserved1

const DefaultBurningEmbers = 10

func NewDestructionWarlock(character *core.Character, options *proto.Player) *DestructionWarlock {
	destroOptions := options.GetDestructionWarlock().Options
	destruction := &DestructionWarlock{
		Warlock: warlock.NewWarlock(character, options, destroOptions.ClassOptions),
	}

	destruction.BurningEmbers = destruction.RegisterNewDefaultSecondaryResourceBar(core.SecondaryResourceConfig{
		Type:    proto.SecondaryResourceType_SecondaryResourceTypeBurningEmbers,
		Max:     40,
		Default: DefaultBurningEmbers,
	})

	return destruction
}

type DestructionWarlock struct {
	*warlock.Warlock

	Conflagrate      *core.Spell
	BurningEmbers    core.SecondaryResourceBar
	FABAura          *core.Aura
	FABImmolate      *core.Spell
	FABConflagrate   *core.Spell
	Havoc            *core.Spell
	HavocChargesAura *core.Aura
	HavocAuras       core.AuraArray
}

func (destruction DestructionWarlock) getGeneratorMasteryBonus() float64 {
	return 0.09 + 0.01*destruction.GetMasteryPoints()
}

func (destruction DestructionWarlock) getSpenderMasteryBonus() float64 {
	return 0.24 + 0.03*destruction.GetMasteryPoints()
}

func (destruction *DestructionWarlock) GetWarlock() *warlock.Warlock {
	return destruction.Warlock
}

func (destruction *DestructionWarlock) Initialize() {
	destruction.Warlock.Initialize()

	destruction.registerDarkSoulInstability()
	destruction.ApplyChaoticEnergy()
	destruction.ApplyMastery()
	destruction.registerIncinerate()
	destruction.registerConflagrate()
	destruction.registerImmolate()
	destruction.registerBackdraft()
	destruction.registerFelflame()
	destruction.registerChaosBolt()
	destruction.registerShadowBurnSpell()
	destruction.registerRainOfFire()
	destruction.registerFireAndBrimstone()
	destruction.registerHavoc()
	destruction.RegisterDrainLife(nil) // no extra callback needed
}

func (destruction *DestructionWarlock) ApplyTalents() {
	destruction.Warlock.ApplyTalents()
}

func (destruction *DestructionWarlock) Reset(sim *core.Simulation) {
	destruction.Warlock.Reset(sim)
}

func (destruction *DestructionWarlock) OnEncounterStart(sim *core.Simulation) {
	destruction.BurningEmbers.ResetBarTo(sim, DefaultBurningEmbers)
	destruction.Warlock.OnEncounterStart(sim)
}

var SpellMaskCinderSpender = warlock.WarlockSpellChaosBolt | warlock.WarlockSpellEmberTap | warlock.WarlockSpellShadowBurn
var SpellMaskCinderGenerator = warlock.WarlockSpellImmolate | warlock.WarlockSpellImmolateDot |
	warlock.WarlockSpellIncinerate | warlock.WarlockSpellFelFlame | warlock.WarlockSpellConflagrate |
	warlock.WarlockSpellFaBIncinerate | warlock.WarlockSpellFaBConflagrate
