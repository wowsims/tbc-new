package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (rogue *Rogue) registerCombatTalents() {
	// Tier 1
	rogue.registerImprovedGouge()
	rogue.registerImprovedSinisterStrike()
	rogue.registerLightningReflexes()

	// Tier 2
	// Improved Slice and Dice implemented in slice_and_dice.go
	rogue.registerDeflection()
	// Improved Sprint NYI

	// Tier 3
	// None in this tier implemented

	// Tier 4
	// Improved Kick NYI
	rogue.registerDaggerSpecialization()
	rogue.registerDualWieldSpecialization()

	// Tier 5
	rogue.registerMaceSpecialization()
	rogue.registerBladeFlurry()
	rogue.registerSwordSpecialization()
	rogue.registerFistWeaponSpecialization()

	// Tier 6
	// Blade Twisting NYI
	rogue.registerWeaponExpertise()
	rogue.registerAggression()

	// Tier 7
	rogue.registerVitality()
	rogue.registerAdrenalineRush()
	// Nerves of Steel NYI

	// Tier 8
	rogue.registerCombatPotency()

	// Tier 9
	rogue.registerSurpriseAttacks()
}

func (rogue *Rogue) registerImprovedGouge() {
	if rogue.Talents.ImprovedGouge == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_Cooldown_Flat,
		ClassMask: RogueSpellGouge,
		TimeValue: time.Millisecond * 500 * time.Duration(rogue.Talents.ImprovedGouge),
	})
}

func (rogue *Rogue) registerImprovedSinisterStrike() {
	if rogue.Talents.ImprovedSinisterStrike == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_PowerCost_Flat,
		ClassMask: RogueSpellSinisterStrike,
		IntValue:  []int32{0, -3, -5}[rogue.Talents.ImprovedSinisterStrike],
	})
}

func (rogue *Rogue) registerLightningReflexes() {
	if rogue.Talents.LightningReflexes == 0 {
		return
	}

	rogue.AddStat(stats.DodgeRating, float64(rogue.Talents.LightningReflexes)*core.DodgeRatingPerDodgePercent)
}

func (rogue *Rogue) registerDeflection() {
	if rogue.Talents.Deflection == 0 {
		return
	}

	rogue.AddStat(stats.ParryRating, float64(rogue.Talents.Deflection)*core.ParryRatingPerParryPercent)
}

func (rogue *Rogue) registerDaggerSpecialization() {
	if rogue.Talents.DaggerSpecialization == 0 {
		return
	}

	if rogue.HasDagger(true) {
		rogue.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_BonusCrit_Percent,
			ProcMask:   core.ProcMaskMeleeMH,
			FloatValue: 1.0 * float64(rogue.Talents.DaggerSpecialization),
		})
	}
	if rogue.HasDagger(false) {
		rogue.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_BonusCrit_Percent,
			ProcMask:   core.ProcMaskMeleeOH,
			FloatValue: 1.0 * float64(rogue.Talents.DaggerSpecialization),
		})
	}
}

func (rogue *Rogue) registerDualWieldSpecialization() {
	if rogue.Talents.DualWieldSpecialization == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ProcMask:   core.ProcMaskMeleeOH,
		FloatValue: 0.1 * float64(rogue.Talents.DualWieldSpecialization),
	})
}

func (rogue *Rogue) registerMaceSpecialization() {
	if rogue.Talents.MaceSpecialization == 0 {
		return
	}

	if rogue.GetMHWeapon() != nil && rogue.GetMHWeapon().WeaponType == proto.WeaponType_WeaponTypeMace {
		rogue.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_CritMultiplier_Flat,
			ProcMask:   core.ProcMaskMeleeMH,
			FloatValue: 0.01 * float64(rogue.Talents.MaceSpecialization),
		})
	}
	if rogue.GetOHWeapon() != nil && rogue.GetOHWeapon().WeaponType == proto.WeaponType_WeaponTypeMace {
		rogue.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_CritMultiplier_Flat,
			ProcMask:   core.ProcMaskMeleeOH,
			FloatValue: 0.01 * float64(rogue.Talents.MaceSpecialization),
		})
	}
}

func (rogue *Rogue) registerBladeFlurry() {
	if !rogue.Talents.BladeFlurry {
		return
	}

	var curDmg float64
	bfHit := rogue.GetOrRegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 22482},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskEmpty, // No proc mask, so it won't proc itself.
		Flags:       core.SpellFlagIgnoreResists | core.SpellFlagIgnoreModifiers | core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, curDmg, spell.OutcomeAlwaysHit)
		},
	})

	rogue.BladeFlurryAura = rogue.GetOrRegisterAura(core.Aura{
		Label:    "Blade Flurry",
		ActionID: core.ActionID{SpellID: 13877},
		Duration: time.Second * 15,

		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if sim.ActiveTargetCount() < 2 {
				return
			}

			if result.Damage == 0 || !spell.ProcMask.Matches(core.ProcMaskMelee) {
				return
			}

			curDmg = result.Damage
			bfHit.Cast(sim, rogue.Env.NextActiveTargetUnit(result.Target))
			bfHit.SpellMetrics[result.Target.UnitIndex].Casts--
		},
	}).AttachMultiplyAttackSpeed(1.2)

	rogue.BladeFlurry = rogue.GetOrRegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 13877},
		ClassSpellMask: RogueSpellBladeFlurry,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Minute * 2,
			},
			IgnoreHaste: true,
		},
		EnergyCost: core.EnergyCostOptions{
			Cost: 25,
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BladeFlurryAura.Activate(sim)
		},
	})

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.BladeFlurry,
		Type:  core.CooldownTypeDPS,
	})
}

