package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (hunter *Hunter) registerSteadyShotSpell() {
	hunter.SteadyShot = hunter.RegisterRangedSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 34120},
		SpellSchool:    core.SpellSchoolPhysical,
		ClassSpellMask: HunterSpellSteadyShot,
		ProcMask:       core.ProcMaskRangedSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: 110,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: time.Millisecond * 1500,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Steady Shot isn't affected by Ammo, Scopes or Adamantite Weightstone
			weaponDamage := hunter.AutoAttacks.Ranged().BaseDamage(sim) - hunter.AmmoDamageBonus

			if ranged := hunter.Ranged(); ranged != nil && ranged.Enchant.EffectID == 2722 {
				weaponDamage -= 10
			} else if ranged != nil && ranged.Enchant.EffectID == 2723 {
				weaponDamage -= 12
			}
			if hunter.Consumables.OhImbueId == 34340 || (hunter.Consumables.MhImbueId == 34340 && !hunter.windFuryEnabled) {
				weaponDamage -= 12
			}

			baseDamage := 0.2*spell.RangedAttackPower(target) +
				weaponDamage*2.8/hunter.AutoAttacks.Ranged().SwingSpeed +
				hunter.talonOfAlarBonus() +
				150

			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}, true)
}
