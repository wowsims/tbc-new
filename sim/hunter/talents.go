package hunter

import (
	"slices"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (hunter *Hunter) ApplyTalents() {
	// Beast Mastery
	hunter.registerImprovedAspectOfTheHawk()
	hunter.registerEnduranceTraining()
	hunter.registerFocusedFire()
	hunter.registerUnleashedFury()
	hunter.registerFerocity()
	// Bestial Discipline handled in pet.go
	hunter.registerAnimalHandler()
	hunter.registerFrenzy()
	hunter.registerFerociousInspiration()
	hunter.registerBestialWrath()
	hunter.registerSerpentsSwiftness()
	hunter.registerTheBeastWithin()

	// Marksmanship
	hunter.registerLethalShots()
	// Improved Hunter's Mark handled in hunters_mark.go
	hunter.registerEfficiency()
	hunter.registerGoForTheThroat()
	hunter.registerImprovedArcaneShot()
	hunter.registerAimedShot()
	hunter.registerRapidKilling()
	hunter.registerImprovedStings()
	hunter.registerMortalShots()
	hunter.registerBarrage()
	hunter.registerCombatExperience()
	hunter.registerRangedWeaponSpecialization()
	hunter.registerCarefulAim()
	// Trueshot Aura handled as a group buff in hunter.go
	hunter.registerImprovedBarrage()
	hunter.registerMasterMarksman()

	// Survival
	hunter.registerSlaying()
	hunter.registerHawkEye()
	hunter.registerSavageStrikes()
	hunter.registerSurvivalist()
	hunter.registerSurefooted()
	hunter.registerSurvivalInstincts()
	hunter.registerKillerInstinct()
	hunter.registerResourcefulness()
	hunter.registerLightningReflexes()
	hunter.registerThrillOfTheHunt()
	hunter.registerExposeWeakness()
	hunter.registerMasterTactician()
	hunter.registerReadiness()

	if hunter.Pet != nil {
		hunter.Pet.ApplyTalents()
	}
}

func (hunter *Hunter) registerImprovedAspectOfTheHawk() {
	if hunter.Talents.ImprovedAspectOfTheHawk == 0 {
		return
	}

	bonus := 1.0 + 0.03*float64(hunter.Talents.ImprovedAspectOfTheHawk)

	quickShots := hunter.RegisterAura(core.Aura{
		Label:    "Quick Shots",
		ActionID: core.ActionID{SpellID: 6150},
		Duration: time.Second * 12,
	}).AttachMultiplyRangedHaste(bonus)

	hunter.OnSpellRegistered(func(spell *core.Spell) {
		if !spell.Matches(HunterSpellAspectOfTheHawk) {
			return
		}

		spell.RelatedSelfBuff.AttachProcTrigger(core.ProcTrigger{
			Name:            "Improved Aspect of the Hawk",
			MetricsActionID: core.ActionID{SpellID: 19556},
			Callback:        core.CallbackOnSpellHitDealt,
			ProcMask:        core.ProcMaskRangedAuto,
			Outcome:         core.OutcomeLanded,
			ProcChance:      0.1,

			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				quickShots.Activate(sim)
			},
		})
	})
}

func (hunter *Hunter) registerEnduranceTraining() {
	if hunter.Pet == nil || hunter.Talents.EnduranceTraining == 0 {
		return
	}

	hunter.Pet.StatDependencyManager.EnableDynamicStatDep(
		hunter.Pet.NewDynamicMultiplyStat(stats.Health, 1+0.02*float64(hunter.Talents.EnduranceTraining)),
	)

	hunter.StatDependencyManager.EnableDynamicStatDep(
		hunter.NewDynamicMultiplyStat(stats.Health, 1+0.01*float64(hunter.Talents.EnduranceTraining)),
	)
}

func (hunter *Hunter) registerFocusedFire() {
	if hunter.Pet == nil || hunter.Talents.FocusedFire == 0 {
		return
	}

	hunter.PseudoStats.DamageDealtMultiplier *= 1.0 + 0.01*float64(hunter.Talents.FocusedFire)
	hunter.Pet.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ClassMask:  HunterSpellKillCommandPet,
		FloatValue: 10.0 * float64(hunter.Talents.FocusedFire),
	})
}

