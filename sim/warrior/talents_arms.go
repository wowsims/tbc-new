package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (war *Warrior) registerArmsTalents() {
	// Tier 1
	war.registerImprovedHeroicStrike()
	war.registerDeflection()
	war.registerImprovedRend()

	// Tier 2
	war.registerImprovedCharge()
	// Iron Will not implemented
	war.registerImprovedThunderClap()

	// Tier 3
	war.registerImprovedOverpower()
	war.registerAngerManagement()
	war.registerDeepWounds()

	// Tier 4
	war.registerTwoHandedWeaponSpecialization()
	war.registerImpale()

	// Tier 5
	war.registerPoleaxeSpecialization()
	war.registerDeathWish()
	war.registerSwordSpecialization()

	// Tier 6
	war.registerImprovedIntercept()
	// Improved Hamstring not implemented
	war.registerImprovedDisciplines()

	// Tier 7
	war.registerBloodFrenzy()
	war.registerMortalStrike()
	// Second Wind not implemented

	// Tier 8
	war.registerImprovedMortalStrike()

	// Tier 9
	war.registerEndlessRage()
}

/*
 * Arms
 */
func (war *Warrior) registerImprovedHeroicStrike() {
	if war.Talents.ImprovedHeroicStrike == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskHeroicStrike,
		Kind:      core.SpellMod_PowerCost_Flat,
		IntValue:  -war.Talents.ImprovedHeroicStrike,
	})
}
func (war *Warrior) registerDeflection() {
	if war.Talents.Deflection == 0 {
		return
	}

	war.PseudoStats.BaseParryChance += float64(war.Talents.Deflection)
}

func (war *Warrior) registerImprovedRend() {
	if war.Talents.ImprovedRend == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask:  SpellMaskRend,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.25 * float64(war.Talents.ImprovedRend),
	})
}

func (war *Warrior) registerImprovedCharge() {
	if war.Talents.ImprovedCharge == 0 {
		return
	}

	war.ChargeRageGain += 3.0 + float64(war.Talents.ImprovedCharge)
}

func (war *Warrior) registerImprovedThunderClap() {
	if war.Talents.ImprovedThunderClap == 0 {
		return
	}

	// Slowing effect implemented in core/debuffs.go

	rageCostReduction := []int32{0, 1, 2, 4}[war.Talents.ImprovedThunderClap]
	damageGain := []float64{0, 0.4, 0.7, 1.0}[war.Talents.ImprovedThunderClap]

	war.AddStaticMod(core.SpellModConfig{
		ClassMask:  SpellMaskThunderClap,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: damageGain,
	})

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskThunderClap,
		Kind:      core.SpellMod_PowerCost_Flat,
		IntValue:  -rageCostReduction,
	})
}

func (war *Warrior) registerImprovedOverpower() {
	if war.Talents.ImprovedOverpower == 0 {
		return
	}

	core.MakePermanent(war.RegisterAura(core.Aura{
		Label:    "Improved Overpower",
		ActionID: core.ActionID{SpellID: 12963}.WithTag(war.Talents.ImprovedOverpower),
	})).AttachSpellMod(core.SpellModConfig{
		ClassMask:  SpellMaskOverpower,
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 25 * float64(war.Talents.ImprovedOverpower),
	})
}

func (war *Warrior) registerAngerManagement() {
	if !war.Talents.AngerManagement {
		return
	}

	rageMetrics := war.NewRageMetrics(core.ActionID{SpellID: 12296})

	war.RegisterResetEffect(func(sim *core.Simulation) {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period: time.Second * 3,
			OnAction: func(sim *core.Simulation) {
				if sim.CurrentTime > 0 {
					war.AddRage(sim, 1, rageMetrics)
				}
			},
		})
	})
}

