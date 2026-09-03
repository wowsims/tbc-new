package paladin

import "github.com/wowsims/tbc/sim/core"

// Spiritual Attunement (Rank 2, SpellID 33776): Whenever you are healed by another character's spell,
// you regain 10% of the amount healed as mana.
// In the sim, this is modeled as mana return from damage taken (since the healing model offsets damage).
func (paladin *Paladin) RegisterSpiritualAttunement() {
	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: 33776})

	paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Spiritual Attunement",
		ActionID:           core.ActionID{SpellID: 33776},
		Callback:           core.CallbackOnSpellHitTaken,
		RequireDamageDealt: true,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			coeff := 0.1
			// Lightbringer Armor 2pc: +10% mana from Spiritual Attunement
			if paladin.T6_4pcAura.IsActive() {
				coeff *= 1.1
			}
			paladin.AddMana(sim, result.Damage*coeff, manaMetrics)
		},
	})
}
