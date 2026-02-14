package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (rogue *Rogue) registerSubtletyTalents() {
	// Tier 1
	// Master of Deception NYI
	rogue.registerOpportunity()

	// Tier 2
	// None in this tier implemented

	// Tier 3
	rogue.registerInitiative()
	rogue.registerGhostlyStrike()
	rogue.registerImprovedAmbush()

	// Tier 4
	// Setup NYI
	rogue.registerElusiveness()
	rogue.registerSerratedBlades()

	// Tier 5
	// Heightened Senses NYI
	rogue.registerPreparation()
	rogue.registerDirtyDeeds()
	rogue.registerHemorrhage()

	// Tier 6
	rogue.registerMasterOfSubtlety()
	rogue.registerDeadliness()

	// Tier 7
	// Enveloping Shadows NYI
	rogue.registerPremeditation()
	// Cheat Death NYI

	// Tier 8
	rogue.registerSinisterCalling()

	// Tier 9
	rogue.registerShadowstep()
}

func (rogue *Rogue) registerOpportunity() {
	if rogue.Talents.Opportunity == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellBackstab | RogueSpellMutilate | RogueSpellAmbush,
		FloatValue: 0.4 * float64(rogue.Talents.Opportunity),
	})
}

func (rogue *Rogue) registerInitiative() {
	if rogue.Talents.Initiative == 0 {
		return
	}

	initMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 13980})

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Initiative Trigger",
		ActionID:       core.ActionID{SpellID: 13980},
		ProcChance:     0.25 * float64(rogue.Talents.Initiative),
		Callback:       core.CallbackOnSpellHitDealt,
		Outcome:        core.OutcomeCrit,
		ClassSpellMask: RogueSpellGarrote | RogueSpellAmbush,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			rogue.AddComboPoints(sim, 1, initMetrics)
		},
	})
}

func (rogue *Rogue) registerGhostlyStrike() {
	if !rogue.Talents.GhostlyStrike {
		return
	}

	pointMetric := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14278})
	rogue.GhostlyStrike = rogue.GetOrRegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 14278},
		ClassSpellMask: RogueSpellGhostlyStrike,
		SpellSchool:    core.SpellSchoolPhysical,
		Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		MaxRange:       core.MaxMeleeRange,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Second * 20,
			},
			IgnoreHaste: true,
		},
		EnergyCost: core.EnergyCostOptions{
			Cost:   40,
			Refund: 0.8,
		},

		DamageMultiplier: 1.25,
		CritMultiplier:   rogue.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)

			// Dodge Aura NYI

			baseDamage := rogue.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			if result.Landed() {
				rogue.AddComboPoints(sim, 1, pointMetric)
			}
		},
	})
}

func (rogue *Rogue) registerImprovedAmbush() {
	if rogue.Talents.ImprovedAmbush == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ClassMask:  RogueSpellAmbush,
		FloatValue: 15 * float64(rogue.Talents.ImprovedAmbush),
	})
}

func (rogue *Rogue) registerElusiveness() {
	if rogue.Talents.Elusiveness == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_Cooldown_Flat,
		ClassMask: RogueSpellVanish,
		TimeValue: time.Second * 45 * time.Duration(rogue.Talents.Elusiveness),
	})
}

func (rogue *Rogue) registerSerratedBlades() {
	if rogue.Talents.SerratedBlades == 0 {
		return
	}

	rogue.AddStat(stats.ArmorPenetration, 186*float64(rogue.Talents.SerratedBlades))
	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellRupture,
		FloatValue: 0.1 * float64(rogue.Talents.SerratedBlades),
	})
}

func (rogue *Rogue) registerPreparation() {
	if !rogue.Talents.Preparation {
		return
	}

	rogue.Preparation = rogue.GetOrRegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 14185},
		ClassSpellMask: RogueSpellPreparation,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Minute * 10,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if rogue.ColdBlood != nil {
				rogue.ColdBlood.CD.Set(0)
			}
			if rogue.Shadowstep != nil {
				rogue.Shadowstep.CD.Set(0)
			}
			if rogue.Premeditation != nil {
				rogue.Premeditation.CD.Set(0)
			}
			if rogue.Vanish != nil {
				rogue.Vanish.CD.Set(0)
			}
		},
	})

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.Preparation,
		Type:  core.CooldownTypeDPS,
	})
}

func (rogue *Rogue) registerDirtyDeeds() {
	if rogue.Talents.DirtyDeeds == 0 {
		return
	}

	ddAura := rogue.GetOrRegisterAura(core.Aura{
		Label:    "Dirty Deeds",
		ActionID: core.ActionID{SpellID: 14083},
		Duration: core.NeverExpires,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellsAll,
		FloatValue: 0.1 * float64(rogue.Talents.DirtyDeeds),
	})

	rogue.RegisterResetEffect(func(sim *core.Simulation) {
		ddAura.Deactivate(sim)
		sim.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
			if isExecute == 35 {
				ddAura.Activate(sim)
			}
		})
	})
}

