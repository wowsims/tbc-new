package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const (
	StarfireBonusCoeff = 1
	StarfireMinDmg     = 550
	StarfireMaxDmg     = 647
)

func (druid *Druid) registerStarfireSpell() {
	druid.Starfire = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26986},
		SpellSchool:    core.SpellSchoolArcane,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: DruidSpellStarfire,
		Flags:          core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: 370,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 3500,
			},
		},

		BonusCoefficient: StarfireBonusCoeff,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   druid.DefaultSpellCritMultiplier(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := druid.CalcAndRollDamageRange(sim, StarfireMinDmg, StarfireMaxDmg)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}
