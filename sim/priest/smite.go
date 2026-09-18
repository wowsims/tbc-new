package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var SmiteRankMap = genRanks.Smite

var smiteCastTimes = map[int32]time.Duration{
	1: 1500 * time.Millisecond,
	2: 2000 * time.Millisecond,
}

func (priest *Priest) registerSmiteSpell(rank shared.RankRow) {
	castTime := 2500 * time.Millisecond
	if ct, ok := smiteCastTimes[rank.Rank]; ok {
		castTime = ct
	}

	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.SpellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellSmite,
		Rank:           rank.Rank,
		ManaCost: core.ManaCostOptions{
			FlatCost: rank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: castTime,
			},
		},

		DamageMultiplier: 1,
		BonusCoefficient: rank.Direct.Coef,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := priest.CalcAndRollDamageRange(sim, rank.Direct.Min, rank.Direct.Max)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}
