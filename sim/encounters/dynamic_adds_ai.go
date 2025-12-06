package encounters

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

const (
	dynamicBossID int32 = 99999
	dynamicAddID  int32 = 99998
)

func addDynamicAddsAI() {
	createDynamicAddsAIPreset()
}

func createDynamicAddsAIPreset() {
	bossName := "Dynamic Boss"
	addName := "Dynamic Add"

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: "Default",

		Config: &proto.Target{
			Id:        dynamicBossID,
			Name:      bossName,
			Level:     73,
			MobType:   proto.MobType_MobTypeMechanical,
			TankIndex: 0,

			Stats: stats.Stats{
				stats.Armor:       7700,
				stats.AttackPower: 0,
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:    2.0,
			MinBaseDamage: 3000,
			DamageSpread:  0.5,
			TargetInputs:  dynamicAddsTargetInputs(),
		},

		AI: makeDynamicAddsAI(true),
	})

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: "Default",

		Config: &proto.Target{
			Id:        dynamicAddID,
			Name:      addName,
			Level:     72,
			MobType:   proto.MobType_MobTypeMechanical,
			TankIndex: 1,

			Stats: stats.Stats{
				stats.Armor:       7700,
				stats.AttackPower: 0,
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:    1.5,
			MinBaseDamage: 700,
			DamageSpread:  0.4,
			TargetInputs:  []*proto.TargetInput{},
		},

		AI: makeDynamicAddsAI(false),
	})

	core.AddPresetEncounter("Dynamic Adds", []string{
		"Default/" + bossName,
		"Default/" + addName,
	})
}

func dynamicAddsTargetInputs() []*proto.TargetInput {
	return []*proto.TargetInput{
		{
			Label:       "Add(s) respawn Time",
			Tooltip:     "Time for add(s) to respawn after previous died (in seconds)",
			InputType:   proto.InputType_Number,
			NumberValue: 10,
		},
		{
			Label:       "Add(s) lifetime",
			Tooltip:     "How long the add(s) stay alive (in seconds)",
			InputType:   proto.InputType_Number,
			NumberValue: 20,
		},
		{
			Label:       "Add(s) spawn delay",
			Tooltip:     "Initial delay before the add(s) spawn (in seconds)",
			InputType:   proto.InputType_Number,
			NumberValue: 10,
		},
	}
}

func makeDynamicAddsAI(isBoss bool) core.AIFactory {
	return func() core.TargetAI {
		return &DynamicAddsAI{
			isBoss: isBoss,
		}
	}
}

type DynamicAddsAI struct {
	Target   *core.Target
	BossUnit *core.Unit
	AddUnits []*core.Unit
	MainTank *core.Unit
	OffTank  *core.Unit

	// Static parameters associated with a given preset
	isBoss bool

	respawnTime time.Duration
	addLifetime time.Duration
	spawnDelay  time.Duration
}

func (ai *DynamicAddsAI) Initialize(target *core.Target, config *proto.Target) {
	ai.Target = target
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = core.TernaryInt32(ai.isBoss, dynamicBossID, dynamicAddID)

	ai.BossUnit = target.Env.Encounter.AllTargetUnits[0]
	ai.AddUnits = target.Env.Encounter.AllTargetUnits[1:]

	ai.MainTank = ai.BossUnit.CurrentTarget
	ai.OffTank = ai.AddUnits[0].CurrentTarget

	if ai.isBoss && len(config.TargetInputs) >= 3 {
		ai.addLifetime = core.DurationFromSeconds(config.TargetInputs[1].NumberValue)
		ai.respawnTime = core.DurationFromSeconds(config.TargetInputs[0].NumberValue)
		ai.spawnDelay = core.DurationFromSeconds(config.TargetInputs[2].NumberValue)
	}
}

func (ai *DynamicAddsAI) Reset(sim *core.Simulation) {
	ai.Target.AutoAttacks.RandomizeMeleeTiming(sim)

	if !ai.isBoss {
		return
	}

	for _, addTarget := range ai.AddUnits {
		sim.DisableTargetUnit(addTarget, true)
	}

	if ai.spawnDelay > 0 {
		pa := sim.GetConsumedPendingActionFromPool()
		pa.NextActionAt = ai.spawnDelay
		pa.Priority = core.ActionPriorityDOT
		pa.OnAction = func(sim *core.Simulation) {
			ai.spawnAdds(sim)
		}
		sim.AddPendingAction(pa)
	} else {
		ai.spawnAdds(sim)
	}
}

func (ai *DynamicAddsAI) spawnAdds(sim *core.Simulation) {
	for _, addUnit := range ai.AddUnits {
		sim.EnableTargetUnit(addUnit)
	}

	if sim.Log != nil {
		sim.Log("Spawned %d adds at %s.", len(ai.AddUnits), sim.CurrentTime)
	}

	pa := sim.GetConsumedPendingActionFromPool()
	pa.NextActionAt = sim.CurrentTime + ai.addLifetime
	pa.Priority = core.ActionPriorityDOT
	pa.OnAction = func(sim *core.Simulation) {
		ai.despawnAdds(sim)
	}
	sim.AddPendingAction(pa)
}

func (ai *DynamicAddsAI) despawnAdds(sim *core.Simulation) {
	for _, addUnit := range ai.AddUnits {
		sim.DisableTargetUnit(addUnit, true)
	}

	if sim.Log != nil {
		sim.Log("Despawned %d adds at %s.", len(ai.AddUnits), sim.CurrentTime)
	}

	// Only schedule next spawn if there is a respawn time
	if ai.respawnTime > 0 {
		pa := sim.GetConsumedPendingActionFromPool()
		pa.NextActionAt = sim.CurrentTime + ai.respawnTime
		pa.Priority = core.ActionPriorityDOT
		pa.OnAction = func(sim *core.Simulation) {
			ai.spawnAdds(sim)
		}
		sim.AddPendingAction(pa)
	} else {
		ai.spawnAdds(sim)
	}
}

func (ai *DynamicAddsAI) ExecuteCustomRotation(sim *core.Simulation) {
	ai.Target.ExtendGCDUntil(sim, sim.CurrentTime+core.BossGCD)
}
