package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const instantImbueID = 26891
const woundImbueID = 27188
const deadlyImbueID = 27186

func (rogue *Rogue) applyPoisons() {
	rogue.applyDeadlyPoison()
	rogue.applyWoundPoison()
	rogue.applyInstantPoison()
}

func (rogue *Rogue) registerDeadlyPoisonSpell() {
	rogue.DeadlyPoison = rogue.GetOrRegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27187},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamageProc,
		ClassSpellMask: RogueSpellDeadlyPoison,
		Flags:          core.SpellFlagPoison | core.SpellFlagPassiveSpell,

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           1,
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
				if stacks := dot.GetStacks(); stacks > 0 {
					dot.SnapshotBaseDamage = 45.0 * float64(dot.GetStacks())
					at := dot.Spell.Unit.AttackTables[target.UnitIndex]
					dot.SnapshotAttackerMultiplier = dot.Spell.AttackerDamageMultiplier(at, true)
				}
			},

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
			if !result.Landed() {
				return
			}

			dot := spell.Dot(target)
			if dot.IsActive() {
				dot.Refresh(sim)
				dot.AddStack(sim)
				dot.TakeSnapshot(sim)
			} else {
				dot.Apply(sim)
				dot.SetStacks(sim, 1)
				dot.TakeSnapshot(sim)
			}
		},
	})

	rogue.ShivDeadlyPoison = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27187, Tag: 1},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamageProc,
		ClassSpellMask: RogueSpellDeadlyPoison,
		Flags:          core.SpellFlagPoison | core.SpellFlagPassiveSpell,

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           1,
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			dot := rogue.DeadlyPoison.Dot(target)
			if dot.IsActive() {
				dot.Refresh(sim)
				dot.AddStack(sim)
				dot.TakeSnapshot(sim)
			} else {
				dot.Apply(sim)
				dot.SetStacks(sim, 1)
				dot.TakeSnapshot(sim)
			}
		},
	})
}

func (rogue *Rogue) registerWoundPoisonSpell() {
	woundPoisonDebuffAura := core.Aura{
		Label:     "Wound Poison",
		ActionID:  core.ActionID{SpellID: 27189},
		Duration:  time.Second * 15,
		MaxStacks: 5,
		// Wound Healing Debuff NYI
	}

	rogue.WoundPoisonDebuffAuras = rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(woundPoisonDebuffAura)
	})

	wpBaseDamage := 65.0

	rogue.WoundPoison = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27189},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamageProc,
		ClassSpellMask: RogueSpellWoundPoison,
		Flags:          core.SpellFlagPoison | core.SpellFlagPassiveSpell,

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           rogue.DefaultSpellCritMultiplier(),
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, wpBaseDamage, spell.OutcomeMagicHitAndCrit)

			if result.Landed() {
				rogue.WoundPoisonDebuffAuras.Get(target).Activate(sim)
			}
		},
	})

	rogue.ShivWoundPoison = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27189, Tag: 1},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamageProc,
		ClassSpellMask: RogueSpellWoundPoison,
		Flags:          core.SpellFlagPoison | core.SpellFlagPassiveSpell,

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           rogue.DefaultSpellCritMultiplier(),
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, wpBaseDamage, spell.OutcomeMagicCrit)
			rogue.WoundPoisonDebuffAuras.Get(target).Activate(sim)
		},
	})
}

func (rogue *Rogue) registerInstantPoisonSpell() {
	ipBaseDamage := 146.0
	ipRange := 48

	rogue.InstantPoison = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26890},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamageProc,
		ClassSpellMask: RogueSpellInstantPoison,
		Flags:          core.SpellFlagPoison | core.SpellFlagPassiveSpell,

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           rogue.DefaultSpellCritMultiplier(),
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, ipBaseDamage+sim.RandomFloat("Instant Poison")*float64(ipRange), spell.OutcomeMagicHitAndCrit)
		},
	})

	rogue.ShivInstantPoison = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26890, Tag: 1},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamageProc,
		ClassSpellMask: RogueSpellInstantPoison,
		Flags:          core.SpellFlagPoison | core.SpellFlagPassiveSpell,

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           rogue.DefaultSpellCritMultiplier(),
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, ipBaseDamage+sim.RandomFloat("Instant Poison")*float64(ipRange), spell.OutcomeMagicCrit)
		},
	})
}

func (rogue *Rogue) getPoisonProcMask(poisonId int32) core.ProcMask {
	var mask core.ProcMask
	if rogue.Consumables.MhImbueId == poisonId {
		mask |= core.ProcMaskMeleeMH
	}
	if rogue.Consumables.OhImbueId == poisonId {
		mask |= core.ProcMaskMeleeOH
	}
	return mask
}

func (rogue *Rogue) applyDeadlyPoison() {
	procMask := rogue.getPoisonProcMask(deadlyImbueID)
	if procMask == core.ProcMaskUnknown {
		return
	}
	pph := 0.3 + 0.02*float64(rogue.Talents.ImprovedPoisons)
	rogue.deadlyPoisonPPHM = rogue.NewFixedProcChanceManager(pph, procMask)

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

func (rogue *Rogue) applyWoundPoison() {
	procMask := rogue.getPoisonProcMask(woundImbueID)
	if procMask == core.ProcMaskUnknown {
		return
	}
	pph := 0.3 + 0.02*float64(rogue.Talents.ImprovedPoisons)
	rogue.woundPoisonPPHM = rogue.NewFixedProcChanceManager(pph, procMask)

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Wound Poison",
		Outcome:            core.OutcomeLanded,
		Callback:           core.CallbackOnSpellHitDealt,
		TriggerImmediately: true,
		ProcMask:           procMask,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if rogue.woundPoisonPPHM.Proc(sim, spell.ProcMask, "Wound Poison") {
				rogue.WoundPoison.Cast(sim, result.Target)
			}
		},
	})
}

func (rogue *Rogue) applyInstantPoison() {
	procMask := rogue.getPoisonProcMask(instantImbueID)
	if procMask == core.ProcMaskUnknown {
		return
	}
	pph := 0.2 + 0.02*float64(rogue.Talents.ImprovedPoisons)
	rogue.instantPoisonPPHM = rogue.NewFixedProcChanceManager(pph, procMask)

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Instant Poison",
		Outcome:            core.OutcomeLanded,
		Callback:           core.CallbackOnSpellHitDealt,
		TriggerImmediately: true,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if rogue.instantPoisonPPHM.Proc(sim, spell.ProcMask, "Instant Poison") {
				rogue.InstantPoison.Cast(sim, result.Target)
			}
		},
	})
}
