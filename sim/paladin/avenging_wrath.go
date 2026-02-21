package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

// Avenging Wrath
// https://www.wowhead.com/tbc/spell=31884
//
// Increases all damage caused by 30% for 20 sec.
// Causes Forebearance, preventing the use of Divine Shield,
// Divine Protection, Blessing of Protection again for 1 min.
func (paladin *Paladin) registerAvengingWrath() {
	if paladin.Level < 70 {
		return
	}

	actionID := core.ActionID{SpellID: 31884}
	aura := paladin.RegisterAura(core.Aura{
		Label:    "Avenging Wrath" + paladin.Label,
		ActionID: actionID,
		Duration: time.Second * 20,
	}).AttachMultiplicativePseudoStatBuff(
		&paladin.PseudoStats.DamageDealtMultiplier, 1.3,
	)

	spell := paladin.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		SpellSchool: core.SpellSchoolHoly,
		ProcMask: core.ProcMaskEmpty,
		Flags: core.SpellFlagAPL,
		ClassSpellMask: SpellMaskAvengingWrath,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute * 3,
			},
		},
		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 8,
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !paladin.Forbearance.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
			paladin.Forbearance.Activate(sim)
		},
	})

	paladin.AvengingWrath = spell
	paladin.AvengingWrathAura = aura

	paladin.AddMajorCooldown(core.MajorCooldown{
		Spell:    spell,
		Priority: int32(core.CooldownTypeDPS),
		Type:     core.CooldownTypeDPS,
	})
}
