package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const uaCoeff = 0.2

func (warlock *Warlock) registerUnstableAffliction() {
	warlock.UnstableAffliction = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 30405},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellUnstableAffliction,

		ManaCost: core.ManaCostOptions{FlatCost: 400},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: 1500 * time.Millisecond,
			},
		},

		CritMultiplier:           1,
		DamageMultiplierAdditive: 1,
		ThreatMultiplier:         1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},
		BonusCoefficient: uaCoeff,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:    "Unstable Affliction",
				Tag:      "Affliction",
				ActionID: core.ActionID{SpellID: 30108},
			},
			NumberOfTicks:       9,
			TickLength:          2 * time.Second,
			AffectedByCastSpeed: true,
			BonusCoefficient:    uaCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, 1050)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			dot := spell.Dot(target)
			if useSnapshot {
				result := dot.CalcSnapshotDamage(sim, target, dot.OutcomeTick)
				result.Damage /= dot.TickPeriod().Seconds()
				return result
			} else {
				result := spell.CalcPeriodicDamage(sim, target, 1050, spell.OutcomeExpectedMagicCrit)
				result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
				return result
			}
		},
	})
}
