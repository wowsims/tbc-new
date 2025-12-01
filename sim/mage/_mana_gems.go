package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerManaGems() {

	var manaGain float64
	actionID := core.ActionID{ItemID: 36799}
	manaMetrics := mage.NewManaMetrics(actionID)
	maxManaGems := 3

	minManaGain := 42750.0
	maxManaGain := 47250.0

	var remainingManaGems int
	mage.RegisterResetEffect(func(sim *core.Simulation) {
		remainingManaGems = maxManaGems
	})

	spell := mage.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL | core.SpellFlagHelpful,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: time.Minute * 2,
			},
		},

		// Don't use if we don't have any gems remaining!
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return remainingManaGems != 0
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			manaGain = sim.Roll(minManaGain, maxManaGain)
			mage.AddMana(sim, manaGain, manaMetrics)

			remainingManaGems--
			if remainingManaGems == 0 {
				mage.GetMajorCooldown(actionID).Disable()
			}
		},
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell:    spell,
		Priority: core.CooldownPriorityDefault,
		Type:     core.CooldownTypeMana,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			// Only pop if we have less than the max mana provided by the gem minus 1mp5 tick.
			totalRegen := character.ManaRegenPerSecondWhileCombat() * 5
			return character.MaxMana()-(character.CurrentMana()+totalRegen) >= maxManaGain
		},
	})
}
