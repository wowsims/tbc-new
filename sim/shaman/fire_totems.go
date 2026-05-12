package shaman

import (
	"math"
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func searingTickCount(offset float64) int32 {
	return int32(math.Ceil(24*(1.0+offset))) - 1
}

func (shaman *Shaman) registerSearingTotemSpell() {
	shaman.SearingTotem = shaman.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 25530},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | SpellFlagShamanSpell | SpellFlagInstant,
		ClassSpellMask: SpellMaskSearingTotem,
		MissileSpeed:   19,
		ManaCost: core.ManaCostOptions{
			FlatCost: 205,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   shaman.DefaultSpellCritMultiplier(),
		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Searing Totem",
			},
			NumberOfTicks: searingTickCount(0),
			// Derived from analysing 100 logs - 1200+ events
			// | Min ms | Avg ms | Median ms | Max ms | Total Delays |
			// | 2001.0 | 2435.6 | 2430.0    | 2954.0 | 1223         |
			TickLength:       time.Millisecond * (2400 + 30),
			BonusCoefficient: 0.16699999571,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				baseDamage := shaman.CalcAndRollDamageRange(sim, 50, 66)
				result := dot.Spell.CalcPeriodicDamage(sim, target, baseDamage, dot.Spell.OutcomeTickMagicHitAndCrit)
				dot.Spell.WaitTravelTime(sim, func(_ *core.Simulation) {
					dot.Spell.DealPeriodicDamage(sim, result)
				})
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			shaman.cancelFireTotems(sim)

			if sim.CurrentTime < 0 {
				dropTime := sim.CurrentTime
				pa := sim.GetConsumedPendingActionFromPool()
				shaman.SearingReplaced = false
				pa.OnAction = func(sim *core.Simulation) {
					if shaman.SearingReplaced {
						return
					}
					spell.Dot(sim.Encounter.ActiveTargetUnits[0]).BaseTickCount = searingTickCount(dropTime.Minutes())
					spell.Dot(sim.Encounter.ActiveTargetUnits[0]).Apply(sim)
				}

				sim.AddPendingAction(pa)
			} else {
				spell.Dot(sim.Encounter.ActiveTargetUnits[0]).BaseTickCount = searingTickCount(0)
				spell.Dot(sim.Encounter.ActiveTargetUnits[0]).Apply(sim)
			}
			duration := time.Second * 60
			shaman.TotemExpirations[FireTotem] = sim.CurrentTime + duration
		},
	})
}

func (shaman *Shaman) registerMagmaTotemSpell() {
	shaman.MagmaTotem = shaman.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 25550},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | SpellFlagShamanSpell | SpellFlagInstant,
		ClassSpellMask: SpellMaskMagmaTotem,
		ManaCost: core.ManaCostOptions{
			FlatCost: 800,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   shaman.DefaultSpellCritMultiplier(),

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: "Magma Totem",
			},
			NumberOfTicks:    10,
			TickLength:       time.Second * 2,
			BonusCoefficient: 0.06700000167,

			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				baseDamage := 97.0
				dot.Spell.CalcPeriodicAoeDamage(sim, baseDamage, dot.Spell.OutcomeTickMagicHitAndCrit)
				dot.Spell.DealBatchedPeriodicDamage(sim)
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			shaman.cancelFireTotems(sim)
			spell.AOEDot().Apply(sim)

			duration := time.Second * 20
			shaman.TotemExpirations[FireTotem] = sim.CurrentTime + duration
		},
	})
}

func (shaman *Shaman) registerFireNovaTotemSpell() {
	shaman.FireNovaTotemPA = &core.PendingAction{}
	duration := time.Duration(4-shaman.Talents.ImprovedFireTotems) * time.Second

	shaman.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 25537},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | SpellFlagShamanSpell | SpellFlagInstant,
		ClassSpellMask: SpellMaskFireNovaTotem,
		ManaCost: core.ManaCostOptions{
			FlatCost: 765,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * 15,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   shaman.DefaultSpellCritMultiplier(),
		BonusCoefficient: 0.21400000155,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			shaman.cancelFireTotems(sim)

			baseDamage := shaman.CalcAndRollDamageRange(sim, 654, 730)
			spell.CalcAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)

			shaman.FireNovaTotemPA.OnAction = func(sim *core.Simulation) {
				spell.DealBatchedAoeDamage(sim)
			}
			shaman.FireNovaTotemPA.NextActionAt = sim.CurrentTime + duration
			sim.AddPendingAction(shaman.FireNovaTotemPA)

			shaman.TotemExpirations[FireTotem] = sim.CurrentTime + duration
		},
	})
}

func (shaman *Shaman) cancelFireTotems(sim *core.Simulation) {
	shaman.MagmaTotem.AOEDot().Deactivate(sim)
	shaman.FireNovaTotemPA.Cancel(sim)
	searingTotemDot := shaman.SearingTotem.Dot(shaman.CurrentTarget)
	if searingTotemDot != nil {
		searingTotemDot.Deactivate(sim)
	}
	shaman.SearingReplaced = true
	if shaman.TotemOfWrath != nil {
		shaman.TotemOfWrath.RelatedSelfBuff.Deactivate(sim)
	}
	if shaman.FireElementalTotem != nil {
		shaman.FireElementalTotem.RelatedSelfBuff.Deactivate(sim)
	}
}
