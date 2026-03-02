package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	// Hand Of Justice
	core.NewItemEffect(11815, func(agent core.Agent) {
		character := agent.GetCharacter()
		var handOfJusticeSpell *core.Spell

		extraAttackDPM := func() *core.DynamicProcManager {
			return character.NewFixedProcChanceManager(
				0.013333,
				character.GetProcMaskForTypes(proto.WeaponType_WeaponTypeSword),
			)
		}

		dpm := extraAttackDPM()

		procTrigger := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Hand Of Justice",
			DPM:                dpm,
			ICD:                time.Second * 2,
			TriggerImmediately: true,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				character.AutoAttacks.MaybeReplaceMHSwing(sim, handOfJusticeSpell).Cast(sim, result.Target)
			},
		})

		procTrigger.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
			config := *character.AutoAttacks.MHConfig()
			config.ActionID = config.ActionID.WithTag(11815)
			config.Flags |= core.SpellFlagPassiveSpell
			handOfJusticeSpell = character.GetOrRegisterSpell(config)
		})

		character.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
			dpm = extraAttackDPM()
		})

		character.ItemSwap.RegisterProc(11815, procTrigger)
	})

	// Badge of the Swarmguard
	core.NewItemEffect(21670, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 30

		arpAura := core.MakeStackingAura(
			character,
			core.StackingStatAura{
				Aura: core.Aura{
					Label:     "Insight of the Qiraji",
					ActionID:  core.ActionID{SpellID: 26481},
					Duration:  duration,
					MaxStacks: 6,
				},
				BonusPerStack: stats.Stats{stats.ArmorPenetration: 200},
			},
		)

		procAura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Badge of the Swarmguard",
			MetricsActionID:    core.ActionID{SpellID: 26480},
			DPM:                character.NewLegacyPPMManager(10, core.ProcMaskMeleeOrRanged),
			Duration:           duration,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			RequireDamageDealt: true,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if arpAura.IsActive() && arpAura.GetStacks() == arpAura.MaxStacks {
					return
				}
				arpAura.Activate(sim)
				arpAura.AddStack(sim)
			},
		})

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID: core.ActionID{ItemID: 21670},
			Flags:    core.SpellFlagNoOnCastComplete,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 3,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				procAura.Activate(sim)
			},

			RelatedSelfBuff: arpAura.Aura,
		})

		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(21670)
		character.AddStatProcBuff(26481, arpAura, false, eligibleSlots)
		character.AddMajorCooldown(core.MajorCooldown{
			Spell:    spell,
			Type:     core.CooldownTypeDPS,
			BuffAura: arpAura,
		})
	})
}
