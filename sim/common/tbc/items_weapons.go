package tbc

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func init() {
	// Blinkstrike
	core.NewItemEffect(31332, func(agent core.Agent) {
		character := agent.GetCharacter()
		var blinkStrikeSpell *core.Spell

		extraAttackDPM := func() *core.DynamicProcManager {
			return character.NewStaticLegacyPPMManager(
				1,
				character.GetProcMaskForTypes(proto.WeaponType_WeaponTypeSword),
			)
		}

		dpm := extraAttackDPM()

		procTrigger := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Blinkstrike",
			DPM:                dpm,
			TriggerImmediately: true,
			Outcome:            core.OutcomeLanded,
			Callback:           core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				character.AutoAttacks.MaybeReplaceMHSwing(sim, blinkStrikeSpell).Cast(sim, result.Target)
			},
		})

		procTrigger.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
			config := *character.AutoAttacks.MHConfig()
			config.ActionID = config.ActionID.WithTag(31332)
			config.Flags |= core.SpellFlagPassiveSpell
			blinkStrikeSpell = character.GetOrRegisterSpell(config)
		})

		character.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
			dpm = extraAttackDPM()
		})

		character.ItemSwap.RegisterProc(31332, procTrigger)
	})
}
