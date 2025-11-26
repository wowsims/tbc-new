package survival

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/hunter"
)

func (survival *SurvivalHunter) ApplyTalents() {
	survival.applyLNL()
	survival.ApplyMods()
	survival.Hunter.ApplyTalents()
}

func (survival *SurvivalHunter) ApplyMods() {
	survival.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		ClassMask:  hunter.HunterSpellSerpentSting,
		FloatValue: 0.5,
	})
}

// Todo: Should we support precasting freezing/ice trap?
func (survival *SurvivalHunter) applyLNL() {
	actionID := core.ActionID{SpellID: 56343}
	procChance := core.TernaryFloat64(survival.CouldHaveSetBonus(hunter.YaungolSlayersBattlegear, 4), 0.40, 0.20)
	has4pcT16 := survival.CouldHaveSetBonus(hunter.BattlegearOfTheUnblinkingVigil, 4)

	icd := core.Cooldown{
		Timer:    survival.NewTimer(),
		Duration: time.Second * 10,
	}

	lnlCostMod := survival.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		ClassMask:  hunter.HunterSpellExplosiveShot,
		FloatValue: -100,
	})

	lnlAura := core.BlockPrepull(survival.RegisterAura(core.Aura{
		Icd:       &icd,
		Label:     "Lock and Load Proc",
		ActionID:  actionID,
		Duration:  time.Second * 12,
		MaxStacks: 2,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			lnlCostMod.Activate()
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			lnlCostMod.Deactivate()
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell == survival.ExplosiveShot {
				survival.ExplosiveShot.CD.Reset()

				// T16 4pc: Explosive Shot casts have a 40% chance to not consume a charge of Lock and Load.
				if has4pcT16 && sim.Proc(0.4, "T16 4pc") {
					return
				}

				aura.RemoveStack(sim)
			}
		},
	}))

	survival.RegisterAura(core.Aura{
		Label:    "Lock and Load",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnPeriodicDamageDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.Matches(hunter.HunterSpellBlackArrow) {
				return
			}

			if !icd.IsReady(sim) {
				return
			}

			if sim.RandomFloat("Lock and Load") < procChance {
				icd.Use(sim)
				lnlAura.Activate(sim)
				lnlAura.SetStacks(sim, 2)
				if survival.ExplosiveShot != nil {
					survival.ExplosiveShot.CD.Reset()
				}
			}
		},
	})
}
