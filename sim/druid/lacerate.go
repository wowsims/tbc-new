package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (druid *Druid) registerLacerateSpell() {
	// Base: 155 damage over 5 ticks = 31 per tick per stack.
	tickDamageBase := 155.0 / 5

	druid.Lacerate = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 33745},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellLacerate,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost:   15,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		CritMultiplier:   druid.FeralCritMultiplier(),
		ThreatMultiplier: 0.5,
		FlatThreatBonus:  267,
		MaxRange:         core.MaxMeleeRange,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:     "Lacerate",
				MaxStacks: 5,
				Duration:  time.Second * 15,
			},
			NumberOfTicks: 5,
			TickLength:    time.Second * 3,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				perStack := tickDamageBase + druid.IdolLacerateBonus + druid.LacerateTickBonus + 0.01*dot.Spell.MeleeAttackPower(target)
				dot.SnapshotPhysical(target, perStack*float64(dot.Aura.GetStacks()))
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := tickDamageBase + druid.IdolLacerateBonus + druid.LacerateTickBonus + 0.01*spell.MeleeAttackPower(target)
			if druid.MangleAuras != nil && druid.MangleAuras.Get(target).IsActive() {
				baseDamage *= 1.3
			}
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				dot := spell.Dot(target)
				if dot.IsActive() {
					dot.Refresh(sim)
					dot.AddStack(sim)
					dot.TakeSnapshot(sim)
				} else {
					dot.Apply(sim)
					dot.SetStacks(sim, 1)
					dot.TakeSnapshot(sim)
				}
			} else {
				spell.IssueRefund(sim)
			}
		},
	})

	druid.Lacerate.ShortName = "Lacerate"
}
