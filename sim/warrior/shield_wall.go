package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (war *Warrior) registerShieldWall() {
	actionID := core.ActionID{SpellID: 871}
	aura := war.RegisterAura(core.Aura{
		Label:    "Shield Wall",
		ActionID: actionID,
		Duration: time.Second * 10,
	}).AttachMultiplicativePseudoStatBuff(
		&war.PseudoStats.DamageTakenMultiplier, 0.25,
	)

	spell := war.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		ClassSpellMask: SpellMaskShieldWall,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
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
			return war.StanceMatches(DefensiveStance) && war.PseudoStats.CanBlock
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},
		RelatedSelfBuff: aura,
	})

	war.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}, func(sim *core.Simulation, slot proto.ItemSlot) {
		if !war.PseudoStats.CanBlock {
			aura.Deactivate(sim)
		}
	})

	war.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeSurvival,
		ShouldActivate: func(s *core.Simulation, c *core.Character) bool {
			return war.CurrentHealthPercent() < 0.4
		},
	})
}
