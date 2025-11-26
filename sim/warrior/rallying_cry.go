package warrior

import (
	"github.com/wowsims/tbc/sim/core"
)

func (war *Warrior) registerRallyingCry() {
	war.RallyingCryAuras = core.RallyingCryAuraArray(&war.Unit, war.Index)

	spell := war.RegisterSpell(core.SpellConfig{
		ActionID:       core.RallyingCryActionID,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskRallyingCry,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: core.RallyingCryCD,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return (war.LastStandAura != nil && !war.LastStandAura.IsActive()) || war.LastStandAura == nil
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			war.RallyingCryAuras.ActivateAllPlayers(sim)
		},
		RelatedAuraArrays: war.RallyingCryAuras.ToMap(),
	})

	war.AddMajorCooldown(core.MajorCooldown{
		Spell:    spell,
		Type:     core.CooldownTypeSurvival,
		Priority: core.CooldownPriorityLow,
		ShouldActivate: func(s *core.Simulation, c *core.Character) bool {
			return war.CurrentHealthPercent() < 0.6
		},
	})
}
