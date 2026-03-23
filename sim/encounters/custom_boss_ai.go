package encounters

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

const bossTankID int32 = 100000

func addCustomBossAI() {
	createCustomBossPreset()
}

func createCustomBossPreset() {
	bossName := "Boss"

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: "Default",

		Config: &proto.Target{
			Id:        bossTankID,
			Name:      bossName,
			Level:     73,
			MobType:   proto.MobType_MobTypeMechanical,
			TankIndex: 0,

			Stats: stats.Stats{
				stats.Armor:       7685,
				stats.AttackPower: 320,
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:    2.0,
			MinBaseDamage: 15113,
			DamageSpread:  0.413,
			ParryHaste:    true,
			TargetInputs:  defaultTankTargetInputs(),
		},

		AI: makeDefaultTankAI(),
	})
	core.AddPresetEncounter("Custom Boss", []string{
		"Default/" + bossName,
	})

}

func defaultTankTargetInputs() []*proto.TargetInput {
	return []*proto.TargetInput{
		{
			Label:     "Has Extra Attacks",
			Tooltip:   "Boss has extra attacks (thrash)",
			InputType: proto.InputType_Bool,
			BoolValue: false,
		},
		{
			Label:       "Extra attack count",
			Tooltip:     "Number of extra attacks (thrash) the boss casts per proc",
			InputType:   proto.InputType_Number,
			NumberValue: 1,
		},
		{
			Label:       "Extra attack proc chance %",
			Tooltip:     "Chance, in percent, for the boss to cast extra attacks (thrash). (ie. 25 = 25% chance to proc on each melee swing)",
			InputType:   proto.InputType_Number,
			NumberValue: 25,
		},
		{
			Label:       "Extra attack cooldown",
			Tooltip:     "Internal cooldown of thrash, in seconds",
			InputType:   proto.InputType_Number,
			NumberValue: 5,
		},
		{
			Label:     "Has Cleave",
			Tooltip:   "Boss has cleave attack",
			InputType: proto.InputType_Bool,
			BoolValue: false,
		},
		{
			Label:       "Cleave cooldown",
			Tooltip:     "Internal cooldown, in seconds, for the boss to cast cleave attack",
			InputType:   proto.InputType_Number,
			NumberValue: 10,
		},
		{
			Label:       "Cleave damage multiplier",
			Tooltip:     "Multiplier for the damage of cleave attacks. (ie. 50 = 50% of min base damage)",
			InputType:   proto.InputType_Number,
			NumberValue: 50,
		},
		{
			Label:     "Has Magic spell",
			Tooltip:   "Boss can cast a magic spell",
			InputType: proto.InputType_Bool,
			BoolValue: false,
		},
		{
			Label:       "Magic spell cast time",
			Tooltip:     "Cast time, in seconds, for the boss to cast the magic spell",
			InputType:   proto.InputType_Number,
			NumberValue: 2,
		},
		{
			Label:       "Magic spell cooldown",
			Tooltip:     "Internal cooldown of the Magic spell, in seconds",
			InputType:   proto.InputType_Number,
			NumberValue: 25,
		},
		{
			Label:       "Magic spell min damage",
			Tooltip:     "Minimum damage of the magic spell cast by the boss",
			InputType:   proto.InputType_Number,
			NumberValue: 3000,
		},
		{
			Label:       "Magic spell max damage",
			Tooltip:     "Maximum damage of the magic spell cast by the boss",
			InputType:   proto.InputType_Number,
			NumberValue: 5000,
		},
		{
			Label:     "Magic Spell School",
			Tooltip:   "Spell school of the magic spell cast by the boss",
			InputType: proto.InputType_Enum,
			EnumValue: 0,
			EnumOptions: []string{
				"Arcane",
				"Fire",
				"Frost",
				"Holy",
				"Nature",
				"Shadow",
			},
		},
	}
}

func makeDefaultTankAI() core.AIFactory {
	return func() core.TargetAI {
		return &DefaultTankAI{}
	}
}

type DefaultTankAIConfig struct {
	HasExtraAttacks       bool
	ExtraAttackCount      int32
	ExtraAttackICD        time.Duration
	ExtraAttackProcChance float64

	HasCleave        bool
	CleaveICD        time.Duration
	CleaveDamageMult float64

	HasMagicSpell       bool
	MagicSpellCastTime  time.Duration
	MagicSpellICD       time.Duration
	MagicSpellMinDamage float64
	MagicSpellMaxDamage float64
	MagicSpellSchool    int32
}

type DefaultTankAI struct {
	Target   *core.Target
	BossUnit *core.Unit
	MainTank *core.Unit

	Config DefaultTankAIConfig

	Cleave     *core.Spell
	MagicSpell *core.Spell
}

