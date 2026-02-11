package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (mage *Mage) registerFlamestrikeSpell() {

	flameStrikeCoefficient := 0.23600000143 // Per https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=exact%253A2120 Field: "BonusCoefficient"
	flameStrikeDotCoefficient := 0.02999999933

	mage.Flamestrike = mage.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 2120},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: MageSpellFlamestrike,

		ManaCost: core.ManaCostOptions{
			FlatCost: 1175,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Second * 3,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   mage.DefaultSpellCritMultiplier(),
		BonusCoefficient: flameStrikeCoefficient,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: "Flamestrike DoT",
			},
			NumberOfTicks:    4,
			TickLength:       time.Second * 2,
			BonusCoefficient: flameStrikeDotCoefficient,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Snapshot(target, 106)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, aoeTarget, dot.OutcomeTick)
				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			baseDamage := mage.CalcAndRollDamageRange(sim, 480, 585)
			spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.AOEDot().Apply(sim)
		},
	})
}