func (hunter *Hunter) registerUnleashedFury() {
	if hunter.Pet == nil || hunter.Talents.UnleashedFury == 0 {
		return
	}

	hunter.Pet.PseudoStats.DamageDealtMultiplier *= 1.0 + 0.04*float64(hunter.Talents.UnleashedFury)
}

func (hunter *Hunter) registerFerocity() {
	if hunter.Pet == nil || hunter.Talents.Ferocity == 0 {
		return
	}

	hunter.Pet.AddStats(stats.Stats{
		stats.PhysicalCritPercent: 2 * float64(hunter.Talents.Ferocity),
		stats.SpellCritPercent:    2 * float64(hunter.Talents.Ferocity),
	})
}

func (hunter *Hunter) registerAnimalHandler() {
	if hunter.Pet == nil || hunter.Talents.AnimalHandler == 0 {
		return
	}

	hunter.Pet.AddStats(stats.Stats{
		stats.PhysicalHitPercent: 2 * float64(hunter.Talents.AnimalHandler),
		stats.SpellHitPercent:    2 * float64(hunter.Talents.AnimalHandler),
	})
}

func (hunter *Hunter) registerFrenzy() {
	if hunter.Pet == nil || hunter.Talents.Frenzy == 0 {
		return
	}

	frenzy := hunter.Pet.RegisterAura(core.Aura{
		Label:    "Frenzy Effect",
		ActionID: core.ActionID{SpellID: 19615},
		Duration: time.Second * 8,
	}).AttachMultiplyMeleeSpeed(1.3)

	hunter.Pet.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Frenzy",
		Callback:   core.CallbackOnSpellHitDealt,
		Outcome:    core.OutcomeCrit,
		ProcChance: 0.2 * float64(hunter.Talents.Frenzy),

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			frenzy.Activate(sim)
		},
	})
}

func (hunter *Hunter) registerFerociousInspiration() {
	if hunter.Pet == nil || hunter.Talents.FerociousInspiration == 0 {
		return
	}

	// TODO
}

func (hunter *Hunter) registerBestialWrath() {
	if hunter.Pet == nil || !hunter.Talents.BestialWrath {
		return
	}

	actionID := core.ActionID{SpellID: 19574}

	hunter.Pet.BestialWrathAura = hunter.Pet.RegisterAura(core.Aura{
		Label:    "Bestial Wrath",
		ActionID: actionID,
		Duration: time.Second * 18,
	}).AttachMultiplicativePseudoStatBuff(
		&hunter.Pet.PseudoStats.DamageDealtMultiplier, 1.5,
	)

	hunter.BestialWrath = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolPhysical,
		ClassSpellMask: HunterSpellBestialWrath,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 10,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Minute * 2,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.GCD.IsReady(sim)
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			hunter.Pet.BestialWrathAura.Activate(sim)
		},
	})

	hunter.AddMajorCooldown(core.MajorCooldown{
		Spell: hunter.BestialWrath,
		Type:  core.CooldownTypeDPS,
	})
}

func (hunter *Hunter) registerSerpentsSwiftness() {
	if hunter.Pet == nil || hunter.Talents.SerpentsSwiftness == 0 {
		return
	}

	hunter.PseudoStats.RangedSpeedMultiplier *= 1 + 0.04*float64(hunter.Talents.SerpentsSwiftness)
	hunter.Pet.PseudoStats.MeleeSpeedMultiplier *= 1 + 0.04*float64(hunter.Talents.SerpentsSwiftness)
}

func (hunter *Hunter) registerTheBeastWithin() {
	if hunter.Pet == nil || !hunter.Talents.BestialWrath || !hunter.Talents.TheBeastWithin {
		return
	}

	hunter.TheBeastWithinAura = hunter.RegisterAura(core.Aura{
		Label:    "The Beast Within",
		ActionID: core.ActionID{SpellID: 34471},
		Duration: time.Second * 18,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		ClassMask:  HunterSpellsAll,
		FloatValue: -0.2,
	}).AttachMultiplicativePseudoStatBuff(
		&hunter.PseudoStats.DamageDealtMultiplier, 1.1,
	)

	hunter.Pet.BestialWrathAura.AttachDependentAura(hunter.TheBeastWithinAura)
}

