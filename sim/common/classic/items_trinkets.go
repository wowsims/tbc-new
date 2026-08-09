package tbc

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
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

}
