package serpentshrinecavern

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

const hydrossMeleeDamageSpread = 0.413
const hydrossTheUnstableID int32 = 21216
const hydrossMarkInterval = time.Second * 15
const hydrossDefaultPhaseShift = 60.0

const hydrossFrostMarkSpellID int32 = 38215
const hydrossNatureMarkSpellID int32 = 38219

// Damage multipliers per stack (10% / 25% / 50% / 100% / 250% / 500%).
var hydrossMarkMultipliers = []float64{1.10, 1.25, 1.50, 2.00, 3.50, 6.00}

func addHydrossTheUnstable(raidPrefix string) {
	createHydrossPreset(raidPrefix, 25, 3_380_792, 5_974)
}

func createHydrossPreset(raidPrefix string, raidSize int32, bossHealth float64, bossMinBaseDamage float64) {
	bossName := fmt.Sprintf("Hydross the Unstable %d", raidSize)

	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: raidPrefix,

		Config: &proto.Target{
			Id:              hydrossTheUnstableID,
			Name:            bossName,
			Level:           73,
			MobType:         proto.MobType_MobTypeElemental,
			TankIndex:       0, //Main Tank (tanks Frost phase).
			SecondTankIndex: 1, //Off Tank (tanks Nature phase).

			Stats: stats.Stats{
				stats.Health:      bossHealth,
				stats.Armor:       7685,
				stats.AttackPower: 320,
			}.ToProtoArray(),

			// Starts in Frost phase; AI switches school on each phase shift.
			SpellSchool:   proto.SpellSchool_SpellSchoolFrost,
			SwingSpeed:    1.5,
			MinBaseDamage: bossMinBaseDamage,
			DamageSpread:  hydrossMeleeDamageSpread,

			ParryHaste: true,

			TargetInputs: hydrossTargetInputs(),
		},

		AI: makeHydrossAI(),
	})

	core.AddPresetEncounter(bossName, []string{
		raidPrefix + "/" + bossName,
	})
}

func hydrossTargetInputs() []*proto.TargetInput {
	return []*proto.TargetInput{
		{
			Label:       "Phase Shift Interval",
			Tooltip:     "Time (in seconds) between Frost and Nature phase shifts. Marks reset on each shift and resume stacking 15s later.",
			InputType:   proto.InputType_Number,
			NumberValue: hydrossDefaultPhaseShift,
		},
	}
}

func makeHydrossAI() core.AIFactory {
	return func() core.TargetAI {
		return &HydrossAI{}
	}
}

type HydrossAI struct {
	Target     *core.Target
	BossUnit   *core.Unit
	MainTank   *core.Unit
	SecondTank *core.Unit

	phaseShiftInterval time.Duration
	inFrostPhase       bool

	frostMarkSpell  *core.Spell
	natureMarkSpell *core.Spell

	nextPhaseShift *core.PendingAction
}

func (ai *HydrossAI) Initialize(target *core.Target, config *proto.Target) {
	ai.Target = target
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = hydrossTheUnstableID
	ai.BossUnit = &target.Unit
	ai.MainTank = ai.BossUnit.CurrentTarget
	ai.SecondTank = ai.BossUnit.SecondaryTarget

	phaseShiftSeconds := hydrossDefaultPhaseShift
	if len(config.TargetInputs) > 0 {
		phaseShiftSeconds = config.TargetInputs[0].NumberValue
	}
	ai.phaseShiftInterval = core.DurationFromSeconds(phaseShiftSeconds)

	ai.registerMarks()
}

