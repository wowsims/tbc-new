package paladin

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

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
// Gives additional armor to party members within 30 yards.
// Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerDevotionAura() {
	actionID := core.ActionID{SpellID: 27149}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskDevotionAura,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// TODO: Implement aura activation
		},
	})
}

// Retribution Aura
// https://www.wowhead.com/tbc/spell=27150
//
// Causes Holy damage to any creature that strikes a party member within 30 yards.
// Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerRetributionAura() {
	actionID := core.ActionID{SpellID: 27150}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskRetributionAura,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// TODO: Implement aura activation
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

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskConcentrationAura,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// TODO: Implement aura activation
		},
	})
}

// Fire Resistance Aura
// https://www.wowhead.com/tbc/spell=27153
//
// Gives additional Fire resistance to all party members within 30 yards.
// Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerFireResistanceAura() {
	actionID := core.ActionID{SpellID: 27153}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskFireResistanceAura,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// TODO: Implement aura activation
		},
	})
}

// Frost Resistance Aura
// https://www.wowhead.com/tbc/spell=27152
//
// Gives additional Frost resistance to all party members within 30 yards.
// Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerFrostResistanceAura() {
	actionID := core.ActionID{SpellID: 27152}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskFrostResistanceAura,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// TODO: Implement aura activation
		},
	})
}

// Shadow Resistance Aura
// https://www.wowhead.com/tbc/spell=27151
//
// Gives additional Shadow resistance to all party members within 30 yards.
// Players may only have one Aura on them per Paladin at any one time.
func (paladin *Paladin) registerShadowResistanceAura() {
	actionID := core.ActionID{SpellID: 27151}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskShadowResistanceAura,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// TODO: Implement aura activation
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
