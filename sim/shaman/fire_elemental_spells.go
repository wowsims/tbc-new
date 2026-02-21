package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (fireElemental *FireElemental) registerFireBlast() {
	fireElemental.FireBlast = fireElemental.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 57984},
		SpellSchool: core.SpellSchoolFire,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShamanSpell,

		ManaCost: core.ManaCostOptions{
			FlatCost: 40,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    fireElemental.NewTimer(),
				Duration: time.Second * 6,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   fireElemental.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 0.42899999022,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 13.8 //Magic number from beta testing
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}

func (fireElemental *FireElemental) registerFireNova() {
	levelScalingMultiplier := 91.517600 / 12.102900
	fireElemental.FireNova = fireElemental.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 117588},
		SpellSchool: core.SpellSchoolFire,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShamanSpell,

		ManaCost: core.ManaCostOptions{
			FlatCost: 30,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: time.Second * 2,
			},
			CD: core.Cooldown{
				Timer:    fireElemental.NewTimer(),
				Duration: time.Second * 10,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   fireElemental.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 1.00,

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, func(sim *core.Simulation, _ *core.Spell) float64 {
				return sim.Roll(49*levelScalingMultiplier, 58*levelScalingMultiplier) //Estimated from beta testing 49 58
			})
		},
	})
}

func (fireElemental *FireElemental) registerImmolate() {
	actionID := core.ActionID{SpellID: 118297}

	fireElemental.Immolate = fireElemental.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolFire,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShamanSpell,

		DamageMultiplier: 1,
		CritMultiplier:   fireElemental.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 1.0,

		ManaCost: core.ManaCostOptions{
			FlatCost: 95,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: time.Second * 2,
			},
			CD: core.Cooldown{
				Timer:    fireElemental.NewTimer(),
				Duration: time.Second * 10,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !fireElemental.IsGuardian()
		},
		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Immolate",
			},
			NumberOfTicks:       7,
			TickLength:          time.Second * 3,
			BonusCoefficient:    0.34999999404,
			AffectedByCastSpeed: true,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, 0.0)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 0.0
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.Dot(target).Apply(sim)
		},
	})
}
