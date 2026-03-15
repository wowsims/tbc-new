package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

const seedTickCoeff = 0.25
const seedPopCoeff = 0.214
const seedExplosionCoeff = 0.143
const seedTriggerBaseDamage = 1044.0

func (warlock *Warlock) registerSeed() {
	warlock.SeedOfCorruptionBonusDamage = 0

	actionID := core.ActionID{SpellID: 27243}
	type seedOptions struct {
		damageTaken float64
		isSoulBurn  bool
	}
	seedPropertyTracker := make([]seedOptions, len(warlock.Env.AllUnits))
	var spell *core.Spell
	seedExplosion := warlock.RegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(2), // actually 27285
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagPassiveSpell | core.SpellFlagIgnoreAttackerModifiers,
		ClassSpellMask: WarlockSpellSeedOfCorruptionExplosion,

		DamageMultiplier: 1,
		CritMultiplier:   warlock.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 0,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			targetCount := sim.Environment.ActiveTargetCount()
			if targetCount < 2 {
				return
			}

			nextTarget := sim.Environment.NextActiveTargetUnit(target)
			maxHits := 0
			hitResults := make([]*core.SpellResult, 0)
			damageResults := make([]*core.SpellResult, 0)

			for range targetCount - 1 {
				result := spell.CalcOutcome(sim, nextTarget, spell.OutcomeMagicHitNoHitCounter)
				hitResults = append(hitResults, result)
				if result.Landed() {
					maxHits++
				}
				nextTarget = sim.Environment.NextActiveTargetUnit(nextTarget)
			}

			maxDamagePerMob := 13580 / float64(maxHits+1)
			for _, result := range hitResults {
				spell.Flags ^= core.SpellFlagIgnoreAttackerModifiers
				attackTable := spell.Unit.AttackTables[target.UnitIndex]
				baseDamage := warlock.CalcAndRollDamageRange(sim, 1110, 1290) + warlock.SeedOfCorruptionBonusDamage
				baseDamage += seedPopCoeff * spell.BonusDamage(attackTable)
				attackerMultiplier := spell.AttackerDamageMultiplier(attackTable, false)
				baseDamage *= attackerMultiplier
				spell.Flags |= core.SpellFlagIgnoreAttackerModifiers
				damageResults = append(damageResults, spell.CalcDamage(sim, result.Target, min(maxDamagePerMob, baseDamage), core.Ternary(result.Landed(), spell.OutcomeMagicCrit, spell.OutcomeAlwaysMiss)))
			}

			for _, result := range damageResults {
				spell.DealDamage(sim, result)
			}
		},
	})

	trySeedPop := func(sim *core.Simulation, target *core.Unit, dmg float64) {
		seedPropertyTracker[target.UnitIndex].damageTaken += dmg
		seedThreshold := seedTriggerBaseDamage + (warlock.GetStat(stats.SpellDamage) + warlock.GetStat(stats.ShadowDamage)*seedPopCoeff)
		if seedPropertyTracker[target.UnitIndex].damageTaken >= seedThreshold {
			spell.Dot(target).Deactivate(sim)
			seedExplosion.Cast(sim, target)
		}
	}

	spell = warlock.RegisterSpell(getSeedSpellConfig(core.SpellConfig{
		ActionID: actionID,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Seed",
				OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if !result.Landed() || !spell.SpellSchool.Matches(core.SpellSchoolShadow) {
						return
					}

					trySeedPop(sim, result.Target, result.Damage)
				},
				OnPeriodicDamageTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if !spell.SpellSchool.Matches(core.SpellSchoolShadow) {
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
				dot.Snapshot(target, seedTriggerBaseDamage/float64(dot.BaseTickCount))
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
					spell.Dot(target).Apply(sim)
				}
			})
		},
	}))

	// Instant seed
	warlock.RegisterSpell(getSeedSpellConfig(core.SpellConfig{
		ActionID: actionID.WithTag(1),
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				if result.Landed() {
					seedExplosion.Cast(sim, target)
				}
			})
		},
	}))
}

func getSeedSpellConfig(config core.SpellConfig) core.SpellConfig {
	return core.SpellConfig{
		ActionID:       config.ActionID,
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

		Dot: config.Dot,

		ApplyEffects: config.ApplyEffects,
	}
}
