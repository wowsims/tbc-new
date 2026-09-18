package warlock

import (
	"github.com/wowsims/tbc/sim/core"
)

var shadowBurnRank = genRanks.Shadowburn.BySpellID(30546)
var shadowBurnCoeff = shadowBurnRank.Direct.BonusCoefficient()

func (warlock *Warlock) registerShadowBurn() {

	warlock.Shadowburn = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 30546},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL | core.SpellFlagBinary,
		ClassSpellMask: WarlockSpellShadowBurn,

		ManaCost: core.ManaCostOptions{FlatCost: shadowBurnRank.Cost},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: shadowBurnRank.GCD,
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: shadowBurnRank.Cooldown,
			},
		},

		DamageMultiplier: 1,
		DefenseType:      core.DefenseTypeMagic,
		ThreatMultiplier: 1,
		BonusCoefficient: shadowBurnCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			dmgRoll := shadowBurnRank.Direct.Damage(sim)
			spell.CalcAndDealDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)

		},
	})
}
