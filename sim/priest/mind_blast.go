package priest

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var MindBlastRankMap = genRanks.MindBlast

func (priest *Priest) registerMindBlastSpell(rank shared.SpellRank, cdTimer *core.Timer) {

	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellMindBlast,
		Rank:           rank.Rank,
		MaxRange:       rank.MaxRange,
		ManaCost: core.ManaCostOptions{
			FlatCost: rank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      rank.GCD,
				CastTime: rank.CastTime,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: rank.Cooldown,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		BonusCoefficient:         rank.Direct.BonusCoefficient(),
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := rank.Direct.Damage(sim)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.DealDamage(sim, result)
		},
	})
}