func (rogue *Rogue) registerSwordSpecialization() {
	if rogue.Talents.SwordSpecialization == 0 {
		return
	}

	var procMask core.ProcMask
	if rogue.GetMHWeapon() != nil && rogue.GetMHWeapon().WeaponType == proto.WeaponType_WeaponTypeSword {
		procMask |= core.ProcMaskMeleeMH
	}
	if rogue.GetOHWeapon() != nil && rogue.GetOHWeapon().WeaponType == proto.WeaponType_WeaponTypeSword {
		procMask |= core.ProcMaskMeleeOH
	}

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Sword Spec Proc Trigger",
		ActionID:           core.ActionID{SpellID: 13964},
		ProcChance:         0.01 * float64(rogue.Talents.SwordSpecialization),
		Callback:           core.CallbackOnSpellHitDealt,
		Outcome:            core.OutcomeLanded,
		ProcMask:           procMask,
		ICD:                time.Millisecond * 500,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			rogue.AutoAttacks.MHAuto().Cast(sim, result.Target)
		},
	})
}

func (rogue *Rogue) registerFistWeaponSpecialization() {
	if rogue.Talents.FistWeaponSpecialization == 0 {
		return
	}

	if rogue.GetMHWeapon() != nil && rogue.GetMHWeapon().WeaponType == proto.WeaponType_WeaponTypeFist {
		rogue.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_BonusCrit_Percent,
			ProcMask:   core.ProcMaskMeleeMH,
			FloatValue: 1 * float64(rogue.Talents.MaceSpecialization),
		})
	}
	if rogue.GetOHWeapon() != nil && rogue.GetOHWeapon().WeaponType == proto.WeaponType_WeaponTypeFist {
		rogue.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_BonusCrit_Percent,
			ProcMask:   core.ProcMaskMeleeOH,
			FloatValue: 1 * float64(rogue.Talents.MaceSpecialization),
		})
	}
}

func (rogue *Rogue) registerWeaponExpertise() {
	if rogue.Talents.WeaponExpertise == 0 {
		return
	}

	rogue.AddStat(stats.ExpertiseRating, core.ExpertisePerQuarterPercentReduction*5*float64(rogue.Talents.WeaponExpertise))
}

func (rogue *Rogue) registerAggression() {
	if rogue.Talents.Aggression == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellSinisterStrike | RogueSpellBackstab | RogueSpellEviscerate,
		FloatValue: 0.06 * float64(rogue.Talents.Aggression),
	})
}

func (rogue *Rogue) registerVitality() {
	if rogue.Talents.Vitality == 0 {
		return
	}

	rogue.MultiplyStat(stats.Agility, 1+0.01*float64(rogue.Talents.Vitality))
	rogue.MultiplyStat(stats.Stamina, 1+0.02*float64(rogue.Talents.Vitality))
}

func (rogue *Rogue) registerAdrenalineRush() {
	if !rogue.Talents.AdrenalineRush {
		return
	}

	rogue.AdrenalineRushAura = rogue.GetOrRegisterAura(core.Aura{
		Label:    "Adrenaline Rush",
		ActionID: core.ActionID{SpellID: 13750},
		Duration: time.Second * 15,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			rogue.MultiplyEnergyRegenSpeed(sim, 2)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			rogue.MultiplyEnergyRegenSpeed(sim, 0.5)
		},
	})

	rogue.AdrenalineRush = rogue.GetOrRegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 13750},
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: RogueSpellAdrenalineRush,

		Cast: core.CastConfig{
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Minute * 5,
			},
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.AdrenalineRushAura.Activate(sim)
		},
	})

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.AdrenalineRush,
		Type:  core.CooldownTypeDPS,
	})
}

func (rogue *Rogue) registerCombatPotency() {
	if rogue.Talents.CombatPotency == 0 {
		return
	}

	potencyMetrics := rogue.NewEnergyMetrics(core.ActionID{SpellID: 35553})

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Combat Potency Trigger",
		ActionID:   core.ActionID{SpellID: 35553},
		ProcChance: 0.2,
		Callback:   core.CallbackOnSpellHitDealt,
		Outcome:    core.OutcomeLanded,
		ProcMask:   core.ProcMaskMeleeOH,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			rogue.AddEnergy(sim, 3.0*float64(rogue.Talents.CombatPotency), potencyMetrics)
		},
	})
}

func (rogue *Rogue) registerSurpriseAttacks() {
	if !rogue.Talents.SurpriseAttacks {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellSinisterStrike | RogueSpellBackstab | RogueSpellShiv | RogueSpellGouge,
		FloatValue: 0.1,
	})

	// Finisher Dodge applied in individual spells
}
