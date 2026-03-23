package druid

import (
	"github.com/wowsims/tbc/sim/core"
)

// Returns the time to wait before the next action, or 0 if innervate is on CD
// or disabled.
func (druid *Druid) registerInnervateCD() {
	innervateTarget := druid.GetUnit(druid.SelfBuffs.InnervateTarget)
	if innervateTarget == nil {
		innervateTarget = &druid.Unit
	}
	innervateTargetChar := druid.Env.Raid.GetPlayerFromUnit(innervateTarget).GetCharacter()

	actionID := core.ActionID{SpellID: 29166, Tag: druid.Index}
	var innervateSpell *DruidSpell

	innervateCD := core.InnervateCD

	amount := 0.05
	if innervateTarget == &druid.Unit {
		amount = 0.2 + float64(druid.Talents.Dreamstate)*0.15
	}

	var innervateAura = core.InnervateAura(innervateTargetChar, amount, actionID.Tag)

	innervateSpell = druid.RegisterSpell(Humanoid|Moonkin|Tree, core.SpellConfig{
		ActionID: actionID,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: innervateCD,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			// If target already has another innervate, don't cast.
			return !innervateTarget.HasActiveAuraWithTag(core.InnervateAuraTag)
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			innervateAura.Activate(sim)
		},
	})

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: innervateSpell.Spell,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			// Require manual APL usage
			return false
		},
	})
}
