package hunter

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var serpentStingRank = genRanks.SerpentSting.BySpellID(27016)

func (hunter *Hunter) registerSerpentStingSpell() {
	serpentStingTick := serpentStingRank.Periodic.(shared.SpellRankPeriodic)

	hunter.SerpentSting = hunter.RegisterRangedSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 27016},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeRanged,
		// A cast, not a proc, but one that must not read as a ranged hit to on-hit listeners; what
		// the sting's application should count as is a separate question. Matches only listeners
		// that state no mask.
		ProcMask:       core.ProcMaskEmpty,
		ClassSpellMask: HunterSpellSerpentSting,
		Flags:          core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: serpentStingRank.Cost,
		},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Serpent Sting",
				Tag:   "Sting",
			},

			NumberOfTicks: serpentStingTick.Ticks,
			TickLength:    serpentStingTick.Period,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				baseDmg := dot.Spell.RangedAttackPower(target)*0.02 + serpentStingRank.Periodic.Damage(sim)
				dot.Snapshot(target, baseDmg)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeRangedHit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				if result.Landed() {
					dot := spell.Dot(target)
					activeSting := target.GetActiveAuraWithTag("Sting")
					if activeSting != nil && activeSting != dot.Aura {
						activeSting.Deactivate(sim)
					}
					dot.Apply(sim)
				}
				spell.DealOutcome(sim, result)
			})
		},
	})
}
