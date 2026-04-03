package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (druid *Druid) registerShredSpell() {
	energyCost := int32(60) - int32(9)*int32(druid.Talents.ShreddingAttacks)

	// Flat damage bonus from gear/idols.
	// Nordrassil Harness 4pc: +75 (applied post-registration via druid.ShredFlatBonus).
	// Everbloom Idol (29390): +88.
	flatDamageBonus := 405.0
	if druid.HasItemEquipped(29390, []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}) {
		flatDamageBonus += 88
	}

	druid.Shred = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27002},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellShred,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost:   energyCost,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},

		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return !druid.PseudoStats.InFrontOfTarget && !druid.CannotShredTarget
		},

		// Weapon damage * 2.25 + flatDamageBonus, boosted by 30% if Mangle is active.
		DamageMultiplier: 2.25,
		CritMultiplier:   druid.FeralCritMultiplier(),
		ThreatMultiplier: 1,
		MaxRange:         core.MaxMeleeRange,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := (flatDamageBonus+druid.ShredFlatBonus)/2.25 + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
			if druid.MangleAuras != nil && druid.MangleAuras.Get(target).IsActive() {
				baseDamage *= 1.3
			}

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				druid.AddComboPoints(sim, 1, spell.ComboPointMetrics())
			} else {
				spell.IssueRefund(sim)
			}
		},

		ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
			baseDamage := (flatDamageBonus+druid.ShredFlatBonus)/2.25 + spell.Unit.AutoAttacks.MH().CalculateAverageWeaponDamage(spell.MeleeAttackPower(target))
			if druid.MangleAuras != nil && druid.MangleAuras.Get(target).IsActive() {
				baseDamage *= 1.3
			}
			return spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMeleeWeaponSpecialHitAndCrit)
		},
	})
}

func (druid *Druid) CurrentShredCost() float64 {
	return druid.Shred.Cost.GetCurrentCost()
}
