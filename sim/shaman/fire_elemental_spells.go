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
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    fireElemental.NewTimer(),
				Duration: time.Second * 6,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   fireElemental.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 535.0 // https://discord.com/channels/260297137554849794/699626629152112730/904088040992026655
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
			CD: core.Cooldown{
				Timer:    fireElemental.NewTimer(),
				Duration: time.Second * 6,
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		CritMultiplier:   fireElemental.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 0.332, // https://discord.com/channels/260297137554849794/699626629152112730/904088040992026655

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, func(sim *core.Simulation, _ *core.Spell) float64 {
				return 703 // https://discord.com/channels/260297137554849794/699626629152112730/904088040992026655
			})
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

		DamageMultiplier: 1,
		CritMultiplier:   fireElemental.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,

		ManaCost: core.ManaCostOptions{
			FlatCost: 95,
		},
		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Fire Shield",
			},
			IsAOE:            true,
			NumberOfTicks:    40,
			TickLength:       time.Second * 3,
			BonusCoefficient: 0.0109, // https://discord.com/channels/260297137554849794/699626629152112730/904088040992026655
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Spell.CalcAndDealAoeDamage(sim, 84, dot.Spell.OutcomeMagicHitAndCrit) // https://discord.com/channels/260297137554849794/699626629152112730/904088040992026655
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		},
	})
}
