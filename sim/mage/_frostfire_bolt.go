package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

const frostfireBoltCoefficient = 1.5        // Per https://wago.tools/db2/SpellEffect?build=5.5.0.60802&filter%5BSpellID%5D=44614 Field "EffetBonusCoefficient"
const frostfireBoltScaling = 1.5            // Per https://wago.tools/db2/SpellEffect?build=5.5.0.60802&filter%5BSpellID%5D=44614 Field "Coefficient"
const frostfireBoltVariance = 0.23999999464 // Per https://wago.tools/db2/SpellEffect?build=5.5.0.60802&filter%5BSpellID%5D=44614 Field "Variance"

func (mage *Mage) frostfireBoltConfig(config core.SpellConfig) core.SpellConfig {
	return core.SpellConfig{
		ActionID:       config.ActionID,
		SpellSchool:    core.SpellSchoolFrostfire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          config.Flags,
		ClassSpellMask: MageSpellFrostfireBolt,
		MissileSpeed:   28,

		ManaCost: config.ManaCost,
		Cast:     config.Cast,

		DamageMultiplier: config.DamageMultiplier,
		CritMultiplier:   mage.DefaultCritMultiplier(),
		BonusCoefficient: frostfireBoltCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: config.ApplyEffects,
	}
}

func (mage *Mage) registerFrostfireBoltSpell() {
	actionID := core.ActionID{SpellID: 44614}
	mageSpecFrost := mage.Spec == proto.Spec_SpecFrostMage
	mageSpecFire := mage.Spec == proto.Spec_SpecFireMage

	mage.RegisterSpell(mage.frostfireBoltConfig(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 4,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 2750,
			},
		},

		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if (mage.BrainFreezeAura == nil || !mage.BrainFreezeAura.IsActive()) && mage.PresenceOfMindAura != nil {
				mage.PresenceOfMindAura.Deactivate(sim)
			}
			damageMultiplier := 1.0

			spell.DamageMultiplier *= damageMultiplier
			baseDamage := mage.CalcAndRollDamageRange(sim, frostfireBoltScaling, frostfireBoltVariance)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.DamageMultiplier /= damageMultiplier

			if result.Landed() && mageSpecFrost {
				mage.ProcFingersOfFrost(sim, spell)
			}

			if mage.BrainFreezeAura != nil {
				mage.BrainFreezeAura.Deactivate(sim)
			}

			if mageSpecFire && spell.TravelTime() > time.Duration(FireSpellMaxTimeUntilResult) {
				pa := sim.GetConsumedPendingActionFromPool()
				pa.NextActionAt = sim.CurrentTime + time.Duration(FireSpellMaxTimeUntilResult)

				pa.OnAction = func(sim *core.Simulation) {
					spell.DealDamage(sim, result)

					mage.HandleHeatingUp(sim, spell, result)
				}

				sim.AddPendingAction(pa)
			} else {
				spell.WaitTravelTime(sim, func(sim *core.Simulation) {
					spell.DealDamage(sim, result)
					if result.Landed() && mageSpecFrost {
						mage.GainIcicle(sim, target, result.Damage)
					}
					if mageSpecFire {
						mage.HandleHeatingUp(sim, spell, result)
					}
				})
			}
		},
	}))
}
