package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var ShadowWordDeathRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 32379, Cost: 243, MinDamage: 450, MaxDamage: 522, Coefficient: 0.429},
	{Rank: 2, SpellID: 32996, Cost: 309, MinDamage: 572, MaxDamage: 664, Coefficient: 0.429},
}

func (priest *Priest) registerShadowWordDeathSpell(rankConfig shared.SpellRankConfig, cdTimer *core.Timer) {
	spell := priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rankConfig.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellShadowWordDeath,
		Rank:           rankConfig.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rankConfig.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: 12 * time.Second, // TODO: verify from wago.tools
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           priest.DefaultSpellCritMultiplier(),
		BonusCoefficient:         rankConfig.Coefficient,
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := priest.CalcAndRollDamageRange(sim, rankConfig.MinDamage, rankConfig.MaxDamage)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.DealDamage(sim, result)

			// Set to always remove for purpose of sim
			priest.RemoveHealth(sim, result.Damage)
		},
	})

	priest.ShadowWordDeath = append(priest.ShadowWordDeath, spell)
}
