package serpentshrinecavern

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

const hydrossMeleeDamageSpread = 0.293
const hydrossTheUnstableID int32 = 21216
const hydrossMarkInterval = time.Second * 15

// Each stack replaces the previous — these are the absolute multipliers on base damage.
// 10% / 25% / 50% / 100% / 250% / 500% increased damage = ×1.10 / ×1.25 / ×1.50 / ×2.00 / ×3.50 / ×6.00
var hydrossFrostMarkSpellIDs = []int32{38215, 38216, 38217, 38218, 38231, 40584}
var hydrossNatureMarkSpellIDs = []int32{38219, 38220, 38221, 38222, 38230, 40583}
var hydrossMarkMultipliers = []float64{1.10, 1.25, 1.50, 2.00, 3.50, 6.00}

func addHydrossTheUnstable(raidPrefix string) {
	createHydrossPreset(raidPrefix, 25, 3_380_792, 7_035)
}

func createHydrossPreset(raidPrefix string, raidSize int32, bossHealth float64, bossMinBaseDamage float64) {
	bossName := fmt.Sprintf("Hydross the Unstable %d", raidSize)

	// Frost form
	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: raidPrefix,

		Config: &proto.Target{
			Id:        hydrossTheUnstableID,
			Name:      bossName + " (Frost)",
			Level:     73,
			MobType:   proto.MobType_MobTypeElemental,
			TankIndex: 0,

			Stats: stats.Stats{
				stats.Health:      bossHealth,
				stats.Armor:       7685,
				stats.AttackPower: 320,
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolFrost,
			SwingSpeed:    1.5,
			MinBaseDamage: bossMinBaseDamage,
			DamageSpread:  hydrossMeleeDamageSpread,

			ParryHaste: true,
		},

		AI: makeHydrossAI(hydrossFrostMarkSpellIDs),
	})

	// Nature form
	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: raidPrefix,

		Config: &proto.Target{
			Id:        hydrossTheUnstableID,
			Name:      bossName + " (Nature)",
			Level:     73,
			MobType:   proto.MobType_MobTypeElemental,
			TankIndex: 0,

			Stats: stats.Stats{
				stats.Health:      bossHealth,
				stats.Armor:       7685,
				stats.AttackPower: 320,
			}.ToProtoArray(),

			SpellSchool:   proto.SpellSchool_SpellSchoolNature,
			SwingSpeed:    1.5,
			MinBaseDamage: bossMinBaseDamage,
			DamageSpread:  hydrossMeleeDamageSpread,

			ParryHaste: true,
		},

		AI: makeHydrossAI(hydrossNatureMarkSpellIDs),
	})

	core.AddPresetEncounter(bossName+" (Frost)", []string{
		raidPrefix + "/" + bossName + " (Frost)",
	})
	core.AddPresetEncounter(bossName+" (Nature)", []string{
		raidPrefix + "/" + bossName + " (Nature)",
	})
}

func makeHydrossAI(markSpellIDs []int32) core.AIFactory {
	return func() core.TargetAI {
		return &HydrossAI{markSpellIDs: markSpellIDs}
	}
}

type HydrossAI struct {
	Target       *core.Target
	BossUnit     *core.Unit
	MainTank     *core.Unit
	markSpellIDs []int32
	markSpell    *core.Spell
}

func (ai *HydrossAI) Initialize(target *core.Target, config *proto.Target) {
	ai.Target = target
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = hydrossTheUnstableID
	ai.BossUnit = &target.Unit
	ai.MainTank = ai.BossUnit.CurrentTarget

	ai.registerMark()
}

func (ai *HydrossAI) registerMark() {
	currentStack := 0

	// Each aura only applies/removes its own multiplier. The mark spell
	// explicitly deactivates the previous stack before activating the next,
	// so OnExpire cleans up before OnGain runs — no double-division.
	markAuras := make([]*core.Aura, len(hydrossMarkMultipliers))
	for i := range hydrossMarkMultipliers {
		stack := i // capture
		markAuras[i] = ai.MainTank.GetOrRegisterAura(core.Aura{
			Label:    fmt.Sprintf("Mark of Hydross - Stack %d", stack+1),
			ActionID: core.ActionID{SpellID: ai.markSpellIDs[stack]},
			Duration: core.NeverExpires,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				ai.BossUnit.PseudoStats.DamageDealtMultiplier *= hydrossMarkMultipliers[stack]
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				ai.BossUnit.PseudoStats.DamageDealtMultiplier /= hydrossMarkMultipliers[stack]
			},
		})
	}

	// The spell is registered once with stack-1's SpellID. Updating ActionID per
	// cast would require mutating spell.ActionID before Cast() is called (the log
	// fires before ApplyEffects), which would mean promoting currentStack to a
	// HydrossAI field. The cosmetic benefit doesn't justify the refactor.
	ai.markSpell = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: ai.markSpellIDs[0]},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.BossGCD,
			},
			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: hydrossMarkInterval,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Deactivate previous stack aura if any
			if currentStack > 0 {
				markAuras[currentStack-1].Deactivate(sim)
			}
			if currentStack < len(markAuras) {
				markAuras[currentStack].Activate(sim)
				currentStack++
			}
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		if currentStack > 0 {
			markAuras[currentStack-1].Deactivate(sim)
		}
		currentStack = 0
		ai.markSpell.CD.Set(hydrossMarkInterval)
	})
}

func (ai *HydrossAI) Reset(sim *core.Simulation) {
	ai.Target.Enable(sim)
	ai.Target.PseudoStats.CanCrush = false
}

func (ai *HydrossAI) ExecuteCustomRotation(sim *core.Simulation) {
	if ai.markSpell.CanCast(sim, ai.MainTank) {
		ai.markSpell.Cast(sim, ai.MainTank)
	}
	ai.Target.ExtendGCDUntil(sim, sim.CurrentTime+core.BossGCD)
}
