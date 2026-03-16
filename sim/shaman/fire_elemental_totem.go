package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (shaman *Shaman) registerFireElementalTotem() {

	actionID := core.ActionID{SpellID: 2894}

	totalDuration := time.Second * 120

	fireElementalAura := shaman.RegisterAura(core.Aura{
		Label:    "Fire Elemental Totem",
		ActionID: actionID,
		Duration: totalDuration,
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			shaman.FireElemental.Disable(sim)
		},
	})

	shaman.FireElementalTotem = shaman.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		Flags:          core.SpellFlagAPL | SpellFlagInstant,
		ClassSpellMask: SpellMaskFireElementalTotem,
		ManaCost: core.ManaCostOptions{
			FlatCost: 680,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second * 1,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Minute * 20,
			},
			SharedCD: core.Cooldown{
				Timer:    shaman.GetOrInitTimer(&shaman.ElementalSharedCDTimer),
				Duration: time.Minute * 1,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			shaman.cancelFireTotems(sim)
			shaman.TotemExpirations[FireTotem] = sim.CurrentTime + fireElementalAura.Duration

			shaman.FireElemental.Disable(sim)
			shaman.FireElemental.EnableWithTimeout(sim, shaman.FireElemental, fireElementalAura.Duration)

			// Add a dummy aura to show in metrics
			fireElementalAura.Activate(sim)
		},
		RelatedSelfBuff: fireElementalAura,
	})

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: shaman.FireElementalTotem,
		Type:  core.CooldownTypeDPS,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			// Fele should only be cast by manual APL intervention
			return false
		},
	})
}