func (war *Warrior) registerDeepWounds() {
	if war.Talents.DeepWounds == 0 {
		return
	}

	war.DeepWounds = war.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 12867},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagIgnoreResists,

		DamageMultiplier: 1,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "DeepWounds",
			},
			NumberOfTicks: 6,
			TickLength:    time.Second * 3,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				baseDamage := dot.Unit.AutoAttacks.MH().CalculateAverageWeaponDamage(dot.Spell.MeleeAttackPower())
				dot.SnapshotPhysical(target, baseDamage/float64(dot.HastedTickCount())*0.2*float64(war.Talents.DeepWounds))
			},

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Deactivate(sim)
			spell.Dot(target).Apply(sim)
		},
	})

	war.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Deep Wounds - Trigger",
		TriggerImmediately: true,
		ProcMaskExclude:    core.ProcMaskEmpty,
		Outcome:            core.OutcomeCrit,
		Callback:           core.CallbackOnSpellHitDealt,
		ExtraCondition: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return spell.SpellSchool.Matches(core.SpellSchoolPhysical)
		},
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			war.DeepWounds.Cast(sim, result.Target)
		},
	})

}

func (war *Warrior) registerTwoHandedWeaponSpecialization() {
	if war.Talents.TwoHandedWeaponSpecialization == 0 {
		return
	}

	weaponMod := war.AddDynamicMod(core.SpellModConfig{
		ClassMask:  SpellMaskDirectDamageSpells,
		School:     core.SpellSchoolPhysical,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.02 * float64(war.Talents.TwoHandedWeaponSpecialization),
	})

	if war.GetHandType() == proto.HandType_HandTypeTwoHand {
		weaponMod.Activate()
	}

	war.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
		if war.GetHandType() == proto.HandType_HandTypeTwoHand {
			weaponMod.Activate()
		} else {
			weaponMod.Deactivate()
		}
	})
}

func (war *Warrior) registerImpale() {
	if war.Talents.Impale == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask:  SpellMaskDamageSpells,
		Kind:       core.SpellMod_CritMultiplier_Flat,
		FloatValue: 0.1 * float64(war.Talents.Impale),
	})
}

func (war *Warrior) registerPoleaxeSpecialization() {
	if war.Talents.PoleaxeSpecialization == 0 {
		return
	}

	isPolearmOrAxe := func(handItem *core.Item) bool {
		return handItem != nil && (handItem.WeaponType == proto.WeaponType_WeaponTypeAxe || handItem.WeaponType == proto.WeaponType_WeaponTypePolearm)
	}

	mhCritMod := war.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ProcMask:   core.ProcMaskMeleeMH,
		FloatValue: 1 * float64(war.Talents.PoleaxeSpecialization),
	})

	ohCritMod := war.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ProcMask:   core.ProcMaskMeleeOH,
		FloatValue: 1 * float64(war.Talents.PoleaxeSpecialization),
	})

	handleEquippedWeapons := func() {
		if isPolearmOrAxe(war.MainHand()) {
			mhCritMod.Activate()
		} else {
			mhCritMod.Deactivate()
		}

		if isPolearmOrAxe(war.OffHand()) {
			ohCritMod.Activate()
		} else {
			ohCritMod.Deactivate()
		}
	}

	war.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
		handleEquippedWeapons()
	})
}

func (war *Warrior) registerDeathWish() {
	if !war.Talents.DeathWish {
		return
	}

	actionID := core.ActionID{SpellID: 12292}

	deathWishAura := war.RegisterAura(core.Aura{
		Label:    "Death Wish",
		ActionID: actionID,
		Duration: time.Second * 30,
	}).
		AttachMultiplicativePseudoStatBuff(
			&war.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical], 1.2,
		).
		AttachMultiplicativePseudoStatBuff(
			&war.PseudoStats.DamageTakenMultiplier, 1.05,
		)

	deathWishSpell := war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		ClassSpellMask: SpellMaskDeathWish,

		RageCost: core.RageCostOptions{
			Cost: 10,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			deathWishAura.Activate(sim)
			war.WaitUntil(sim, sim.CurrentTime+core.GCDDefault)
		},

		RelatedSelfBuff: deathWishAura,
	})

	war.AddMajorCooldown(core.MajorCooldown{
		Spell: deathWishSpell,
		Type:  core.CooldownTypeDPS,
	})
}

