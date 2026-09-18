package priest

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

// Devouring Plague - Undead Racial
// Shadow school DoT, 3 min cooldown, 24s duration

var DevouringPlagueRankMap = genRanks.DevouringPlague

func (priest *Priest) registerDevouringPlagueSpell(rank shared.SpellRank, cdTimer *core.Timer) {
	healthMetrics := priest.NewHealthMetrics(core.ActionID{SpellID: rank.SpellID}.WithTag(1))

	priest.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.SpellID},
		SpellSchool:    core.SpellSchoolShadow,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: PriestSpellDevouringPlague,
		Rank:           rank.Rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: rank.Cost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: 180 * time.Second,
			},
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: fmt.Sprintf("DevouringPlague-%d", rank.Rank),
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					aura.AttachProcTrigger(core.ProcTrigger{
						Name:               "DevouringPlague-Heal",
						Callback:           core.CallbackOnPeriodicDamageTaken,
						ClassSpellMask:     PriestSpellDevouringPlague,
						RequireDamageDealt: true,
						Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
							priest.GainHealth(sim, result.Damage, healthMetrics)
						},
					})
				},
			},
			NumberOfTicks:       8,
			TickLength:          3 * time.Second,
			AffectedByCastSpeed: false,
			BonusCoefficient:    rank.Periodic.Coef,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, rank.Periodic.Tick)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
				spell.Dot(target).TickOnce(sim)
			}
			spell.DealOutcome(sim, result)
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			if useSnapshot {
				dot := spell.Dot(target)
				return dot.CalcSnapshotDamage(sim, target, spell.OutcomeExpectedMagicHit)
			}
			return spell.CalcPeriodicDamage(sim, target, rank.Periodic.Tick, spell.OutcomeExpectedMagicHit)
		},
	})
}
