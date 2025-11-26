package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const SwpScaleCoeff = 0.629660992297
const SwpSpellCoeff = 0.310169488695

func (priest *Priest) registerShadowWordPainSpell() {
	priest.ShadowWordPain = priest.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 589},
		SpellSchool:      core.SpellSchoolShadow,
		ProcMask:         core.ProcMaskSpellDamage,
		Flags:            core.SpellFlagAPL,
		ClassSpellMask:   PriestSpellShadowWordPain,
		BonusCoefficient: SwpSpellCoeff,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 4.4,
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           priest.DefaultCritMultiplier(),

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "ShadowWordPain",
			},

			NumberOfTicks:       6,
			TickLength:          time.Second * 3,
			AffectedByCastSpeed: true,

			BonusCoefficient: SwpSpellCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, priest.CalcScalingSpellDmg(SwpScaleCoeff))
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeSnapshotCrit)
			},
		},

		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				dot := spell.Dot(target)
				dot.Apply(sim)
				dot.TickOnce(sim)

				// Custom code for tracking T15 2PC extension logic
				// which recalculates the snapshot if you extend after
				// the initial duration.
				priest.T15_2PC_ExtensionTracker[target.Index].Swp = spell.Dot(target).ExpiresAt()
			}
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			dot := spell.Dot(target)
			if useSnapshot {
				result := dot.CalcSnapshotDamage(sim, target, dot.OutcomeExpectedSnapshotCrit)
				result.Damage /= dot.TickPeriod().Seconds()
				return result
			} else {
				result := spell.CalcPeriodicDamage(sim, target, priest.CalcScalingSpellDmg(SwpScaleCoeff), spell.OutcomeExpectedMagicCrit)
				result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
				return result
			}
		},
	})
}
