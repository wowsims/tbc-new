package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const ShoutExpirationThreshold = time.Second * 3

func (warrior *Warrior) MakeShoutSpellHelper(actionID core.ActionID, spellMask int64, allyAuras core.AuraArray) *core.Spell {
	shoutMetrics := warrior.NewRageMetrics(actionID)
	rageGen := 20.0
	duration := time.Minute * 1

	return warrior.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: spellMask,

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
			warrior.AddRage(sim, rageGen, shoutMetrics)
			allyAuras.ActivateAllPlayers(sim)
		},

		RelatedAuraArrays: allyAuras.ToMap(),
	})
}

func (warrior *Warrior) registerShouts() {
	commandingPresenceMultiplier := 1.0 + 0.05*float64(warrior.Talents.CommandingPresence)
	hasSolarianSapphire := warrior.HasItemEquipped(30446, core.TrinketSlots())

	warrior.BattleShout = warrior.MakeShoutSpellHelper(core.ActionID{SpellID: 6673}, SpellMaskBattleShout, warrior.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
		return core.BattleShoutAura(warrior.GetCharacter(), commandingPresenceMultiplier, hasSolarianSapphire)
	}))

	warrior.CommandingShout = warrior.MakeShoutSpellHelper(core.ActionID{SpellID: 469}, SpellMaskCommandingShout, warrior.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
		return core.CommandingShoutAura(warrior.GetCharacter(), commandingPresenceMultiplier, warrior.T6Tank2P != nil && warrior.T6Tank2P.IsActive())
	}))
}
