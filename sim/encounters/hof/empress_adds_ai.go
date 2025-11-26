package hof

import (
	"fmt"
	"math"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

const windbladeMeleeDamageSpread = 1.9462
const windbladeAddID int32 = 64453

func addEmpress(raidPrefix string) {
	createEmpressAddsPreset(raidPrefix, 25, true, 53_120_592, 128_513)
}

func createEmpressAddsPreset(raidPrefix string, raidSize int32, isHeroic bool, addHealth float64, addMinBaseDamage float64) {
	bossName := fmt.Sprintf("Grand Empress Shek'zeer %d", raidSize)
	addName := fmt.Sprintf("Set'thik Windblade %d", raidSize)

	if isHeroic {
		bossName += " H"
		addName += " H"
	}

	targetPathNames := []string{}

	for addIdx := int32(1); addIdx <= 6; addIdx++ {
		currentAddName := addName + fmt.Sprintf(" - %d", addIdx)

		core.AddPresetTarget(&core.PresetTarget{
			PathPrefix: raidPrefix,

			Config: &proto.Target{
				Id:        windbladeAddID*100 + addIdx, // hack to guarantee distinct IDs for each add
				Name:      currentAddName,
				Level:     92,
				MobType:   proto.MobType_MobTypeHumanoid,
				TankIndex: 0,

				Stats: stats.Stats{
					stats.Health:      addHealth,
					stats.Armor:       24835, // TODO: verify add armor
					stats.AttackPower: 0,     // actual value doesn't matter in MoP, as long as damage parameters are fit consistently
				}.ToProtoArray(),

				SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
				SwingSpeed:    1.5,
				MinBaseDamage: addMinBaseDamage,
				DualWield:     true,
				DamageSpread:  windbladeMeleeDamageSpread,
				TargetInputs:  []*proto.TargetInput{},
			},

			AI: makeEmpressAI(raidSize, isHeroic, addIdx),
		})

		targetPathNames = append(targetPathNames, raidPrefix+"/"+currentAddName)
	}

	core.AddPresetEncounter(bossName+" P2 Adds", targetPathNames)
}

func makeEmpressAI(raidSize int32, isHeroic bool, addIdx int32) core.AIFactory {
	return func() core.TargetAI {
		return &WindbladeAI{
			raidSize: raidSize,
			isHeroic: isHeroic,
			addIdx:   addIdx,
		}
	}
}

type WindbladeAI struct {
	// Unit references
	Target   *core.Target
	AddUnits []*core.Unit
	TankUnit *core.Unit

	// Static parameters associated with a given preset
	raidSize int32
	isHeroic bool
	addIdx   int32

	// Spell + aura references
	SonicBlade  *core.Spell
	BandOfValor *core.Aura
}

func (ai *WindbladeAI) Initialize(target *core.Target, config *proto.Target) {
	// Save unit references
	ai.Target = target
	ai.AddUnits = target.Env.Encounter.AllTargetUnits[:]
	ai.TankUnit = target.CurrentTarget

	// Hack for UI results parsing
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = windbladeAddID*100 + ai.addIdx
	ai.Target.AutoAttacks.OHConfig().ActionID.Tag = windbladeAddID*100 + ai.addIdx

	// Register relevant spells and auras
	ai.registerSonicBlade()
	ai.registerBandOfValor()
}

func (ai *WindbladeAI) registerSonicBlade() {
	ai.SonicBlade = ai.Target.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 125886},
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		Flags:            core.SpellFlagMeleeMetrics,
		DamageMultiplier: 1.5,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.BossGCD,
			},

			CD: core.Cooldown{
				Timer:    ai.Target.NewTimer(),
				Duration: time.Second * 20,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.Unit.AutoAttacks.MH().EnemyWeaponDamage(sim, spell.MeleeAttackPower(), windbladeMeleeDamageSpread)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeEnemyMeleeWhite)
		},
	})
}

func (ai *WindbladeAI) registerBandOfValor() {
	ai.BandOfValor = ai.Target.GetOrRegisterAura(core.Aura{
		Label:     "Band of Valor",
		ActionID:  core.ActionID{SpellID: 125422},
		Duration:  core.NeverExpires,
		MaxStacks: math.MaxInt32,

		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.SetStacks(sim, int32(len(ai.AddUnits)-1))
		},

		OnStacksChange: func(aura *core.Aura, _ *core.Simulation, oldStacks int32, newStacks int32) {
			aura.Unit.PseudoStats.DamageDealtMultiplier *= (1.0 + 0.3*float64(newStacks)) / (1.0 + 0.3*float64(oldStacks))
		},
	})
}

func (ai *WindbladeAI) Reset(sim *core.Simulation) {
	aa := &ai.Target.AutoAttacks
	aa.RandomizeMeleeTiming(sim)
	aa.SetOffhandSwingAt(aa.NextAttackAt() + core.DurationFromSeconds(sim.RandomFloat("OH Desync")*aa.MainhandSwingSpeed().Seconds()))
	ai.SonicBlade.CD.Set(time.Second * 10)
}

func (ai *WindbladeAI) ExecuteCustomRotation(sim *core.Simulation) {
	if ai.TankUnit == nil {
		return
	}

	if ai.SonicBlade.IsReady(sim) && sim.Proc(0.5, "Sonic Blade Timing") {
		ai.SonicBlade.Cast(sim, ai.TankUnit)
	} else {
		ai.Target.ExtendGCDUntil(sim, sim.CurrentTime+core.BossGCD)
	}
}
