package dps

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
	"github.com/wowsims/tbc/sim/warrior"
)

func RegisterDpsWarrior() {
	core.RegisterAgentFactory(
		proto.Player_DpsWarrior{},
		proto.Spec_SpecDPSWarrior,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewDpsWarrior(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_DpsWarrior)
			if !ok {
				panic("Invalid spec value for Dps Warrior!")
			}
			player.Spec = playerSpec
		},
	)
}

type DpsWarrior struct {
	*warrior.Warrior

	Options *proto.DPSWarrior_Options

	BloodsurgeAura  *core.Aura
	MeatCleaverAura *core.Aura
}

// ApplyTalents implements core.Agent.
func (war *DpsWarrior) ApplyTalents() {
	// panic("unimplemented")
}

func NewDpsWarrior(character *core.Character, options *proto.Player) *DpsWarrior {
	dpsOptions := options.GetDpsWarrior().Options

	war := &DpsWarrior{
		Warrior: warrior.NewWarrior(character, dpsOptions.ClassOptions, options.TalentsString, warrior.WarriorInputs{
			StanceSnapshot: dpsOptions.StanceSnapshot,
		}),
		Options: dpsOptions,
	}

	war.ApplySyncType(dpsOptions.SyncType)

	return war
}

func (war *DpsWarrior) GetWarrior() *warrior.Warrior {
	return war.Warrior
}

func (war *DpsWarrior) Initialize() {
	war.Warrior.Initialize()
	war.registerPassives()
	// war.registerBloodthirst()
}

func (war *DpsWarrior) registerPassives() {
	war.ApplyArmorSpecializationEffect(stats.Strength, proto.ArmorType_ArmorTypePlate, 86526)

	// war.registerCrazedBerserker()
	// war.registerFlurry()
	// war.registerBloodsurge()
	// war.registerMeatCleaver()
	// war.registerSingleMindedFuryOrTitansGrip()
	// war.registerUnshackledFury()
}

func (war *DpsWarrior) Reset(sim *core.Simulation) {
	war.Warrior.Reset(sim)
}

func (war *DpsWarrior) OnEncounterStart(sim *core.Simulation) {
	war.ResetRageBar(sim, 25+war.PrePullChargeGain)
	war.Warrior.OnEncounterStart(sim)
}

func (war *DpsWarrior) ApplySyncType(syncType proto.WarriorSyncType) {
	if syncType == proto.WarriorSyncType_WarriorSyncMainhandOffhandSwings {
		war.AutoAttacks.SetReplaceMHSwing(func(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
			aa := &war.AutoAttacks
			if nextMHSwingAt := sim.CurrentTime + aa.MainhandSwingSpeed(); nextMHSwingAt > aa.OffhandSwingAt() {
				aa.SetOffhandSwingAt(nextMHSwingAt)
			}

			return mhSwingSpell
		})

	} else {
		war.AutoAttacks.SetReplaceMHSwing(nil)
	}
}
