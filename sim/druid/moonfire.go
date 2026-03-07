package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const (
	MoonfireDotBonusCoeff    = 0.12999999523
	MoonfireDotTickLength    = time.Second * 3
	MoonfireDotNumberOfTicks = 4
	MoonfireDotTotalDamage   = 600

	MoonfireImpactBonusCoeff = 0.15000000596
	MoonfireImpactMinDmg     = 305
	MoonfireImpactMaxDmg     = 357
)

func (druid *Druid) registerMoonfireSpell() {
	druid.registerMoonfireImpactSpell()
	druid.registerMoonfireDoTSpell()
}

func (druid *Druid) registerMoonfireDoTSpell() {
	druid.Moonfire.RelatedDotSpell = druid.Unit.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26988}.WithTag(1),
		SpellSchool:    core.SpellSchoolArcane,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: DruidSpellMoonfireDoT,
		Flags:          core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Moonfire",
			},
			NumberOfTicks:       MoonfireDotNumberOfTicks,
			TickLength:          MoonfireDotTickLength,
			AffectedByCastSpeed: false,
			BonusCoefficient:    MoonfireDotBonusCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, MoonfireDotTotalDamage)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeAlwaysHitNoHitCounter)

			spell.Dot(target).Apply(sim)
			spell.DealOutcome(sim, result)
		},
	})
}

func (druid *Druid) registerMoonfireImpactSpell() {
	druid.Moonfire = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 26988},
		SpellSchool:    core.SpellSchoolArcane,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: DruidSpellMoonfire,
		Flags:          core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: 495,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		BonusCoefficient: MoonfireImpactBonusCoeff,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   druid.DefaultSpellCritMultiplier(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := druid.CalcAndRollDamageRange(sim, MoonfireImpactMinDmg, MoonfireImpactMaxDmg)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			if result.Landed() {
				druid.Moonfire.RelatedDotSpell.Cast(sim, target)
			}

			spell.DealDamage(sim, result)
		},
	})
}
