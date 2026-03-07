package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const (
	WrathBonusCoeff = 0.57099997997
	WrathMinDmg     = 383
	WrathMaxDmg     = 432
)

func (druid *Druid) registerWrathSpell() {
	druid.Wrath = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26985},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: DruidSpellWrath,
		Flags:          core.SpellFlagAPL,
		MissileSpeed:   20,

		ManaCost: core.ManaCostOptions{
			FlatCost: 255,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 2000,
			},
		},

		BonusCoefficient: WrathBonusCoeff,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   druid.DefaultSpellCritMultiplier(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := druid.CalcAndRollDamageRange(sim, WrathMinDmg, WrathMaxDmg)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