func (hunter *Hunter) registerLethalShots() {
	if hunter.Talents.LethalShots == 0 {
		return
	}

	hunter.AddStat(stats.RangedCritPercent, float64(hunter.Talents.LethalShots))
}

func (hunter *Hunter) registerEfficiency() {
	if hunter.Talents.Efficiency == 0 {
		return
	}

	hunter.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		ClassMask:  HunterSpellsShotsAndStings,
		FloatValue: -0.02 * float64(hunter.Talents.Efficiency),
	})
}

func (hunter *Hunter) registerGoForTheThroat() {
	if hunter.Pet == nil || hunter.Talents.GoForTheThroat == 0 {
		return
	}

	amount := 25.0 * float64(hunter.Talents.GoForTheThroat)
	metricsSpellID := []int32{0, 34952, 34953}
	metrics := hunter.Pet.NewFocusMetrics(core.ActionID{SpellID: metricsSpellID[hunter.Talents.GoForTheThroat]})

	hunter.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Go for the Throat",
		Callback: core.CallbackOnSpellHitDealt,
		Outcome:  core.OutcomeCrit,
		ProcMask: core.ProcMaskRanged,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			hunter.Pet.AddFocus(sim, amount, metrics)
		},
	})
}

func (hunter *Hunter) registerImprovedArcaneShot() {
	if hunter.Talents.ImprovedArcaneShot == 0 {
		return
	}

	hunter.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_Cooldown_Flat,
		ClassMask: HunterSpellArcaneShot,
		TimeValue: -core.DurationFromSeconds(0.2 * float64(hunter.Talents.ImprovedArcaneShot)),
	})
}

func (hunter *Hunter) registerAimedShot() {
	hunter.AimedShot = hunter.RegisterRangedSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27065},
		SpellSchool:    core.SpellSchoolPhysical,
		ClassSpellMask: HunterSpellAimedShot,
		ProcMask:       core.ProcMaskRangedSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: 370,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: time.Millisecond * 3000,
			},
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 6,
			},
		},

		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 0.2*spell.RangedAttackPower(target) +
				hunter.AutoAttacks.Ranged().BaseDamage(sim) +
				hunter.talonOfAlarBonus() +
				870

			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}, true)
}

func (hunter *Hunter) registerRapidKilling() {
	if hunter.Talents.RapidKilling == 0 {
		return
	}

	hunter.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_Cooldown_Flat,
		ClassMask: HunterSpellRapidFire,
		TimeValue: -core.DurationFromSeconds(60 * float64(hunter.Talents.RapidKilling)),
	})
}

func (hunter *Hunter) registerImprovedStings() {
	if hunter.Talents.ImprovedStings == 0 {
		return
	}

	hunter.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		ClassMask:  HunterSpellSerpentSting,
		FloatValue: 0.06 * float64(hunter.Talents.ImprovedStings),
	})
}

func (hunter *Hunter) registerMortalShots() {
	if hunter.Talents.MortalShots == 0 {
		return
	}

	hunter.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_CritMultiplier_Flat,
		ProcMask:   core.ProcMaskRanged,
		FloatValue: 0.06 * float64(hunter.Talents.MortalShots),
	})
}

func (hunter *Hunter) registerBarrage() {
	if hunter.Talents.Barrage == 0 {
		return
	}

	hunter.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		ClassMask:  HunterSpellMultiShot | HunterSpellVolley,
		FloatValue: 0.04 * float64(hunter.Talents.Barrage),
	})
}

func (hunter *Hunter) registerCombatExperience() {
	if hunter.Talents.CombatExperience == 0 {
		return
	}

	hunter.MultiplyStat(stats.Agility, 1+0.01*float64(hunter.Talents.CombatExperience))
	hunter.MultiplyStat(stats.Intellect, 1+0.03*float64(hunter.Talents.CombatExperience))
}