func (ai *HydrossAI) registerMarks() {
	maxStacks := int32(len(hydrossMarkMultipliers))

	// Build a stacking mark aura on a unit. OnStacksChange multiplies/divides
	// boss damage dealt by the appropriate multiplier for the current stack count.
	registerMarkAura := func(unit *core.Unit, spellID int32, label string) *core.Aura {
		return unit.GetOrRegisterAura(core.Aura{
			Label:     label,
			ActionID:  core.ActionID{SpellID: spellID},
			Duration:  core.NeverExpires,
			MaxStacks: maxStacks,
			OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
				if oldStacks > 0 {
					ai.BossUnit.PseudoStats.DamageDealtMultiplier /= hydrossMarkMultipliers[oldStacks-1]
				}
				if newStacks > 0 {
					ai.BossUnit.PseudoStats.DamageDealtMultiplier *= hydrossMarkMultipliers[newStacks-1]
				}
			},
		})
	}

	markTargets := ai.BossUnit.Env.Raid.AllPlayerUnits

	frostMarkAuras := make([]*core.Aura, len(markTargets))
	natureMarkAuras := make([]*core.Aura, len(markTargets))
	for i, unit := range markTargets {
		frostMarkAuras[i] = registerMarkAura(unit, hydrossFrostMarkSpellID, "Mark of Hydross (Frost)")
		natureMarkAuras[i] = registerMarkAura(unit, hydrossNatureMarkSpellID, "Mark of Hydross (Nature)")
	}

	applyMarkStack := func(sim *core.Simulation, auras []*core.Aura) {
		for _, aura := range auras {
			if aura.GetStacks() < maxStacks {
				aura.Activate(sim)
				aura.AddStack(sim)
			}
		}
	}

	dropMarks := func(sim *core.Simulation, auras []*core.Aura) {
		for _, aura := range auras {
			aura.Deactivate(sim)
		}
	}

	// Frost mark spell — cast every 15s while in Frost phase.
	ai.frostMarkSpell = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: hydrossFrostMarkSpellID},
		SpellSchool: core.SpellSchoolFrost,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.BossGCD},
			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: hydrossMarkInterval,
			},
			IgnoreHaste: true,
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			applyMarkStack(sim, frostMarkAuras)
		},
	})

	// Nature mark spell — cast every 15s while in Nature phase.
	ai.natureMarkSpell = ai.BossUnit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: hydrossNatureMarkSpellID},
		SpellSchool: core.SpellSchoolNature,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{GCD: core.BossGCD},
			CD: core.Cooldown{
				Timer:    ai.BossUnit.NewTimer(),
				Duration: hydrossMarkInterval,
			},
			IgnoreHaste: true,
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			applyMarkStack(sim, natureMarkAuras)
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		dropMarks(sim, frostMarkAuras)
		dropMarks(sim, natureMarkAuras)
		// Both mark CDs start at full so first stack lands 15s into the fight.
		ai.frostMarkSpell.CD.Set(sim.CurrentTime + hydrossMarkInterval)
		ai.natureMarkSpell.CD.Set(sim.CurrentTime + hydrossMarkInterval)
		ai.inFrostPhase = true
		ai.BossUnit.AutoAttacks.MHAuto().SpellSchool = core.SpellSchoolFrost
		ai.BossUnit.CurrentTarget = ai.MainTank
		// If no main tank is assigned, suppress auto-attacks until a valid target exists.
		if ai.MainTank == nil {
			ai.BossUnit.AutoAttacks.CancelMeleeSwing(sim)
		}
		if ai.nextPhaseShift != nil {
			ai.nextPhaseShift.Cancel(sim)
			ai.nextPhaseShift = nil
		}
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		ai.schedulePhaseShift(sim, frostMarkAuras, natureMarkAuras, dropMarks)
	})
}

func (ai *HydrossAI) schedulePhaseShift(
	sim *core.Simulation,
	frostMarkAuras []*core.Aura,
	natureMarkAuras []*core.Aura,
	dropMarks func(*core.Simulation, []*core.Aura),
) {
	ai.nextPhaseShift = &core.PendingAction{
		NextActionAt: sim.CurrentTime + ai.phaseShiftInterval,
		Priority:     core.ActionPriorityAuto,
		OnAction: func(sim *core.Simulation) {
			if ai.inFrostPhase {
				// Shift to Nature: drop frost marks, switch to Nature attacks,
				// point boss at the off-tank. First nature mark lands 15s later.
				dropMarks(sim, frostMarkAuras)
				ai.BossUnit.AutoAttacks.MHAuto().SpellSchool = core.SpellSchoolNature
				ai.BossUnit.CurrentTarget = ai.SecondTank
				if ai.SecondTank != nil {
					ai.BossUnit.AutoAttacks.EnableMeleeSwing(sim)
				} else {
					ai.BossUnit.AutoAttacks.CancelMeleeSwing(sim)
				}
				ai.natureMarkSpell.CD.Set(sim.CurrentTime + hydrossMarkInterval)
				ai.inFrostPhase = false
			} else {
				// Shift to Frost: drop nature marks, switch to Frost attacks,
				// point boss at the main tank. First frost mark lands 15s later.
				dropMarks(sim, natureMarkAuras)
				ai.BossUnit.AutoAttacks.MHAuto().SpellSchool = core.SpellSchoolFrost
				ai.BossUnit.CurrentTarget = ai.MainTank
				if ai.MainTank != nil {
					ai.BossUnit.AutoAttacks.EnableMeleeSwing(sim)
				} else {
					ai.BossUnit.AutoAttacks.CancelMeleeSwing(sim)
				}
				ai.frostMarkSpell.CD.Set(sim.CurrentTime + hydrossMarkInterval)
				ai.inFrostPhase = true
			}
			ai.schedulePhaseShift(sim, frostMarkAuras, natureMarkAuras, dropMarks)
		},
	}
	sim.AddPendingAction(ai.nextPhaseShift)
}

func (ai *HydrossAI) Reset(sim *core.Simulation) {
	ai.Target.Enable(sim)
	ai.Target.PseudoStats.CanCrush = false
	ai.inFrostPhase = true
}

func (ai *HydrossAI) ExecuteCustomRotation(sim *core.Simulation) {
	// Marks apply to all players regardless of tank assignment.
	// Use the current target as the spell target; fall back to MainTank if current target is nil.
	castTarget := ai.BossUnit.CurrentTarget
	if castTarget == nil {
		castTarget = ai.MainTank
	}
	if castTarget != nil {
		if ai.inFrostPhase {
			if ai.frostMarkSpell.CanCast(sim, castTarget) {
				ai.frostMarkSpell.Cast(sim, castTarget)
			}
		} else {
			if ai.natureMarkSpell.CanCast(sim, castTarget) {
				ai.natureMarkSpell.Cast(sim, castTarget)
			}
		}
	}
	ai.Target.ExtendGCDUntil(sim, sim.CurrentTime+core.BossGCD)
}
