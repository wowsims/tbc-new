package protection

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/warrior"
)

func RegisterProtectionWarrior() {
	core.RegisterAgentFactory(
		proto.Player_ProtectionWarrior{},
		proto.Spec_SpecProtectionWarrior,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewProtectionWarrior(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_ProtectionWarrior)
			if !ok {
				panic("Invalid spec value for Protection Warrior!")
			}
			player.Spec = playerSpec
		},
	)
}

type ProtectionWarrior struct {
	*warrior.Warrior

	Options *proto.ProtectionWarrior_Options
}

func (war *ProtectionWarrior) ApplyTalents() {
	war.Warrior.ApplyTalents()
}

func NewProtectionWarrior(character *core.Character, options *proto.Player) *ProtectionWarrior {
	protOptions := options.GetProtectionWarrior().Options

	war := &ProtectionWarrior{
		Warrior: warrior.NewWarrior(character, protOptions.ClassOptions, options.TalentsString, warrior.WarriorInputs{
			StanceSnapshot: protOptions.StanceSnapshot,
			QueueDelay:     protOptions.QueueDelay,
			Stance:         protOptions.Stance,
		}),
		Options: protOptions,
	}

	return war
}

func (war *ProtectionWarrior) GetWarrior() *warrior.Warrior {
	return war.Warrior
}

func (war *ProtectionWarrior) Initialize() {
	war.Warrior.Initialize()
}

func (war *ProtectionWarrior) Reset(sim *core.Simulation) {
	war.Warrior.Reset(sim)
}

func (war *ProtectionWarrior) OnEncounterStart(sim *core.Simulation) {
	war.Warrior.OnEncounterStart(sim)
}
