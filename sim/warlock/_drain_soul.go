package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const drainSoulScale = 0.257
const drainSoulCoeff = 0.257

func (warlock *Warlock) registerDrainSoul() {
	warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 1120},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL | core.SpellFlagChanneled,
		ClassSpellMask: WarlockSpellDrainSoul,

		ManaCost: core.ManaCostOptions{BaseCostPercent: 1.5},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		DamageMultiplierAdditive: 1,
		CritMultiplier:           warlock.DefaultSpellCritMultiplier(),
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura:                 core.Aura{Label: "DrainSoul"},
			NumberOfTicks:        6,
			TickLength:           2 * time.Second,
			AffectedByCastSpeed:  true,
			HasteReducesDuration: true,
			BonusCoefficient:     drainSoulCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, warlock.CalcAndRollDamageRange(drainSoulScale))
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				result := dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)

				if !result.Landed() || !sim.IsExecutePhase20() {
					return
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			dot := spell.Dot(target)
			if useSnapshot {
				result := dot.CalcSnapshotDamage(sim, target, dot.OutcomeTick)
				result.Damage /= dot.TickPeriod().Seconds()
				return result
			} else {
				result := spell.CalcPeriodicDamage(sim, target, warlock.CalcAndDealDamage(drainSoulScale), spell.OutcomeExpectedMagicCrit)
				result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
				return result
			}
		},
	})

	dmgMode := warlock.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 1,
		ClassMask:  warlock.WarlockSpellDrainSoul,
	})

	warlock.RegisterResetEffect(func(s *core.Simulation) {
		dmgMode.Deactivate()
		s.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
			if isExecute > 20 {
				return
			}

			dmgMode.Activate()
		})
	})
}
