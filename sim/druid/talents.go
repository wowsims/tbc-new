package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

// FeralCritMultiplier returns the melee crit multiplier for cat/bear form abilities,
// including the bonus from Predatory Instincts (+2% crit damage per rank).
func (druid *Druid) FeralCritMultiplier() float64 {
	return druid.DefaultMeleeCritMultiplier() * (1 + 0.02*float64(druid.Talents.PredatoryInstincts))
}

// ApplyBalanceTalents applies all Balance tree talents. Call this from Balance spec only.
func (druid *Druid) ApplyBalanceTalents() {
	druid.applyStarlightWrath()
	druid.applyFocusedStarlight()
	druid.applyImprovedMoonfire()
	druid.applyBrambles()
	druid.applyInsectSwarm()
	druid.applyNaturesReach()
	druid.applyVengeance()
	druid.applyCelestialFocus()
	druid.applyLunarGuidance()
	druid.applyNaturesGrace()
	druid.applyMoonglow()
	druid.applyMoonfury()
	druid.applyBalanceOfPower()
	druid.applyDreamstate()
	druid.applyImprovedFaerieFire()
	druid.applyWrathOfCenarius()
	druid.applyForceOfNature()
}

// ApplyFeralTalents applies Feral tree talents and Restoration tree talents used by
// feral specs (cat and bear). Call this from FeralCat and FeralBear specs.
func (druid *Druid) ApplyFeralTalents() {
	druid.applyNaturalShapeshifter()
	druid.applySubtlety()
	druid.applyNaturalist()
	druid.applyIntensity()
	druid.applyPrimalFury()
	druid.applyLeaderOfThePack()
	druid.applyPredatoryStrikes()
	druid.applySharpenedClaws()
	druid.applyFeralSwiftness()
	druid.applyHeartOfTheWild()
	druid.applySurvivalOfTheFittest()
}

func (druid *Druid) applyForceOfNature() {
	if !druid.Talents.ForceOfNature {
		return
	}

	druid.registerForceOfNatureCD()
}

func (druid *Druid) applyWrathOfCenarius() {
	if druid.Talents.WrathOfCenarius == 0 {
		return
	}

	druid.AddStaticMod(core.SpellModConfig{
		ClassMask:  DruidSpellWrath,
		Kind:       core.SpellMod_BonusCoeffecient_Flat,
		FloatValue: 0.02 * float64(druid.Talents.WrathOfCenarius),
	})

	druid.AddStaticMod(core.SpellModConfig{
		ClassMask:  DruidSpellStarfire,
		Kind:       core.SpellMod_BonusCoeffecient_Flat,
		FloatValue: 0.04 * float64(druid.Talents.WrathOfCenarius),
	})
}

func (druid *Druid) applyImprovedFaerieFire() {
	if druid.Talents.ImprovedFaerieFire == 0 {
		return
	}
}

func (druid *Druid) applyDreamstate() {
	if druid.Talents.Dreamstate == 0 {
		return
	}

	mp5Bonus := []float64{0, 0.04, 0.07, 0.1}[druid.Talents.LunarGuidance]
	druid.AddStatDependency(stats.Intellect, stats.MP5, mp5Bonus)
}

func (druid *Druid) applyBalanceOfPower() {
	if druid.Talents.BalanceOfPower == 0 {
		return
	}

	druid.AddStaticMod(core.SpellModConfig{
		// See https://www.wowhead.com/tbc/spell=33596/balance-of-power
		ClassMask:  DruidSpellWrath | DruidSpellStarfire | DruidSpellMoonfire,
		Kind:       core.SpellMod_BonusHit_Percent,
		FloatValue: 2.0 * float64(druid.Talents.BalanceOfPower),
	})
}

func (druid *Druid) applyMoonfury() {
	if druid.Talents.Moonfury == 0 {
		return
	}

	druid.AddStaticMod(core.SpellModConfig{
		ClassMask:  DruidSpellWrath | DruidSpellStarfire | DruidSpellMoonfire,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.02 * float64(druid.Talents.Moonfury),
	})
}

func (druid *Druid) applyMoonglow() {
	if druid.Talents.Moonglow == 0 {
		return
	}

	druid.AddStaticMod(core.SpellModConfig{
		ClassMask:  DruidSpellMoonfire | DruidSpellStarfire | DruidSpellWrath | DruidSpellHealingTouch | DruidSpellRegrowth | DruidSpellRejuvenation,
		FloatValue: -0.03 * float64(druid.Talents.Moonglow),
		Kind:       core.SpellMod_PowerCost_Pct,
	})
}

