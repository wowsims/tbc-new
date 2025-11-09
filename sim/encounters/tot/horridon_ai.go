package tot

import (
	"fmt"
	"math"
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
)

const horridonID int32 = 68476
const jalakID int32 = 69374

func addHorridon(raidPrefix string) {
	createHorridonPreset(raidPrefix, 25, true, 1_962_616_500, 512_867, 78_504_660, 485_334)
	createHorridonPreset(raidPrefix, 10, true, 654_205_500, 491_480, 26_168_220, 423_170)
}

func createHorridonPreset(raidPrefix string, raidSize int32, isHeroic bool, horridonHealth float64, horridonMinBaseDamage float64, jalakHealth float64, jalakMinBaseDamage float64) {
	bossName := fmt.Sprintf("Horridon %d", raidSize)
	addName := fmt.Sprintf("War-God Jalak %d", raidSize)

	if isHeroic {
		bossName += " H"
		addName += " H"
	}

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: raidPrefix,

		Config: &proto.Target{
			Id:        horridonID,
			Name:      bossName,
			Level:     93,
			MobType:   proto.MobType_MobTypeBeast,
			TankIndex: 0,

			Stats: stats.Stats{
				stats.Health:      horridonHealth,
				stats.Armor:       24835,
				stats.AttackPower: 0, // actual value doesn't matter in MoP, as long as damage parameters are fit consistently
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:    2.0,
			MinBaseDamage: horridonMinBaseDamage,
			DamageSpread:  0.5508,
			TargetInputs:  horridonTargetInputs(),
		},

		AI: makeHorridonAI(raidSize, isHeroic, true),
	})

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: raidPrefix,

		Config: &proto.Target{
			Id:        jalakID,
			Name:      addName,
			Level:     93,
			MobType:   proto.MobType_MobTypeHumanoid,
			TankIndex: 1,

			Stats: stats.Stats{
				stats.Health:      jalakHealth,
				stats.Armor:       24835,
				stats.AttackPower: 0, // actual value doesn't matter in MoP, as long as damage parameters are fit consistently
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:    2.0,
			MinBaseDamage: jalakMinBaseDamage,
			DamageSpread:  0.4668,
			TargetInputs:  []*proto.TargetInput{},
		},

		AI: makeHorridonAI(raidSize, isHeroic, false),
	})

	core.AddPresetEncounter(bossName+" P2", []string{
		raidPrefix + "/" + bossName,
		raidPrefix + "/" + addName,
	})
}

func horridonTargetInputs() []*proto.TargetInput {
	return []*proto.TargetInput{
		{
			Label:       "Jalak death time",
			Tooltip:     "Simulation time (in seconds) at which to disable War-God Jalak and trigger the Rampage buff on Horridon. If set longer than the simulated fight length, then Jalak will be tanked the entire time and Rampage will never be triggered.",
			InputType:   proto.InputType_Number,
			NumberValue: 33,
		},
		{
			Label:     "Taunt swap for Triple Puncture",
			Tooltip:   "If checked, taunt swap upon Jalak's death and on every other Triple Puncture application afterwards in order to limit stack accumulation on a single tank.",
			InputType: proto.InputType_Bool,
			BoolValue: true,
		},
	}
}

func makeHorridonAI(raidSize int32, isHeroic bool, isBoss bool) core.AIFactory {
	return func() core.TargetAI {
		return &HorridonAI{
			raidSize: raidSize,
			isHeroic: isHeroic,
			isBoss:   isBoss,
		}
	}
}

type HorridonAI struct {
	// Unit references
	Target     *core.Target
	BossUnit   *core.Unit
	AddUnit    *core.Unit
	MainTank   *core.Unit
	OffTank    *core.Unit
	ValidTanks []*core.Unit

	// Static parameters associated with a given preset
	raidSize int32
	isHeroic bool
	isBoss   bool

	// Dynamic parameters taken from user inputs
	disableAddAt time.Duration
	tankSwap     bool
	lastTankSwap time.Duration

	// Spell + aura references
	TriplePuncture *core.Spell
	DireCallAura   *core.Aura
	DireCall       *core.Spell
	BestialCryAura *core.Aura
	BestialCry     *core.Spell
	RampageAura    *core.Aura
}

