package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (war *Warrior) registerFuryTalents() {
	// Tier 1
	// Booming Voice implemented in shouts.go
	war.registerCruelty()

	// Tier 2
	// Improved Demoralizing Shout implemented in demoralizing_shout.go
	war.registerUnbridledWrath()

	// Tier 3
	war.registerImprovedCleave()
	// Piercing Howl not implemented
	// Blood Craze not implemented
	// Commanding Presence implemented in shouts.go

	// Tier 4
	war.registerDualWieldSpecialization()
	war.registerImprovedExecute()
	war.registerEnrage()

	// Tier 5
	war.registerImprovedSlam()
	war.registerSweepingStrikes()
	war.registerWeaponMastery()

	// Tier 6
	// Improved Berserker Rage implemented in berserker_rage.go
	war.registerFlurry()

	// Tier 7
	war.registerPrecision()
	war.registerBloodthirst()
	war.registerImprovedWhirlwind()

	// Tier 8
	war.registerImprovedBerserkerStance()

	// Tier 9
	war.registerRampage()
}

func (war *Warrior) registerCruelty() {
	if war.Talents.Cruelty == 0 {
		return
	}

	war.AddStat(stats.PhysicalCritPercent, 1+0.01*float64(war.Talents.Cruelty))
}

func (war *Warrior) registerUnbridledWrath() {
	if war.Talents.UnbridledWrath == 0 {
		return
	}

	rageMetrics := war.NewRageMetrics(core.ActionID{SpellID: 13002})

	war.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Unbridled Wrath",
		DPM:                war.NewStaticLegacyPPMManager(3*float64(war.Talents.UnbridledWrath), core.ProcMaskMeleeWhiteHit),
		RequireDamageDealt: true,
		Outcome:            core.OutcomeLanded,
		Callback:           core.CallbackOnSpellHitDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			war.AddRage(sim, 1, rageMetrics)
		},
	})
}

func (war *Warrior) registerImprovedCleave() {
	if war.Talents.ImprovedCleave == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask:  SpellMaskCleave,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.4 * float64(war.Talents.ImprovedCleave),
	})
}

func (war *Warrior) registerDualWieldSpecialization() {
	if war.Talents.DualWieldSpecialization == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ProcMask:   core.ProcMaskMeleeOH,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.05 * float64(war.Talents.DualWieldSpecialization),
	})
}

func (war *Warrior) registerImprovedExecute() {
	if war.Talents.ImprovedExecute == 0 {
		return
	}

	rageCostReduction := []float64{0, 2, 5}[war.Talents.ImprovedExecute]

	war.AddStaticMod(core.SpellModConfig{
		ClassMask:  SpellMaskExecute,
		Kind:       core.SpellMod_PowerCost_Flat,
		FloatValue: -rageCostReduction,
	})
}

func (war *Warrior) registerEnrage() {
	if war.Talents.Enrage == 0 {
		return
	}

	war.EnrageAura = war.GetOrRegisterAura(core.Aura{
		Label:     "Enrage",
		ActionID:  core.ActionID{SpellID: 12317},
		Duration:  time.Second * 12,
		MaxStacks: 12,
	}).AttachSpellMod(core.SpellModConfig{
		School:     core.SpellSchoolPhysical,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.05 * float64(war.Talents.Enrage),
	}).AttachProcTrigger(core.ProcTrigger{
		Name:     "Enrage - Spend",
		ProcMask: core.ProcMaskMelee,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			war.EnrageAura.RemoveStack(sim)
		},
	})

	war.EnrageAura.NewExclusiveEffect("Enrage", true, core.ExclusiveEffect{Priority: 5 * float64(war.Talents.Enrage)})

	war.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Enrage - Trigger",
		ProcMask: core.ProcMaskMelee,
		Outcome:  core.OutcomeCrit,
		Callback: core.CallbackOnSpellHitTaken,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			war.EnrageAura.Activate(sim)
			war.EnrageAura.SetStacks(sim, 12)
		},
	})
}

func (war *Warrior) registerImprovedSlam() {
	if war.Talents.ImprovedSlam == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskSlam,
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: -time.Millisecond * time.Duration(500*war.Talents.ImprovedSlam),
	})
}

