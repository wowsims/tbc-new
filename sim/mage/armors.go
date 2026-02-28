package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (mage *Mage) registerArmorSpells() {

	mageArmorEffectCategory := "MageArmors"

	moltenArmorActionId := core.ActionID{SpellID: 30482}
	moltenArmor := mage.RegisterAura(core.Aura{
		Label:    "Molten Armor",
		ActionID: moltenArmorActionId,
		Duration: time.Minute * 30,
	}).AttachStatBuff(stats.SpellCritPercent, 3)

	moltenArmor.NewExclusiveEffect(mageArmorEffectCategory, true, core.ExclusiveEffect{})

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       moltenArmorActionId,
		SpellSchool:    core.SpellSchoolFire,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: MageSpellMoltenArmor,
		ManaCost: core.ManaCostOptions{
			FlatCost: 630,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !moltenArmor.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			moltenArmor.Activate(sim)
		},
	})

	mageArmorActionId := core.ActionID{SpellID: 27125}
	mageArmor := mage.RegisterAura(core.Aura{
		ActionID: mageArmorActionId,
		Label:    "Mage Armor",
		Duration: time.Minute * 30,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			mage.PseudoStats.SpiritRegenRateCombat += .3
			mage.UpdateManaRegenRates()
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			mage.PseudoStats.SpiritRegenRateCombat -= .3
			mage.UpdateManaRegenRates()
		},
	})

	mageArmor.NewExclusiveEffect(mageArmorEffectCategory, true, core.ExclusiveEffect{})

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       mageArmorActionId,
		SpellSchool:    core.SpellSchoolArcane,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: MageSpellMageArmor,
		ManaCost: core.ManaCostOptions{
			FlatCost: 575,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !mageArmor.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			mageArmor.Activate(sim)
		},
	})

	//Frost armor/IceArmor gives no benefit to dps, merely armor and slow on hit
	iceArmorActionId := core.ActionID{SpellID: 27124}
	iceArmor := mage.RegisterAura(core.Aura{
		ActionID: iceArmorActionId,
		Label:    "Frost Armor",
		Duration: time.Minute * 30,
	})

	iceArmor.NewExclusiveEffect(mageArmorEffectCategory, true, core.ExclusiveEffect{})

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       iceArmorActionId,
		SpellSchool:    core.SpellSchoolFrost,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: MageSpellFrostArmor,
		ManaCost: core.ManaCostOptions{
			FlatCost: 630,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !iceArmor.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			iceArmor.Activate(sim)
		},
	})

	switch mage.Options.DefaultMageArmor {
	case proto.MageArmor_MageArmorFrostArmor:
		core.MakePermanent(iceArmor)
	case proto.MageArmor_MageArmorMageArmor:
		core.MakePermanent(mageArmor)
	case proto.MageArmor_MageArmorMoltenArmor:
		core.MakePermanent(moltenArmor)
	}
}
