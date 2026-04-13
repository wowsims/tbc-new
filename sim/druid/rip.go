package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (druid *Druid) registerRipSpell() {
	idolBonus := func(cp int32) float64 {
		return druid.IdolRipBonus * float64(cp)
	}

	druid.Rip = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27008},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellRip,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost:   30,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return druid.ComboPoints() > 0
		},

		DamageMultiplier: 1,
		CritMultiplier:   druid.FeralCritMultiplier(),
		ThreatMultiplier: 1,
		MaxRange:         core.MaxMeleeRange,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Rip",
			},
			NumberOfTicks: 6,
			TickLength:    time.Second * 2,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				cp := druid.ComboPoints()
				ap := dot.Spell.MeleeAttackPower(target)

				var tickDamage float64
				switch {
				case cp <= 3:
					tickDamage = 990 + 0.18*ap
				case cp == 4:
					tickDamage = 1272 + 0.24*ap
				default: // 5
					tickDamage = 1554 + 0.24*ap
				}
				tickDamage = tickDamage/6 + idolBonus(cp)

				dot.SnapshotPhysical(target, tickDamage)
				druid.UpdateBleedPower(druid.Rip, sim, target, true, true)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
				druid.SpendComboPoints(sim, spell.ComboPointMetrics())
			} else {
				spell.IssueRefund(sim)
			}
			spell.DealOutcome(sim, result)
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			if useSnapshot {
				dot := spell.Dot(target)
				return dot.CalcSnapshotDamage(sim, target, dot.OutcomeTick)
			}
			// Assume 5 CP for projections.
			ap := spell.MeleeAttackPower(target)
			tickDamage := (1554+0.24*ap)/6 + idolBonus(5)
			result := spell.CalcPeriodicDamage(sim, target, tickDamage, spell.OutcomeExpectedMagicAlwaysHit)
			attackTable := spell.Unit.AttackTables[target.UnitIndex]
			critChance := spell.PhysicalCritChance(attackTable)
			result.Damage *= 1 + critChance*(spell.CritMultiplier-1)
			return result
		},
	})

	druid.Rip.ShortName = "Rip"
}

func (druid *Druid) CurrentRipCost() float64 {
	return druid.Rip.Cost.GetCurrentCost()
}