func (war *Warrior) registerSweepingStrikes() {
	if !war.Talents.SweepingStrikes {
		return
	}

	ssSpellActionID := core.ActionID{SpellID: 12723}
	ssHitActionID := core.ActionID{SpellID: 12723}

	var curDmg float64
	hitSpell := war.RegisterSpell(core.SpellConfig{
		ActionID:    ssHitActionID,
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskEmpty, // No proc mask, so it won't proc itself.
		Flags:       core.SpellFlagIgnoreResists | core.SpellFlagIgnoreModifiers | core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, curDmg, spell.OutcomeAlwaysHit)
		},
	})

	ssAura := war.RegisterAura(core.Aura{
		Label:     "Sweeping Strikes",
		ActionID:  ssHitActionID,
		Duration:  core.NeverExpires,
		MaxStacks: 10,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.SetStacks(sim, 5)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if aura.GetStacks() == 0 || result.Damage <= 0 || !spell.ProcMask.Matches(core.ProcMaskMelee) {
				return
			}

			if (spell.Matches(SpellMaskExecute) && !sim.IsExecutePhase20()) || spell.Matches(SpellMaskWhirlwind) {
				curDmg = spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			} else {
				curDmg = result.Damage
			}

			hitSpell.Cast(sim, war.Env.NextActiveTargetUnit(result.Target))
			aura.RemoveStack(sim)
		},
	})

	ssCD := war.RegisterSpell(core.SpellConfig{
		ActionID:    ssSpellActionID,
		SpellSchool: core.SpellSchoolPhysical,

		RageCost: core.RageCostOptions{
			Cost: 30,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Second * 30,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(BattleStance|BerserkerStance) || sim.ActiveTargetCount() > 1
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			ssAura.Activate(sim)
		},
	})

	war.AddMajorCooldown(core.MajorCooldown{
		Spell: ssCD,
		Type:  core.CooldownTypeDPS,
	})
}

func (war *Warrior) registerWeaponMastery() {
	if war.Talents.WeaponMastery == 0 {
		return
	}

	war.PseudoStats.DodgeReduction += 0.01 * float64(war.Talents.WeaponMastery)
}

func (war *Warrior) registerFlurry() {
	if war.Talents.Flurry == 0 {
		return
	}

	flurryAura := war.RegisterAura(core.Aura{
		Label:     "Flurry",
		ActionID:  core.ActionID{SpellID: 12966},
		Duration:  15 * time.Second,
		MaxStacks: 3,
	}).AttachMultiplyMeleeSpeed(1 + 0.05*float64(war.Talents.Flurry))

	war.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Flurry - Trigger",
		ActionID:           core.ActionID{SpellID: 12319},
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMelee,
		Outcome:            core.OutcomeLanded,
		TriggerImmediately: true,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.Matches(SpellMaskWhirlwindOh) {
				return
			}

			if result.Outcome.Matches(core.OutcomeCrit) {
				flurryAura.Activate(sim)
				flurryAura.SetStacks(sim, 3)
				return
			}

			if flurryAura.IsActive() && spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
				flurryAura.RemoveStack(sim)
			}
		},
	})
}

func (war *Warrior) registerPrecision() {
	if war.Talents.Precision == 0 {
		return
	}

	war.AddStat(stats.PhysicalHitPercent, 0.01*float64(war.Talents.Precision))
	war.AddStat(stats.RangedHitPercent, 0.01*float64(war.Talents.Precision))
}

func (war *Warrior) registerBloodthirst() {
	actionID := core.ActionID{SpellID: 23881}

	war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: SpellMaskBloodthirst,
		MaxRange:       core.MaxMeleeRange,

		RageCost: core.RageCostOptions{
			Cost:   30,
			Refund: 0.8,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Second * 6,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.MeleeAttackPower() * 0.45
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}

func (war *Warrior) registerImprovedWhirlwind() {
	if war.Talents.ImprovedWhirlwind == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskWhirlwind,
		Kind:      core.SpellMod_Cooldown_Flat,
		TimeValue: -time.Second * time.Duration(war.Talents.ImprovedWhirlwind),
	})
}

func (war *Warrior) registerImprovedBerserkerStance() {
	if war.Talents.ImprovedBerserkerStance == 0 {
		return
	}

	apDep := war.NewDynamicMultiplyStat(stats.AttackPower, 1+0.02*float64(war.Talents.ImprovedBerserkerStance))
	war.OnSpellRegistered(func(spell *core.Spell) {
		if !spell.Matches(SpellMaskBerserkerStance) {
			war.BerserkerStanceAura.AttachStatDependency(apDep)
		}
	})

}

func (war *Warrior) registerRampage() {
	actionID := core.ActionID{SpellID: 29801}

	validUntil := time.Duration(0)

	aura := core.MakeStackingAura(&war.Character, core.StackingStatAura{
		Aura: core.Aura{
			Label:     "Rampage",
			ActionID:  actionID,
			Duration:  time.Second * 30,
			MaxStacks: 5,
		},
		BonusPerStack: stats.Stats{stats.AttackPower: 50},
	})

	war.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Rampage - Trigger",
		Outcome:  core.OutcomeLanded,
		Callback: core.CallbackOnSpellHitDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Outcome.Matches(core.OutcomeCrit) {
				validUntil = sim.CurrentTime + time.Second*5
			}

			if spell.ProcMask.Matches(core.ProcMaskMelee) {
				if aura.IsActive() {
					aura.AddStack(sim)
				}
			}
		},
	})

	spell := war.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Second * 90,
			},
		},

		RageCost: core.RageCostOptions{
			Cost: 20,
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return sim.CurrentTime < validUntil
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			validUntil = 0
			aura.Activate(sim)
			aura.AddStack(sim)
		},
	})

	war.AddMajorCooldown(core.MajorCooldown{
		Type:  core.CooldownTypeDPS,
		Spell: spell,
	})
}
