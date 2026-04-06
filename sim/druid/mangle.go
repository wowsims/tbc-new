package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (druid *Druid) registerMangleAuras() {
	if druid.MangleAuras != nil {
		return
	}
	druid.MangleAuras = druid.NewEnemyAuraArray(core.MangleAura)
}

func (druid *Druid) registerMangleCatSpell() {
	if !druid.Talents.Mangle {
		return
	}

	druid.registerMangleAuras()

	// Idol of the Wild (28064): +24 flat damage to Mangle (Cat).
	flatDamageBonus := 264.0
	if druid.HasItemEquipped(28064, []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}) {
		flatDamageBonus += 24
	}

	druid.MangleCat = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 33983},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellMangleCat,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost:   45,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1.6,
		CritMultiplier:   druid.FeralCritMultiplier(),
		ThreatMultiplier: 1,
		MaxRange:         core.MaxMeleeRange,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := flatDamageBonus/1.6 + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				druid.AddComboPoints(sim, 1, spell.ComboPointMetrics())
				druid.MangleAuras.Get(target).Activate(sim)
			} else {
				spell.IssueRefund(sim)
			}
		},

		ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			baseDamage := flatDamageBonus/1.6 + spell.Unit.AutoAttacks.MH().CalculateAverageWeaponDamage(spell.MeleeAttackPower(target))
			return spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMeleeWeaponSpecialHitAndCrit)
		},
	})
}

func (druid *Druid) registerMangleBearSpell() {
	if !druid.Talents.Mangle {
		return
	}

	druid.registerMangleAuras()

	// Idol of the Wild (28064): +52 flat damage to Mangle (Bear).
	flatDamageBonus := 155.0
	if druid.HasItemEquipped(28064, []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}) {
		flatDamageBonus += 52
	}

	druid.MangleBear = druid.RegisterSpell(Bear, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 33987},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellMangleBear,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		RageCost: core.RageCostOptions{
			Cost:   20,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Second * 6,
			},
		},

		DamageMultiplier: 1.15,
		CritMultiplier:   druid.FeralCritMultiplier(),
		ThreatMultiplier: 1.5 / 1.15,
		MaxRange:         core.MaxMeleeRange,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := flatDamageBonus/1.15 + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				druid.MangleAuras.Get(target).Activate(sim)
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}

func (druid *Druid) CurrentMangleCatCost() float64 {
	return druid.MangleCat.Cost.GetCurrentCost()
}
