package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (mage *Mage) registerArmorSpells() {

	mageArmorEffectCategory := "MageArmors"

	moltenArmor := mage.RegisterAura(core.Aura{
		Label:      "Molten Armor",
		ActionID:   core.ActionID{SpellID: 30482},
		Duration:   time.Minute * 30,
		BuildPhase: core.Ternary(mage.Options.DefaultMageArmor == proto.MageArmor_MageArmorMoltenArmor, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone),
	}).AttachStatBuff(stats.SpellCritPercent, 3)

	moltenArmor.NewExclusiveEffect(mageArmorEffectCategory, true, core.ExclusiveEffect{})

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 30482},
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

	mageArmorActionId := core.ActionID{SpellID: 6117}
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
		BuildPhase: core.Ternary(mage.Options.DefaultMageArmor == proto.MageArmor_MageArmorMageArmor, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone),
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
	frostArmor := mage.RegisterAura(core.Aura{
		ActionID:   core.ActionID{SpellID: 7302},
		Label:      "Frost Armor",
		Duration:   time.Minute * 30,
		BuildPhase: core.Ternary(mage.Options.DefaultMageArmor == proto.MageArmor_MageArmorFrostArmor, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone),
	})

	frostArmor.NewExclusiveEffect(mageArmorEffectCategory, true, core.ExclusiveEffect{})

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 7302},
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
			return !frostArmor.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			frostArmor.Activate(sim)
		},
	})

	switch mage.Options.DefaultMageArmor {
	case proto.MageArmor_MageArmorFrostArmor:
		core.MakePermanent(frostArmor)
	case proto.MageArmor_MageArmorMageArmor:
		core.MakePermanent(mageArmor)
	case proto.MageArmor_MageArmorMoltenArmor:
		core.MakePermanent(moltenArmor)
	}
}
