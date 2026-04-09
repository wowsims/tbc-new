package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (hunter *Hunter) registerRapidFireCD() {
	actionID := core.ActionID{SpellID: 3045}

	hunter.RapidFire = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolArcane,
		ClassSpellMask: HunterSpellRapidFire,

		ManaCost: core.ManaCostOptions{
			FlatCost: 100,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Minute * 5,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.GCD.IsReady(sim) && !hunter.RapidFire.RelatedSelfBuff.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},

		RelatedSelfBuff: hunter.RegisterAura(core.Aura{
			Label:    "Rapid Fire",
			ActionID: actionID,
			Duration: time.Second * 15,
		}).AttachMultiplyRangedHaste(1.4),
	})

	hunter.AddMajorCooldown(core.MajorCooldown{
		Spell: hunter.RapidFire,
		Type:  core.CooldownTypeDPS,

		ShouldActivate: func(s *core.Simulation, c *core.Character) bool {
			return !hunter.RapidFire.RelatedSelfBuff.IsActive()
		},
	})
}
