package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (rogue *Rogue) registerShivSpell() {
	var baseCost float64
	if ohWeapon := rogue.GetOHWeapon(); ohWeapon != nil {
		baseCost = 20 + 10*ohWeapon.SwingSpeed
	}

	rogue.Shiv = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 5938},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeOHSpecial,
		Flags:          core.SpellFlagCannotBeDodged | core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
		ClassSpellMask: RogueSpellShiv,

		EnergyCost: core.EnergyCostOptions{
			Cost: int32(baseCost),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           rogue.DefaultMeleeCritMultiplier(),
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := spell.Unit.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialNoParry)

			if result.Landed() {
				rogue.AddComboPoints(sim, 1, spell.ComboPointMetrics())

				switch rogue.Consumables.OhImbueId {
				case deadlyImbueID:
					rogue.ShivDeadlyPoison.Cast(sim, target)
				case instantImbueID:
					rogue.ShivInstantPoison.Cast(sim, target)
				case woundImbueID:
					rogue.ShivWoundPoison.Cast(sim, target)
				}
			}
		},
	})
}
