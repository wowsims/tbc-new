package protection

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/paladin"
)

/*
	Tooltip:

Sends bolts of power in all directions, causing ((8127 + 9075) / 2) / 2 + <AP> * 0.91 Holy damage

to your target
----------

, stunning Demons

and Undead for 3 sec.
*/
func (prot *ProtectionPaladin) registerHolyWrath() {
	maxTargets := prot.Env.TotalTargetCount()

	prot.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 119072},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: paladin.SpellMaskHolyWrath,

		MissileSpeed: 40,
		MaxRange:     10,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 5,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    prot.NewTimer(),
				Duration: 9 * time.Second,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   prot.DefaultCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Ingame tooltip is ((<MIN> + <MAX>) / 2) / 2
			// This is the same as, <AVG> / 2 which is the same as just halving the coef
			baseDamage := prot.CalcScalingSpellDmg(7.53200006485/2) + 0.91*spell.MeleeAttackPower()

			// Damage is split between all mobs, each hit rolls for hit/crit separately
			numTargets := min(maxTargets, sim.Environment.ActiveTargetCount())
			baseDamage /= float64(numTargets)

			multiplier := spell.DamageMultiplier

			spell.CalcCleaveDamage(sim, target, numTargets, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.DamageMultiplier = multiplier

			spell.WaitTravelTime(sim, func(simulation *core.Simulation) {
				spell.DealBatchedAoeDamage(sim)
			})
		},
	})
}
