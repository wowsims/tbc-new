package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (paladin *Paladin) registerHealingSpells() {
	paladin.registerHolyLight()
	paladin.registerFlashOfLight()
	paladin.registerLayOnHands()
}

// Holy Light
// https://www.wowhead.com/tbc/spell=27136
//
// Heals a friendly target for a large amount.
func (paladin *Paladin) registerHolyLight() {
	actionID := core.ActionID{SpellID: 27136}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskHolyLight,

		MaxRange: 40,

		ManaCost: core.ManaCostOptions{
			FlatCost: 840, // Rank 11
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 2500,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   paladin.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// TODO: Implement healing calculation
			// Base heal: 2196 to 2446
		},
	})
}

// Flash of Light
// https://www.wowhead.com/tbc/spell=27137
//
// Heals a friendly target for a small amount.
func (paladin *Paladin) registerFlashOfLight() {
	actionID := core.ActionID{SpellID: 27137}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskFlashOfLight,

		MaxRange: 40,

		ManaCost: core.ManaCostOptions{
			FlatCost: 180, // Rank 7
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 1500,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   paladin.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// TODO: Implement healing calculation
			// Base heal: 458 to 513
		},
	})
}

// Lay on Hands
// https://www.wowhead.com/tbc/spell=27154
//
// Heals a friendly target for an amount equal to the Paladin's maximum health
// and restores mana to the target. Causes Forbearance for 1 min.
func (paladin *Paladin) registerLayOnHands() {
	actionID := core.ActionID{SpellID: 27154}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskLayOnHands,

		MaxRange: 40,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Hour, // 1 hour cooldown in TBC (reduced by talents)
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   paladin.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// TODO: Implement healing for paladin's max health + mana restore
			// Also apply Forbearance debuff
		},
	})
}

// Holy Shock (Talent) - Healing component
// https://www.wowhead.com/tbc/spell=33072
//
// Blasts the target with Holy energy, causing Holy damage to an enemy,
// or healing to an ally.
//
// NOTE: Implementation is in holy_shock.go, registered via talents.go when talent is taken
