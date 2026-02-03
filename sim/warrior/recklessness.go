package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (war *Warrior) registerRecklessness() {
	actionID := core.ActionID{SpellID: 1719}

	aura := war.RegisterAura(core.Aura{
		Label:    "Recklessness",
		ActionID: actionID,
		Duration: time.Second * 15,
	}).AttachSpellMod(core.SpellModConfig{
		ProcMask:   core.ProcMaskMeleeSpecial,
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 100,
	}).AttachMultiplicativePseudoStatBuff(
		&war.PseudoStats.DamageTakenMultiplier, 1.2,
	)

	spell := war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskRecklessness,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Minute * 30,
			},
			SharedCD: core.Cooldown{
				Timer:    war.sharedMCD,
				Duration: time.Minute * 30,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.StanceMatches(BerserkerStance)
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},

		RelatedSelfBuff: aura,
	})

	war.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}
