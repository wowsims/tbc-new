package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func (druid *Druid) applyOmenOfClarity() {
	if !druid.Talents.OmenOfClarity {
		return
	}

	const ppm = 2.0

	const clearcastingSpells = DruidSpellMangleCat | DruidSpellRake | DruidSpellRip |
		DruidSpellFerociousBite | DruidSpellShred |
		DruidSpellMangleBear | DruidSpellMaul | DruidSpellLacerate |
		DruidSpellSwipe | DruidSpellDemoralizingRoar

	// White auto-attacks use the form's paw weapon speed (1.0s cat, 2.5s bear).
	// Yellow specials use the equipped weapon's true swing speed. Cache the
	// special proc chance and refresh on item swaps.
	autoProcChance := ppm * druid.AutoAttacks.MH().SwingSpeed / 60.0
	specialProcChance := autoProcChance

	updateSpecialProcChance := func() {
		if weapon := druid.GetMHWeapon(); weapon != nil {
			specialProcChance = ppm * weapon.SwingSpeed / 60.0
		} else {
			specialProcChance = autoProcChance
		}
	}
	updateSpecialProcChance()

	druid.ClearcastingAura = druid.RegisterAura(core.Aura{
		Label:    "Clearcasting",
		ActionID: core.ActionID{SpellID: 16870},
		Duration: time.Second * 15,

		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Matches(clearcastingSpells) {
				aura.Deactivate(sim)
			}
		},
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		ClassMask:  clearcastingSpells,
		FloatValue: -2,
	})

	druid.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Omen of Clarity",
		ActionID: core.ActionID{SpellID: 16864},
		Callback: core.CallbackOnSpellHitDealt,
		ProcMask: core.ProcMaskMelee,
		ICD:      time.Second * 10,
		ExtraCondition: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			// Yellow specials use the equipped weapon swing speed.
			// White auto attacks use the paw swing speed.
			var procChance float64
			if spell.ProcMask.Matches(core.ProcMaskMeleeMHAuto) {
				procChance = autoProcChance
			} else {
				procChance = specialProcChance
			}
			return sim.RandomFloat("Omen of Clarity") < procChance
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			druid.ClearcastingAura.Activate(sim)
		},
	})

	// Re-compute the special proc chance whenever the equipped weapon changes.
	druid.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(_ *core.Simulation, _ proto.ItemSlot) {
		updateSpecialProcChance()
	})
}
