package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (priest *Priest) applyForceOfWill() {
	if priest.Talents.ForceOfWill == 0 {
		return
	}
	// +1% damage per rank
	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.01 * float64(priest.Talents.ForceOfWill),
		ClassMask:  PriestSpellsAll,
	})
	// +1% crit per rank
	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 1.0 * float64(priest.Talents.ForceOfWill),
		ClassMask:  PriestSpellsAll,
	})
}

func (priest *Priest) applyPowerInfusion() {
	if !priest.Talents.PowerInfusion {
		return
	}

	piAura := core.PowerInfusionAura(priest.GetCharacter(), 0)

	piSpell := priest.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 10060},
		SpellSchool: core.SpellSchoolHoly,
		Flags:       core.SpellFlagHelpful,
		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 16,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    priest.NewTimer(),
				Duration: core.PowerInfusionCD,
			},
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			piAura.Activate(sim)
		},
	})

	priest.AddMajorCooldown(core.MajorCooldown{
		Spell:    piSpell,
		Priority: core.CooldownPriorityBloodlust,
		Type:     core.CooldownTypeMana,
	})
}

func (priest *Priest) applyFocusedPower() {
	if priest.Talents.FocusedPower == 0 {
		return
	}
	// +2% hit per rank (2 ranks = 4%)
	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusHit_Percent,
		FloatValue: 2.0 * float64(priest.Talents.FocusedPower),
		ClassMask:  PriestSpellSmite | PriestSpellMindBlast,
	})
}

func (priest *Priest) applyEnlightenment() {
	if priest.Talents.Enlightenment == 0 {
		return
	}
	// +1% per rank
	multiplier := 1.0 + 0.01*float64(priest.Talents.Enlightenment)
	priest.MultiplyStat(stats.Stamina, multiplier)
	priest.MultiplyStat(stats.Intellect, multiplier)
	priest.MultiplyStat(stats.Spirit, multiplier)
}

func (priest *Priest) applyMentalStrength() {
	if priest.Talents.MentalStrength == 0 {
		return
	}
	// +2% mana per rank
	priest.MultiplyStat(stats.Mana, 1.0+0.02*float64(priest.Talents.MentalStrength))
}

func (priest *Priest) applySpiritualGuidance() {
	if priest.Talents.SpiritualGuidance == 0 {
		return
	}
	// 5% of Spirit added to spell damage per rank
	coeff := 0.05 * float64(priest.Talents.SpiritualGuidance)
	priest.AddStatDependency(stats.Spirit, stats.SpellDamage, coeff) // Only scaling damage for now since no healing sim....yet!
}

func (priest *Priest) applyDivineFury() {
	if priest.Talents.DivineFury == 0 {
		return
	}
	// -0.1s per rank
	priest.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Millisecond * time.Duration(-100*priest.Talents.DivineFury),
		ClassMask: PriestSpellSmite | PriestSpellHolyFire,
	})
}

func (priest *Priest) applySearingLight() {
	if priest.Talents.SearingLight == 0 {
		return
	}
	// +5% damage per rank
	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.05 * float64(priest.Talents.SearingLight),
		ClassMask:  PriestSpellSmite | PriestSpellHolyFire,
	})
}

func (priest *Priest) applySurgeOfLight() {
	if priest.Talents.SurgeOfLight == 0 {
		return
	}

	// Dynamic mods activated by the aura
	castTimeMod := priest.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_CastTime_Pct,
		FloatValue: -1.0, // -100% = instant cast
		ClassMask:  PriestSpellSmite,
	})
	manaCostMod := priest.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -1.0, // -100% = free
		ClassMask:  PriestSpellSmite,
	})
	critMod := priest.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: -100.0, // unable to crit
		ClassMask:  PriestSpellSmite,
	})

	solAura := priest.RegisterAura(core.Aura{
		Label:    "Surge of Light",
		ActionID: core.ActionID{SpellID: 33151},
		Duration: 10 * time.Second,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			castTimeMod.Activate()
			manaCostMod.Activate()
			critMod.Activate()
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			castTimeMod.Deactivate()
			manaCostMod.Deactivate()
			critMod.Deactivate()
		},
		// Consumed on next Smite cast
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Matches(PriestSpellSmite) {
				aura.Deactivate(sim)
			}
		},
	})

	// 25% proc chance at rank 1, 50% at rank 2
	procChance := 0.25 * float64(priest.Talents.SurgeOfLight)

	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Surge of Light Trigger",
		Callback:   core.CallbackOnSpellHitDealt,
		Outcome:    core.OutcomeCrit,
		ProcChance: procChance,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			solAura.Activate(sim)
		},
	})
}

