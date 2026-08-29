package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

var pvpTrinketIDs = []int32{
	28234, 28235, 28236, 28237, 28238, // Medallion of the Alliance
	28239, 28240, 28241, 28242, 28243, // Medallion of the Horde
	30343, 30344, 30345, 30346, // Medallion of the Horde
	30348, 30349, 30350, 30351, // Medallion of the Alliance
	37864, // Medallion of the Alliance
	37865, // Medallion of the Horde
}

func init() {
	core.AddEffectsToTest = false
	for _, itemID := range pvpTrinketIDs {
		core.NewItemEffect(itemID, makePvPTrinketEffect(itemID))
	}
	core.AddEffectsToTest = true
}

func makePvPTrinketEffect(itemID int32) core.ApplyEffect {
	return func(agent core.Agent) {
		character := agent.GetCharacter()

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: itemID},
			SpellSchool: core.SpellSchoolPhysical,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagCastWhileIncapacitated,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				character.BreakFear(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeSurvival,
			ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
				return character.PseudoStats.Incapacitated
			},
		})
	}
}
