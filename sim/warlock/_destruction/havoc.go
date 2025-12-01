package destruction

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/warlock"
)

func (destruction *DestructionWarlock) spellMatches(aura *core.Aura, sim *core.Simulation, spell *core.Spell, target *core.Unit) {
	if !destruction.HavocAuras.Get(target).IsActive() { //If the target of the calling spell does NOT have the HavocDebuff
		//How many stacks are meant to be removed
		var stacks int32
		if spell.Matches(warlock.WarlockSpellFelFlame | warlock.WarlockSpellImmolate | warlock.WarlockSpellIncinerate |
			warlock.WarlockSpellShadowBurn | warlock.WarlockSpellConflagrate) {
			stacks = 1
		} else if spell.Matches(warlock.WarlockSpellChaosBolt) {
			stacks = 3
		} else {
			return
		}

		for _, havocAuras := range destruction.HavocAuras.ToMap() {
			for _, havocAura := range havocAuras {
				if havocAura != nil {
					if havocAura.IsActive() {
						aura.RemoveStacks(sim, stacks)
						//AddHavocFlag
						spell.Flags |= SpellFlagDestructionHavoc
						spell.Proc(sim, havocAura.Unit)
						//RemoveHavocFlag
						spell.Flags &^= SpellFlagDestructionHavoc
					}
				}
			}
		}

	}
}

func (destruction *DestructionWarlock) registerHavoc() {
	havocDebuffAura := core.Aura{
		Label:    "Havoc",
		ActionID: core.ActionID{SpellID: 80240},
		Duration: time.Second * 15,
	}

	destruction.HavocAuras = destruction.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(havocDebuffAura)
	})

	var havocCharges int32 = 3
	var cooldown = 25

	actionID := core.ActionID{SpellID: 80240}
	destruction.HavocChargesAura = destruction.RegisterAura(core.Aura{
		Label:     "Havoc Charges Aura",
		ActionID:  actionID,
		Duration:  time.Second * 15,
		MaxStacks: havocCharges,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			destruction.HavocChargesAura.AddStacks(sim, havocCharges)
		},

		OnApplyEffects: func(aura *core.Aura, sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			destruction.spellMatches(aura, sim, spell, target)

			if aura.GetStacks() == 0 {
				aura.Deactivate(sim)
				destruction.HavocAuras.DeactivateAll(sim)
			}
		},

		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Deactivate(sim)
		},
	})

	destruction.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: warlock.WarlockSpellHavoc,

		ManaCost: core.ManaCostOptions{BaseCostPercent: 4},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCDMin: time.Millisecond * 500,
				GCD:    core.GCDMin,
			},
			CD: core.Cooldown{
				Timer:    destruction.NewTimer(),
				Duration: time.Duration(cooldown) * time.Second,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			destruction.HavocChargesAura.Activate(sim)
			destruction.HavocAuras.Get(target).Activate(sim)
		},
	})
}
