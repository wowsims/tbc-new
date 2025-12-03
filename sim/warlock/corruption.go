package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const corruptionCoeff = 0.156

func (warlock *Warlock) registerCorruption() *core.Spell {
	resultSlice := make(core.SpellResultSlice, 1)

	warlock.Corruption = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 172},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellCorruption,

		ManaCost: core.ManaCostOptions{FlatCost: 370},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 2000,
			},
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:    "Corruption",
				Tag:      "Affliction",
				ActionID: core.ActionID{SpellID: 172},
			},
			NumberOfTicks:       6,
			TickLength:          3 * time.Second,
			AffectedByCastSpeed: false,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, warlock.CalcScalingSpellDmg(corruptionCoeff))
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				resultSlice[0] = dot.CalcSnapshotDamage(sim, target, dot.OutcomeSnapshotCrit)

				// if onTickCallback != nil {
				// 	onTickCallback(resultSlice, dot.Spell, sim)
				// // }

				dot.Spell.DealPeriodicDamage(sim, resultSlice[0])
			},
		},

		// ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		// 	result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		// 	if onApplyCallback != nil {
		// 		resultSlice[0] = result
		// 		onApplyCallback(resultSlice, spell, sim)
		// 	}
		// 	spell.DealOutcome(sim, result)
		// },
		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			dot := spell.Dot(target)
			if useSnapshot {
				result := dot.CalcSnapshotDamage(sim, target, dot.OutcomeExpectedSnapshotCrit)
				result.Damage /= dot.TickPeriod().Seconds()
				return result
			} else {
				result := spell.CalcPeriodicDamage(sim, target, warlock.CalcScalingSpellDmg(corruptionCoeff), spell.OutcomeExpectedMagicCrit)
				result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
				return result
			}
		},
	})

	return warlock.Corruption
}
