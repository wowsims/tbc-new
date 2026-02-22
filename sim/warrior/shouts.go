package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
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

		FlatThreatBonus: 68,

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			allyAuras.ActivateAllPlayers(sim)
		},

		RelatedAuraArrays: allyAuras.ToMap(),
	})
}

func (warrior *Warrior) registerShouts() {
	commandingPresenceMultiplier := 1.0 + 0.05*float64(warrior.Talents.CommandingPresence)

	warrior.registerDemoralizingShout()

	warrior.BattleShout = warrior.MakeShoutSpellHelper(
		core.ActionID{SpellID: 2048},
		SpellMaskBattleShout,
		warrior.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
			aura := core.BattleShoutAura(
				warrior.GetCharacter(),
				warrior.DefaultShout != proto.WarriorShout_WarriorShoutNone,
				warrior.Talents.BoomingVoice,
				commandingPresenceMultiplier,
				warrior.HasBsSolarianSapphire,
				warrior.HasBsT2,
			)
			aura.BuildPhase = core.Ternary(warrior.DefaultShout == proto.WarriorShout_WarriorShoutBattle, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone)
			return aura
		}),
	)

	warrior.CommandingShout = warrior.MakeShoutSpellHelper(
		core.ActionID{SpellID: 469},
		SpellMaskCommandingShout,
		warrior.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
			aura := core.CommandingShoutAura(
				warrior.GetCharacter(),
				warrior.DefaultShout != proto.WarriorShout_WarriorShoutNone,
				warrior.Talents.BoomingVoice,
				commandingPresenceMultiplier,
				warrior.CouldHaveSetBonus(ItemSetOnslaughtArmor, 2),
			)
			aura.BuildPhase = core.Ternary(warrior.DefaultShout == proto.WarriorShout_WarriorShoutCommanding, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone)
			return aura
		}),
	)
}
