package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (hunter *Hunter) registerRaptorStrikeSpell() {
	hunter.RaptorStrike = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27014},
		SpellSchool:    core.SpellSchoolPhysical,
		ClassSpellMask: HunterSpellRaptorStrike,
		ProcMask:       core.ProcMaskMeleeMH,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

		MaxRange: core.MaxMeleeRange,

		ManaCost: core.ManaCostOptions{
			FlatCost: 120,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 6,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   hunter.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Emit an "auto delayed" log line whenever the mh auto fired
			// later than it would have in an uncontested rotation. Below 1ms
			// is treated as rounding noise so the common case stays silent.
			delay := hunter.AutoAttacks.MainHandPendingSwingDelay()
			readyAt := sim.CurrentTime - delay
			if sim.Log != nil && delay > time.Millisecond && readyAt > 0 {
				hunter.Log(sim, "%s delayed by %s, was ready at %s", spell.ActionID, delay, readyAt)
			}

			baseDamage := hunter.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) + 170
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})

	hunter.RegisterAura(core.Aura{
		Label:    "Raptor Strike",
		ActionID: core.ActionID{SpellID: 27014}.WithTag(2),
		Icd:      &hunter.RaptorStrike.CD,
	})
}

// Returns true if the regular melee swing should be used, false otherwise.
func (hunter *Hunter) TryRaptorStrike(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if mhSwingSpell.ActionID.Tag != 1 || !hunter.RaptorStrike.CanCast(sim, hunter.CurrentTarget) {
		return mhSwingSpell
	}

	return hunter.RaptorStrike
}