func (priest *Priest) applySilentResolve() {
	if priest.Talents.SilentResolve == 0 {
		return
	}
	// -4% threat per rank for discipline and holy spells
	threatReduction := []float64{0, -0.04, -0.08, -0.12, -0.16, -0.20}[priest.Talents.SilentResolve]
	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_ThreatMultiplier_Pct,
		FloatValue: threatReduction,
		ClassMask:  PriestHolySpells,
	})
}

func (priest *Priest) applyHolyNova() {
	if !priest.Talents.HolyNova {
		return
	}
	HolyNovaRankMap.RegisterAll(priest.registerHolyNovaSpell)
}

func (priest *Priest) applyVampiricTouch() {
	if !priest.Talents.VampiricTouch {
		return
	}
	VampiricTouchRankMap.RegisterAll(priest.registerVampiricTouchSpell)
}

func (priest *Priest) applyMindFlay() {
	if !priest.Talents.MindFlay {
		return
	}
	MindFlayRankMap.RegisterAll(priest.registerMindFlaySpell)
}

func (priest *Priest) applyImprovedMindBlast() {
	if priest.Talents.ImprovedMindBlast == 0 {
		return
	}

	priest.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_Cooldown_Flat,
		TimeValue: time.Millisecond * time.Duration(-500*priest.Talents.ImprovedMindBlast),
		ClassMask: PriestSpellMindBlast,
	})
}

func (priest *Priest) applyInnerFocus() {
	if !priest.Talents.InnerFocus {
		return
	}

	critMod := priest.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 25.0,
		ClassMask:  PriestSpellsAll,
	})

	var innerFocusSpell *core.Spell
	priest.InnerFocusAura = priest.RegisterAura(core.Aura{
		Label:    "Inner Focus",
		ActionID: core.ActionID{SpellID: 14751},
		Duration: time.Hour,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SpellCostPercentModifier -= 100
			critMod.Activate()
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SpellCostPercentModifier += 100
			critMod.Deactivate()
			innerFocusSpell.CD.Use(sim)
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !spell.Matches(PriestSpellsAll) {
				return
			}
			aura.Deactivate(sim)
		},
	})

	innerFocusSpell = priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 14751},
		Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
		ClassSpellMask: PriestSpellFlagNone,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    priest.NewTimer(),
				Duration: time.Second * 180,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			priest.InnerFocusAura.Activate(sim)
		},
		RelatedSelfBuff: priest.InnerFocusAura,
	})

	priest.AddMajorCooldown(core.MajorCooldown{
		Spell: innerFocusSpell,
		Type:  core.CooldownTypeMana,
	})
}

func (priest *Priest) applyMeditation() {
	if priest.Talents.Meditation == 0 {
		return
	}

	priest.PseudoStats.SpiritRegenRateCasting += 0.10 * float64(priest.Talents.Meditation)
	priest.UpdateManaRegenRates()
}

func (priest *Priest) applyMentalAgility() {
	if priest.Talents.MentalAgility == 0 {
		return
	}

	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -0.02 * float64(priest.Talents.MentalAgility),
		ClassMask:  PriestSpellInstant,
	})
}

func (priest *Priest) applyDarkness() {
	if priest.Talents.Darkness == 0 {
		return
	}

	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.02 * float64(priest.Talents.Darkness),
		ClassMask:  PriestShadowSpells,
	})
}

func (priest *Priest) applyShadowFocus() {
	if priest.Talents.ShadowFocus == 0 {
		return
	}

	priest.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexShadow] += 2 * float64(priest.Talents.ShadowFocus)

}

func (priest *Priest) applyImprovedShadowWordPain() {
	if priest.Talents.ImprovedShadowWordPain == 0 {
		return
	}

	priest.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_DotNumberOfTicks_Flat,
		IntValue:  int32(priest.Talents.ImprovedShadowWordPain),
		ClassMask: PriestSpellShadowWordPain,
	})
}

