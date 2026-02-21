package tbc

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {

	// Crusader
	// EffectID: 1900, Proc SpellID: 20007
	// PPM: 1, ICD: 0
	// Permanently enchant a melee weapon so that often when attacking in melee
	// it heals for 75 to 125 and increases Strength by 100 for 15 sec.
	// Has a reduced effect for players above level 60.
	core.NewEnchantEffect(1900, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 15
		actionID := core.ActionID{SpellID: 20007}
		healthMetrics := character.NewHealthMetrics(actionID)

		createCrusaderAuras := func(tag int32) *core.StatBuffAura {
			labelSuffix := core.Ternary(tag == 1, "(MH)", "(OH)")
			slot := core.Ternary(tag == 1, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand)
			aura := character.NewTemporaryStatsAura(
				fmt.Sprintf("Holy Strength %s", labelSuffix),
				actionID.WithTag(tag),
				stats.Stats{stats.Strength: 60},
				duration,
			)
			character.AddStatProcBuff(1900, aura, true, []proto.ItemSlot{slot})
			character.ItemSwap.RegisterWeaponEnchantBuff(aura.Aura, 1900)
			return aura
		}

		mhAura := createCrusaderAuras(1)
		ohAura := createCrusaderAuras(2)

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Enchant Weapon - Crusader",
			ActionID: actionID,
			DPM:      character.NewDynamicLegacyProcForEnchant(1900, 1.0, 0),
			Outcome:  core.OutcomeLanded,
			Callback: core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				core.Ternary(spell.IsOH(), ohAura, mhAura).Activate(sim)
				character.GainHealth(sim, sim.Roll(45, 75), healthMetrics)
			},
		})
	})
}
