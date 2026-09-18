package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var agonyRank = genRanks.CurseOfAgony.BySpellID(27218)
var agonyTick = agonyRank.Periodic.(shared.SpellRankPeriodic)
var agonyCoeff = agonyTick.Coef

func (warlock *Warlock) registerCurseOfAgony() {

	calculateBaseDamage := func(sim *core.Simulation, dot *core.Dot) float64 {
		damageMultiplier := core.TernaryFloat64(warlock.AmplifyCurseAura != nil && warlock.AmplifyCurseAura.IsActive(), 1.5, 1.0)
		return agonyTick.Tick * damageMultiplier
	}

	warlock.CurseOfAgony = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27218},
		Flags:          core.SpellFlagAPL,
		ProcMask:       core.ProcMaskSpellDamage,
		SpellSchool:    core.SpellSchoolShadow,
		DefenseType:    core.DefenseTypeMagic,
		ClassSpellMask: WarlockSpellCurseOfAgony,

		ThreatMultiplier: 1,
		DamageMultiplier: 1,
		BonusCoefficient: agonyCoeff,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: agonyRank.GCD,
			},
		},

		ManaCost: core.ManaCostOptions{
			FlatCost: agonyRank.Cost,
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				warlock.DeactivateOtherCurses(sim, spell, target)
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Agony",
				Tag:   "Affliction",
			},

			TickLength:               agonyTick.Period,
			NumberOfTicks:            agonyTick.Ticks,
			PeriodicDamageMultiplier: 1,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, calculateBaseDamage(sim, dot))
			},

			BonusCoefficient: agonyCoeff,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			dot := spell.Dot(target)

			// Always compare fully stacked agony damage
			if useSnapshot {
				result := dot.CalcSnapshotDamage(sim, target, dot.OutcomeTick)
				result.Damage *= 10
				result.Damage /= dot.TickPeriod().Seconds()
				return result
			} else {
				result := spell.CalcPeriodicDamage(sim, target, calculateBaseDamage(sim, dot), spell.OutcomeExpectedMagicHit)
				result.Damage *= 10
				result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
				return result
			}
		},
	})
}