func (priest *Priest) applyFocusedMind() {
	if priest.Talents.FocusedMind == 0 {
		return
	}

	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -0.05 * float64(priest.Talents.FocusedMind),
		ClassMask:  PriestSpellMindBlast | PriestSpellMindFlay,
	})
}

func (priest *Priest) applyShadowAffinity() {
	if priest.Talents.ShadowAffinity == 0 {
		return
	}

	threatReduction := []float64{0, -0.08, -0.16, -0.25}[priest.Talents.ShadowAffinity]

	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_ThreatMultiplier_Pct,
		FloatValue: threatReduction,
		ClassMask:  PriestShadowSpells,
	})
}

func (priest *Priest) applyShadowPower() {
	if priest.Talents.ShadowPower == 0 {
		return
	}

	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 2.0 * float64(priest.Talents.ShadowPower),
		ClassMask:  PriestSpellMindBlast | PriestSpellShadowWordDeath,
	})
}

func (priest *Priest) applyShadowWeaving() {
	if priest.Talents.ShadowWeaving == 0 {
		return
	}

	swAuras := priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ShadowWeavingAura(target)
	})
	procChance := 0.20 * float64(priest.Talents.ShadowWeaving)

	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Shadow Weaving Trigger",
		ClassSpellMask: PriestShadowSpells,
		Callback:       core.CallbackOnSpellHitDealt,
		Outcome:        core.OutcomeLanded,
		ProcChance:     procChance,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			swAuras.Get(result.Target).Activate(sim)
			swAuras.Get(result.Target).AddStack(sim)
		},
	})
}

func (priest *Priest) applyMisery() {
	if priest.Talents.Misery == 0 {
		return
	}

	miseryAuras := priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.MiseryAura(target, priest.Talents.Misery)
	})

	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Misery Trigger",
		ClassSpellMask: PriestSpellShadowWordPain | PriestSpellVampiricTouch | PriestSpellMindFlay,
		Outcome:        core.OutcomeLanded,
		Callback:       core.CallbackOnSpellHitDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura := miseryAuras.Get(result.Target)
			dotDuration := spell.Dot(result.Target).RemainingDuration(sim)
			currentRemaining := aura.RemainingDuration(sim)

			if dotDuration > currentRemaining {
				aura.Duration = dotDuration
			} else {
				aura.Duration = currentRemaining
			}

			aura.Activate(sim)

		},
	})
}

func (priest *Priest) applyShadowform() {
	if !priest.Talents.Shadowform {
		return
	}

	shadowformAura := priest.RegisterAura(core.Aura{
		Label:    "Shadowform",
		ActionID: core.ActionID{SpellID: 15473},
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			if priest.SelfBuffs.PreShadowform {
				aura.Activate(sim)
			}
		},
		// Casting any holy-school spell breaks Shadowform.
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolHoly) {
				aura.Deactivate(sim)
			}
		},
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.15,
		ClassMask:  PriestShadowSpells,
	}).AttachMultiplicativePseudoStatBuff(
		&priest.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical], 0.85,
	)

	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 15473},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellShadowform,
		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 32,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			shadowformAura.Activate(sim)
		},
	})
}

func (priest *Priest) applyVampiricEmbrace() {
	if !priest.Talents.VampiricEmbrace {
		return
	}

	healPct := 0.15 + 0.05*float64(priest.Talents.ImprovedVampiricEmbrace)
	healthMetrics := priest.NewHealthMetrics(core.ActionID{SpellID: 15286})

	veDebuffAuras := priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		aura := target.RegisterAura(core.Aura{
			Label:    "Vampiric Embrace",
			ActionID: core.ActionID{SpellID: 15286},
			Duration: time.Second * 60,
		})
		aura.AttachProcTriggerCallback(target, core.ProcTrigger{
			Name:               "Vampiric Embrace Proc",
			Callback:           core.CallbackOnSpellHitTaken | core.CallbackOnPeriodicDamageTaken,
			ClassSpellMask:     PriestShadowSpells,
			RequireDamageDealt: true,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				priest.GainHealth(sim, result.Damage*healPct, healthMetrics)
			},
		})
		return aura
	})

	priest.VampiricEmbrace = priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 15286},
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellVampiricEmbrace,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 2,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    priest.NewTimer(),
				Duration: time.Second * 10,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			veDebuffAuras.Get(target).Activate(sim)
		},

		RelatedAuraArrays: veDebuffAuras.ToMap(),
	})
}