func (war *Warrior) registerSwordSpecialization() {
	if war.Talents.SwordSpecialization == 0 {
		return
	}

	var swordSpecializationSpell *core.Spell
	procChance := 0.01 * float64(war.Talents.SwordSpecialization)

	newSwordSpecializationDPM := func() *core.DynamicProcManager {
		return war.NewFixedProcChanceManager(
			procChance,
			war.GetProcMaskForTypes(proto.WeaponType_WeaponTypeSword),
		)
	}

	dpm := newSwordSpecializationDPM()

	procTrigger := war.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Sword Specialization",
		DPM:                dpm,
		ICD:                time.Millisecond * 500,
		TriggerImmediately: true,
		Outcome:            core.OutcomeLanded,
		Callback:           core.CallbackOnSpellHitDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// OH WW hits can't proc this
			if spell.Matches(SpellMaskWhirlwindOh) {
				return
			}
			war.AutoAttacks.MaybeReplaceMHSwing(sim, swordSpecializationSpell).Cast(sim, result.Target)
		},
	})

	procTrigger.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
		config := *war.AutoAttacks.MHConfig()
		config.ActionID = core.ActionID{SpellID: 12281}
		swordSpecializationSpell = war.GetOrRegisterSpell(config)
	})

	war.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
		dpm = newSwordSpecializationDPM()
	})
}

func (war *Warrior) registerImprovedIntercept() {
	if war.Talents.ImprovedIntercept == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskIntercept,
		Kind:      core.SpellMod_Cooldown_Flat,
		TimeValue: -time.Second * time.Duration(5*war.Talents.ImprovedIntercept),
	})
}

func (war *Warrior) registerImprovedDisciplines() {
	if war.Talents.ImprovedDisciplines == 0 {
		return
	}

	cooldownReduction := []time.Duration{
		0,
		4 * time.Minute,
		7 * time.Minute,
		10 * time.Minute,
	}[war.Talents.ImprovedDisciplines]

	durationIncrease := time.Second * time.Duration(2*war.Talents.ImprovedDisciplines)

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskRetaliation | SpellMaskRecklessness | SpellMaskShieldWall,
		Kind:      core.SpellMod_Cooldown_Flat,
		TimeValue: -cooldownReduction,
	})

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskRetaliation | SpellMaskRecklessness | SpellMaskShieldWall,
		Kind:      core.SpellMod_BuffDuration_Flat,
		TimeValue: durationIncrease,
	})
}

func (war *Warrior) registerBloodFrenzy() {
	if war.Talents.BloodFrenzy == 0 {
		return
	}

	bfAuras := war.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.BloodFrenzyAura(target, war.Talents.BloodFrenzy)
	})

	war.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Blood Frenzy",
		ClassSpellMask: SpellMaskRend | SpellMaskDeepWounds,
		Outcome:        core.OutcomeLanded,
		Callback:       core.CallbackOnSpellHitDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura := bfAuras.Get(result.Target)
			aura.Duration = spell.Dot(result.Target).RemainingDuration(sim)
			aura.Activate(sim)
		},
	})
}

func (war *Warrior) registerMortalStrike() {
	if !war.Talents.MortalStrike {
		return
	}

	war.MortalStrike = war.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 12294},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics,
		ClassSpellMask: SpellMaskMortalStrike,
		MaxRange:       core.MaxMeleeRange,

		RageCost: core.RageCostOptions{
			Cost:   30,
			Refund: 0.8,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Second * 6,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 210 + spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}

func (war *Warrior) registerImprovedMortalStrike() {
	if war.Talents.ImprovedMortalStrike == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskMortalStrike,
		Kind:      core.SpellMod_Cooldown_Flat,
		TimeValue: -time.Second * time.Duration(0.2*float64(war.Talents.ImprovedMortalStrike)),
	})

	war.AddStaticMod(core.SpellModConfig{
		ClassMask:  SpellMaskMortalStrike,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: float64(war.Talents.ImprovedMortalStrike),
	})
}

func (war *Warrior) registerEndlessRage() {
	if !war.Talents.EndlessRage {
		return
	}

	war.MultiplyAutoAttackRageGen(1.25)
}
