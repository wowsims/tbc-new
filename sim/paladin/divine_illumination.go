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
	paladin.DivineIlluminationAura = paladin.RegisterAura(core.Aura{
		Label:    "Divine Illumination" + paladin.Name,
		ActionID: actionId,
		Duration: time.Second * 15,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -0.5,
	})

	paladin.DivineIlluminationSpell = paladin.RegisterSpell(core.SpellConfig{
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
			paladin.DivineIlluminationAura.Activate(sim)
		},
	})

	paladin.AddMajorCooldown(core.MajorCooldown{
		Spell:    paladin.DivineIlluminationSpell,
		Priority: core.CooldownPriorityLow,
		Type:     core.CooldownTypeMana,
	})
}
