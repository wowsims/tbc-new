package priest

import (
	"fmt"
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var VampiricTouchRankMap = genRanks.VampiricTouch

func (priest *Priest) registerVampiricTouchSpell(rank shared.SpellRank) {
	tick := rank.Periodic.(shared.SpellRankPeriodic)

	manaMetrics := priest.NewManaMetrics(core.ActionID{SpellID: rank.SpellID}.WithTag(1))

	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellVampiricTouch,
		Rank:           rank.Rank,
		MaxRange:       rank.MaxRange,

		ManaCost: core.ManaCostOptions{
			FlatCost: rank.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      rank.GCD,
				CastTime: rank.CastTime,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("VampiricTouch-%d", rank.Rank),
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					aura.AttachProcTrigger(core.ProcTrigger{
						Name:               "VampiricTouch-ManaReturn",
						CanProcFromProcs:   true, // 34914, 34916, 34917 carry the bit.
						Callback:           core.CallbackOnSpellHitTaken | core.CallbackOnPeriodicDamageTaken,
						ClassSpellMask:     PriestShadowSpells,
						RequireDamageDealt: true,
						Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
							priest.AddMana(sim, result.Damage*0.05, manaMetrics)
						},
					})
				},
			},
			NumberOfTicks:       tick.NumberOfTicks,
			TickLength:          tick.TickLength,
			AffectedByCastSpeed: false,
			BonusCoefficient:    rank.Periodic.BonusCoefficient(),

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, shared.SpellRankMin(rank.Periodic))
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			if useSnapshot {
				dot := spell.Dot(target)
				return dot.CalcSnapshotDamage(sim, target, spell.OutcomeExpectedMagicHit)
			}
			return spell.CalcPeriodicDamage(sim, target, shared.SpellRankMin(rank.Periodic), spell.OutcomeExpectedMagicHit)
		},
	})
}
