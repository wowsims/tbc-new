package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

type PetAbilityType int

const (
	Unknown PetAbilityType = iota

	Bite
	Claw
	Gore
	LightningBreath
	Screech
)

func (hp *HunterPet) NewPetAbility(abilityType PetAbilityType) *core.Spell {
	switch abilityType {

	case Bite:
		return hp.newBite()
	case Claw:
		return hp.newClaw()
	case Gore:
		return hp.newGore()
	case LightningBreath:
		return hp.newLightningBreath()
	case Screech:
		return hp.newScreech()

	case Unknown:
		return nil
	default:
		panic("Invalid pet ability type")
	}
}

func (hp *HunterPet) registerKillCommandSpell() {
	hp.KillCommand = hp.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 34027},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: HunterSpellKillCommandPet,
		MaxRange:       core.MaxMeleeRange,

		FocusCost: core.FocusCostOptions{
			Cost: 0,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
		},

		DamageMultiplier: hp.config.DamageMultiplier,
		CritMultiplier:   hp.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := hp.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) + 127
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})
}

func (hp *HunterPet) newBite() *core.Spell {
	return hp.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 27050},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics,
		MaxRange:    core.MaxMeleeRange,

		FocusCost: core.FocusCostOptions{
			Cost: 35,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    hp.NewTimer(),
				Duration: time.Second * 10,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier:         1,
		CritMultiplier:           hp.DefaultMeleeCritMultiplier(),
		ThreatMultiplier:         1,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hp.IsEnabled()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := hp.CalcAndRollDamageRange(sim, 108, 132)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})
}

func (hp *HunterPet) newClaw() *core.Spell {
	return hp.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 27049},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics,
		MaxRange:    core.MaxMeleeRange,

		FocusCost: core.FocusCostOptions{
			Cost: 25,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier:         1,
		CritMultiplier:           hp.DefaultMeleeCritMultiplier(),
		ThreatMultiplier:         1,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hp.IsEnabled()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := hp.CalcAndRollDamageRange(sim, 54, 76)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})
}

func (hp *HunterPet) newGore() *core.Spell {
	return hp.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 35298},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics,
		MaxRange:    core.MaxMeleeRange,

		FocusCost: core.FocusCostOptions{
			Cost: 25,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier:         1,
		CritMultiplier:           hp.DefaultMeleeCritMultiplier(),
		ThreatMultiplier:         1,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hp.IsEnabled()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := hp.CalcAndRollDamageRange(sim, 37, 61)
			if sim.Proc(0.5, "Gore") {
				baseDamage *= 2
			}
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})
}

func (hp *HunterPet) newLightningBreath() *core.Spell {
	return hp.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 25012},
		SpellSchool: core.SpellSchoolNature,
		ProcMask:    core.ProcMaskSpellDamage,
		MaxRange:    20,

		FocusCost: core.FocusCostOptions{
			Cost: 50,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier:         1,
		CritMultiplier:           hp.DefaultSpellCritMultiplier(),
		ThreatMultiplier:         1,
		BonusCoefficient:         0.05,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hp.IsEnabled()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := hp.CalcAndRollDamageRange(sim, 101, 116)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}

func (hp *HunterPet) newScreech() *core.Spell {
	auraArray := hp.NewEnemyAuraArray(core.ScreechAura)
	return hp.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 27051},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics,
		MaxRange:    core.MaxMeleeRange,

		FocusCost: core.FocusCostOptions{
			Cost: 20,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier:         1,
		CritMultiplier:           hp.DefaultMeleeCritMultiplier(),
		ThreatMultiplier:         1,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hp.IsEnabled()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := hp.CalcAndRollDamageRange(sim, 33, 61)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if result.Landed() {
				aura := auraArray.Get(target)
				aura.Activate(sim)
			}

			spell.DealDamage(sim, result)
		},

		RelatedAuraArrays: auraArray.ToMap(),
	})
}

func (hp *HunterPet) registerDash() {
	actionID := core.ActionID{SpellID: 23110}

	dashAura := hp.RegisterAura(core.Aura{
		Label:    "Dash",
		ActionID: actionID,
		Duration: time.Second * 15,
	})
	dashAura.NewActiveMovementSpeedEffect(0.8)

	hp.Dash = hp.RegisterSpell(core.SpellConfig{
		ActionID: actionID,

		FocusCost: core.FocusCostOptions{
			Cost: 20,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    hp.NewTimer(),
				Duration: time.Second * 30,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hp.IsEnabled()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			dashAura.Activate(sim)
		},
	})
}
