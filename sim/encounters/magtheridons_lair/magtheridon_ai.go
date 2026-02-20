package magtheridonslair

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

const magtheridonMeleeDamageSpread = 0.413
const magtheridonID int32 = 17257

func addMagtheridon(raidPrefix string) {
	createMagtheridonPreset(raidPrefix, 25, 4_818_380, 14_603)
}

func createMagtheridonPreset(raidPrefix string, raidSize int32, bossHealth float64, bossMinBaseDamage float64) {
	bossName := fmt.Sprintf("Magtheridon %d", raidSize)

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: raidPrefix,

		Config: &proto.Target{
			Id:        magtheridonID,
			Name:      bossName,
			Level:     73,
			MobType:   proto.MobType_MobTypeDemon,
			TankIndex: 0,

			Stats: stats.Stats{
				stats.Health:      bossHealth,
				stats.Armor:       7685,
				stats.AttackPower: 320,
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:    2,
			MinBaseDamage: bossMinBaseDamage,
			DamageSpread:  magtheridonMeleeDamageSpread,

			ParryHaste: true,
		},

		AI: makeMagtheridonAI(),
	})

	core.AddPresetEncounter(bossName, []string{
		raidPrefix + "/" + bossName,
	})
}

func makeMagtheridonAI() core.AIFactory {
	return func() core.TargetAI {
		return &MagtheridonAI{}
	}
}

type MagtheridonAI struct {
	// Unit references
	Target     *core.Target
	BossUnit   *core.Unit
	MainTank   *core.Unit
	OffTank    *core.Unit
	ValidTanks []*core.Unit

	// Spell + aura references
	Cleave          *core.Spell
	Quake           *core.Spell
	BlastNova       *core.Spell
	ShadowGraspAura *core.Aura
}

func (ai *MagtheridonAI) Initialize(target *core.Target, config *proto.Target) {
	// Save unit references
	ai.Target = target
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = magtheridonID

	ai.BossUnit = &target.Unit

	ai.MainTank = ai.BossUnit.CurrentTarget

	ai.ValidTanks = core.FilterSlice([]*core.Unit{ai.MainTank, ai.OffTank}, func(unit *core.Unit) bool {
		return unit != nil
	})

	// Register relevant spells and auras
	ai.registerBlastNova()
	ai.registerQuake()
	ai.registerCleave()

	ai.BossUnit.AutoAttacks.SetReplaceMHSwing(ai.TryCleave)
}

func (ai *MagtheridonAI) registerBlastNova() {
	ai.ShadowGraspAura = ai.BossUnit.RegisterAura(core.Aura{
		Label:    "Shadow Grasp",
		ActionID: core.ActionID{SpellID: 30410},
		Duration: time.Second * 10,
	}).
		AttachMultiplicativePseudoStatBuff(&ai.BossUnit.PseudoStats.DamageTakenMultiplier, 3)

	ai.BlastNova = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 30616},
		SpellSchool: core.SpellSchoolFire,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.BossGCD,
				CastTime: time.Second * 2,
			},

			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				cast.CastTime = core.DurationFromSeconds(sim.Roll(0.2, 2))
				meleeDelay := sim.CurrentTime + cast.CastTime + ai.ShadowGraspAura.Duration
				spell.Unit.AutoAttacks.StopMeleeUntil(sim, meleeDelay)
				cleaveDelay := max(
					spell.Unit.AutoAttacks.MainhandSwingSpeed()+1,
					core.DurationFromSeconds(sim.RandomFloat("Cleave delay")*ai.Cleave.CD.Duration.Seconds()),
				)
				ai.Cleave.CD.Set(meleeDelay + cleaveDelay)
			},

			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: time.Minute * 1,
			},

			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysMiss)
			ai.ShadowGraspAura.Activate(sim)
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		ai.BlastNova.CD.Set(core.DurationFromSeconds(sim.RandomFloat("Blast Nova Timing")) + ai.BlastNova.CD.Duration)
	})
}

func (ai *MagtheridonAI) registerQuake() {
	ai.Quake = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 30576},
		SpellSchool: core.SpellSchoolPhysical,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.BossGCD,
				CastTime: time.Second * 7,
			},

			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				meleeDelay := sim.CurrentTime + cast.CastTime
				spell.Unit.AutoAttacks.StopMeleeUntil(sim, meleeDelay)
				cleaveDelay := max(
					spell.Unit.AutoAttacks.MainhandSwingSpeed()+1,
					core.DurationFromSeconds(sim.RandomFloat("Cleave delay")*spell.Unit.AutoAttacks.MainhandSwingSpeed().Seconds()*3),
				)
				ai.Cleave.CD.Set(meleeDelay + cleaveDelay)
			},

			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: time.Minute * 1,
			},

			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		ai.Quake.CD.Set(core.DurationFromSeconds(sim.RandomFloat("Quake Timing")) + time.Second*40)
	})
}

func (ai *MagtheridonAI) registerCleave() {
	ai.Cleave = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 30619},
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		Flags:            core.SpellFlagMeleeMetrics,
		DamageMultiplier: 1.5,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: time.Second * 10,
			},

			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, tankTarget *core.Unit, spell *core.Spell) {
			baseDamage := spell.Unit.AutoAttacks.MH().EnemyWeaponDamage(sim, spell.MeleeAttackPower(), magtheridonMeleeDamageSpread)
			spell.CalcAndDealDamage(sim, tankTarget, baseDamage, spell.OutcomeEnemyMeleeWhite)
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		ai.Cleave.CD.Set(core.DurationFromSeconds(sim.RandomFloat("Cleave Timing")) + ai.Cleave.CD.Duration)
	})
}

func (ai *MagtheridonAI) Reset(sim *core.Simulation) {
	// Randomize GCD and swing timings to prevent fake APL-Haste couplings.
	ai.Target.Enable(sim)
}

func (ai *MagtheridonAI) ExecuteCustomRotation(sim *core.Simulation) {
	target := ai.Target.CurrentTarget
	if target == nil {
		// For individual non tank sims we still want abilities to work
		target = &ai.Target.Env.Raid.Parties[0].Players[0].GetCharacter().Unit
	}

	if ai.BlastNova.CanCast(sim, target) {
		ai.BlastNova.Cast(sim, target)
	}

	if ai.Quake.CanCast(sim, target) {
		ai.Quake.Cast(sim, target)
	}

	ai.Target.ExtendGCDUntil(sim, sim.CurrentTime+core.BossGCD)
}

func (ai *MagtheridonAI) TryCleave(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if !ai.Cleave.CanCast(sim, ai.MainTank) {
		return mhSwingSpell
	}

	return ai.Cleave
}
