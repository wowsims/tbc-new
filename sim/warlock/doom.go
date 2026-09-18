package warlock

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var doomRank = genRanks.CurseOfDoom.BySpellID(30910)
var doomTick = doomRank.Periodic.(shared.SpellRankPeriodic)
var doomCoeff = doomTick.Coef

func (warlock *Warlock) registerCurseOfDoom() {

	calculateBaseDamage := func() float64 {
		damageMultiplier := core.TernaryFloat64(warlock.AmplifyCurseAura != nil && warlock.AmplifyCurseAura.IsActive(), 1.5, 1.0)
		return doomTick.Tick * damageMultiplier
	}

	warlock.CurseOfDoom = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 30910},
		SpellSchool:    core.SpellSchoolShadow,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellCurseOfDoom,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: doomRank.GCD,
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: doomRank.Cooldown,
			},
		},

		ThreatMultiplier: 1,
		DamageMultiplier: 1,
		BonusCoefficient: doomCoeff,

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
			NumberOfTicks:            doomTick.Ticks,
			TickLength:               doomTick.Period,
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
				return spell.CalcPeriodicDamage(sim, target, calculateBaseDamage(), spell.OutcomeExpectedMagicHit)
			}
		},
	})
}
