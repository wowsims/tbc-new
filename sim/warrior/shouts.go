package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const ShoutExpirationThreshold = time.Second * 3

func (warrior *Warrior) MakeShoutSpellHelper(actionID core.ActionID, spellMask int64, allyAuras core.AuraArray) *core.Spell {
	duration := time.Minute * 1

	return warrior.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: spellMask,

		RageCost: core.RageCostOptions{
			Cost: 10,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    warrior.sharedShoutsCD,
				Duration: duration,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			allyAuras.ActivateAllPlayers(sim)
		},

		RelatedAuraArrays: allyAuras.ToMap(),
	})
}

func (warrior *Warrior) registerShouts() {
	commandingPresenceMultiplier := 1.0 + 0.05*float64(warrior.Talents.CommandingPresence)
	hasSolarianSapphire := warrior.HasItemEquipped(30446, core.TrinketSlots())

	warrior.registerDemoralizingShout()

	warrior.BattleShout = warrior.MakeShoutSpellHelper(
		core.ActionID{SpellID: 6673},
		SpellMaskBattleShout,
		warrior.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
			return core.BattleShoutAura(warrior.GetCharacter(), warrior.Talents.BoomingVoice, commandingPresenceMultiplier, hasSolarianSapphire)
		}),
	)

	warrior.CommandingShout = warrior.MakeShoutSpellHelper(
		core.ActionID{SpellID: 469},
		SpellMaskCommandingShout,
		warrior.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
			return core.CommandingShoutAura(warrior.GetCharacter(), warrior.Talents.BoomingVoice, commandingPresenceMultiplier, warrior.T6Tank2P != nil && warrior.T6Tank2P.IsActive())
		}),
	)
}
