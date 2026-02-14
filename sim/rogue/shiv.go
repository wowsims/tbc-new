package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (rogue *Rogue) registerShivSpell() {
	shivCostMod := rogue.AddDynamicMod(core.SpellModConfig{
		Kind:      core.SpellMod_PowerCost_Flat,
		ClassMask: RogueSpellShiv,
		IntValue:  rogue.getShivCostModifier(),
	})
	shivCostMod.Activate()

	rogue.Shiv = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 5938},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeOHSpecial,
		Flags:          core.SpellFlagCannotBeDodged | core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
		ClassSpellMask: RogueSpellShiv,

		EnergyCost: core.EnergyCostOptions{
			Cost: 20,
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

	rogue.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(s *core.Simulation, is proto.ItemSlot) {
		shivCostMod.UpdateIntValue(rogue.getShivCostModifier())
		shivCostMod.Activate()
	})
}

func (rogue *Rogue) getShivCostModifier() int32 {
	if ohWeapon := rogue.GetOHWeapon(); ohWeapon != nil {
		return int32(10 * ohWeapon.SwingSpeed)
	}

	return 0
}
