package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerBlizzardSpell() {

	// https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=42208
	blizzardCoefficient := 0.11900000274

	blizzardTickSpell := mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 42208},
		SpellSchool:    core.SpellSchoolFrost,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: MageSpellBlizzard,

		DamageMultiplier: 1,
		CritMultiplier:   mage.DefaultSpellCritMultiplier(),
		BonusCoefficient: blizzardCoefficient,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.CalcAndDealAoeDamage(sim, 184, spell.OutcomeMagicHitAndCrit)
		},
	})

	mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 10},
		SpellSchool:    core.SpellSchoolFrost,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagChanneled | core.SpellFlagAPL | core.SpellFlagBinary,
		ClassSpellMask: MageSpellBlizzard,
		ManaCost: core.ManaCostOptions{
			FlatCost: 1645,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label:    "Blizzard",
				ActionID: core.ActionID{SpellID: 10},
			},
			NumberOfTicks:        8,
			TickLength:           time.Second * 1,
			AffectedByCastSpeed:  true,
			HasteReducesDuration: true,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				blizzardTickSpell.Cast(sim, target)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		},
	})
}