func (druid *Druid) applyNaturesGrace() {
	if !druid.Talents.NaturesGrace {
		return
	}

	aura := druid.RegisterAura(core.Aura{
		Label:    "Nature's Grace",
		ActionID: core.ActionID{SpellID: 16886},
		Duration: time.Second * 3,
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.CurCast.CastTime == 0 {
				return
			}

			aura.Deactivate(sim)
		},
	}).AttachSpellMod(core.SpellModConfig{
		ClassMask: DruidSpellStarfire | DruidSpellWrath,
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Millisecond * time.Duration(-500),
	})

	druid.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Nature's Grace Trigger",
		Duration:       core.NeverExpires,
		ClassSpellMask: DruidSpellWrath | DruidSpellStarfire,
		Outcome:        core.OutcomeCrit,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura.Activate(sim)
		},
	})
}

func (druid *Druid) applyLunarGuidance() {
	if druid.Talents.LunarGuidance == 0 {
		return
	}

	spellDamageBonus := []float64{0, 0.08, 0.16, 0.25}[druid.Talents.LunarGuidance]
	druid.AddStatDependency(stats.Intellect, stats.SpellDamage, spellDamageBonus)
}

func (druid *Druid) applyCelestialFocus() {
	if druid.Talents.CelestialFocus == 0 {
		return
	}

	panic("unimplemented")
}

func (druid *Druid) applyVengeance() {
	if druid.Talents.Vengeance == 0 {
		return
	}

	druid.AddStaticMod(core.SpellModConfig{
		ClassMask:  DruidSpellWrath | DruidSpellStarfire | DruidSpellMoonfire,
		Kind:       core.SpellMod_CritMultiplier_Flat,
		FloatValue: 0.2 * float64(druid.Talents.Vengeance),
	})
}

func (druid *Druid) applyNaturesReach() {
	if druid.Talents.NaturesReach == 0 {
		return
	}

	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  DruidSpellsBalance | DruidSpellFearieFireFeral,
	// 	Kind:       ****BONUS RANGE**** most likely irrelevant for sim
	// 	FloatValue: 10.0 * float64(druid.Talents.NaturesReach),
	// })
}

func (druid *Druid) applyInsectSwarm() {
	if !druid.Talents.InsectSwarm {
		return
	}

	druid.registerInsectSwarmSpell()
}

func (druid *Druid) applyBrambles() {
	if druid.Talents.Brambles == 0 {
		return
	}

	druid.AddStaticMod(core.SpellModConfig{
		ClassMask:  DruidSpellThorns | DruidSpellEntanglingRoots,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.25 * float64(druid.Talents.Brambles),
	})
}

func (druid *Druid) applyStarlightWrath() {
	if druid.Talents.StarlightWrath == 0 {
		return
	}

	druid.AddStaticMod(core.SpellModConfig{
		ClassMask: DruidSpellStarfire | DruidSpellWrath,
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Millisecond * time.Duration(-100*druid.Talents.StarlightWrath),
	})
}

func (druid *Druid) applyFocusedStarlight() {
	if druid.Talents.FocusedStarlight == 0 {
		return
	}

	druid.AddStaticMod(core.SpellModConfig{
		ClassMask:  DruidSpellStarfire | DruidSpellWrath,
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 2 * float64(druid.Talents.FocusedStarlight),
	})
}

func (druid *Druid) applyImprovedMoonfire() {
	if druid.Talents.ImprovedMoonfire == 0 {
		return
	}

	// 5% per point damage increase to Moonfire and its DoT
	druid.AddStaticMod(core.SpellModConfig{
		ClassMask:  DruidSpellMoonfire,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.05 * float64(druid.Talents.ImprovedMoonfire),
	})

	// 5% per point chance to crit with Moonfire
	druid.AddStaticMod(core.SpellModConfig{
		ClassMask:  DruidSpellMoonfire,
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 5 * float64(druid.Talents.ImprovedMoonfire),
	})
}

func (druid *Druid) applyNaturalShapeshifter() {
	if druid.Talents.NaturalShapeshifter == 0 {
		return
	}

	druid.AddStaticMod(core.SpellModConfig{
		ClassMask:  DruidSpellCatForm | DruidSpellBearForm,
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -0.1 * float64(druid.Talents.NaturalShapeshifter),
	})
}

func (druid *Druid) applySubtlety() {
	if druid.Talents.Subtlety == 0 {
		return
	}

	druid.PseudoStats.ThreatMultiplier *= 1 - 0.04*float64(druid.Talents.Subtlety)
}

func (druid *Druid) applyNaturalist() {
	if druid.Talents.Naturalist == 0 {
		return
	}

	druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= 1 + 0.02*float64(druid.Talents.Naturalist)
}

func (druid *Druid) applyIntensity() {
	if druid.Talents.Intensity == 0 {
		return
	}

	druid.PseudoStats.SpiritRegenRateCasting += float64(druid.Talents.Intensity) * 0.1
}

