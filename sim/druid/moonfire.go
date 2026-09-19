package druid

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var moonfireRank = genRanks.Moonfire.BySpellID(26988)
var moonfireTick = moonfireRank.Periodic.(shared.SpellRankPeriodic)

func (druid *Druid) registerMoonfireSpell() {
	druid.registerMoonfireImpactSpell()
	druid.registerMoonfireDoTSpell()
}

func (druid *Druid) registerMoonfireDoTSpell() {
	druid.Moonfire.RelatedDotSpell = druid.Unit.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: moonfireRank.SpellID}.WithTag(1),
		SpellSchool:    core.SpellSchoolArcane,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: DruidSpellMoonfireDoT,
		Flags:          core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Moonfire",
			},
			NumberOfTicks:       moonfireTick.NumberOfTicks,
			TickLength:          moonfireTick.TickLength,
			AffectedByCastSpeed: false,
			BonusCoefficient:    moonfireTick.Coef,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, moonfireTick.Tick)
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
		ActionID:       core.ActionID{SpellID: moonfireRank.SpellID},
		SpellSchool:    core.SpellSchoolArcane,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: DruidSpellMoonfire,
		Flags:          core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: moonfireRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: moonfireRank.GCD,
			},
		},

		BonusCoefficient: moonfireRank.Direct.BonusCoefficient(),
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		MaxRange:         moonfireRank.MaxRange,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := moonfireRank.Direct.Damage(sim)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			if result.Landed() {
				druid.Moonfire.RelatedDotSpell.Cast(sim, target)
			}

			spell.DealDamage(sim, result)
		},
	})
}
