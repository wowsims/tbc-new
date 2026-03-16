package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (hunter *Hunter) registerSerpentStingSpell() {
	hunter.SerpentSting = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27016},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskProc,
		ClassSpellMask: HunterSpellSerpentSting,
		Flags:          core.SpellFlagAPL,

		MissileSpeed: 40,
		MinRange:     core.MaxMeleeRange,
		MaxRange:     HunterBaseMaxRange,

		ManaCost: core.ManaCostOptions{
			FlatCost: 275,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		CritMultiplier:   0,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Serpent Sting",
				Tag:   "Sting",
			},

			NumberOfTicks: 5,
			TickLength:    time.Second * 3,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				baseDmg := dot.Spell.RangedAttackPower(target)*0.1 + 132
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
