package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const seedTickCoeff = 0.25
const seedExploCoeff = 0.143

func (warlock *Warlock) registerSeed() {
	actionID := core.ActionID{SpellID: 27243}
	type seedOptions struct {
		damageTaken float64
		isSoulBurn  bool
	}
	seedPropertyTracker := make([]seedOptions, len(warlock.Env.AllUnits))

	seedExplosion := warlock.RegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(1), // actually 27285
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagPassiveSpell,
		ClassSpellMask: WarlockSpellSeedOfCorruptionExplosion,

		DamageMultiplier: 1,
		CritMultiplier:   warlock.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: seedExploCoeff,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDmg := warlock.CalcAndRollDamageRange(sim, 1110, 1290)
			for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
				result := spell.CalcAndDealDamage(sim, aoeTarget, baseDmg, spell.OutcomeMagicHitAndCrit)
				if result.Landed() {
					if warlock.Talents.ShadowEmbrace > 0 {
						warlock.ShadowEmbraceAura.Activate(sim)
					}
				}
			}
		},
	})

	trySeedPop := func(sim *core.Simulation, target *core.Unit, dmg float64) {
		seedPropertyTracker[target.UnitIndex].damageTaken += dmg
		if seedPropertyTracker[target.UnitIndex].damageTaken >= 1044 {
			warlock.Seed.Dot(target).Deactivate(sim)
			seedExplosion.Cast(sim, target)
		}
	}

	warlock.Seed = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		MissileSpeed:   28,
		ClassSpellMask: WarlockSpellSeedOfCorruption,

		ManaCost: core.ManaCostOptions{BaseCostPercent: 6},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: 2000 * time.Millisecond,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Seed",
				OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if !result.Landed() || spell.SpellSchool != core.SpellSchoolShadow {
						return
					}

					trySeedPop(sim, result.Target, result.Damage)
				},
				OnPeriodicDamageTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if spell.SpellSchool != core.SpellSchoolShadow {
						return
					}
					trySeedPop(sim, result.Target, result.Damage)
				},
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					seedPropertyTracker[aura.Unit.UnitIndex].damageTaken = 0
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					seedPropertyTracker[aura.Unit.UnitIndex].damageTaken = 0
				},
			},

			NumberOfTicks:    6,
			TickLength:       3 * time.Second,
			BonusCoefficient: seedTickCoeff,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, 174)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				result := dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				trySeedPop(sim, target, result.Damage)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				if result.Landed() {
					if warlock.Options.DetonateSeed {
						seedExplosion.Cast(sim, target)
					} else {
						spell.Dot(target).Apply(sim)
					}
				}
			})
		},
	})
}
