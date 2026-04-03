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

	icd := core.Cooldown{
		Timer:    druid.NewTimer(),
		Duration: time.Second * 10,
	}

	const ppm = 2.0

	// For feral druids in cat form, white auto attacks use the cat paw speed (1.0s),
	// but yellow special attacks (Shred, Mangle, etc.) use the actual equipped weapon
	// swing speed
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

	var affectedSpells []*DruidSpell

	druid.ClearcastingAura = druid.RegisterAura(core.Aura{
		Label:    "Clearcasting",
		ActionID: core.ActionID{SpellID: 16870},
		Duration: time.Second * 15,

		OnInit: func(_ *core.Aura, _ *core.Simulation) {
			affectedSpells = core.FilterSlice([]*DruidSpell{
				druid.MangleCat,
				druid.Rake,
				druid.Rip,
				druid.FerociousBite,
				druid.Shred,
			}, func(spell *DruidSpell) bool { return spell != nil })
		},

		OnGain: func(_ *core.Aura, _ *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.Cost.PercentModifier *= -1
			}
		},

		OnExpire: func(_ *core.Aura, _ *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.Cost.PercentModifier /= -1
			}
		},

		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			for _, as := range affectedSpells {
				if as.IsEqual(spell) {
					aura.Deactivate(sim)
					break
				}
			}
		},
	})

	druid.RegisterAura(core.Aura{
		Label:    "Omen of Clarity",
		ActionID: core.ActionID{SpellID: 16864},
		Duration: core.NeverExpires,

		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
			updateSpecialProcChance()
		},

		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMelee) || !icd.IsReady(sim) {
				return
			}

			// Yellow specials use the equipped weapon swing speed (not cat paw speed).
			// White auto attacks use the cat paw swing speed.
			var procChance float64
			if spell.ProcMask.Matches(core.ProcMaskMeleeMHAuto) {
				procChance = autoProcChance
			} else {
				procChance = specialProcChance
			}

			if sim.RandomFloat("Omen of Clarity") < procChance {
				icd.Use(sim)
				druid.ClearcastingAura.Activate(sim)
			}
		},
	})

	// Re-compute the special proc chance whenever the equipped weapon changes.
	druid.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(_ *core.Simulation, _ proto.ItemSlot) {
		updateSpecialProcChance()
	})
}
