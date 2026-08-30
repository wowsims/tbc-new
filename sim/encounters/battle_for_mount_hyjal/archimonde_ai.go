package battleformounthyjal

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

const archimondeName = "Archimonde"
const archimondeMeleeDamageSpread = 0.413
const archimondeID int32 = 17968

const archimondeFearSpellID int32 = 31970
const archimondeFearCastTime = time.Millisecond * 1500

// Fear lasts 8s, but is usually broken well before that.
const archimondeDefaultFearDuration = 3.0

func addArchimonde(raidPrefix string) {
	createArchimondePreset(raidPrefix, 25, 4_552_500, 20_304)
}

func createArchimondePreset(raidPrefix string, raidSize int32, bossHealth float64, bossMinBaseDamage float64) {
	bossName := fmt.Sprintf("%s %d", archimondeName, raidSize)

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: raidPrefix,

		Config: &proto.Target{
			Id:        archimondeID,
			Name:      bossName,
			Level:     73,
			MobType:   proto.MobType_MobTypeDemon,
			TankIndex: 0,

			Stats: stats.Stats{
				stats.Health:      bossHealth,
				stats.Armor:       6193,
				stats.AttackPower: 320,
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:    1.5,
			MinBaseDamage: bossMinBaseDamage,
			DamageSpread:  archimondeMeleeDamageSpread,

			ParryHaste: true,
			CanCrush:   false,

			TargetInputs: archimondeTargetInputs(),
		},

		AI: makeArchimondeAI(),
	})

	core.AddPresetEncounter(bossName, []string{
		raidPrefix + "/" + bossName,
	})
}

func archimondeTargetInputs() []*proto.TargetInput {
	return []*proto.TargetInput{
		{
			Label:       "Fear Duration",
			Tooltip:     "Time (in seconds) that Fear keeps players from acting. Fear lasts 8s, but is usually broken after ~3s. Set to 0 to disable.",
			InputType:   proto.InputType_Number,
			NumberValue: archimondeDefaultFearDuration,
		},
	}
}

func makeArchimondeAI() core.AIFactory {
	return func() core.TargetAI {
		return &ArchimondeAI{}
	}
}

// Archimonde's other abilities (Air Burst, Doomfire) do not really hit the tank,
// so only his melee and Fear are modeled.
type ArchimondeAI struct {
	Target   *core.Target
	BossUnit *core.Unit

	Fear      *core.Spell
	FearAuras core.AuraArray
}

func (ai *ArchimondeAI) Initialize(target *core.Target, config *proto.Target) {
	ai.Target = target
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = archimondeID

	ai.BossUnit = &target.Unit

	ai.registerFear(config.TargetInputs[0].NumberValue)
}

func (ai *ArchimondeAI) registerFear(fearDurationSeconds float64) {
	cooldown := time.Second * 40
	rollFearCD := func(sim *core.Simulation) time.Duration {
		return cooldown + core.DurationFromSeconds(sim.Roll(0, 13))
	}

	fearDuration := core.DurationFromSeconds(fearDurationSeconds)

	if fearDuration > 0 {
		ai.FearAuras = ai.BossUnit.NewAllyAuraArray(func(allyUnit *core.Unit) *core.Aura {
			return allyUnit.RegisterFearAura(
				fmt.Sprintf("Fear (%s)", archimondeName),
				core.ActionID{SpellID: archimondeFearSpellID},
				fearDuration,
			)
		})
	}

	ai.Fear = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: archimondeFearSpellID},
		SpellSchool: core.SpellSchoolShadow,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.BossGCD,
				CastTime: archimondeFearCastTime,
			},
			IgnoreHaste: true,

			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				// Archimonde does not swing while casting Fear.
				spell.Unit.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+cast.CastTime)
			},

			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: cooldown,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			ai.FearAuras.ApplyFearToAllUnits(sim)
			spell.CD.Set(sim.CurrentTime + rollFearCD(sim))
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		ai.Fear.CD.Set(rollFearCD(sim))
	})
}

func (ai *ArchimondeAI) Reset(sim *core.Simulation) {
	// Randomize GCD and swing timings to prevent fake APL-Haste couplings.
	ai.Target.Enable(sim)
}

func (ai *ArchimondeAI) ExecuteCustomRotation(sim *core.Simulation) {
	target := ai.Target.CurrentTarget
	if target == nil {
		// For individual non tank sims we still want abilities to work
		target = &ai.Target.Env.Raid.Parties[0].Players[0].GetCharacter().Unit
	}

	if ai.Fear.CanCast(sim, target) {
		ai.Fear.Cast(sim, target)
	}

	ai.Target.ExtendGCDUntil(sim, sim.CurrentTime+core.BossGCD)
}
