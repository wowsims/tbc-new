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
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ClassMask:  SpellMaskHolyLight | SpellMaskFlashOfLight | SpellMaskHolyShock,
		FloatValue: 100,
	}).AttachProcTrigger(core.ProcTrigger{
		Name:               "Divine Favor - Consume",
		Callback:           core.CallbackOnCastComplete,
		ClassSpellMask:     SpellMaskHolyLight | SpellMaskFlashOfLight | SpellMaskHolyShock,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			paladin.DivineFavorAura.Deactivate(sim)
		},
	})

	divineFavor := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskDivineFavor,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

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
