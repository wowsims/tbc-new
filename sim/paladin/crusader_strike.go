package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

// Crusader Strike
// https://www.wowhead.com/tbc/spell=35395
//
// An instant strike that causes 110% weapon damage and refreshes all Judgements on the target.
func (paladin *Paladin) registerCrusaderStrike() {
	actionID := core.ActionID{SpellID: 35395}
	paladin.CrusaderStrike = paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,
		ClassSpellMask: SpellMaskCrusaderStrike,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Second * 6,
			},
		},

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 8,
		},

		MaxRange:         core.MaxMeleeRange,
		DamageMultiplier: 1.1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultMeleeCritMultiplier(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		},
	})
}
