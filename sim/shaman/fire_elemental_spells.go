package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (fireElemental *FireElemental) registerFireBlast() {
	fireElemental.FireBlast = fireElemental.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 13339},
		SpellSchool: core.SpellSchoolFire,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShamanSpell,

		ManaCost: core.ManaCostOptions{
			FlatCost: 120,
		},

		DamageMultiplier: 1,
		CritMultiplier:   fireElemental.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			scalingCoeff := 46.977695 / 11.334235
			baseDamage := fireElemental.CalcAndRollDamageRange(sim, 110*scalingCoeff, 130*scalingCoeff)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}

func (fireElemental *FireElemental) registerFireNova() {
	fireElemental.FireNova = fireElemental.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 12470},
		SpellSchool: core.SpellSchoolFire,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShamanSpell,

		ManaCost: core.ManaCostOptions{
			FlatCost: 95,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: time.Second * 2,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		CritMultiplier:   fireElemental.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 0.332, // https://discord.com/channels/260297137554849794/699626629152112730/904088040992026655

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			scalingCoeff := 46.977695 / 11.334235
			baseDamage := fireElemental.CalcAndRollDamageRange(sim, 148*scalingCoeff, 170*scalingCoeff)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	})
}

func (fireElemental *FireElemental) registerFireShield() {
	actionID := core.ActionID{SpellID: 13376}

	fireElemental.FireShield = fireElemental.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolFire,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShamanSpell,

		ManaCost: core.ManaCostOptions{
			FlatCost: 95,
		},

		DamageMultiplier: 1,
		CritMultiplier:   fireElemental.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 0.0109, // https://discord.com/channels/260297137554849794/699626629152112730/904088040992026655
		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Fire Shield",
			},
			IsAOE:         true,
			NumberOfTicks: 40,
			TickLength:    time.Second * 3,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				baseDamage := 1.20000004768 * (70 - 1)
				dot.Spell.CalcAndDealAoeDamage(sim, baseDamage, dot.Spell.OutcomeMagicHitAndCrit)
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		},
	})
}