func (druid *Druid) applyPredatoryStrikes() {
	if druid.Talents.PredatoryStrikes == 0 || !druid.InForm(Bear|Cat) {
		return
	}

	druid.AddStat(stats.AttackPower, float64(druid.Talents.PredatoryStrikes)*0.5*core.CharacterLevel)
}

func (druid *Druid) applySharpenedClaws() {
	if druid.Talents.SharpenedClaws == 0 || !druid.InForm(Bear|Cat) {
		return
	}

	druid.AddStat(stats.PhysicalCritPercent, float64(druid.Talents.SharpenedClaws)*2.0)
}

func (druid *Druid) applyFeralSwiftness() {
	if druid.Talents.FeralSwiftness == 0 || !druid.InForm(Bear|Cat) {
		return
	}

	druid.AddStat(stats.DodgeRating, core.DodgeRatingPerDodgePercent*2.0*float64(druid.Talents.FeralSwiftness))
}

func (druid *Druid) applyHeartOfTheWild() {
	if druid.Talents.HeartOfTheWild == 0 {
		return
	}

	ranks := float64(druid.Talents.HeartOfTheWild)
	// +4% Intellect per rank (all forms)
	druid.MultiplyStat(stats.Intellect, 1+0.04*ranks)
	if druid.InForm(Cat) {
		// +2% Attack Power per rank in Cat Form (+10% at 5/5)
		druid.MultiplyStat(stats.AttackPower, 1+0.02*ranks)
	} else if druid.InForm(Bear) {
		// +4% Stamina per rank in Bear Form (+20% at 5/5)
		druid.MultiplyStat(stats.Stamina, 1+0.04*ranks)
	}
}

func (druid *Druid) applySurvivalOfTheFittest() {
	if druid.Talents.SurvivalOfTheFittest == 0 {
		return
	}

	mult := 1 + 0.01*float64(druid.Talents.SurvivalOfTheFittest)
	for _, s := range []stats.Stat{stats.Stamina, stats.Strength, stats.Agility, stats.Intellect, stats.Spirit} {
		druid.MultiplyStat(s, mult)
	}
	druid.PseudoStats.ReducedCritTakenChance += 0.01 * float64(druid.Talents.SurvivalOfTheFittest)
}

func (druid *Druid) applyPrimalFury() {
	if druid.Talents.PrimalFury == 0 {
		return
	}

	procChance := 0.5 * float64(druid.Talents.PrimalFury)
	actionID := core.ActionID{SpellID: 37117}
	rageMetrics := druid.NewRageMetrics(actionID)
	cpMetrics := druid.NewComboPointMetrics(actionID)

	// Cat form: +1 combo point on builder crits.
	druid.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Primal Fury (Cat)",
		ActionID:       actionID,
		Callback:       core.CallbackOnSpellHitDealt,
		ClassSpellMask: DruidSpellBuilder,
		Outcome:        core.OutcomeCrit,
		ProcChance:     procChance,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			druid.AddComboPoints(sim, 1, cpMetrics)
		},
	})

	// Bear form: +5 rage on Mangle crits.
	druid.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Primal Fury (Bear Mangle)",
		ActionID:       actionID,
		Callback:       core.CallbackOnSpellHitDealt,
		ClassSpellMask: DruidSpellMangleBear,
		Outcome:        core.OutcomeCrit,
		ProcChance:     procChance,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			druid.AddRage(sim, 5, rageMetrics)
		},
	})

	// Bear form: +5 rage on auto-attack crits.
	druid.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Primal Fury (Bear Auto)",
		ActionID:   actionID,
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskMeleeMHAuto,
		Outcome:    core.OutcomeCrit,
		ProcChance: procChance,
		ExtraCondition: func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
			return druid.InForm(Bear)
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			druid.AddRage(sim, 5, rageMetrics)
		},
	})
}

func (druid *Druid) applyLeaderOfThePack() {
	if !druid.Talents.LeaderOfThePack {
		return
	}
	if druid.Talents.ImprovedLeaderOfThePack == 0 {
		return
	}

	// Improved LotP: crits heal the druid for 4% max health, 6s ICD.
	healthRestore := 0.04

	healingSpell := druid.RegisterSpell(Cat|Bear, core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 34299},
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagIgnoreModifiers,
		DamageMultiplier: 1,
		ThreatMultiplier: 0,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, healthRestore*spell.Unit.MaxHealth(), spell.OutcomeHealing)
		},
	})

	druid.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Improved Leader of the Pack",
		Callback: core.CallbackOnSpellHitDealt,
		ProcMask: core.ProcMaskMeleeOrRanged,
		Outcome:  core.OutcomeCrit,
		ICD:      time.Second * 6,
		ExtraCondition: func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
			return druid.InForm(Cat | Bear)
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			healingSpell.Cast(sim, &druid.Unit)
		},
	})
}
