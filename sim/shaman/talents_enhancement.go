package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (shaman *Shaman) ApplyEnhancementTalents() {

	shaman.applyAncestralKnowledge()
	shaman.applyDualWield()
	shaman.applyDualWieldSpecialization()
	shaman.applyElementalWeapons()
	shaman.applyEnhancingTotems()
	shaman.applyFlurry()
	shaman.applyImprovedLightningShield()
	shaman.applyImprovedWeaponTotems()
	shaman.applyMentalQuickness()
	shaman.applyShamanisticFocus()
	shaman.applyShamanisticRage()
	shaman.applySpiritWeapons()
	shaman.applyStormstrike()
	shaman.applyThunderingStrikes()
	shaman.applyUnleashedRage()
	shaman.applyWeaponMastery()
}

func (shaman *Shaman) applyAncestralKnowledge() {
	if shaman.Talents.AncestralKnowledge == 0 {
		return
	}
	shaman.MultiplyStat(stats.Mana, 1+(0.01*float64(shaman.Talents.AncestralKnowledge)))
}

func (shaman *Shaman) applyDualWield() {
	if !shaman.Talents.DualWield {
		return
	}
	// TODO ?
}

func (shaman *Shaman) applyDualWieldSpecialization() {
	if shaman.Talents.DualWieldSpecialization == 0 {
		return
	}
	if shaman.AutoAttacks.IsDualWielding {
		shaman.AddStaticMod(core.SpellModConfig{
			Kind:       core.SpellMod_BonusHit_Percent,
			FloatValue: 2 * float64(shaman.Talents.DualWieldSpecialization),
			ProcMask:   core.ProcMaskMeleeOrRanged,
		})
	}
	//TODO weapon swap
}

func (shaman *Shaman) applyElementalWeapons() {
	if shaman.Talents.ElementalWeapons == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: [4]float64{0.0, 0.07, 0.14, 0.2}[shaman.Talents.ElementalWeapons],
		ClassMask:  SpellMaskRockbiterWeapon,
	})
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: [4]float64{0.0, 0.13, 0.27, 0.4}[shaman.Talents.ElementalWeapons],
		ClassMask:  SpellMaskWindfuryWeapon,
	})
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.05 * float64(shaman.Talents.ElementalWeapons),
		ClassMask:  SpellMaskFlametongueWeapon | SpellMaskFrostbrandWeapon,
	})
}

func (shaman *Shaman) applyEnhancingTotems() {
	if shaman.Talents.EnhancingTotems == 0 {
		return
	}
	// TODO
}

func (shaman *Shaman) applyFlurry() {
	if shaman.Talents.Flurry == 0 {
		return
	}

	flurryICD := &core.Cooldown{
		Timer:    shaman.NewTimer(),
		Duration: 500 * time.Millisecond,
	}

	flurryAura := shaman.RegisterAura(core.Aura{
		ActionID:  core.ActionID{SpellID: 16284},
		Label:     "Flurry",
		Duration:  time.Second * 15,
		MaxStacks: 3,
	}).AttachMultiplyMeleeSpeed(0.05 + 0.05*float64(shaman.Talents.Flurry))

	shaman.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Flurry Trigger",
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMelee | core.ProcMaskMeleeProc,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Outcome.Matches(core.OutcomeCrit) {
				flurryAura.Activate(sim)
				flurryAura.SetStacks(sim, 3)
				return
			}

			// Remove a stack.
			if flurryAura.IsActive() && spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && flurryICD.IsReady(sim) {
				flurryICD.Use(sim)
				flurryAura.RemoveStack(sim)
			}
		},
	})

}

func (shaman *Shaman) applyImprovedLightningShield() {
	if shaman.Talents.ImprovedLightningShield == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.05 * float64(shaman.Talents.ImprovedLightningShield),
	})
}

func (shaman *Shaman) applyImprovedWeaponTotems() {
	if shaman.Talents.ImprovedWeaponTotems == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.06 * float64(shaman.Talents.ImprovedWeaponTotems),
		ClassMask:  SpellMaskFlametongueTotem,
	})
	//TODO WF totem bonus
}

func (shaman *Shaman) applyMentalQuickness() {
	if shaman.Talents.MentalQuickness == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -0.02 * float64(shaman.Talents.MentalQuickness),
		SpellFlag:  SpellFlagInstant,
	})
	shaman.AddStatDependency(stats.AttackPower, stats.SpellDamage, 0.1*float64(shaman.Talents.MentalQuickness))
}

func (shaman *Shaman) applyShamanisticFocus() {
	if !shaman.Talents.ShamanisticFocus {
		return
	}
	sfAura := shaman.RegisterAura(core.Aura{
		Label:    "Focused",
		ActionID: core.ActionID{SpellID: 43339},
		Duration: time.Second * 15,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -0.6,
		ClassMask:  SpellMaskShock,
	})

	shaman.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Shamanistic Focus Trigger",
		Callback: core.CallbackOnSpellHitDealt,
		ProcMask: core.ProcMaskMeleeOrMeleeProc,
		Outcome:  core.OutcomeCrit,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			sfAura.Activate(sim)
		},
	})

	shaman.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Shamanistic Focus Untrigger",
		Callback: core.CallbackOnCastComplete,
		Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			if !spell.Matches(SpellMaskShock) {
				return
			}
			sfAura.Deactivate(sim)
		},
	})
}

func (shaman *Shaman) applyShamanisticRage() {
	if !shaman.Talents.ShamanisticRage {
		return
	}
	actionId := core.ActionID{SpellID: 30823}
	srManaMetric := shaman.NewManaMetrics(actionId)
	shamRageAura := shaman.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Shamanistic Rage",
		MetricsActionID:    actionId,
		Duration:           time.Second * 15,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMeleeWhiteHit,
		Outcome:            core.OutcomeLanded,
		RequireDamageDealt: true,
		DPM:                shaman.NewLegacyPPMManager(15.0, core.ProcMaskMeleeWhiteHit),
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			shaman.AddMana(sim, 0.3*shaman.GetAttackPowerValue(spell), srManaMetric)
		},
	})

	shaman.RegisterSpell(core.SpellConfig{
		ActionID:       actionId,
		SpellSchool:    core.SpellSchoolPhysical,
		Flags:          SpellFlagInstant,
		ClassSpellMask: SpellMaskShamanisticRage,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * 120,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			shamRageAura.Activate(sim)
		},
	})
}

func (shaman *Shaman) applySpiritWeapons() {
	if !shaman.Talents.SpiritWeapons {
		return
	}
	//TODO
}

func (shaman *Shaman) applyStormstrike() {
	if !shaman.Talents.Stormstrike {
		return
	}
	// TODO Need to merge Dora's changes
}

func (shaman *Shaman) applyThunderingStrikes() {
	if shaman.Talents.ThunderingStrikes == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 1 * float64(shaman.Talents.ThunderingStrikes),
		ProcMask:   core.ProcMaskMelee,
	})
}

func (shaman *Shaman) applyUnleashedRage() {
	if shaman.Talents.UnleashedRage == 0 {
		return
	}
	// TODO
}

func (shaman *Shaman) applyWeaponMastery() {
	if shaman.Talents.WeaponMastery == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.02 * float64(shaman.Talents.WeaponMastery),
		ProcMask:   core.ProcMaskMelee,
	})
}
