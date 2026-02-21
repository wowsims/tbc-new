package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (shaman *Shaman) ApplyRestorationTalents() {
	shaman.applyNaturesGuidance()
	shaman.applyNaturesSwiftness()
	shaman.applyRestorativeTotems()
	shaman.applyTidalMastery()
	shaman.applyTotemicFocus()
}

func (shaman *Shaman) applyNaturesGuidance() {
	if shaman.Talents.NaturesGuidance == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusHit_Percent,
		FloatValue: 1 * float64(shaman.Talents.NaturesGuidance),
	})
}

func (shaman *Shaman) applyNaturesSwiftness() {
	if !shaman.Talents.NaturesSwiftness {
		return
	}
	nsAura := shaman.RegisterAura(core.Aura{
		ActionID: core.ActionID{SpellID: 16188},
		Label:    "Nature's Swiftness",
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !spell.Matches(SpellMaskChainLightning | SpellMaskLightningBolt) {
				return
			}
			aura.Deactivate(sim)
		},
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_CastTime_Pct,
		FloatValue: -100,
		ClassMask:  SpellMaskChainLightning | SpellMaskLightningBolt,
	})

	shaman.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 16188},
		SpellSchool: core.SpellSchoolPhysical,
		Flags:       core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * 180,
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			nsAura.Activate(sim)
		},
	})
}

func (shaman *Shaman) applyRestorativeTotems() {
	if shaman.Talents.RestorativeTotems == 0 {
		return
	}
	// TODO
}

func (shaman *Shaman) applyTidalMastery() {
	if shaman.Talents.TidalMastery == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_CritMultiplier_Flat,
		FloatValue: 0.01 * float64(shaman.Talents.TidalMastery),
		ClassMask:  SpellMaskChainLightning | SpellMaskLightningBolt | SpellMaskLightningShield,
	})
}

func (shaman *Shaman) applyTotemicFocus() {
	if shaman.Talents.TotemicFocus == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -0.05 * float64(shaman.Talents.TotemicFocus),
		ClassMask:  SpellMaskTotem,
	})
}
