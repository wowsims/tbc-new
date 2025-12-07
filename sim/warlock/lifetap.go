package warlock

import (
	"math"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (warlock *Warlock) registerLifeTap() {
	actionID := core.ActionID{SpellID: 1454}
	manaMetrics := warlock.NewManaMetrics(actionID)
	healthCost := 582.0
	baseRestore := healthCost * (1.0 + 0.1*float64(warlock.Talents.ImprovedLifeTap))

	petRestore := 0.3333 * float64(warlock.Talents.ManaFeed)
	var petManaMetrics []*core.ResourceMetrics
	if warlock.Talents.ManaFeed > 0 {
		petManaMetrics = append(petManaMetrics, warlock.ActivePet.NewManaMetrics(actionID))
	}

	warlock.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellLifeTap,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Life tap adds 0.8*sp to mana restore
			restore := baseRestore + (math.Max(warlock.GetStat(stats.SpellPower), warlock.GetStat(stats.ShadowPower)) * 0.8)
			warlock.RemoveHealth(sim, healthCost)
			warlock.AddMana(sim, restore, manaMetrics)

			if warlock.Talents.ManaFeed > 0 {
				warlock.ActivePet.AddMana(sim, restore*petRestore, petManaMetrics[0])
			}
		},
	})
}
