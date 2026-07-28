package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerManaGems() {

	var manaGain float64
	actionID := core.ActionID{ItemID: 22044}
	manaMetrics := mage.NewManaMetrics(actionID)

	minManaGain := 2340.0
	maxManaGain := 2460.0

	manaGemAura := core.MakePermanent(mage.GetOrRegisterAura(core.Aura{
		Label:     "Mana Gem Charges",
		ActionID:  actionID,
		MaxStacks: 3,
	})).ApplyOnReset(func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, 3)
	})

	manaGem := mage.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: MageSpellManaGem,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    mage.GetConjuredCD(),
				Duration: time.Minute * 2,
			},
			SharedCD: core.Cooldown{
				Timer:    mage.GetCombatConsumableCD(),
				Duration: time.Minute * 2,
			},
		},

		// Don't use if we don't have any gems remaining!
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return manaGemAura.GetStacks() > 0
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			manaGemAura.RemoveStack(sim)
			manaGain = sim.Roll(minManaGain, maxManaGain)
			if mage.SerpentCoilBraid.IsActive() {
				manaGain *= 1.25
			}
			mage.AddMana(sim, manaGain, manaMetrics)
		},
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: manaGem,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(sim *core.Simulation, char *core.Character) bool {
			manaGain = sim.Roll(minManaGain, maxManaGain)
			if mage.SerpentCoilBraid.IsActive() {
				manaGain *= 1.25
			}

			return char.CurrentMana()+manaGain+char.SpiritManaRegenPerSecond() <= char.MaxMana()
		},
	})
}
