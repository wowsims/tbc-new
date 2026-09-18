package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

var shadowFuryRank = genRanks.Shadowfury.BySpellID(30414)
var shadowFuryCoeff = shadowFuryRank.Direct.BonusCoefficient()

func (warlock *Warlock) registerShadowfury() {

	warlock.Shadowfury = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 30414},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellShadowFury,

		ManaCost: core.ManaCostOptions{FlatCost: shadowFuryRank.Cost},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      shadowFuryRank.GCD,
				GCDMin:   500 * time.Millisecond,
				CastTime: shadowFuryRank.CastTime,
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: shadowFuryRank.Cooldown,
			},
		},

		DamageMultiplier: 1,
		DefenseType:      core.DefenseTypeMagic,
		ThreatMultiplier: 1,
		BonusCoefficient: shadowFuryCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			dmgRoll := shadowFuryRank.Direct.Damage(sim)
			result := spell.CalcDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
