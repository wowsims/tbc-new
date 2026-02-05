package paladin

// Holy Shock
// https://www.wowhead.com/tbc/spell=33072
//
// Blasts the target with Holy energy, causing 721 to 779 Holy damage to an
// enemy, or 931 to 987 healing to an ally.

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (paladin *Paladin) registerHolyShock() {
	// Holy Shock Damage
	holyShockDamage := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 33072}.WithTag(1),
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHolyShock,

		MaxRange: 20,

		ManaCost: core.ManaCostOptions{
			FlatCost: 520,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Second * 15,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   paladin.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		BonusCoefficient: 0.4286,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(721, 779)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})

	// Holy Shock Healing
	holyShockHeal := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 33072}.WithTag(2),
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskHolyShock,

		MaxRange: 40,

		ManaCost: core.ManaCostOptions{
			FlatCost: 520,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    holyShockDamage.CD.Timer,
				Duration: time.Second * 15,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   paladin.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		BonusCoefficient: 0.4286,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if target.IsOpponent(&paladin.Unit) {
				target = &paladin.Unit
			}
			baseHealing := sim.Roll(931, 987)
			spell.CalcAndDealHealing(sim, target, baseHealing, spell.OutcomeHealingCrit)
		},
	})

	_ = holyShockHeal
}
