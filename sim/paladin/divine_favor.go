package paladin

// Divine Favor
// https://www.wowhead.com/tbc/spell=20216
//
// When activated, gives your next Flash of Light, Holy Light, or Holy Shock
// spell a 100% critical strike chance.

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (paladin *Paladin) registerDivineFavor() {
	actionID := core.ActionID{SpellID: 20216}

	paladin.DivineFavorAura = paladin.RegisterAura(core.Aura{
		Label:    "Divine Favor" + paladin.Label,
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range paladin.HolyLights {
				spell.BonusCritPercent += 100
			}
			for _, spell := range paladin.FlashOfLights {
				spell.BonusCritPercent += 100
			}
			for _, spell := range paladin.HolyShocks {
				spell.BonusCritPercent += 100
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range paladin.HolyLights {
				spell.BonusCritPercent -= 100
			}
			for _, spell := range paladin.FlashOfLights {
				spell.BonusCritPercent -= 100
			}
			for _, spell := range paladin.HolyShocks {
				spell.BonusCritPercent -= 100
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// Consume after Flash of Light, Holy Light, or Holy Shock
			if spell.Matches(SpellMaskFlashOfLight | SpellMaskHolyLight | SpellMaskHolyShock) {
				aura.Deactivate(sim)
			}
		},
	})

	divineFavor := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskDivineFavor,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 3,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: 2 * time.Minute,
			},
		},

		CritMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},

		RelatedSelfBuff: paladin.DivineFavorAura,
	})

	paladin.AddMajorCooldown(core.MajorCooldown{
		Spell: divineFavor,
		Type:  core.CooldownTypeDPS,
	})
}
