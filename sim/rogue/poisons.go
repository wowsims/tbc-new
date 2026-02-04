package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (rogue *Rogue) applyPoisons() {
	rogue.applyDeadlyPoison()
	rogue.applyWoundPoison()
	rogue.applyInstantPoison()
}

func (rogue *Rogue) registerDeadlyPoisonSpell() {
	pph := 0.3 + 0.02*float64(rogue.Talents.ImprovedPoisons)
	rogue.deadlyPoisonPPHM = rogue.NewFixedProcChanceManager(pph, core.ProcMaskMelee)

	rogue.DeadlyPoison = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27187},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamageProc,
		ClassSpellMask: RogueSpellDeadlyPoison,
		Flags:          core.SpellFlagPassiveSpell,

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           0,
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:     "Deadly Poison",
				MaxStacks: 5,
				Duration:  time.Second * 12,
			},
			NumberOfTicks: 4,
			TickLength:    time.Second * 3,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.SnapshotBaseDamage = 45
				at := dot.Spell.Unit.AttackTables[target.UnitIndex]
				dot.SnapshotAttackerMultiplier = dot.Spell.AttackerDamageMultiplier(at, true)
			},

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialNoBlockDodgeParry)
			if !result.Landed() {
				return
			}

			dot := spell.Dot(target)
			if dot.IsActive() {
				dot.AddStack(sim)
				dot.Refresh(sim)
			} else {
				dot.Apply(sim)
				dot.Refresh(sim)
			}
		},
	})
}

func (rogue *Rogue) registerWoundPoisonSpell() {
	pph := 0.3 + 0.02*float64(rogue.Talents.ImprovedPoisons)
	rogue.woundPoisonPPHM = rogue.NewFixedProcChanceManager(pph, core.ProcMaskMelee)

	woundPoisonDebuffAura := core.Aura{
		Label:     "Wound Poison",
		ActionID:  core.ActionID{SpellID: 8680},
		Duration:  time.Second * 15,
		MaxStacks: 5,
		// Wound Healing Debuff NYI
	}

	rogue.WoundPoisonDebuffAuras = rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(woundPoisonDebuffAura)
	})

	wpBaseDamage := 65.0

	rogue.WoundPoison = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 8680},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamageProc,
		ClassSpellMask: RogueSpellWoundPoison,

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           rogue.DefaultSpellCritMultiplier(),
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, wpBaseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if result.Landed() {
				rogue.WoundPoisonDebuffAuras.Get(target).Activate(sim)
			}
		},
	})
}

func (rogue *Rogue) registerInstantPoisonSpell() {
	pph := 0.2 + 0.02*float64(rogue.Talents.ImprovedPoisons)
	rogue.instantPoisonPPHM = rogue.NewFixedProcChanceManager(pph, core.ProcMaskMelee)
	ipBaseDamage := 146.0

	rogue.WoundPoison = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26890},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamageProc,
		ClassSpellMask: RogueSpellWoundPoison,

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           rogue.DefaultSpellCritMultiplier(),
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, ipBaseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if result.Landed() {
				rogue.WoundPoisonDebuffAuras.Get(target).Activate(sim)
			}
		},
	})
}

func (rogue *Rogue) applyDeadlyPoison() {
	if rogue.Options.LethalPoison == proto.RogueOptions_DeadlyPoison {
		rogue.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Deadly Poison",
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			TriggerImmediately: true,

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if rogue.deadlyPoisonPPHM.Proc(sim, spell.ProcMask, "Deadly Poison") {
					rogue.DeadlyPoison.Cast(sim, result.Target)
				}
			},
		})
	}
}

func (rogue *Rogue) applyWoundPoison() {
	if rogue.Options.LethalPoison == proto.RogueOptions_WoundPoison {
		rogue.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Wound Poison",
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			TriggerImmediately: true,

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if rogue.woundPoisonPPHM.Proc(sim, spell.ProcMask, "Wound Poison") {
					rogue.WoundPoison.Cast(sim, result.Target)
				}
			},
		})
	}
}

func (rogue *Rogue) applyInstantPoison() {
	if rogue.Options.LethalPoison == proto.RogueOptions_WoundPoison {
		rogue.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Instant Poison",
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			TriggerImmediately: true,

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if rogue.instantPoisonPPHM.Proc(sim, spell.ProcMask, "Instant Poison") {
					rogue.WoundPoison.Cast(sim, result.Target)
				}
			},
		})
	}
}
