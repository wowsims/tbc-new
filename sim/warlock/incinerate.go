package warlock

import (
	"github.com/wowsims/tbc/sim/core"
)

var incinerateRank = genRanks.Incinerate.BySpellID(32231)
var incinerateCoeff = incinerateRank.Direct.BonusCoefficient()

func (warlock *Warlock) registerIncinerate() {
	warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 32231},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		MissileSpeed:   incinerateRank.MissileSpeed,
		ClassSpellMask: WarlockSpellIncinerate,

		ManaCost: core.ManaCostOptions{FlatCost: incinerateRank.Cost},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      incinerateRank.GCD,
				CastTime: incinerateRank.CastTime,
			},
		},

		DamageMultiplierAdditive: 1,
		DefenseType:              core.DefenseTypeMagic,
		ThreatMultiplier:         1,
		BonusCoefficient:         incinerateCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := incinerateRank.Direct.Damage(sim)
			if warlock.Immolate.Dot(target).IsActive() {
				baseDamage += sim.Roll(111, 128)
			}
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
