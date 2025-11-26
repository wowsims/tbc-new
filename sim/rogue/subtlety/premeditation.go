package subtlety

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/rogue"
)

func (subRogue *SubtletyRogue) registerPremeditation() {
	comboMetrics := subRogue.NewComboPointMetrics(core.ActionID{SpellID: 14183})
	shouldTimeout := false

	premedAura := subRogue.RegisterAura(core.Aura{
		Label:    "Premed Timeout Aura",
		Duration: time.Second * 18,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			shouldTimeout = true
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Flags.Matches(rogue.SpellFlagFinisher) && spell.ClassSpellMask == rogue.RogueSpellSliceAndDice {
				shouldTimeout = false
			}
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.Flags.Matches(rogue.SpellFlagFinisher) && result.Landed() {
				shouldTimeout = false
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			// Remove 2 points because no finisher was casted
			if shouldTimeout {
				subRogue.AddComboPoints(sim, -2, comboMetrics)
				shouldTimeout = false
			}
		},
		OnEncounterStart: func(aura *core.Aura, sim *core.Simulation) {
			// Reset Premed back to 20s CD on EncounterStart
			if !subRogue.Premeditation.CD.IsReady(sim) {
				subRogue.Premeditation.CD.Set(time.Second * 20)
			}

			// PENDING VALIDATION: If SnD is active but was casted before our last Premed, then we can't have any points at all
			if subRogue.SliceAndDiceAura.IsActive() && shouldTimeout {
				subRogue.ResetComboPoints(sim, 0)
				shouldTimeout = false
			}
		},
	})

	subRogue.Premeditation = subRogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 14183},
		Flags:          core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,
		ClassSpellMask: rogue.RogueSpellPremeditation,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				Cost: 0,
				GCD:  0,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    subRogue.NewTimer(),
				Duration: time.Second * 20,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return subRogue.IsStealthed() || subRogue.HasActiveAura("Shadowmeld")
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			subRogue.AddComboPointsOrAnticipation(sim, 2, comboMetrics)
			premedAura.Activate(sim)
		},
	})

	subRogue.AddMajorCooldown(core.MajorCooldown{
		Spell:              subRogue.Premeditation,
		Type:               core.CooldownTypeDPS,
		Priority:           core.CooldownPriorityLow,
		AllowSpellQueueing: true,
	})
}
