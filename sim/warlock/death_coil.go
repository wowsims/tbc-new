package warlock

import (
	"github.com/wowsims/tbc/sim/core"
)

var deathCoilRank = genRanks.DeathCoil.BySpellID(27223)

func (warlock *Warlock) registerDeathCoil() {

	warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27223},
		SpellSchool:    core.SpellSchoolShadow,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellDeathCoil,
		MissileSpeed:   24,
		MaxRange:       deathCoilRank.MaxRange,

		ManaCost: core.ManaCostOptions{FlatCost: deathCoilRank.Cost},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: deathCoilRank.GCD,
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: deathCoilRank.Cooldown,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 0.214,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 526, spell.OutcomeMagicHit)
		},
	})
}