func (ai *HorridonAI) Initialize(target *core.Target, config *proto.Target) {
	// Save unit references
	ai.Target = target
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = core.TernaryInt32(ai.isBoss, horridonID, jalakID)

	if ai.isBoss {
		ai.BossUnit = &target.Unit
		ai.AddUnit = &target.NextActiveTarget().Unit
	} else {
		ai.AddUnit = &target.Unit
		ai.BossUnit = &target.NextActiveTarget().Unit
	}

	ai.MainTank = ai.BossUnit.CurrentTarget
	ai.OffTank = ai.AddUnit.CurrentTarget

	ai.ValidTanks = core.FilterSlice([]*core.Unit{ai.MainTank, ai.OffTank}, func(unit *core.Unit) bool {
		return unit != nil
	})

	// Save user input parameters
	if ai.isBoss {
		ai.disableAddAt = core.DurationFromSeconds(config.TargetInputs[0].NumberValue)
		ai.tankSwap = config.TargetInputs[1].BoolValue
	}

	// Register relevant spells and auras
	if ai.isBoss {
		ai.registerTriplePuncture()
		ai.registerDireCall()
		ai.registerRampage()
	} else {
		ai.registerBestialCry()
	}
}

func (ai *HorridonAI) registerTriplePuncture() {
	// 0 - 10N, 1 - 25N, 2 - 10H, 3 - 25H
	scalingIndex := core.TernaryInt(ai.raidSize == 10, core.TernaryInt(ai.isHeroic, 2, 0), core.TernaryInt(ai.isHeroic, 3, 1))
	triplePunctureBase := []float64{370000, 462500, 462500, 555000}[scalingIndex]
	triplePunctureVariance := []float64{60000, 75000, 75000, 90000}[scalingIndex]

	ai.TriplePuncture = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 136767},
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		Flags:            core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		DamageMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.BossGCD,
			},

			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: time.Second * 10,
			},

			IgnoreHaste: true,
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:     "Triple Puncture",
				MaxStacks: math.MaxInt32,
				Duration:  time.Second * 90,
			},

			NumberOfTicks: 1,
			TickLength:    time.Second * 90,

			OnSnapshot: func(_ *core.Simulation, _ *core.Unit, _ *core.Dot, _ bool) {
			},

			OnTick: func(_ *core.Simulation, _ *core.Unit, _ *core.Dot) {
			},
		},

		ApplyEffects: func(sim *core.Simulation, tankTarget *core.Unit, spell *core.Spell) {
			if tankTarget == ai.BossUnit.CurrentTarget {
				damageRoll := triplePunctureBase + triplePunctureVariance*sim.RandomFloat("Triple Puncture Damage")
				dot := spell.Dot(tankTarget)

				if dot.IsActive() {
					damageRoll *= 1.0 + 0.1*float64(dot.GetStacks())
				}

				spell.CalcAndDealDamage(sim, tankTarget, damageRoll, spell.OutcomeAlwaysHit)

				if dot.IsActive() {
					dot.Refresh(sim)
					dot.AddStack(sim)
				} else {
					dot.Apply(sim)
					dot.SetStacks(sim, 1)
				}
			}

			if ai.tankSwap && !ai.AddUnit.IsEnabled() && (sim.CurrentTime-ai.lastTankSwap > time.Second*20) {
				ai.tauntSwap(sim)
			}
		},
	})
}

func (ai *HorridonAI) tauntSwap(sim *core.Simulation) {
	newTankTarget := core.Ternary(ai.BossUnit.CurrentTarget == ai.MainTank, ai.OffTank, ai.MainTank)
	ai.BossUnit.AutoAttacks.CancelAutoSwing(sim)
	ai.BossUnit.CurrentTarget = newTankTarget
	ai.BossUnit.AutoAttacks.EnableAutoSwing(sim)
	ai.BossUnit.AutoAttacks.RandomizeMeleeTiming(sim)
	ai.lastTankSwap = sim.CurrentTime
}

func (ai *HorridonAI) registerDireCall() {
	if !ai.isHeroic {
		return
	}

	actionID := core.ActionID{SpellID: 137458}
	direCallBase := core.TernaryFloat64(ai.raidSize == 10, 250000, 270000)

	ai.DireCallAura = ai.AddUnit.RegisterAura(core.Aura{
		Label:    "Dire Call",
		ActionID: actionID,
		Duration: time.Second * 20,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.MultiplyMeleeSpeed(sim, 1.5)
		},

		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.MultiplyMeleeSpeed(sim, 1.0/1.5)
		},
	})

	ai.DireCall = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:         actionID,
		SpellSchool:      core.SpellSchoolPhysical,
		Flags:            core.SpellFlagAPL | core.SpellFlagIgnoreArmor,
		ProcMask:         core.ProcMaskSpellDamage,
		DamageMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.BossGCD * 2,
				CastTime: time.Second * 2,
			},

			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: time.Minute,
			},

			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Raid.AllPlayerUnits {
				spell.CalcAndDealDamage(sim, aoeTarget, direCallBase, spell.OutcomeAlwaysHit)
			}

			if ai.AddUnit.IsEnabled() {
				ai.DireCallAura.Activate(sim)
			}
		},
	})
}