func (hunter *Hunter) registerRangedWeaponSpecialization() {
	if hunter.Talents.RangedWeaponSpecialization == 0 {
		return
	}

	hunter.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		ProcMask:   core.ProcMaskRanged,
		FloatValue: 0.01 * float64(hunter.Talents.RangedWeaponSpecialization),
	})
}

func (hunter *Hunter) registerCarefulAim() {
	if hunter.Talents.CarefulAim == 0 {
		return
	}

	hunter.AddStatDependency(stats.Intellect, stats.RangedAttackPower, 0.15*float64(hunter.Talents.CarefulAim))
}

func (hunter *Hunter) registerImprovedBarrage() {
	if hunter.Talents.ImprovedBarrage == 0 {
		return
	}

	hunter.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ClassMask:  HunterSpellMultiShot,
		FloatValue: 4 * float64(hunter.Talents.ImprovedBarrage),
	})
}

func (hunter *Hunter) registerMasterMarksman() {
	if hunter.Talents.MasterMarksman == 0 {
		return
	}

	hunter.MultiplyStat(stats.RangedAttackPower, 1+0.02*float64(hunter.Talents.MasterMarksman))
}

func (hunter *Hunter) registerSlaying() {
	if hunter.Talents.MonsterSlaying == 0 && hunter.Talents.HumanoidSlaying == 0 {
		return
	}

	var beastMultiplier float64 = 1.0 + 0.01*float64(hunter.Talents.MonsterSlaying)
	var humanoidMultiplier float64 = 1.0 + 0.01*float64(hunter.Talents.HumanoidSlaying)
	hunter.Env.RegisterPostFinalizeEffect(func() {
		for _, at := range hunter.AttackTables {
			if at.Defender.MobType == proto.MobType_MobTypeHumanoid {
				at.DamageDealtMultiplier *= humanoidMultiplier
				at.CritMultiplier *= humanoidMultiplier
			} else if slices.Contains([]proto.MobType{proto.MobType_MobTypeBeast, proto.MobType_MobTypeGiant, proto.MobType_MobTypeDragonkin}, at.Defender.MobType) {
				at.DamageDealtMultiplier *= beastMultiplier
				at.CritMultiplier *= beastMultiplier
			}
		}
	})
}

func (hunter *Hunter) registerHawkEye() {
	if hunter.Talents.HawkEye == 0 {
		return
	}

	bonusRange := float64(hunter.Talents.HawkEye) * 2
	ranged := hunter.AutoAttacks.Ranged()

	if ranged != nil {
		ranged.MaxRange += bonusRange
	}

	hunter.AddStaticMod(core.SpellModConfig{
		Kind:     core.SpellMod_Custom,
		ProcMask: core.ProcMaskRanged,
		ApplyCustom: func(mod *core.SpellMod, spell *core.Spell) {
			if spell.MaxRange > 0 {
				spell.MaxRange += bonusRange
			}
		},
		RemoveCustom: func(mod *core.SpellMod, spell *core.Spell) {
			if spell.MaxRange > 0 {
				spell.MaxRange -= bonusRange
			}
		},
	})
}

func (hunter *Hunter) registerSavageStrikes() {
	if hunter.Talents.SavageStrikes == 0 {
		return
	}

	hunter.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ClassMask:  HunterSpellRaptorStrike,
		FloatValue: 10 * float64(hunter.Talents.SavageStrikes),
	})
}

func (hunter *Hunter) registerSurvivalist() {
	if hunter.Talents.Survivalist == 0 {
		return
	}

	hunter.MultiplyStat(stats.Health, 1+0.02*float64(hunter.Talents.Survivalist))
}

func (hunter *Hunter) registerSurefooted() {
	if hunter.Talents.Surefooted == 0 {
		return
	}

	hunter.AddStat(stats.PhysicalHitPercent, float64(hunter.Talents.Surefooted))
}

