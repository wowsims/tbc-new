package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

var multiShotRank = genRanks.MultiShot.BySpellID(27021)

func (hunter *Hunter) registerMultiShotSpell() {
	hunter.MultiShot = hunter.RegisterRangedSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27021},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeRanged,
		ProcMask:       core.ProcMaskRangedSpecial,
		ClassSpellMask: HunterSpellMultiShot,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		MissileSpeed: 30,

		ManaCost: core.ManaCostOptions{
			FlatCost: multiShotRank.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: time.Millisecond * 500,
			},
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: multiShotRank.Cooldown,
			},
		},

		BonusCoefficient: multiShotRank.Direct.BonusCoefficient(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.RangedAttackPower(target)*0.2 +
				hunter.AutoAttacks.Ranged().BaseDamage(sim) +
				hunter.talonOfAlarBonus() +
				multiShotRank.Direct.Damage(sim)

			spell.CalcAoeDamage(sim, baseDamage, spell.OutcomeRangedHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealBatchedAoeDamage(sim)
			})
		},
	})
}
