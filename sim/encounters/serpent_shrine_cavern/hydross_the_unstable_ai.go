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

// Damage bonuses per stack (+10% / +25% / +50% / +100% / +250% / +500%).
var hydrossMarkDamageBonuses = []float64{0.10, 0.25, 0.50, 1.00, 2.50, 5.00}

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

	markDamageMod *core.SpellMod

	frostMarkAuras  core.AuraArray
	frostMarkSpell  *core.Spell
	frostMeleeSpell *core.Spell

	natureMarkAuras  core.AuraArray
	natureMarkSpell  *core.Spell
	natureMeleeSpell *core.Spell

	nextPhaseShift *core.PendingAction
}

func (ai *HydrossAI) Initialize(target *core.Target, config *proto.Target) {
	ai.Target = target
	ai.Target.AutoAttacks.MHConfig().ActionID.Tag = hydrossTheUnstableID
	ai.BossUnit = &target.Unit
	ai.MainTank = ai.BossUnit.CurrentTarget
	ai.SecondTank = ai.BossUnit.SecondaryTarget

	ai.frostMeleeSpell = ai.BossUnit.GetOrRegisterSpell(*ai.BossUnit.AutoAttacks.MHConfig())

	natureMeleeSpellConfig := *ai.BossUnit.AutoAttacks.MHConfig()
	natureMeleeSpellConfig.ActionID.Tag = hydrossTheUnstableID + 1
	natureMeleeSpellConfig.SpellSchool = core.SpellSchoolNature

	ai.natureMeleeSpell = ai.BossUnit.GetOrRegisterSpell(natureMeleeSpellConfig)

	phaseShiftSeconds := hydrossDefaultPhaseShift
	if len(config.TargetInputs) > 0 {
		phaseShiftSeconds = config.TargetInputs[0].NumberValue
	}
	ai.phaseShiftInterval = core.DurationFromSeconds(phaseShiftSeconds)

	ai.registerMarks()
}

func (ai *HydrossAI) registerMarks() {
	// Build a stacking mark aura on a unit. OnStacksChange multiplies/divides
	// boss damage dealt by the appropriate multiplier for the current stack count.
	registerMarkAura := func(unit *core.Unit, spellID int32, spellSchool core.SpellSchool, label string) *core.Aura {
		var markDamageMod = ai.BossUnit.AddDynamicMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Pct,
			School:     spellSchool,
			FloatValue: 0,
		})
		return unit.GetOrRegisterAura(core.Aura{
			Label:     label,
			ActionID:  core.ActionID{SpellID: spellID},
			Duration:  time.Second * 30,
			MaxStacks: int32(len(hydrossMarkDamageBonuses)),
			OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
				if newStacks == 0 {
					markDamageMod.Deactivate()
					return
				}
				markDamageMod.UpdateFloatValue(hydrossMarkDamageBonuses[newStacks-1])
				markDamageMod.Activate()
			},
		})
	}

	ai.frostMarkAuras = ai.BossUnit.NewAllyAuraArray(func(allyUnit *core.Unit) *core.Aura {
		return registerMarkAura(allyUnit, hydrossFrostMarkSpellID, core.SpellSchoolFrost, "Mark of Hydross")
	})

	ai.natureMarkAuras = ai.BossUnit.NewAllyAuraArray(func(allyUnit *core.Unit) *core.Aura {
		return registerMarkAura(allyUnit, hydrossNatureMarkSpellID, core.SpellSchoolNature, "Mark of Corruption")
	})

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
			ai.applyMarkStack(sim)
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
			ai.applyMarkStack(sim)
		},
	})

	ai.BossUnit.RegisterResetEffect(func(sim *core.Simulation) {
		ai.frostMarkAuras.DeactivateAll(sim)
		ai.natureMarkAuras.DeactivateAll(sim)

		// Both mark CDs start at full so first stack lands 15s into the fight.
		ai.frostMarkSpell.CD.Set(sim.CurrentTime + hydrossMarkInterval)
		ai.natureMarkSpell.CD.Set(sim.CurrentTime + hydrossMarkInterval)
		ai.inFrostPhase = true
		ai.BossUnit.AutoAttacks.SetMHSpell(ai.frostMeleeSpell)
		ai.BossUnit.CurrentTarget = ai.MainTank
		// If no main tank is assigned, suppress auto-attacks until a valid target exists.
		if ai.MainTank == nil {
			ai.BossUnit.AutoAttacks.CancelMeleeSwing(sim)
		}
		if ai.nextPhaseShift != nil {
			ai.nextPhaseShift.Cancel(sim)
			ai.nextPhaseShift = nil
		}

		ai.schedulePhaseShift(sim)
	})
}

func (ai *HydrossAI) schedulePhaseShift(
	sim *core.Simulation,
) {
	ai.nextPhaseShift = &core.PendingAction{
		NextActionAt: sim.CurrentTime + ai.phaseShiftInterval,
		Priority:     core.ActionPriorityAuto,
		OnAction: func(sim *core.Simulation) {
			if ai.inFrostPhase {
				// Shift to Nature: drop frost marks, switch to Nature attacks,
				// point boss at the off-tank. First nature mark lands 15s later.
				ai.BossUnit.AutoAttacks.SetMHSpell(ai.natureMeleeSpell)
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
				ai.BossUnit.AutoAttacks.SetMHSpell(ai.frostMeleeSpell)
				ai.BossUnit.CurrentTarget = ai.MainTank
				if ai.MainTank != nil {
					ai.BossUnit.AutoAttacks.EnableMeleeSwing(sim)
				} else {
					ai.BossUnit.AutoAttacks.CancelMeleeSwing(sim)
				}
				ai.frostMarkSpell.CD.Set(sim.CurrentTime + hydrossMarkInterval)
				ai.inFrostPhase = true
			}
			ai.schedulePhaseShift(sim)
		},
	}
	sim.AddPendingAction(ai.nextPhaseShift)
}

func (ai *HydrossAI) applyMarkStack(sim *core.Simulation) {
	auras := core.Ternary(ai.inFrostPhase, ai.frostMarkAuras, ai.natureMarkAuras)
	for _, aura := range auras {
		if aura == nil {
			continue
		}
		if aura.GetStacks() < aura.MaxStacks {
			aura.Activate(sim)
			aura.AddStack(sim)
		}
	}
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