func (hunter *Hunter) registerSurvivalInstincts() {
	if hunter.Talents.SurvivalInstincts == 0 {
		return
	}

	hunter.MultiplyStat(stats.AttackPower, 1+0.02*float64(hunter.Talents.SurvivalInstincts))
	hunter.MultiplyStat(stats.RangedAttackPower, 1+0.02*float64(hunter.Talents.SurvivalInstincts))
}

func (hunter *Hunter) registerKillerInstinct() {
	if hunter.Talents.KillerInstinct == 0 {
		return
	}

	hunter.AddStat(stats.PhysicalCritPercent, float64(hunter.Talents.KillerInstinct))
}

func (hunter *Hunter) registerResourcefulness() {
	if hunter.Talents.Resourcefulness == 0 {
		return
	}

	hunter.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		ClassMask:  HunterSpellRaptorStrike,
		FloatValue: -0.2 * float64(hunter.Talents.Resourcefulness),
	})
}

func (hunter *Hunter) registerLightningReflexes() {
	if hunter.Talents.LightningReflexes == 0 {
		return
	}

	hunter.MultiplyStat(stats.Agility, 1+0.03*float64(hunter.Talents.LightningReflexes))
}

func (hunter *Hunter) registerThrillOfTheHunt() {
	if hunter.Talents.ThrillOfTheHunt == 0 {
		return
	}

	metrics := hunter.NewManaMetrics(core.ActionID{SpellID: 34720})

	hunter.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Thrill of the Hunt",
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskRangedSpecial,
		Outcome:    core.OutcomeCrit,
		ProcChance: float64(hunter.Talents.ThrillOfTheHunt) / 3,

		ExtraCondition: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			return spell.CurCast.Cost > 0
		},

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			hunter.AddMana(sim, spell.CurCast.Cost*0.4, metrics)
		},
	})
}

func (hunter *Hunter) registerExposeWeakness() {
	if hunter.Talents.ExposeWeakness == 0 {
		return
	}

	auraArray := hunter.NewEnemyAuraArray(func(unit *core.Unit) *core.Aura {
		return core.ExposeWeaknessAura(unit, func() float64 {
			return hunter.GetStat(stats.Agility)
		})
	})

	hunter.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Expose Weakness",
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskRanged,
		Outcome:    core.OutcomeCrit,
		ProcChance: float64(hunter.Talents.ExposeWeakness) / 3,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura := auraArray.Get(result.Target)
			aura.Activate(sim)
		},
	})
}

func (hunter *Hunter) registerMasterTactician() {
	if hunter.Talents.MasterTactician == 0 {
		return
	}

	statBuff := hunter.NewTemporaryStatsAura(
		"Master Tactician",
		core.ActionID{SpellID: 34837},
		stats.Stats{stats.PhysicalCritPercent: 2 * float64(hunter.Talents.MasterTactician)},
		time.Second*8)

	hunter.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Master Tactician",
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskRanged,
		Outcome:    core.OutcomeLanded,
		ProcChance: 0.06,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			statBuff.Activate(sim)
		},
	})
}

func (hunter *Hunter) registerReadiness() {
	if !hunter.Talents.Readiness {
		return
	}

	hunter.Readiness = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 23989},
		SpellSchool:    core.SpellSchoolPhysical,
		ClassSpellMask: HunterSpellReadiness,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Minute * 5,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !hunter.RapidFire.IsReady(sim) ||
				!hunter.MultiShot.IsReady(sim) ||
				!hunter.ArcaneShot.IsReady(sim) ||
				!hunter.KillCommand.IsReady(sim) ||
				!hunter.RaptorStrike.IsReady(sim)
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			hunter.RapidFire.CD.Reset()
			hunter.MultiShot.CD.Reset()
			hunter.ArcaneShot.CD.Reset()
			hunter.KillCommand.CD.Reset()
			hunter.RaptorStrike.CD.Reset()
		},
	})

	hunter.AddMajorCooldown(core.MajorCooldown{
		Spell: hunter.Readiness,
		Type:  core.CooldownTypeDPS,

		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return !hunter.RapidFire.IsReady(sim)
		},
	})
}
