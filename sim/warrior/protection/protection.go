package protection

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
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

	SwordAndBoardAura *core.Aura
}

// ApplyTalents implements core.Agent.
func (war *ProtectionWarrior) ApplyTalents() {
	// panic("unimplemented")
}

func NewProtectionWarrior(character *core.Character, options *proto.Player) *ProtectionWarrior {
	protOptions := options.GetProtectionWarrior().Options

	war := &ProtectionWarrior{
		Warrior: warrior.NewWarrior(character, protOptions.ClassOptions, options.TalentsString, warrior.WarriorInputs{}),
		Options: protOptions,
	}

	return war
}

func (war *ProtectionWarrior) GetWarrior() *warrior.Warrior {
	return war.Warrior
}

func (war *ProtectionWarrior) Initialize() {
	war.Warrior.Initialize()
	war.registerPassives()

	// war.registerRevenge()
	// war.registerShieldSlam()
	// war.registerShieldBlock()
	// war.registerDemoralizingShout()
	// war.registerLastStand()
}

func (war *ProtectionWarrior) registerPassives() {
	war.ApplyArmorSpecializationEffect(stats.Stamina, proto.ArmorType_ArmorTypePlate, 86526)

	// war.registerUnwaveringSentinel()
	// war.registerBastionOfDefense()
	// war.registerSwordAndBoard()
	// war.registerUltimatum()
	// war.registerRiposte()
}

func (war *ProtectionWarrior) Reset(sim *core.Simulation) {
	war.Warrior.Reset(sim)
}

func (war *ProtectionWarrior) OnEncounterStart(sim *core.Simulation) {
	war.ResetRageBar(sim, core.TernaryFloat64(war.ShieldBarrierAura.IsActive(), 5, 25)+war.PrePullChargeGain)
	war.Warrior.OnEncounterStart(sim)
}
