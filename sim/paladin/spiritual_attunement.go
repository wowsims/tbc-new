package paladin

import "github.com/wowsims/tbc/sim/core"

// Spiritual Attunement (Rank 2, SpellID 33776): Whenever you are healed by another character's spell,
// you regain 10% of the amount healed as mana.
// In the sim, this is modeled as mana return from damage taken (since the healing model offsets damage).
func (paladin *Paladin) RegisterSpiritualAttunement() {
	coeff := 0.1
	if paladin.GetAuraByID(core.ActionID{SpellID: 38426}).IsActive() {
		coeff *= 1.1 // Lightbringer Armor 2pc: +10% mana from Spiritual Attunement
	}

	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: 33776})

	core.MakePermanent(paladin.RegisterAura(core.Aura{
		Label: "Spiritual Attunement",
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Damage > 0 {
				paladin.AddMana(sim, result.Damage*coeff, manaMetrics)
			}
		},
	}))
}
