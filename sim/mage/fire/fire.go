package fire

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/mage"
)

const (
	DDBC_Pyromaniac int = iota
	DDBC_Total
)

func RegisterFireMage() {
	core.RegisterAgentFactory(
		proto.Player_FireMage{},
		proto.Spec_SpecFireMage,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewFireMage(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_FireMage)
			if !ok {
				panic("Invalid spec value for Fire Mage!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewFireMage(character *core.Character, options *proto.Player) *FireMage {
	fireOptions := options.GetFireMage().Options

	fireMage := &FireMage{
		Mage: mage.NewMage(character, options, fireOptions.ClassOptions),
	}
	fireMage.FireOptions = fireOptions
	fireMage.combustionDotDamageMultiplier = 0.2
	fireMage.criticalMassMultiplier = 0.3

	return fireMage
}

type FireMage struct {
	*mage.Mage

	Combustion   *core.Spell
	Ignite       *core.Spell
	Pyroblast    *core.Spell
	InfernoBlast *core.Spell

	pyromaniacAuras core.AuraArray

	criticalMassMultiplier        float64
	combustionDotDamageMultiplier float64
	combustionDotEstimate         int32
}

func (fireMage *FireMage) GetMage() *mage.Mage {
	return fireMage.Mage
}

func (fireMage *FireMage) Reset(sim *core.Simulation) {
	fireMage.Mage.Reset(sim)
}

func (fireMage *FireMage) Initialize() {
	fireMage.Mage.Initialize()

	fireMage.registerPassives()
	fireMage.registerSpells()
	fireMage.registerHotfixes()
}

func (fireMage *FireMage) registerPassives() {
	fireMage.registerMastery()
	fireMage.registerCriticalMass()
	fireMage.registerPyromaniac()
}

func (fireMage *FireMage) registerSpells() {
	fireMage.registerCombustionSpell()
	fireMage.registerFireballSpell()
	fireMage.registerInfernoBlastSpell()
	fireMage.registerDragonsBreathSpell()
	fireMage.registerPyroblastSpell()
	fireMage.registerScorchSpell()
}
