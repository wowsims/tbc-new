package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (mage *Mage) registerSummonWaterElementalSpell() {
	if !mage.Talents.SummonWaterElemental {
		return
	}

	mage.SummonWaterElemental = mage.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 31687},
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 16,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			mage.waterElemental.Enable(sim, mage.waterElemental)
		},
	})
}

type WaterElemental struct {
	core.Pet

	mageOwner *Mage

	Waterbolt *core.Spell
}

func (mage *Mage) NewWaterElemental() *WaterElemental {
	waterElementalStatInheritance := func(ownerStats stats.Stats) stats.Stats {
		// Water elemental usually has about half the HP of the caster
		return stats.Stats{
			stats.Stamina:          ownerStats[stats.Stamina] * 0.3,
			stats.Intellect:        ownerStats[stats.Intellect] * 0.3,
			stats.SpellDamage:      ownerStats[stats.FrostDamage] * 0.33,
			stats.SpellHitRating:   ownerStats[stats.SpellHitRating],
			stats.SpellPenetration: ownerStats[stats.SpellPenetration],
			stats.SpellCritPercent: ownerStats[stats.SpellCritPercent],
			// this (crit) needs to be tested more thoroughly when pet hit is not bugged
		}
	}

	waterElementalBaseStats := stats.Stats{
		stats.Health: 1596,
		stats.Mana:   1893,
	}

	waterElemental := &WaterElemental{
		Pet: core.NewPet(core.PetConfig{
			Name:                           "Water Elemental",
			Owner:                          &mage.Character,
			BaseStats:                      waterElementalBaseStats,
			NonHitExpStatInheritance:       waterElementalStatInheritance,
			HasDynamicCastSpeedInheritance: true,
			EnabledOnStart:                 true,
			IsGuardian:                     true,
		}),
		mageOwner: mage,
	}
	waterElemental.EnableManaBar()

	mage.AddPet(waterElemental)

	return waterElemental
}

func (we *WaterElemental) GetPet() *core.Pet {
	return &we.Pet
}

func (we *WaterElemental) Initialize() {
	we.registerWaterboltSpell()
}

func (we *WaterElemental) Reset(_ *core.Simulation) {
}

func (we *WaterElemental) OnEncounterStart(_ *core.Simulation) {
}

func (we *WaterElemental) ExecuteCustomRotation(sim *core.Simulation) {
	spell := we.Waterbolt
	spell.Cast(sim, we.CurrentTarget)
}

func (we *WaterElemental) registerWaterboltSpell() {

	waterboltCoefficient := 0.83300000429 // Per https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=31707 Field: "BonusCoefficient"

	we.Waterbolt = we.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 31707},
		SpellSchool:    core.SpellSchoolFrost,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: MageWaterElementalSpellWaterBolt,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 10,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				GCDMin:   core.GCDDefault,
				CastTime: time.Millisecond * 2500,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   we.mageOwner.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		BonusCoefficient: waterboltCoefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := we.CalcAndRollDamageRange(sim, 256, 328)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
