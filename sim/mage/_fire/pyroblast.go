package fire

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/mage"
)

func (fire *FireMage) registerPyroblastSpell() {
	actionID := core.ActionID{SpellID: 11366}
	pyroblastVariance := 0.23800000548    // Per https://wago.tools/db2/SpellEffect?build=5.5.0.61217&filter%5BSpellID%5D=11366 Field: "Variance"
	pyroblastScaling := 1.98000001907     // Per https://wago.tools/db2/SpellEffect?build=5.5.0.61217&filter%5BSpellID%5D=11366 Field: "Coefficient"
	pyroblastCoefficient := 1.98000001907 // Per https://wago.tools/db2/SpellEffect?build=5.5.0.61217&filter%5BSpellID%5D=11366 Field: "BonusCoefficient"
	pyroblastDotScaling := 0.36000001431
	pyroblastDotCoefficient := 0.36000001431

	instantPyroblastDotMod := fire.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: .25,
		ClassMask:  mage.MageSpellPyroblastDot,
	})

	fire.Pyroblast = fire.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: mage.MageSpellPyroblast,
		MissileSpeed:   24,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 5,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 3500,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   fire.DefaultCritMultiplier(),
		BonusCoefficient: pyroblastCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			hasInstantPyroblast := fire.InstantPyroblastAura.IsActive()
			if !hasInstantPyroblast && fire.PresenceOfMindAura != nil {
				fire.PresenceOfMindAura.Deactivate(sim)
			}
			baseDamage := fire.CalcAndRollDamageRange(sim, pyroblastScaling, pyroblastVariance)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			if hasInstantPyroblast {
				fire.InstantPyroblastAura.Deactivate(sim)
			}
			fire.HeatingUpSpellHandler(sim, spell, result, func() {
				if hasInstantPyroblast {
					instantPyroblastDotMod.Activate()
				}
				spell.RelatedDotSpell.Cast(sim, target)
				if hasInstantPyroblast {
					instantPyroblastDotMod.Deactivate()
				}
				spell.DealDamage(sim, result)
			})
		},
	})

	fire.Pyroblast.RelatedDotSpell = fire.RegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(1),
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: mage.MageSpellPyroblastDot,
		Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		CritMultiplier:   fire.DefaultCritMultiplier(),
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "PyroblastDoT",
			},
			NumberOfTicks:       6,
			TickLength:          time.Second * 3,
			BonusCoefficient:    pyroblastDotCoefficient,
			AffectedByCastSpeed: true,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, fire.CalcScalingSpellDmg(pyroblastDotScaling))
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Apply(sim)
		},
	})
}
