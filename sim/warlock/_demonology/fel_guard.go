package demonology

import (
	"math"
	"slices"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/warlock"
)

func (demo *DemonologyWarlock) registerFelguard() *warlock.WarlockPet {
	name := proto.WarlockOptions_Summon_name[int32(proto.WarlockOptions_Felguard)]
	enabledOnStart := proto.WarlockOptions_Felguard == demo.Options.Summon
	return demo.registerFelguardWithName(name, enabledOnStart, false, false)
}

func (demo *DemonologyWarlock) registerFelguardWithName(name string, enabledOnStart bool, autoCastFelstorm bool, isGuardian bool) *warlock.WarlockPet {
	pet := demo.RegisterPet(proto.WarlockOptions_Felguard, 2, 3.5, name, enabledOnStart, isGuardian)
	felStorm := registerFelstorm(pet, demo, autoCastFelstorm)
	legionStrike := registerLegionStrikeSpell(pet, demo)
	pet.MinEnergy = 120

	if !isGuardian {
		demo.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 89751},
			SpellSchool: core.SpellSchoolPhysical,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagAPL | core.SpellFlagNoMetrics,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    demo.NewTimer(),
					Duration: time.Second * 45,
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				pet.AutoCastAbilities = slices.Insert(pet.AutoCastAbilities, 0, felStorm)
			},
		})

		oldEnable := pet.OnPetEnable
		pet.OnPetEnable = func(sim *core.Simulation) {
			if oldEnable != nil {
				oldEnable(sim)
			}

			if len(pet.AutoCastAbilities) > 1 {
				pet.AutoCastAbilities = pet.AutoCastAbilities[1:]
			}
		}
	} else {
		oldEnable := pet.OnPetEnable
		pet.OnPetEnable = func(sim *core.Simulation) {
			if oldEnable != nil {
				oldEnable(sim)
			}
			felStorm.CD.Set(sim.CurrentTime + core.DurationFromSeconds(legionStrike.CD.Duration.Seconds()*math.Round(sim.RollWithLabel(1, 3, "Felstorm Delay"))))
		}
	}

	return pet
}

var legionStrikePetAction = core.ActionID{SpellID: 30213}

func registerLegionStrikeSpell(pet *warlock.WarlockPet, demo *DemonologyWarlock) *core.Spell {
	legionStrike := pet.RegisterSpell(core.SpellConfig{
		ActionID:       legionStrikePetAction,
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: warlock.WarlockSpellFelGuardLegionStrike,

		EnergyCost: core.EnergyCostOptions{
			Cost: 60,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second * 1,
			},

			CD: core.Cooldown{
				Timer:    pet.NewTimer(),
				Duration: time.Millisecond * 1300, // add small cooldown to allow for proper rotation of abilities
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   2,

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			baseDmg := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower()) * 1.3
			baseDmg /= float64(sim.Environment.ActiveTargetCount())
			spell.CalcAndDealAoeDamage(sim, baseDmg, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			// Pets are not affected by Fury gain modifiers
			demo.DemonicFury.Gain(sim, 12, core.ActionID{SpellID: 30213})
		},
	})

	pet.AutoCastAbilities = append(pet.AutoCastAbilities, legionStrike)

	return legionStrike
}

func registerFelstorm(pet *warlock.WarlockPet, _ *DemonologyWarlock, autoCast bool) *core.Spell {
	felStorm := pet.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 89751},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagAoE | core.SpellFlagMeleeMetrics | core.SpellFlagAPL | core.SpellFlagChanneled,
		EnergyCost: core.EnergyCostOptions{
			Cost: 60,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    pet.NewTimer(),
				Duration: time.Second * 45,
			},
		},
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   2,
		Dot: core.DotConfig{
			IsAOE:         true,
			Aura:          core.Aura{Label: "Felstorm"},
			NumberOfTicks: 6,
			TickLength:    time.Second,
			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				baseDamage := dot.Spell.Unit.MHWeaponDamage(sim, dot.Spell.MeleeAttackPower()) + dot.Spell.Unit.OHWeaponDamage(sim, dot.Spell.MeleeAttackPower())
				dot.Spell.CalcAndDealAoeDamage(sim, baseDamage, dot.Spell.OutcomeMeleeSpecialBlockAndCritNoHitCounter)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
			spell.AOEDot().TickOnce(sim)
			pet.AutoAttacks.DelayMeleeBy(sim, spell.AOEDot().BaseDuration())

			// remove from auto cast again to trigger it once
			if !pet.IsGuardian() {
				pet.AutoCastAbilities = pet.AutoCastAbilities[1:]
			}
		},
	})

	if autoCast {
		pet.AutoCastAbilities = append(pet.AutoCastAbilities, felStorm)
	}

	return felStorm
}
