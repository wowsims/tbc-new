package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (shaman *Shaman) ApplyElementalTalents() {
	shaman.applyCallOfFlame()
	shaman.applyCallOfThunder()
	shaman.applyConcussion()
	shaman.applyConvection()
	shaman.applyElementalDevastation()
	shaman.applyElementalFocus()
	shaman.applyElementalFury()
	shaman.applyElementalMastery()
	shaman.applyElementalPrecision()
	shaman.applyImprovedFireTotems()
	shaman.applyLightningMastery()
	shaman.applyLightningOverload()
	shaman.applyReverberation()
	shaman.applyTotemOfWrath()
	shaman.applyUnrelentingStorm()
}

func (shaman *Shaman) applyCallOfFlame() {
	if shaman.Talents.CallOfFlame == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.05 * float64(shaman.Talents.CallOfFlame),
		ClassMask:  SpellMaskFireTotem,
	})
}
func (shaman *Shaman) applyCallOfThunder() {
	if shaman.Talents.CallOfThunder == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 1 * float64(shaman.Talents.CallOfThunder),
		ClassMask:  SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload,
	})
}
func (shaman *Shaman) applyConcussion() {
	if shaman.Talents.Concussion == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.01 * float64(shaman.Talents.Concussion),
		ClassMask:  SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload | SpellMaskShock,
	})
}
func (shaman *Shaman) applyConvection() {
	if shaman.Talents.Convection == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -0.02 * float64(shaman.Talents.Convection),
		ClassMask:  SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload | SpellMaskShock,
	})
}
func (shaman *Shaman) applyElementalDevastation() {
	if shaman.Talents.ElementalDevastation == 0 {
		return
	}
	critBuffAura := shaman.RegisterAura(core.Aura{
		Label:    "Elemental Devastation",
		ActionID: core.ActionID{SpellID: 29178},
		Duration: time.Second * 10,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 3 * float64(shaman.Talents.ElementalDevastation),
		ProcMask:   core.ProcMaskMelee,
	})
	shaman.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Elemental Devastation Trigger",
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskSpellDamage,
		Outcome:            core.OutcomeCrit,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			critBuffAura.Activate(sim)
		},
	})

}
func (shaman *Shaman) applyElementalFocus() {
	if !shaman.Talents.ElementalFocus {
		return
	}
	var triggeringSpell *core.Spell
	var triggerTime time.Duration

	canConsumeSpells := SpellMaskLightningBolt | SpellMaskChainLightning | (SpellMaskShock & ^SpellMaskFlameShockDot)

	maxStacks := int32(2)

	clearcastingAura := shaman.RegisterAura(core.Aura{
		Label:     "Clearcasting",
		ActionID:  core.ActionID{SpellID: 16246},
		Duration:  time.Second * 15,
		MaxStacks: maxStacks,
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !spell.Matches(canConsumeSpells) {
				return
			}
			if spell == triggeringSpell && sim.CurrentTime == triggerTime {
				return
			}
			aura.RemoveStack(sim)
		},
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		ClassMask:  canConsumeSpells,
		FloatValue: -0.4,
	})

	shaman.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Elemental Focus",
		Callback:           core.CallbackOnSpellHitDealt,
		Outcome:            core.OutcomeCrit,
		TriggerImmediately: true,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.SpellSchool.Matches(core.SpellSchoolElemental) {
				return
			}
			triggeringSpell = spell
			triggerTime = sim.CurrentTime
			clearcastingAura.Activate(sim)
			clearcastingAura.SetStacks(sim, maxStacks)
		},
	})
}
func (shaman *Shaman) applyElementalFury() {
	if !shaman.Talents.ElementalFury {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_CritMultiplier_Flat,
		FloatValue: 1.0,
		ClassMask:  SpellMaskFireTotem | SpellMaskFire | SpellMaskNature | SpellMaskFrost,
	})
}
func (shaman *Shaman) applyElementalMastery() {
	if !shaman.Talents.ElementalMastery {
		return
	}

	emAura := shaman.RegisterAura(core.Aura{
		ActionID: core.ActionID{SpellID: 16166},
		Label:    "Elemental Mastery",
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !spell.Matches(SpellMaskFire | SpellMaskFrost | SpellMaskNature) {
				return
			}
			aura.Deactivate(sim)
		},
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 100,
		ClassMask:  SpellMaskFire | SpellMaskFrost | SpellMaskNature,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -2,
		ClassMask:  SpellMaskFire | SpellMaskFrost | SpellMaskNature,
	})

	shaman.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 16166},
		SpellSchool: core.SpellSchoolNature,
		Flags:       core.SpellFlagAPL | core.SpellFlagNoOnCastComplete | SpellFlagInstant,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * 180,
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			emAura.Activate(sim)
		},
	})
}
func (shaman *Shaman) applyElementalPrecision() {
	if shaman.Talents.ElementalPrecision == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusHit_Percent,
		FloatValue: 2 * float64(shaman.Talents.ElementalPrecision),
		School:     core.SpellSchoolElemental,
	})
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_ThreatMultiplier_Pct,
		FloatValue: [4]float64{0, -0.04, -0.07, -0.1}[shaman.Talents.ElementalPrecision],
	})
}
func (shaman *Shaman) applyImprovedFireTotems() {
	if shaman.Talents.ImprovedFireTotems == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_ThreatMultiplier_Pct,
		FloatValue: -0.25 * float64(shaman.Talents.ImprovedFireTotems),
		ClassMask:  SpellMaskMagmaTotem,
	})
	// Reduction to Fire Nova Activation Delay in fire_totems.go
}
func (shaman *Shaman) applyLightningMastery() {
	if shaman.Talents.LightningMastery == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: -time.Duration(100*shaman.Talents.LightningMastery) * time.Millisecond,
		ClassMask: SpellMaskLightningBolt | SpellMaskChainLightning,
	})
}
func (shaman *Shaman) applyLightningOverload() {
	if shaman.Talents.LightningOverload == 0 {
		return
	}
	// In shaman.go -> GetOverloadChance()
}
func (shaman *Shaman) applyReverberation() {
	if shaman.Talents.Reverberation == 0 {
		return
	}
	shaman.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_Cooldown_Flat,
		TimeValue: time.Duration(-200*shaman.Talents.Reverberation) * time.Millisecond,
		ClassMask: SpellMaskShock,
	})
}
func (shaman *Shaman) applyTotemOfWrath() {
	if !shaman.Talents.TotemOfWrath {
		return
	}
	// TODO
}
func (shaman *Shaman) applyUnrelentingStorm() {
	if shaman.Talents.UnrelentingStorm == 0 {
		return
	}
	shaman.AddStatDependency(stats.Intellect, stats.MP5, 0.02*float64(shaman.Talents.UnrelentingStorm))
}
