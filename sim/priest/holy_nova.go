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
				GCD: rank.GCD,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		BonusCoefficient:         rank.Direct.BonusCoefficient(),
		ThreatMultiplier:         0,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := rank.Direct.Damage(sim)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)

			baseHeal := rank.Direct.Damage(sim)
			spell.CalcAndDealHealing(sim, spell.Unit, baseHeal, spell.OutcomeHealing)
		},
	})
}
