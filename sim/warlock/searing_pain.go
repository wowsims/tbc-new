package warlock

import (
	"github.com/wowsims/tbc/sim/core"
)

var searingPainRank = genRanks.SearingPain.BySpellID(30459)
var searingPainCoeff = searingPainRank.Direct.BonusCoefficient()

func (warlock *Warlock) registerSearingPain() {

	warlock.Shadowburn = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 30459},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellSearingPain,
		MaxRange:       searingPainRank.MaxRange,

		ManaCost: core.ManaCostOptions{FlatCost: searingPainRank.Cost},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      searingPainRank.GCD,
				CastTime: searingPainRank.CastTime,
			},
		},

		DamageMultiplier: 1,
		DefenseType:      core.DefenseTypeMagic,
		ThreatMultiplier: 2,
		BonusCoefficient: searingPainCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			dmgRoll := searingPainRank.Direct.Damage(sim)
			spell.CalcAndDealDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)
		},
	})
}