func (ai *HorridonAI) registerRampage() {
	damageMulti := core.TernaryFloat64(ai.isHeroic, 2, 1.5)

	ai.RampageAura = ai.BossUnit.RegisterAura(core.Aura{
		Label:    "Rampage",
		ActionID: core.ActionID{SpellID: 136821},
		Duration: core.NeverExpires,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageDealtMultiplier *= damageMulti
			aura.Unit.MultiplyMeleeSpeed(sim, 1.5)
		},

		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageDealtMultiplier /= damageMulti
			aura.Unit.MultiplyMeleeSpeed(sim, 1.0/1.5)
		},

		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			if ai.disableAddAt < sim.Duration {
				pa := sim.GetConsumedPendingActionFromPool()
				pa.NextActionAt = ai.disableAddAt
				pa.Priority = core.ActionPriorityDOT

				pa.OnAction = func(sim *core.Simulation) {
					sim.DisableTargetUnit(ai.AddUnit, true)
					aura.Activate(sim)

					if ai.OffTank != nil {
						ai.OffTank.CurrentTarget = ai.BossUnit
					}

					if ai.tankSwap {
						ai.tauntSwap(sim)
					}
				}

				sim.AddPendingAction(pa)
			}
		},
	})
}

func (ai *HorridonAI) registerBestialCry() {
	actionID := core.ActionID{SpellID: 136817}

	ai.BestialCryAura = ai.AddUnit.RegisterAura(core.Aura{
		Label:     "Bestial Cry",
		ActionID:  actionID,
		Duration:  core.NeverExpires,
		MaxStacks: math.MaxInt32,

		OnStacksChange: func(aura *core.Aura, _ *core.Simulation, oldStacks int32, newStacks int32) {
			aura.Unit.PseudoStats.DamageDealtMultiplier *= (1.0 + 0.5*float64(newStacks)) / (1.0 + 0.5*float64(oldStacks))
		},
	})

	// 0 - 10N, 1 - 25N, 2 - 10H, 3 - 25H
	scalingIndex := core.TernaryInt(ai.raidSize == 10, core.TernaryInt(ai.isHeroic, 2, 0), core.TernaryInt(ai.isHeroic, 3, 1))
	bestialCryBase := []float64{100000, 125000, 180000, 200000}[scalingIndex]

	ai.BestialCry = ai.AddUnit.RegisterSpell(core.SpellConfig{
		ActionID:         actionID,
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskSpellDamage,
		Flags:            core.SpellFlagAPL | core.SpellFlagIgnoreArmor,
		DamageMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.BossGCD,
			},

			CD: core.Cooldown{
				Timer:    ai.AddUnit.NewTimer(),
				Duration: time.Second * 10,
			},

			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			for _, aoeTarget := range sim.Raid.AllPlayerUnits {
				spell.CalcAndDealDamage(sim, aoeTarget, bestialCryBase, spell.OutcomeAlwaysHit)
			}

			spell.Unit.AutoAttacks.PauseMeleeBy(sim, core.BossGCD)
			ai.BestialCryAura.Activate(sim)
			ai.BestialCryAura.AddStack(sim)
		},
	})
}

func (ai *HorridonAI) Reset(sim *core.Simulation) {
	ai.Target.Enable(sim)

	if ai.isBoss {
		ai.TriplePuncture.CD.Set(core.DurationFromSeconds(sim.RandomFloat("Triple Puncture Timing") * ai.TriplePuncture.CD.Duration.Seconds()))
		ai.DireCall.CD.Set(core.DurationFromSeconds(sim.RandomFloat("Dire Call Timing") * ai.DireCall.CD.Duration.Seconds()))
	} else {
		ai.BestialCry.CD.Set(core.DurationFromSeconds(sim.RandomFloat("Bestial Cry Timing") * 0.5 * ai.BestialCry.CD.Duration.Seconds()))
	}

	ai.lastTankSwap = -core.NeverExpires
}

func (ai *HorridonAI) ExecuteCustomRotation(sim *core.Simulation) {
	target := ai.Target.CurrentTarget
	if target == nil {
		// For individual non tank sims we still want abilities to work
		target = &ai.Target.Env.Raid.Parties[0].Players[0].GetCharacter().Unit
	}

	if ai.isBoss && ai.TriplePuncture.IsReady(sim) && sim.Proc(0.75, "Triple Puncture Timing") {
		ai.TriplePuncture.Cast(sim, target)
	} else if ai.isBoss && ai.DireCall.IsReady(sim) {
		ai.DireCall.Cast(sim, target)
	} else if !ai.isBoss && ai.BestialCry.IsReady(sim) {
		ai.BestialCry.Cast(sim, target)
	} else {
		ai.Target.ExtendGCDUntil(sim, sim.CurrentTime+core.BossGCD)
	}
}
