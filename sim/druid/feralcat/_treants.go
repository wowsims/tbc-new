package feral

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
	"github.com/wowsims/tbc/sim/druid"
)

type FeralTreant struct {
	*druid.DefaultTreantImpl

	owner *FeralDruid

	Rake *core.Spell
}

func (cat *FeralDruid) newTreant() *FeralTreant {
	treant := &FeralTreant{
		DefaultTreantImpl: cat.NewDefaultTreant(druid.TreantConfig{
			NonHitExpStatInheritance: func(ownerStats stats.Stats) stats.Stats {
				return stats.Stats{
					stats.Health:              0.4 * ownerStats[stats.Health],
					stats.AttackPower:         ownerStats[stats.AttackPower],
					stats.PhysicalCritPercent: ownerStats[stats.PhysicalCritPercent],
					stats.HasteRating:         ownerStats[stats.HasteRating],
					stats.MasteryRating:       ownerStats[stats.MasteryRating],
				}
			},

			EnableAutos:             true,
			WeaponDamageCoefficient: 2,
		}),

		owner: cat,
	}

	cat.AddPet(treant)

	return treant
}

func (cat *FeralDruid) registerTreants() {
	for idx := range cat.Treants {
		cat.Treants[idx] = cat.newTreant()
	}
}

func (treant *FeralTreant) Initialize() {
	// Raw parameter from spell database
	const coefficient = 0.02999999933
	const bonusCoefficientFromAP = 0.10000000149

	// Scaled parameters for spell code
	flatBaseDamage := coefficient * treant.owner.ClassSpellScaling // ~32.8422

	treant.Rake = treant.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 150017},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagIgnoreArmor,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second * 15,
			},

			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		CritMultiplier:   treant.DefaultCritMultiplier(),
		ThreatMultiplier: 1,
		MaxRange:         core.MaxMeleeRange,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:    fmt.Sprintf("Rake (Treant %d)", treant.UnitIndex),
				Duration: time.Second * 15,
			},

			NumberOfTicks: 5,
			TickLength:    time.Second * 3,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.SnapshotPhysical(target, flatBaseDamage+bonusCoefficientFromAP*dot.Spell.MeleeAttackPower())
			},

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := flatBaseDamage + bonusCoefficientFromAP*spell.MeleeAttackPower()
			spell.DamageMultiplier = 1.0 + BaseMasteryMod + MasteryModPerPoint*treant.GetMasteryPoints()
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
		},
	})
}

func (treant *FeralTreant) ExecuteCustomRotation(sim *core.Simulation) {
	if treant.GCD.IsReady(sim) {
		treant.Rake.Cast(sim, treant.CurrentTarget)
	}
}
