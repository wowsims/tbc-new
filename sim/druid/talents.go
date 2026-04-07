package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (druid *Druid) ApplyTalents() {
	// Balance
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

	// Restoration
	druid.applyIntensity()
	// Feral
	// druid.registerFelineSwiftness()
	// druid.registerDisplacerBeast()
	// druid.registerWildCharge()

	// druid.registerYserasGift()
	// druid.registerRenewal()
	// druid.registerCenarionWard()

	// druid.registerForceOfNature()

	// druid.registerHeartOfTheWild()
	// druid.registerNaturesVigil()
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
		Kind:       core.SpellMod_DamageDone_Flat,
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

func (druid *Druid) applyIntensity() {
	if druid.Talents.Intensity == 0 {
		return
	}

	// Allows 10% per rank of mana regeneration to continue while casting
	druid.PseudoStats.SpiritRegenRateCasting += 0.10 * float64(druid.Talents.Intensity)
	druid.UpdateManaRegenRates()

	// Enrage instantly generates additional rage per rank (4/7/10)
	druid.IntensityEnrageRageBonus = []float64{0, 4, 7, 10}[druid.Talents.Intensity]
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

// func (druid *Druid) registerFelineSwiftness() {

// 	druid.PseudoStats.MovementSpeedMultiplier *= 1.15
// }

// func (druid *Druid) registerDisplacerBeast() {

// 	druid.DisplacerBeastAura = druid.RegisterAura(core.Aura{
// 		Label:    "Displacer Beast",
// 		ActionID: core.ActionID{SpellID: 137452},
// 		Duration: time.Second * 4,
// 	})

// 	exclusiveSpeedEffect := druid.DisplacerBeastAura.NewActiveMovementSpeedEffect(0.5)

// 	druid.DisplacerBeast = druid.RegisterSpell(Any, core.SpellConfig{
// 		ActionID:        core.ActionID{SpellID: 102280},
// 		RelatedSelfBuff: druid.DisplacerBeastAura,

// 		Cast: core.CastConfig{
// 			DefaultCast: core.Cast{
// 				GCD: core.GCDDefault,
// 			},

// 			IgnoreHaste: true,

// 			CD: core.Cooldown{
// 				Timer:    druid.NewTimer(),
// 				Duration: time.Second * 30,
// 			},
// 		},

// 		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
// 			if !druid.InForm(Cat) {
// 				druid.CatFormAura.Activate(sim)
// 			}

// 			druid.DistanceFromTarget = math.Abs(druid.DistanceFromTarget - 20)
// 			druid.MoveDuration(core.SpellBatchWindow, sim)

// 			if !exclusiveSpeedEffect.Category.AnyActive() {
// 				druid.DisplacerBeastAura.Activate(sim)
// 			}

// 			// Displacer Beast is a full swing timer reset based on in-game measurements.
// 			if druid.DistanceFromTarget > core.MaxMeleeRange {
// 				return
// 			}

// 			druid.AutoAttacks.CancelMeleeSwing(sim)
// 			pa := sim.GetConsumedPendingActionFromPool()
// 			pa.NextActionAt = sim.CurrentTime + druid.AutoAttacks.MainhandSwingSpeed()
// 			pa.Priority = core.ActionPriorityDOT

// 			pa.OnAction = func(sim *core.Simulation) {
// 				druid.AutoAttacks.EnableMeleeSwing(sim)
// 			}

// 			sim.AddPendingAction(pa)
// 		},
// 	})

// 	druid.AddMajorCooldown(core.MajorCooldown{
// 		Spell: druid.DisplacerBeast.Spell,
// 		Type:  core.CooldownTypeDPS,

// 		ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
// 			return (character.DistanceFromTarget >= 20-core.MaxMeleeRange) && (character.GetAura("Nitro Boosts") == nil)
// 		},
// 	})
// }

// func (druid *Druid) registerWildCharge() {

// 	sharedCD := core.Cooldown{
// 		Timer:    druid.NewTimer(),
// 		Duration: time.Second * 15,
// 	}

// 	if druid.CatFormAura != nil {
// 		druid.registerCatCharge(sharedCD)
// 	}

// 	// TODO: Bear and Moonkin versions
// }

// func (druid *Druid) registerCatCharge(sharedCD core.Cooldown) {
// 	druid.CatCharge = druid.RegisterSpell(Cat, core.SpellConfig{
// 		ActionID: core.ActionID{SpellID: 49376},
// 		Flags:    core.SpellFlagAPL,
// 		MinRange: 8,
// 		MaxRange: 25,

// 		Cast: core.CastConfig{
// 			SharedCD:    sharedCD,
// 			IgnoreHaste: true,
// 		},

// 		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
// 			return !druid.PseudoStats.InFrontOfTarget && !druid.CannotShredTarget
// 		},

// 		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
// 			// Leap speed is around 80 yards/second according to measurements
// 			// from boЯsch. This is too fast to be modeled accurately using
// 			// movement aura stacks, so do it directly here by setting the
// 			// position to 0 instantaneously but introducing a GCD delay based
// 			// on the distance traveled.
// 			travelTime := core.DurationFromSeconds(druid.DistanceFromTarget / 80)
// 			druid.ExtendGCDUntil(sim, max(druid.NextGCDAt(), sim.CurrentTime+travelTime))
// 			druid.DistanceFromTarget = 0
// 			druid.MoveDuration(travelTime, sim)

// 			// Measurements from boЯsch indicate that while travel speed (and
// 			// therefore special ability delays) is fairly consistent, there
// 			// is an additional variable delay on auto-attacks after landing,
// 			// likely due to the server needing to perform positional checks.
// 			const minAutoDelaySeconds = 0.150
// 			const autoDelaySpreadSeconds = 0.6

// 			randomDelayTime := core.DurationFromSeconds(minAutoDelaySeconds + sim.RandomFloat("Cat Charge")*autoDelaySpreadSeconds)
// 			druid.AutoAttacks.CancelMeleeSwing(sim)
// 			pa := sim.GetConsumedPendingActionFromPool()
// 			pa.NextActionAt = sim.CurrentTime + travelTime + randomDelayTime
// 			pa.Priority = core.ActionPriorityDOT

// 			pa.OnAction = func(sim *core.Simulation) {
// 				druid.AutoAttacks.EnableMeleeSwing(sim)
// 			}

// 			sim.AddPendingAction(pa)
// 		},
// 	})
// }

// func (druid *Druid) registerHeartOfTheWild() {

// 	// Apply 6% increase to Stamina, Agility, and Intellect
// 	statMultiplier := 1.06
// 	druid.MultiplyStat(stats.Stamina, statMultiplier)
// 	druid.MultiplyStat(stats.Agility, statMultiplier)
// 	druid.MultiplyStat(stats.Intellect, statMultiplier)

// 	// The activation spec specific effects are implemented in individual spec packages.
// }

// func (druid *Druid) RegisterSharedFeralHotwMods() (*core.SpellMod, *core.SpellMod, *core.SpellMod) {
// 	healingMask := DruidSpellTranquility | DruidSpellRejuvenation | DruidSpellHealingTouch | DruidSpellCenarionWard

// 	healingMod := druid.AddDynamicMod(core.SpellModConfig{
// 		ClassMask:  healingMask,
// 		Kind:       core.SpellMod_DamageDone_Pct,
// 		FloatValue: 1,
// 	})

// 	damageMask := DruidSpellWrath | DruidSpellMoonfire | DruidSpellMoonfireDoT | DruidSpellHurricane

// 	damageMod := druid.AddDynamicMod(core.SpellModConfig{
// 		ClassMask:  damageMask,
// 		Kind:       core.SpellMod_DamageDone_Pct,
// 		FloatValue: 3.2,
// 	})

// 	costMod := druid.AddDynamicMod(core.SpellModConfig{
// 		ClassMask:  healingMask | damageMask,
// 		Kind:       core.SpellMod_PowerCost_Pct,
// 		FloatValue: -2,
// 	})

// 	return healingMod, damageMod, costMod
// }

// func (druid *Druid) registerNaturesVigil() {

// 	var smartHealStrength float64

// 	smartHealSpell := druid.RegisterSpell(Any, core.SpellConfig{
// 		ActionID:         core.ActionID{SpellID: 124988},
// 		SpellSchool:      core.SpellSchoolNature,
// 		ProcMask:         core.ProcMaskSpellHealing,
// 		Flags:            core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagIgnoreAttackerModifiers,
// 		DamageMultiplier: 1,
// 		ThreatMultiplier: 0,

// 		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
// 			spell.CalcAndDealHealing(sim, target, 0.25*smartHealStrength, spell.OutcomeHealing)
// 		},
// 	})

// 	actionID := core.ActionID{SpellID: 124974}
// 	numAllyTargets := float64(len(druid.Env.Raid.AllPlayerUnits) - 1)

// 	naturesVigilAura := druid.RegisterAura(core.Aura{
// 		Label:    "Nature's Vigil",
// 		ActionID: actionID,
// 		Duration: time.Second * 30,

// 		OnGain: func(aura *core.Aura, sim *core.Simulation) {
// 			aura.Unit.PseudoStats.DamageDealtMultiplier *= 1.12
// 			aura.Unit.PseudoStats.HealingDealtMultiplier *= 1.12
// 			smartHealStrength = 0

// 			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
// 				Period:   time.Millisecond * 250,
// 				NumTicks: 119,
// 				Priority: core.ActionPriorityDOT,

// 				OnAction: func(sim *core.Simulation) {
// 					if smartHealStrength == 0 {
// 						return
// 					}

// 					// Assume that the target is randomly selected from 25 raid members.
// 					if sim.Proc(1.0/25.0, "Nature's Vigil") {
// 						smartHealSpell.Cast(sim, aura.Unit)
// 					} else if numAllyTargets > 0 {
// 						targetIdx := 1 + int(sim.RandomFloat("Nature's Vigil")*numAllyTargets)
// 						smartHealSpell.Cast(sim, sim.Raid.AllPlayerUnits[targetIdx])
// 					}

// 					smartHealStrength = 0
// 				},
// 			})
// 		},

// 		OnExpire: func(aura *core.Aura, _ *core.Simulation) {
// 			aura.Unit.PseudoStats.DamageDealtMultiplier /= 1.12
// 			aura.Unit.PseudoStats.HealingDealtMultiplier /= 1.12
// 		},

// 		OnSpellHitDealt: func(_ *core.Aura, _ *core.Simulation, spell *core.Spell, result *core.SpellResult) {
// 			// if !spell.Flags.Matches(core.SpellFlagAoE) {
// 			// 	smartHealStrength = max(smartHealStrength, result.Damage)
// 			// }
// 		},
// 	})

// 	naturesVigilSpell := druid.RegisterSpell(Any, core.SpellConfig{
// 		ActionID:        actionID,
// 		Flags:           core.SpellFlagAPL,
// 		RelatedSelfBuff: naturesVigilAura,

// 		Cast: core.CastConfig{
// 			DefaultCast: core.Cast{
// 				GCD: 0,
// 			},
// 			IgnoreHaste: true,
// 			CD: core.Cooldown{
// 				Timer:    druid.NewTimer(),
// 				Duration: time.Second * 90,
// 			},
// 		},

// 		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
// 			spell.RelatedSelfBuff.Activate(sim)
// 		},
// 	})

// 	druid.AddMajorCooldown(core.MajorCooldown{
// 		Spell: naturesVigilSpell.Spell,
// 		Type:  core.CooldownTypeDPS,
// 	})
// }

// func (druid *Druid) registerYserasGift() {

// 	healingSpell := druid.RegisterSpell(Any, core.SpellConfig{
// 		ActionID:         core.ActionID{SpellID: 145108},
// 		SpellSchool:      core.SpellSchoolNature,
// 		ProcMask:         core.ProcMaskSpellHealing,
// 		Flags:            core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
// 		DamageMultiplier: 1,
// 		ThreatMultiplier: 1,
// 		CritMultiplier:   druid.DefaultSpellCritMultiplier(),

// 		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
// 			spell.CalcAndDealHealing(sim, target, 0.05*spell.Unit.MaxHealth(), spell.OutcomeHealing)
// 		},
// 	})

// 	druid.RegisterResetEffect(func(sim *core.Simulation) {
// 		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
// 			Period:   time.Second * 5,
// 			Priority: core.ActionPriorityDOT,

// 			OnAction: func(sim *core.Simulation) {
// 				healingSpell.Cast(sim, &druid.Unit)
// 			},
// 		})
// 	})
// }

// func (druid *Druid) registerRenewal() {

// 	renewalSpell := druid.RegisterSpell(Any, core.SpellConfig{
// 		ActionID:         core.ActionID{SpellID: 108238},
// 		SpellSchool:      core.SpellSchoolNature,
// 		ProcMask:         core.ProcMaskSpellHealing,
// 		Flags:            core.SpellFlagHelpful | core.SpellFlagAPL | core.SpellFlagIgnoreModifiers,
// 		DamageMultiplier: 1,
// 		ThreatMultiplier: 1,

// 		Cast: core.CastConfig{
// 			CD: core.Cooldown{
// 				Timer:    druid.NewTimer(),
// 				Duration: time.Minute * 2,
// 			},
// 		},

// 		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
// 			spell.CalcAndDealHealing(sim, spell.Unit, 0.3*spell.Unit.MaxHealth(), spell.OutcomeHealing)
// 		},
// 	})

// 	druid.AddMajorCooldown(core.MajorCooldown{
// 		Spell: renewalSpell.Spell,
// 		Type:  core.CooldownTypeSurvival,
// 	})
// }

// func (druid *Druid) registerCenarionWard() {

// 	// First register the HoT spell that gets triggered when the target takes damage.
// 	baseTickDamage := 11.27999973297 // ~12349

// 	// SP is snapshot at the time of the original buff cast according to simc
// 	var spSnapshot float64

// 	cenarionWardHot := druid.RegisterSpell(Any, core.SpellConfig{
// 		ActionID:         core.ActionID{SpellID: 102352},
// 		SpellSchool:      core.SpellSchoolNature,
// 		ProcMask:         core.ProcMaskSpellHealing,
// 		Flags:            core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete,
// 		DamageMultiplier: 1,
// 		ThreatMultiplier: 1,
// 		CritMultiplier:   druid.DefaultSpellCritMultiplier(),
// 		ClassSpellMask:   DruidSpellCenarionWard,

// 		Hot: core.DotConfig{
// 			Aura: core.Aura{
// 				Label: "Cenarion Ward (HoT)",
// 			},

// 			NumberOfTicks: 3,
// 			TickLength:    time.Second * 2,

// 			OnSnapshot: func(_ *core.Simulation, _ *core.Unit, dot *core.Dot) {
// 				dot.SnapshotBaseDamage = baseTickDamage + spSnapshot*1.04
// 				dot.SnapshotAttackerMultiplier = dot.CasterPeriodicHealingMultiplier()
// 			},

// 			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
// 				dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
// 			},
// 		},

// 		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
// 			spell.Hot(target).Apply(sim)
// 		},
// 	})

// 	// Then register the buff that triggers the HoT upon taking damage.
// 	buffActionID := core.ActionID{SpellID: 102351}

// 	buffConfig := core.Aura{
// 		Label:    "Cenarion Ward",
// 		ActionID: buffActionID,
// 		Duration: time.Second * 30,

// 		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
// 			if result.Damage > 0 {
// 				cenarionWardHot.Cast(sim, aura.Unit)
// 				aura.Deactivate(sim)
// 			}
// 		},
// 	}

// 	cenarionWardBuffs := druid.NewAllyAuraArray(func(target *core.Unit) *core.Aura {
// 		return target.GetOrRegisterAura(buffConfig)
// 	})

// 	// Finally, register the spell that applies the buff.
// 	druid.RegisterSpell(Any, core.SpellConfig{
// 		ActionID: buffActionID,
// 		ProcMask: core.ProcMaskEmpty,
// 		Flags:    core.SpellFlagHelpful | core.SpellFlagAPL,

// 		ManaCost: core.ManaCostOptions{
// 			BaseCostPercent: 14.8,
// 		},

// 		Cast: core.CastConfig{
// 			DefaultCast: core.Cast{
// 				GCD: core.GCDDefault,
// 			},

// 			CD: core.Cooldown{
// 				Timer:    druid.NewTimer(),
// 				Duration: time.Second * 30,
// 			},
// 		},

// 		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
// 			spSnapshot = cenarionWardHot.HealingPower(target)
// 			cenarionWardBuffs.Get(target).Activate(sim)
// 		},
// 	})
// }

// func (druid *Druid) registerForceOfNature() {
// 	if !druid.Talents.ForceOfNature {
// 		return
// 	}

// 	druid.ForceOfNature = druid.RegisterSpell(Any, core.SpellConfig{
// 		ActionID:     core.ActionID{SpellID: 106737},
// 		Flags:        core.SpellFlagAPL,
// 		Charges:      3,
// 		RechargeTime: time.Second * 20,

// 		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
// 			druid.Treants[spell.GetNumCharges()].Enable(sim)
// 		},
// 	})

// 	druid.AddMajorCooldown(core.MajorCooldown{
// 		Spell: druid.ForceOfNature.Spell,
// 		Type:  core.CooldownTypeDPS,
// 	})
// }
