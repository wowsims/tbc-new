package warrior

import (
	"github.com/wowsims/tbc/sim/core"
)

func (war *Warrior) registerHeroicStrike() {
	spell := war.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 78},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics,
		ClassSpellMask: SpellMaskHeroicStrike,
		MaxRange:       core.MaxMeleeRange,

		RageCost: core.RageCostOptions{
			Cost:   15,
			Refund: 0.8,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),
		FlatThreatBonus:  194,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 176 + war.MHWeaponDamage(sim, spell.MeleeAttackPower())
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}

			if war.curQueueAura != nil {
				war.curQueueAura.Deactivate(sim)
			}
		},
	})
	war.makeQueueSpellsAndAura(spell)
}

func (war *Warrior) registerCleave() {
	const maxTargets int32 = 2
	flatDamage := 70 * (1 + 0.4*float64(war.Talents.ImprovedCleave))

	spell := war.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 845},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics,
		ClassSpellMask: SpellMaskCleave,
		MaxRange:       core.MaxMeleeRange,

		RageCost: core.RageCostOptions{
			Cost: 20,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
		},

		DamageMultiplier: 0.82,
		ThreatMultiplier: 1,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),
		FlatThreatBonus:  125,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := flatDamage + war.MHWeaponDamage(sim, spell.MeleeAttackPower())
			spell.CalcCleaveDamage(sim, target, maxTargets, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			spell.DealBatchedAoeDamage(sim)

			if war.curQueueAura != nil {
				war.curQueueAura.Deactivate(sim)
			}
		},
	})
	war.makeQueueSpellsAndAura(spell)
}

func (war *Warrior) makeQueueSpellsAndAura(srcSpell *core.Spell) *core.Spell {
	isQueueQueued := false

	queueAura := war.RegisterAura(core.Aura{
		Label:    "HS/Cleave Queue Aura-" + srcSpell.ActionID.String(),
		ActionID: srcSpell.ActionID.WithTag(1),
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			isQueueQueued = false
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			if war.curQueueAura != nil {
				war.curQueueAura.Deactivate(sim)
			}
			war.PseudoStats.DisableDWMissPenalty = true
			war.curQueueAura = aura
			war.curQueuedAutoSpell = srcSpell
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			war.PseudoStats.DisableDWMissPenalty = false
			war.curQueueAura = nil
			war.curQueuedAutoSpell = nil
		},
	})

	queueSpell := war.RegisterSpell(core.SpellConfig{
		ActionID:    srcSpell.ActionID.WithTag(1),
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.curQueueAura == nil &&
				!isQueueQueued &&
				war.CurrentRage() >= srcSpell.Cost.GetCurrentCost() &&
				war.queuedRealismICD.IsReady(sim)
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if war.queuedRealismICD.IsReady(sim) {
				isQueueQueued = true
				war.queuedRealismICD.Use(sim)
				sim.AddPendingAction(&core.PendingAction{
					NextActionAt: sim.CurrentTime + war.queuedRealismICD.Duration,
					OnAction: func(sim *core.Simulation) {
						queueAura.Activate(sim)
						isQueueQueued = false
					},
				})
			}
		},
	})

	return queueSpell
}

// Returns true if the regular melee swing should be used, false otherwise.
func (war *Warrior) TryHSOrCleave(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if !war.curQueueAura.IsActive() {
		return mhSwingSpell
	}

	if !war.curQueuedAutoSpell.CanCast(sim, war.CurrentTarget) {
		war.curQueueAura.Deactivate(sim)
		return mhSwingSpell
	}

	return war.curQueuedAutoSpell
}
