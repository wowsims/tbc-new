package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

// Divine Illumination (Talent)
// https://www.wowhead.com/tbc/spell=31842
//
// Reduces the mana cost of all spells by 50% for 15 sec.
func (paladin *Paladin) registerDivineIllumination() {
	actionId := core.ActionID{SpellID: 31842}
	aura := paladin.RegisterAura(core.Aura{
		Label:    "Divine Illumination" + paladin.Name,
		ActionID: actionId,
		Duration: time.Second * 15,

		// TODO: Spell says reduces cost of all spells but wowhead shows it only reduces holy shock.
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range paladin.Spellbook {
				spell.Cost.PercentModifier *= 0.5
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range paladin.Spellbook {
				spell.Cost.PercentModifier /= 0.5
			}
		},
	})

	spell := paladin.RegisterSpell(core.SpellConfig{
		ActionID: actionId,
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskDivineIllumination,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},
	})

	paladin.DivineIlluminationSpell = spell
	paladin.DivineIlluminationAura = aura

	paladin.AddMajorCooldown(core.MajorCooldown{
		Spell:    spell,
		Priority: core.CooldownPriorityLow,
		Type:     core.CooldownTypeMana,
	})
}
