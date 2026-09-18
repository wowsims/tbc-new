package mage

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var fireballRank = genRanks.Fireball.BySpellID(27070)

func (mage *Mage) registerFireballSpell() {
	fireballTick := fireballRank.Periodic.(shared.SpellRankPeriodic)

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27070},
		SpellSchool:    core.SpellSchoolFire,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: MageSpellFireball,
		MissileSpeed:   24,

		ManaCost: core.ManaCostOptions{
			FlatCost: fireballRank.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      fireballRank.GCD,
				CastTime: fireballRank.CastTime,
			},
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "FireballDoT",
			},
			NumberOfTicks: fireballTick.Ticks,
			TickLength:    fireballTick.Period,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, 21)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		DamageMultiplier: 1,
		BonusCoefficient: fireballRank.Direct.BonusCoefficient(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := fireballRank.Direct.Damage(sim)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			spell.WaitTravelTime(sim, func(s *core.Simulation) {
				spell.DealDamage(sim, result)
				if result.Landed() {
					spell.Dot(target).Apply(sim)
				}
			})

		},
	})
}
