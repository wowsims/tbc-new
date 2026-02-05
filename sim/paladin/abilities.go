package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

// Exorcism
// https://www.wowhead.com/tbc/spell=10314
//
// Causes Holy damage to an Undead or Demon target.
func (paladin *Paladin) registerExorcism() {
	actionID := core.ActionID{SpellID: 10314}

	paladin.Exorcism = paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskExorcism,

		MaxRange: 30,

		ManaCost: core.ManaCostOptions{
			FlatCost: 295, // Rank 7
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 1500,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Second * 15,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   paladin.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// TODO: Implement damage calculation (only works on Undead/Demon)
		},
	})
}

// Hammer of Justice
// https://www.wowhead.com/tbc/spell=10308
//
// Stuns the target for 6 sec.
func (paladin *Paladin) registerHammerOfJustice() {
	actionID := core.ActionID{SpellID: 10308}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHammerOfJustice,

		MaxRange: 10,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 3,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// TODO: Implement stun
		},
	})
}

// Hammer of Wrath
// https://www.wowhead.com/tbc/spell=27180
//
// Hurls a hammer that strikes an enemy for Holy damage.
// Only usable on enemies that have 20% or less health.
func (paladin *Paladin) registerHammerOfWrath() {
	actionID := core.ActionID{SpellID: 27180}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHammerOfWrath,

		MaxRange: 30,

		ManaCost: core.ManaCostOptions{
			FlatCost: 665, // Rank 4
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Second,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Second * 6,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			// Only usable on enemies below 20% health
			return sim.IsExecutePhase20()
		},

		DamageMultiplier: 1,
		CritMultiplier:   paladin.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// TODO: Implement damage calculation
		},
	})
}

// Divine Shield
// https://www.wowhead.com/tbc/spell=642
//
// Protects the Paladin from all damage and spells for 12 sec, but reduces
// all damage you deal by 50%. Once protected, the target cannot be protected
// by Divine Shield, Divine Protection, or Blessing of Protection again for 1 min.
func (paladin *Paladin) registerDivineShield() {
	actionID := core.ActionID{SpellID: 642}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskDivineShield,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute * 5,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// TODO: Implement immunity and damage reduction
		},
	})
}

// Divine Protection
// https://www.wowhead.com/tbc/spell=27147
//
// Protects the Paladin from all physical attacks for 8 sec, but reduces
// all damage you deal by 50%. Once protected, the target cannot be protected
// by Divine Shield, Divine Protection, or Blessing of Protection again for 1 min.
func (paladin *Paladin) registerDivineProtection() {
	actionID := core.ActionID{SpellID: 27147}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskDivineProtection,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 3,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// TODO: Implement physical immunity and damage reduction
		},
	})
}

// Cleanse
// https://www.wowhead.com/tbc/spell=4987
//
// Cleanses a friendly target, removing 1 poison effect, 1 disease effect,
// and 1 magic effect.
func (paladin *Paladin) registerCleanse() {
	actionID := core.ActionID{SpellID: 4987}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskCleanse,

		MaxRange: 40,

		ManaCost: core.ManaCostOptions{
			FlatCost: 120,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// TODO: Implement dispel
		},
	})
}

// Holy Shield (Talent)
// https://www.wowhead.com/tbc/spell=27179
//
// Increases chance to block by 30% for 10 sec, and deals Holy damage
// for each attack blocked while active. Damage caused by Holy Shield causes
// 35% additional threat. Each block expends a charge. 8 charges.
func (paladin *Paladin) registerHolyShield() {
	actionID := core.ActionID{SpellID: 27179}

	paladin.HolyShieldAura = paladin.RegisterAura(core.Aura{
		Label:     "Holy Shield" + paladin.Label,
		ActionID:  actionID,
		Duration:  time.Second * 10,
		MaxStacks: 8,
	})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHolyShield,

		ManaCost: core.ManaCostOptions{
			FlatCost: 170, // Rank 4
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Second * 10,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// TODO: Activate Holy Shield aura with block chance increase
		},
	})
}

// Avenger's Shield (Talent)
// https://www.wowhead.com/tbc/spell=32700
//
// Hurls a holy shield at the enemy, dealing Holy damage, dazing them and
// then jumping to additional nearby enemies. Affects 3 total targets.
func (paladin *Paladin) registerAvengersShield() {
	actionID := core.ActionID{SpellID: 32700}

	paladin.AvengersShield = paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskAvengersShield,

		MaxRange:     30,
		MissileSpeed: 35,

		ManaCost: core.ManaCostOptions{
			FlatCost: 780,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Second * 30,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return paladin.PseudoStats.CanBlock
		},

		DamageMultiplier: 1,
		CritMultiplier:   paladin.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// TODO: Implement damage and bounce to additional targets
		},
	})
}

// Repentance (Talent)
// https://www.wowhead.com/tbc/spell=20066
//
// Puts the enemy target in a state of meditation, incapacitating them for up to 1 min.
// Any damage caused will awaken the target. Usable against Demons, Dragonkin,
// Giants, Humanoids and Undead.
func (paladin *Paladin) registerRepentance() {
	actionID := core.ActionID{SpellID: 20066}

	paladin.Repentance = paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskRepentance,

		MaxRange: 20,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 9,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// TODO: Implement CC
		},
	})
}

// Divine Illumination (Talent)
// https://www.wowhead.com/tbc/spell=31842
//
// Reduces the mana cost of all spells by 50% for 15 sec.
func (paladin *Paladin) registerDivineIllumination() {
	actionID := core.ActionID{SpellID: 31842}

	paladin.DivineIlluminationAura = paladin.RegisterAura(core.Aura{
		Label:    "Divine Illumination" + paladin.Label,
		ActionID: actionID,
		Duration: time.Second * 15,
	})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskDivineIllumination,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// TODO: Activate mana cost reduction aura
		},
	})
}
