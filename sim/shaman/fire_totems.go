package shaman

import (
	"math"
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func searingTickCount(offset float64) int32 {
	return int32(math.Ceil(40*(1.0+offset))) - 1
}

func (shaman *Shaman) registerSearingTotemSpell() {
	shaman.SearingTotem = shaman.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 25530},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | SpellFlagShamanSpell,
		ClassSpellMask: SpellMaskSearingTotem,
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
		BonusCoefficient: 0.16699999571,
		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Searing Totem",
			},
			// Actual searing totem cast in game is currently 1500 milliseconds with a slight random
			// delay inbetween each cast so using an extra 20 milliseconds to account for the delay
			// subtracting 1 tick so that it doesn't shoot after its actual expiration
			NumberOfTicks: searingTickCount(0),
			TickLength:    time.Millisecond * (1500 + 20), // TODO
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				baseDamage := shaman.CalcAndRollDamageRange(sim, 50, 66)
				dot.Spell.CalcAndDealDamage(sim, target, baseDamage, dot.Spell.OutcomeMagicHitAndCrit)
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			shaman.MagmaTotem.AOEDot().Deactivate(sim)
			shaman.FireElemental.Disable(sim)
			shaman.FireNovaTotemPA.Cancel(sim)
			if sim.CurrentTime < 0 {
				dropTime := sim.CurrentTime
				pa := sim.GetConsumedPendingActionFromPool()

				pa.OnAction = func(sim *core.Simulation) {
					spell.Dot(sim.Encounter.ActiveTargetUnits[0]).BaseTickCount = searingTickCount(dropTime.Minutes())
					spell.Dot(sim.Encounter.ActiveTargetUnits[0]).Apply(sim)
				}

				sim.AddPendingAction(pa)
			} else {
				spell.Dot(sim.Encounter.ActiveTargetUnits[0]).BaseTickCount = searingTickCount(0)
				spell.Dot(sim.Encounter.ActiveTargetUnits[0]).Apply(sim)
			}
			duration := 60
			shaman.TotemExpirations[FireTotem] = sim.CurrentTime + time.Duration(duration)*time.Second
		},
	})
}

func (shaman *Shaman) registerMagmaTotemSpell() {
	shaman.MagmaTotem = shaman.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 25550},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | SpellFlagShamanSpell,
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
				dot.Spell.CalcPeriodicAoeDamage(sim, baseDamage, dot.Spell.OutcomeMagicHitAndCrit)
				dot.Spell.DealBatchedPeriodicDamage(sim)
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			shaman.SearingTotem.Dot(shaman.CurrentTarget).Deactivate(sim)
			shaman.FireElemental.Disable(sim)
			shaman.FireNovaTotemPA.Cancel(sim)
			spell.AOEDot().Apply(sim)

			duration := 20
			shaman.TotemExpirations[FireTotem] = sim.CurrentTime + time.Duration(duration)*time.Second
		},
	})
}

func (shaman *Shaman) registerFireNovaTotemSpell() {
	shaman.FireNovaTotemPA = &core.PendingAction{}

	shaman.MagmaTotem = shaman.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 25537},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | SpellFlagShamanSpell,
		ClassSpellMask: SpellMaskMagmaTotem,
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
			shaman.SearingTotem.Dot(shaman.CurrentTarget).Deactivate(sim)
			shaman.MagmaTotem.AOEDot().Deactivate(sim)
			shaman.FireElemental.Disable(sim)
			duration := 5 * time.Second

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
