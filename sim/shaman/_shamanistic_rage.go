package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (shaman *Shaman) registerShamanisticRageSpell() {

	actionID := core.ActionID{SpellID: 30823}
	srAura := shaman.RegisterAura(core.Aura{
		Label:    "Shamanistic Rage",
		ActionID: actionID,
		Duration: time.Second * 15,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -2,
	})

	spell := shaman.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		ClassSpellMask: SpellMaskShamanisticRage,
		Flags:          core.SpellFlagReadinessTrinket,
		Cast: core.CastConfig{
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Minute * 1,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			srAura.Activate(sim)
		},
		RelatedSelfBuff: srAura,
	})

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(s *core.Simulation, c *core.Character) bool {
			return shaman.CurrentManaPercent() < 0.05
		},
	})
}
