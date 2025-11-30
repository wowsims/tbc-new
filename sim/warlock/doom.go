package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const doomScale = 0.9375
const doomCoeff = 0.9375

func (warlock *Warlock) registerDoom() {
	warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 603},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: warlock.WarlockSpellDoom,

		Cast: core.CastConfig{DefaultCast: core.Cast{GCD: core.GCDDefault}},

		DamageMultiplierAdditive: 1,
		CritMultiplier:           warlock.DefaultCritMultiplier(),
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Doom",
			},
			NumberOfTicks:       4,
			TickLength:          15 * time.Second,
			AffectedByCastSpeed: true,
			BonusCoefficient:    doomCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, warlock.CalcScalingSpellDmg(doomScale))
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeSnapshotCrit)
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warlock.IsInMeta() && warlock.CanSpendDemonicFury(60)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				warlock.SpendDemonicFury(sim, 60, spell.ActionID)
				warlock.ApplyDotWithPandemic(spell.Dot(target), sim)
			}
			spell.DealOutcome(sim, result)
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			dot := spell.Dot(target)
			if useSnapshot {
				result := dot.CalcSnapshotDamage(sim, target, dot.OutcomeExpectedSnapshotCrit)
				result.Damage /= dot.TickPeriod().Seconds()
				return result
			} else {
				result := spell.CalcPeriodicDamage(sim, target, warlock.CalcScalingSpellDmg(doomScale), spell.OutcomeExpectedMagicCrit)
				result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
				return result
			}
		},
	})
}
