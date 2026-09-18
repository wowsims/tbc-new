package hunter

import (
	"github.com/wowsims/tbc/sim/core"
)

var arcaneShotRank = genRanks.ArcaneShot.BySpellID(27019)

func (hunter *Hunter) registerArcaneShotSpell() {
	hunter.ArcaneShot = hunter.RegisterRangedSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27019},
		SpellSchool:    core.SpellSchoolArcane,
		DefenseType:    core.DefenseTypeRanged,
		ClassSpellMask: HunterSpellArcaneShot,
		ProcMask:       core.ProcMaskRangedSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: arcaneShotRank.Cost,
		},

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: arcaneShotRank.Cooldown,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.RangedAttackPower(target)*0.15 +
				hunter.talonOfAlarBonus() +
				arcaneShotRank.Direct.Damage(sim)

			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
