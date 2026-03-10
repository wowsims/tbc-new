package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

const ShoutExpirationThreshold = time.Second * 3

type ShoutHelperConfig struct {
	ActionID           core.ActionID
	SpellMask          int64
	ThreatBonus        float64
	AllyAuras          core.AuraArray
	ExtraCastCondition core.CanCastCondition
}

func (warrior *Warrior) MakeShoutSpellHelper(config ShoutHelperConfig) *core.Spell {
	duration := time.Minute * 1

	return warrior.RegisterSpell(core.SpellConfig{
		ActionID:       config.ActionID,
		ClassSpellMask: config.SpellMask,
		SpellSchool:    core.SpellSchoolPhysical,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ProcMask:       core.ProcMaskEmpty,

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

		ThreatMultiplier: 1,
		FlatThreatBonus:  config.ThreatBonus,

		ExtraCastCondition: config.ExtraCastCondition,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Assuming full party, thus multiplying by 5
			spell.FlatThreatBonus = core.TernaryFloat64(sim.CurrentTime > 0, config.ThreatBonus*5/float64(sim.Environment.ActiveTargetCount()), 0)
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
			config.AllyAuras.ActivateAllPlayers(sim)
		},

		RelatedAuraArrays: config.AllyAuras.ToMap(),
	})
}

func (warrior *Warrior) registerShouts() {
	commandingPresenceMultiplier := 1.0 + 0.05*float64(warrior.Talents.CommandingPresence)

	warrior.registerDemoralizingShout()

	battleShoutAuras := warrior.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
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
	})

	warrior.BattleShout = warrior.MakeShoutSpellHelper(ShoutHelperConfig{
		ActionID:    core.ActionID{SpellID: 2048},
		SpellMask:   SpellMaskBattleShout,
		ThreatBonus: 69,
		ExtraCastCondition: func(sim *core.Simulation, _ *core.Unit) bool {
			aura := battleShoutAuras.Get(&warrior.Unit)
			return !aura.IsActive() || aura.ExclusiveEffects[0].Priority <= core.GetBattleShoutValue(warrior.Talents.BoomingVoice, commandingPresenceMultiplier, warrior.HasBsSolarianSapphire, warrior.HasBsT2, sim.CurrentTime < 0)
		},
		AllyAuras: battleShoutAuras,
	})

	hasT6Tank2P := warrior.CouldHaveSetBonus(ItemSetOnslaughtArmor, 2)
	commandingShoutAuras := warrior.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
		aura := core.CommandingShoutAura(
			warrior.GetCharacter(),
			warrior.DefaultShout != proto.WarriorShout_WarriorShoutNone,
			warrior.Talents.BoomingVoice,
			commandingPresenceMultiplier,
			hasT6Tank2P,
		)
		aura.BuildPhase = core.Ternary(warrior.DefaultShout == proto.WarriorShout_WarriorShoutCommanding, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone)
		return aura
	})

	warrior.CommandingShout = warrior.MakeShoutSpellHelper(ShoutHelperConfig{
		ActionID:    core.ActionID{SpellID: 469},
		SpellMask:   SpellMaskCommandingShout,
		ThreatBonus: 68,
		ExtraCastCondition: func(sim *core.Simulation, _ *core.Unit) bool {
			aura := commandingShoutAuras.Get(&warrior.Unit)
			return !aura.IsActive() || aura.ExclusiveEffects[0].Priority <= core.GetCommandingShoutValue(warrior.Talents.BoomingVoice, commandingPresenceMultiplier, hasT6Tank2P, sim.CurrentTime < 0)
		},
		AllyAuras: commandingShoutAuras,
	})
}