func (rogue *Rogue) registerHemorrhage() {
	if !rogue.Talents.Hemorrhage {
		return
	}

	pointMetric := rogue.NewComboPointMetrics(core.ActionID{SpellID: 26864})
	rogue.Hemorrhage = rogue.GetOrRegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26864},
		ClassSpellMask: RogueSpellHemorrhage,
		SpellSchool:    core.SpellSchoolPhysical,
		Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		MaxRange:       core.MaxMeleeRange,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		EnergyCost: core.EnergyCostOptions{
			Cost:   35,
			Refund: 0.8,
		},

		DamageMultiplier: 1.1,
		CritMultiplier:   rogue.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)

			// Making an executive decision for modeling Hemorrhage as a solo player.
			// Since the aura won't ever get fully consumed while solo, I'm adding the full value of all 10 stacks into the baseDamage.
			// This more accurately models each individual Hemo's DPS contribution without needing to somehow consume the stacks.
			// It's also easier for me to just leave it that way :)

			baseDamage := rogue.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
			if result.Landed() {
				rogue.AddComboPoints(sim, 1, pointMetric)
			}
		},
	})
}

func (rogue *Rogue) registerMasterOfSubtlety() {
	if rogue.Talents.MasterOfSubtlety == 0 {
		return
	}

	rogue.MasterOfSubtletyAura = rogue.GetOrRegisterAura(core.Aura{
		Label:    "Master of Subtlety",
		ActionID: core.ActionID{SpellID: 31223},
		Duration: time.Second * 6,
	}).AttachAdditivePseudoStatBuff(&rogue.PseudoStats.DamageDealtMultiplier, 1.1)

	// Activated in stealth.go
}

func (rogue *Rogue) registerDeadliness() {
	if rogue.Talents.Deadliness == 0 {
		return
	}

	rogue.MultiplyStat(stats.AttackPower, 1+0.2*float64(rogue.Talents.Deadliness))
}

func (rogue *Rogue) registerPremeditation() {
	if !rogue.Talents.Premeditation {
		return
	}

	comboMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14183})
	shouldTimeout := false

	premedAura := rogue.RegisterAura(core.Aura{
		Label:    "Premed Timeout Aura",
		Duration: time.Second * 10,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			shouldTimeout = true
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagFinisher) && spell.ClassSpellMask == RogueSpellSliceAndDice {
				shouldTimeout = false
			}
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.Flags.Matches(SpellFlagFinisher) && result.Landed() {
				shouldTimeout = false
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			// Remove 2 points because no finisher was casted
			if shouldTimeout {
				rogue.AddComboPoints(sim, -2, comboMetrics)
				shouldTimeout = false
			}
		},
	})

	rogue.Premeditation = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 14183},
		Flags:          core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,
		ClassSpellMask: RogueSpellPremeditation,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				Cost: 0,
				GCD:  0,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Minute * 2,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return rogue.IsStealthed()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			rogue.AddComboPoints(sim, 2, comboMetrics)
			premedAura.Activate(sim)
		},
	})

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell:              rogue.Premeditation,
		Type:               core.CooldownTypeDPS,
		Priority:           core.CooldownPriorityLow,
		AllowSpellQueueing: true,
	})
}

func (rogue *Rogue) registerSinisterCalling() {
	if rogue.Talents.SinisterCalling == 0 {
		return
	}

	rogue.MultiplyStat(stats.Agility, 1+0.03*float64(rogue.Talents.SinisterCalling))
	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellHemorrhage | RogueSpellBackstab,
		FloatValue: 0.01 * float64(rogue.Talents.SinisterCalling),
	})
}

func (rogue *Rogue) registerShadowstep() {
	if !rogue.Talents.Shadowstep {
		return
	}

	actionID := core.ActionID{SpellID: 36554}

	rogue.ShadowstepAura = rogue.RegisterAura(core.Aura{
		Label:    "Shadowstep",
		ActionID: actionID,
		Duration: time.Second * 10,
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ClassSpellMask&RogueSpellsAll != 0 {
				aura.Deactivate(sim)
			}
		},
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellsAll,
		FloatValue: 0.2,
	})

	rogue.Shadowstep = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: RogueSpellShadowstep,

		Cast: core.CastConfig{
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Second * 30,
			},
		},
		EnergyCost: core.EnergyCostOptions{
			Cost: 10,
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// TODO: Teleport?
			spell.RelatedSelfBuff.Activate(sim)
		},
		RelatedSelfBuff: rogue.ShadowstepAura,
	})
}
