package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (hunter *Hunter) registerMultiShotSpell() {
	hunter.MultiShot = hunter.RegisterRangedSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27021},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskRangedSpecial,
		ClassSpellMask: HunterSpellMultiShot,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		MissileSpeed: 30,
		MinRange:     core.MaxMeleeRange,
		MaxRange:     HunterBaseMaxRange,

		ManaCost: core.ManaCostOptions{
			FlatCost: 275,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 500,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 10,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   hunter.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.RangedAttackPower(target)*0.2 +
				hunter.AutoAttacks.Ranged().BaseDamage(sim) +
				hunter.talonOfAlarBonus() +
				205

			spell.CalcAoeDamage(sim, baseDamage, spell.OutcomeRangedHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealBatchedAoeDamage(sim)
			})
		},
	})
}
