package warlock

import (
	"math"
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const drainLifeCoeff = 0.143

func (warlock *Warlock) registerDrainLife() {
	healthMetric := warlock.NewHealthMetrics(core.ActionID{SpellID: 689})
	resultSlice := make(core.SpellResultSlice, 1)

	warlock.DrainLife = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 689},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagChanneled | core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellDrainLife,

		ManaCost: core.ManaCostOptions{FlatCost: 425},
		Cast:     core.CastConfig{DefaultCast: core.Cast{GCD: core.GCDDefault}},

		DamageMultiplierAdditive: 1,
		ThreatMultiplier:         1,
		BonusCoefficient:         drainLifeCoeff,

		Dot: core.DotConfig{
			Aura:                 core.Aura{Label: "Drain Life"},
			NumberOfTicks:        6,
			TickLength:           1 * time.Second,
			AffectedByCastSpeed:  true,
			HasteReducesDuration: true,
			BonusCoefficient:     drainLifeCoeff,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, 108)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				if warlock.Talents.SoulSiphon > 0 {
					cappedDmgBonus := 0.24
					if warlock.Talents.SoulSiphon == 2 {
						cappedDmgBonus = 0.60
					}
					dot.PeriodicDamageMultiplier *= 1 + math.Min((0.02*float64(warlock.Talents.SoulSiphon)), cappedDmgBonus)
				}

				resultSlice[0] = dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)

				warlock.GainHealth(sim, resultSlice[0].Damage*warlock.PseudoStats.SelfHealingMultiplier, healthMetric)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
				spell.DealOutcome(sim, result)
			}
		},
	})
}