func (ai *DefaultTankAI) Initialize(target *core.Target, config *proto.Target) {
	ai.Target = target
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = bossTankID

	ai.BossUnit = &target.Unit
	ai.MainTank = ai.BossUnit.CurrentTarget

	ai.Config = DefaultTankAIConfig{
		HasExtraAttacks:       config.TargetInputs[0].BoolValue,
		ExtraAttackCount:      int32(config.TargetInputs[1].NumberValue),
		ExtraAttackProcChance: config.TargetInputs[2].NumberValue / 100,
		ExtraAttackICD:        time.Duration(config.TargetInputs[3].NumberValue) * time.Second,

		HasCleave:        config.TargetInputs[4].BoolValue,
		CleaveICD:        time.Duration(config.TargetInputs[5].NumberValue) * time.Second,
		CleaveDamageMult: config.TargetInputs[6].NumberValue / 100,

		HasMagicSpell:       config.TargetInputs[7].BoolValue,
		MagicSpellCastTime:  time.Duration(config.TargetInputs[8].NumberValue) * time.Second,
		MagicSpellICD:       time.Duration(config.TargetInputs[9].NumberValue) * time.Second,
		MagicSpellMinDamage: config.TargetInputs[10].NumberValue,
		MagicSpellMaxDamage: config.TargetInputs[11].NumberValue,
		MagicSpellSchool:    config.TargetInputs[12].EnumValue,
	}

	if ai.Config.HasExtraAttacks {
		ai.registerThrash()
	}

	if ai.Config.HasCleave {
		ai.registerCleave()
		ai.BossUnit.AutoAttacks.SetReplaceMHSwing(ai.TryCleave)
	}

	if ai.Config.HasMagicSpell {
		ai.registerMagicSpell()
	}
}

func (ai *DefaultTankAI) registerThrash() {
	var thrashSpell *core.Spell

	procTrigger := ai.BossUnit.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Thrash - Listener",
		ProcChance:         ai.Config.ExtraAttackProcChance,
		ICD:                ai.Config.ExtraAttackICD,
		ProcMask:           core.ProcMaskMeleeWhiteHit,
		TriggerImmediately: true,
		Callback:           core.CallbackOnSpellHitDealt,

		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			for range ai.Config.ExtraAttackCount {
				thrashSpell.Cast(sim, result.Target)
			}
		},
	})

	procTrigger.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
		config := *ai.BossUnit.AutoAttacks.MHConfig()
		config.ActionID = config.ActionID.WithTag(bossTankID + 1)
		config.Flags |= core.SpellFlagPassiveSpell
		thrashSpell = ai.BossUnit.GetOrRegisterSpell(config)
	})
}

func (ai *DefaultTankAI) registerCleave() {
	ai.Cleave = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 30619},
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		Flags:            core.SpellFlagMeleeMetrics,
		DamageMultiplier: 1 + ai.Config.CleaveDamageMult,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: ai.Config.CleaveICD,
			},

			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, tankTarget *core.Unit, spell *core.Spell) {
			baseDamage := spell.Unit.AutoAttacks.MH().EnemyWeaponDamage(sim, spell.MeleeAttackPower(tankTarget), ai.Target.PseudoStats.DamageSpread)
			spell.CalcAndDealDamage(sim, tankTarget, baseDamage, spell.OutcomeEnemyMeleeWhite)
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		ai.Cleave.CD.Set(core.DurationFromSeconds(sim.RandomFloat("Cleave Timing")) + ai.Cleave.CD.Duration)
	})
}

func (ai *DefaultTankAI) registerMagicSpell() {
	spellSchool := core.SpellSchoolArcane
	switch ai.Config.MagicSpellSchool {
	case 0:
		spellSchool = core.SpellSchoolArcane
	case 1:
		spellSchool = core.SpellSchoolFire
	case 2:
		spellSchool = core.SpellSchoolFrost
	case 3:
		spellSchool = core.SpellSchoolHoly
	case 4:
		spellSchool = core.SpellSchoolNature
	case 5:
		spellSchool = core.SpellSchoolShadow
	}

	ai.MagicSpell = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 18282}.WithTag(bossTankID + ai.Config.MagicSpellSchool),
		SpellSchool: spellSchool,

		ProcMask: core.ProcMaskEmpty,
		Flags:    core.SpellFlagBinary | core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.BossGCD,
				CastTime: ai.Config.MagicSpellCastTime,
			},

			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				meleeDelay := sim.CurrentTime + cast.CastTime
				spell.Unit.AutoAttacks.StopMeleeUntil(sim, meleeDelay)
				if ai.Config.HasCleave {
					cleaveDelay := max(
						spell.Unit.AutoAttacks.MainhandSwingSpeed()+1,
						core.DurationFromSeconds(sim.RandomFloat("Cleave delay")*spell.Unit.AutoAttacks.MainhandSwingSpeed().Seconds()*3),
					)
					ai.Cleave.CD.Set(meleeDelay + cleaveDelay)
				}
			},

			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: ai.Config.MagicSpellICD,
			},

			IgnoreHaste: true,
		},

		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.RollWithLabel(ai.Config.MagicSpellMinDamage, ai.Config.MagicSpellMaxDamage, "Dummy Spell")
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeAlwaysHit)
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		ai.MagicSpell.CD.Set(core.DurationFromSeconds(sim.RandomFloat("Magic Spell Timing")) + ai.Config.MagicSpellICD)
	})
}

func (ai *DefaultTankAI) Reset(sim *core.Simulation) {
	ai.BossUnit.AutoAttacks.RandomizeMeleeTiming(sim)
}

func (ai *DefaultTankAI) ExecuteCustomRotation(sim *core.Simulation) {
	target := ai.BossUnit.CurrentTarget
	if target == nil {
		// For individual non tank sims we still want abilities to work
		target = &ai.Target.Env.Raid.Parties[0].Players[0].GetCharacter().Unit
	}

	if ai.Config.HasMagicSpell && ai.MagicSpell.CanCast(sim, target) {
		ai.MagicSpell.Cast(sim, target)
	}

	ai.BossUnit.ExtendGCDUntil(sim, sim.CurrentTime+core.BossGCD)
}

func (ai *DefaultTankAI) TryCleave(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if !ai.Cleave.CanCast(sim, ai.MainTank) {
		return mhSwingSpell
	}

	return ai.Cleave
}
