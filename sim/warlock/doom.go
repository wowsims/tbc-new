package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const doomCoeff = 2

func (warlock *Warlock) registerCurseOfDoom() {

	calculateBaseDamage := func() float64 {
		damageMultiplier := core.TernaryFloat64(warlock.AmplifyCurseAura != nil && warlock.AmplifyCurseAura.IsActive(), 1.5, 1.0)
		return 4200 * damageMultiplier
	}

	warlock.CurseOfDoom = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 30910},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellCurseOfDoom,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: time.Second * 60,
			},
		},

		ThreatMultiplier: 1,
		DamageMultiplier: 1,
		BonusCoefficient: doomCoeff,
		CritMultiplier:   1,

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
				Label: "Doom",
				Tag:   "Affliction",
			},
			NumberOfTicks:            1,
			TickLength:               1 * time.Minute,
			BonusCoefficient:         doomCoeff,
			PeriodicDamageMultiplier: 1,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, calculateBaseDamage())
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			dot := spell.Dot(target)
			if useSnapshot {
				return dot.CalcSnapshotDamage(sim, target, dot.OutcomeTick)
			} else {
				return spell.CalcPeriodicDamage(sim, target, calculateBaseDamage(), spell.OutcomeExpectedMagicCrit)
			}
		},
	})
}
