package serpentshrinecavern

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

const morogrimMeleeDamageSpread = 0.413
const morogrimTidewalkerID int32 = 21213

func addMorogrimTidewalker(raidPrefix string) {
	createMorogrimPreset(raidPrefix, 25, 5_691_000, 12_744)
}

func createMorogrimPreset(raidPrefix string, raidSize int32, bossHealth float64, bossMinBaseDamage float64) {
	bossName := fmt.Sprintf("Morogrim Tidewalker %d", raidSize)

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: raidPrefix,

		Config: &proto.Target{
			Id:        morogrimTidewalkerID,
			Name:      bossName,
			Level:     73,
			MobType:   proto.MobType_MobTypeGiant,
			TankIndex: 0,

			Stats: stats.Stats{
				stats.Health:      bossHealth,
				stats.Armor:       7685,
				stats.AttackPower: 320,
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:    1.6,
			MinBaseDamage: bossMinBaseDamage,
			DamageSpread:  morogrimMeleeDamageSpread,

			ParryHaste: true,

			TargetInputs: morogrimTargetInputs(),
		},

		AI: makeMorogrimAI(),
	})

	core.AddPresetEncounter(bossName, []string{
		raidPrefix + "/" + bossName,
	})
}

func morogrimTargetInputs() []*proto.TargetInput {
	return []*proto.TargetInput{
		{
			Label:     "Disable Tidal Wave 400% Attack Speed Slow",
			Tooltip:   "This will disable the attack speed slow from Tidal Wave.",
			InputType: proto.InputType_Bool,
			BoolValue: false,
		},
	}
}

func makeMorogrimAI() core.AIFactory {
	return func() core.TargetAI {
		return &MorogrimAI{}
	}
}

type MorogrimAI struct {
	// Unit references
	Target   *core.Target
	BossUnit *core.Unit
	MainTank *core.Unit

	// Spell + aura references
	Thrash        *core.Spell
	Earthquake    *core.Spell
	TidalWave     *core.Spell
	TidalWaveAura *core.Aura
}

func (ai *MorogrimAI) Initialize(target *core.Target, config *proto.Target) {
	// Save unit references
	ai.Target = target
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = morogrimTidewalkerID

	ai.BossUnit = &target.Unit

	ai.MainTank = ai.BossUnit.CurrentTarget

	// Register relevant spells and auras
	ai.registerTidalWave(config.TargetInputs[0].BoolValue)
	ai.registerEarthquake()
	ai.registerThrash()
}

func (ai *MorogrimAI) registerTidalWave(disableSlow bool) {
	duration := time.Second * 15

	ai.TidalWaveAura = ai.MainTank.RegisterAura(core.Aura{
		Label:    "Tidal Wave",
		ActionID: core.ActionID{SpellID: 37730},
		Duration: duration,
	})

	if !disableSlow {
		ai.TidalWaveAura.AttachMultiplyMeleeSpeed(1.0 / 4.0)
	}

	rollTidalWaveCD := func(sim *core.Simulation) time.Duration {
		// The median across 100 logs is ~30s, with a minimum of ~22s.
		return duration + core.DurationFromSeconds(sim.Roll(10, 30))
	}

	ai.TidalWave = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 37730},
		SpellSchool: core.SpellSchoolFrost,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagBinary | core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.BossGCD,
				CastTime: time.Second * 2,
			},

			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: duration,
			},

			IgnoreHaste: true,
		},

		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, sim.Roll(3938, 5062), spell.OutcomeMagicHit)
			ai.TidalWaveAura.Activate(sim)
			spell.CD.Set(sim.CurrentTime + rollTidalWaveCD(sim))
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		// The median across 100 logs is ~16s, with a minimum of ~11s.
		ai.TidalWave.CD.Set(core.DurationFromSeconds(12 + sim.Roll(0, 8)))
	})
}

func (ai *MorogrimAI) registerEarthquake() {
	cooldown := time.Second * 45
	rollEarthquakeCD := func(sim *core.Simulation) time.Duration {
		return cooldown + core.DurationFromSeconds(sim.Roll(0, 15))
	}

	ai.Earthquake = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 37764},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskSpellDamage,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.BossGCD,
			},

			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: cooldown,
			},

			IgnoreHaste: true,
		},

		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, sim.Roll(4000, 4200), spell.OutcomeMagicHit)
			ai.Earthquake.CD.Set(sim.CurrentTime + rollEarthquakeCD(sim))
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		ai.Earthquake.CD.Set(rollEarthquakeCD(sim))
	})
}

func (ai *MorogrimAI) registerThrash() {
	var thrashSpell *core.Spell

	procTrigger := ai.Target.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Thrash - Listener",
		ProcChance:         0.5,
		ICD:                time.Second * 3,
		ProcMask:           core.ProcMaskMeleeWhiteHit,
		TriggerImmediately: true,
		Callback:           core.CallbackOnSpellHitDealt,

		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			thrashSpell.Cast(sim, result.Target)
		},
	})

	procTrigger.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
		config := *ai.Target.AutoAttacks.MHConfig()
		config.ActionID = config.ActionID.WithTag(morogrimTidewalkerID + 18943)
		config.Flags |= core.SpellFlagPassiveSpell
		thrashSpell = ai.Target.GetOrRegisterSpell(config)
	})
}

func (ai *MorogrimAI) Reset(sim *core.Simulation) {
	// Randomize GCD and swing timings to prevent fake APL-Haste couplings.
	ai.Target.Enable(sim)
}

func (ai *MorogrimAI) ExecuteCustomRotation(sim *core.Simulation) {
	target := ai.Target.CurrentTarget
	if target == nil {
		// For individual non tank sims we still want abilities to work
		target = &ai.Target.Env.Raid.Parties[0].Players[0].GetCharacter().Unit
	}

	if ai.TidalWave.CanCast(sim, target) {
		ai.TidalWave.Cast(sim, target)
	}

	if ai.Earthquake.CanCast(sim, target) {
		ai.Earthquake.Cast(sim, target)
	}

	ai.Target.ExtendGCDUntil(sim, sim.CurrentTime+core.BossGCD)
}
