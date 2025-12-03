package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const immolateCoeff = 0.13

func (warlock *Warlock) registerImmolate() {
	actionID := core.ActionID{SpellID: 348}
	warlock.Immolate = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellImmolate,

		ManaCost: core.ManaCostOptions{BaseCostPercent: 3},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: 1500 * time.Millisecond,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   warlock.DefaultCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: immolateCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, warlock.CalcScalingSpellDmg(immolateCoeff), spell.OutcomeMagicHitAndCrit)
			if result.Landed() {
				spell.RelatedDotSpell.Cast(sim, target)
			}

			spell.DealDamage(sim, result)
		},
	})

	warlock.Immolate.RelatedDotSpell = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 348}.WithTag(1),
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: WarlockSpellImmolateDot,
		Flags:          core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		CritMultiplier:   warlock.DefaultCritMultiplier(),

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Immolate (DoT)",
				// OnGain: func(aura *core.Aura, sim *core.Simulation) {
				// 	core.EnableDamageDoneByCaster(DDBC_Immolate, DDBC_Total, warlock.AttackTables[aura.Unit.UnitIndex], immolateDamageDoneByCasterHandler)
				// },
				// OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				// 	core.DisableDamageDoneByCaster(DDBC_Immolate, warlock.AttackTables[aura.Unit.UnitIndex])
				// },
			},
			NumberOfTicks:    5,
			TickLength:       3 * time.Second,
			BonusCoefficient: immolateCoeff,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, warlock.CalcScalingSpellDmg(immolateCoeff))
			},
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			dot := spell.Dot(target)
			if useSnapshot {
				result := dot.CalcSnapshotDamage(sim, target, dot.OutcomeExpectedSnapshotCrit)
				result.Damage /= dot.TickPeriod().Seconds()
				return result
			} else {
				result := spell.CalcPeriodicDamage(sim, target, warlock.CalcScalingSpellDmg(immolateCoeff), spell.OutcomeExpectedMagicCrit)
				result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
				return result
			}
		},
	})
}
