package priest

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var HolyNovaRankMap = genRanks.HolyNova

func (priest *Priest) registerHolyNovaSpell(rank shared.SpellRank) {
	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.SpellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellHolyNova,
		Rank:           rank.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rank.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		BonusCoefficient:         shared.SpellRankCoef(rank.Direct),
		ThreatMultiplier:         0,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := priest.CalcAndRollDamageRange(sim, shared.SpellRankMin(rank.Direct), shared.SpellRankMax(rank.Direct))
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)

			baseHeal := priest.CalcAndRollDamageRange(sim, shared.SpellRankMin(rank.Direct), shared.SpellRankMax(rank.Direct))
			spell.CalcAndDealHealing(sim, spell.Unit, baseHeal, spell.OutcomeHealing)
		},
	})
}
