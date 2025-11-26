package feral

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/druid"
)

func (cat *FeralDruid) applyPredatorySwiftness() {
	cat.PredatorySwiftnessAura = core.BlockPrepull(cat.RegisterAura(core.Aura{
		Label:    "Predatory Swiftness",
		ActionID: core.ActionID{SpellID: 69369},
		Duration: time.Second * 12,

		OnGain: func(_ *core.Aura, _ *core.Simulation) {
			cat.HealingTouch.CastTimeMultiplier -= 1
			cat.HealingTouch.Cost.PercentModifier *= -1
			cat.HealingTouch.FormMask |= druid.Cat
		},

		OnExpire: func(_ *core.Aura, _ *core.Simulation) {
			cat.HealingTouch.CastTimeMultiplier += 1
			cat.HealingTouch.Cost.PercentModifier /= -1
			cat.HealingTouch.FormMask ^= druid.Cat
		},

		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if cat.HealingTouch.IsEqual(spell) {
				aura.Deactivate(sim)
			}
		},
	}))

	// Predatory Swiftness only procs off successfully landed hits, but the
	// CPs (which we need for calculating the proc chance) have already been
	// spent by the time OnSpellHitDealt is called, so we need to cache the
	// CP value in an additional OnApplyEffects callback.
	var cpSnapshot int32

	procPredatorySwiftness := func(sim *core.Simulation) {
		procChance := 0.2 * float64(cpSnapshot)

		if sim.Proc(procChance, "Predatory Swiftness") {
			cat.PredatorySwiftnessAura.Activate(sim)
		}

		cpSnapshot = 0
	}

	cat.RegisterAura(core.Aura{
		Label:    "Predatory Swiftness Trigger",
		Duration: core.NeverExpires,

		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},

		OnApplyEffects: func(aura *core.Aura, _ *core.Simulation, _ *core.Unit, spell *core.Spell) {
			if spell.Matches(druid.DruidSpellFinisher) {
				cpSnapshot = aura.Unit.ComboPoints()
			}
		},

		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.Matches(druid.DruidSpellFinisher) && result.Landed() {
				procPredatorySwiftness(sim)
			}
		},

		OnCastComplete: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Matches(druid.DruidSpellSavageRoar) {
				procPredatorySwiftness(sim)
			}
		},
	})
}
