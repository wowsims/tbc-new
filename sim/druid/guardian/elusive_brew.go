package guardian

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/druid"
)

func (bear *GuardianDruid) registerElusiveBrewSpell() {
	actionID := core.ActionID{SpellID: 126453}

	elusiveBrewAura := bear.RegisterAura(core.Aura{
		Label:    "Elusive Brew",
		ActionID: actionID,
		Duration: time.Second * 8,

		OnGain: func(aura *core.Aura, _ *core.Simulation) {
			aura.Unit.PseudoStats.BaseDodgeChance += 0.1
		},

		OnExpire: func(aura *core.Aura, _ *core.Simulation) {
			aura.Unit.PseudoStats.BaseDodgeChance -= 0.1
		},
	})

	elusiveBrewSpell := bear.RegisterSpell(druid.Any, core.SpellConfig{
		ActionID:        actionID,
		SpellSchool:     core.SpellSchoolPhysical,
		ProcMask:        core.ProcMaskEmpty,
		Flags:           core.SpellFlagAPL,
		RelatedSelfBuff: elusiveBrewAura,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    bear.NewTimer(),
				Duration: time.Minute,
			},

			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},
	})

	bear.AddMajorCooldown(core.MajorCooldown{
		Spell: elusiveBrewSpell.Spell,
		Type:  core.CooldownTypeSurvival,
	})
}
