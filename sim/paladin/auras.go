package paladin

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

const auraEffectCategory = "PaladinAura"

func (paladin *Paladin) registerAuras() {
	paladin.registerDevotionAura()
	paladin.registerRetributionAura()
	paladin.registerConcentrationAura()
	paladin.registerFireResistanceAura()
	paladin.registerFrostResistanceAura()
	paladin.registerShadowResistanceAura()
}

// Devotion Aura
// https://www.wowhead.com/tbc/spell=27149
//
// Gives 861 additional armor to party members within 30 yards.
// Improved Devotion Aura talent increases the armor bonus by up to 40%.
// Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerDevotionAura() {
	actionID := core.ActionID{SpellID: 27149}
	armorBuff := 861.0 * (1 + 0.08*float64(paladin.Talents.ImprovedDevotionAura))

	aura := paladin.RegisterAura(core.Aura{
		Label:    "Devotion Aura" + paladin.Label,
		ActionID: actionID,
		Duration: core.NeverExpires,
	}).AttachStatBuff(stats.BonusArmor, armorBuff)

	aura.NewExclusiveEffect(auraEffectCategory, true, core.ExclusiveEffect{})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskDevotionAura,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},
	})
}

// Retribution Aura
// https://www.wowhead.com/tbc/spell=27150
//
// Causes 26 Holy damage to any creature that strikes a party member within 30 yards.
// Improved Retribution Aura talent increases damage by up to 50%.
// Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerRetributionAura() {
	actionID := core.ActionID{SpellID: 27150}
	impRetAuraMultiplier := 1 + 0.25*float64(paladin.Talents.ImprovedRetributionAura)

	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagBinary | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 26*impRetAuraMultiplier, spell.OutcomeAlwaysHit)
		},
	})

	aura := paladin.RegisterAura(core.Aura{
		Label:    "Retribution Aura" + paladin.Label,
		ActionID: actionID,
		Duration: core.NeverExpires,
	}).AttachProcTrigger(core.ProcTrigger{
		Name:     "Retribution Aura Damage",
		Callback: core.CallbackOnSpellHitTaken,
		Outcome:  core.OutcomeLanded,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.SpellSchool == core.SpellSchoolPhysical {
				procSpell.Cast(sim, spell.Unit)
			}
		},
	})

	aura.NewExclusiveEffect(auraEffectCategory, true, core.ExclusiveEffect{})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskRetributionAura,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},
	})
}

// Concentration Aura
// https://www.wowhead.com/tbc/spell=19746
//
// All party members within 30 yards lose 35% less casting or channeling time
// when damaged. Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerConcentrationAura() {
	actionID := core.ActionID{SpellID: 19746}

	aura := paladin.RegisterAura(core.Aura{
		Label:    "Concentration Aura" + paladin.Label,
		ActionID: actionID,
		Duration: core.NeverExpires,
	})

	aura.NewExclusiveEffect(auraEffectCategory, true, core.ExclusiveEffect{})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskConcentrationAura,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},
	})
}

// Fire Resistance Aura
// https://www.wowhead.com/tbc/spell=27153
func (paladin *Paladin) registerFireResistanceAura() {
	actionID := core.ActionID{SpellID: 27153}

	aura := paladin.RegisterAura(core.Aura{
		Label:    "Fire Resistance Aura" + paladin.Label,
		ActionID: actionID,
		Duration: core.NeverExpires,
	}).AttachStatBuff(stats.FireResistance, 70)

	aura.NewExclusiveEffect(auraEffectCategory, true, core.ExclusiveEffect{})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskFireResistanceAura,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},
	})
}

// Frost Resistance Aura
// https://www.wowhead.com/tbc/spell=27152
func (paladin *Paladin) registerFrostResistanceAura() {
	actionID := core.ActionID{SpellID: 27152}

	aura := paladin.RegisterAura(core.Aura{
		Label:    "Frost Resistance Aura" + paladin.Label,
		ActionID: actionID,
		Duration: core.NeverExpires,
	}).AttachStatBuff(stats.FrostResistance, 70)

	aura.NewExclusiveEffect(auraEffectCategory, true, core.ExclusiveEffect{})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskFrostResistanceAura,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},
	})
}

// Shadow Resistance Aura
// https://www.wowhead.com/tbc/spell=27151
func (paladin *Paladin) registerShadowResistanceAura() {
	actionID := core.ActionID{SpellID: 27151}

	aura := paladin.RegisterAura(core.Aura{
		Label:    "Shadow Resistance Aura" + paladin.Label,
		ActionID: actionID,
		Duration: core.NeverExpires,
	}).AttachStatBuff(stats.ShadowResistance, 70)

	aura.NewExclusiveEffect(auraEffectCategory, true, core.ExclusiveEffect{})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskShadowResistanceAura,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},
	})
}

// Sanctity Aura (Talent)
// https://www.wowhead.com/tbc/spell=20218
//
// Increases Holy damage done by party members within 30 yards by 10%.
// Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerSanctityAura() {
	actionID := core.ActionID{SpellID: 20218}

	aura := paladin.RegisterAura(core.Aura{
		Label:    "Sanctity Aura" + paladin.Label,
		ActionID: actionID,
		Duration: core.NeverExpires,
	}).AttachMultiplicativePseudoStatBuff(
		&paladin.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexHoly], 1.1,
	)

	aura.NewExclusiveEffect(auraEffectCategory, true, core.ExclusiveEffect{})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskSanctityAura,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},
	})
}
