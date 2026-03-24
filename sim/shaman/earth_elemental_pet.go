package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type EarthElemental struct {
	core.Pet

	shamanOwner *Shaman
}

var EarthElementalSpellPowerScaling = 0.5

func (shaman *Shaman) NewEarthElemental() *EarthElemental {
	earthElemental := &EarthElemental{
		Pet: core.NewPet(core.PetConfig{
			Name:            "Greater Earth Elemental",
			Owner:           &shaman.Character,
			BaseStats:       shaman.earthElementalBaseStats(),
			StatInheritance: shaman.earthElementalStatInheritance(),
			EnabledOnStart:  false,
			IsGuardian:      true,
		}),
		shamanOwner: shaman,
	}

	earthElemental.AddStatDependency(stats.Strength, stats.AttackPower, 2)
	earthElemental.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[proto.Class_ClassWarrior])

	earthElemental.EnableAutoAttacks(earthElemental, core.AutoAttackOptions{
		MainHand: core.Weapon{
			// https://discord.com/channels/260297137554849794/1474479843428139101/1480955121394520237
			BaseDamageMin:  174,
			BaseDamageMax:  196,
			SwingSpeed:     2,
			CritMultiplier: earthElemental.DefaultMeleeCritMultiplier(),
			SpellSchool:    core.SpellSchoolPhysical,
		},
		AutoSwingMelee: true,
	})

	earthElemental.OnPetEnable = earthElemental.enable()
	earthElemental.OnPetDisable = earthElemental.disable

	shaman.AddPet(earthElemental)

	return earthElemental
}

func (earthElemental *EarthElemental) enable() func(*core.Simulation) {
	return func(sim *core.Simulation) {
	}
}

func (earthElemental *EarthElemental) disable(sim *core.Simulation) {
}

func (earthElemental *EarthElemental) GetPet() *core.Pet {
	return &earthElemental.Pet
}

func (earthElemental *EarthElemental) Initialize() {
}

func (earthElemental *EarthElemental) Reset(_ *core.Simulation) {
}

func (earthElemental *EarthElemental) OnEncounterStart(_ *core.Simulation) {
}

func (earthElemental *EarthElemental) ExecuteCustomRotation(sim *core.Simulation) {
	earthElemental.ExtendGCDUntil(sim, sim.CurrentTime+time.Second)
}

func (earthElemental *EarthElemental) TryCast(sim *core.Simulation, target *core.Unit, spell *core.Spell) bool {
	if !spell.Cast(sim, target) {
		return false
	}
	// all spell casts reset the elemental's swing timer
	earthElemental.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+spell.CurCast.CastTime)
	return true
}

func (shaman *Shaman) earthElementalBaseStats() stats.Stats {
	// Assuming warrior stats for now with 5% of each crit type.
	// Logs suggest at least the crit chances are probably correct
	// and damage value are looking reliable right now
	return core.ClassBaseStats[proto.Class_ClassWarrior].Add(stats.Stats{
		stats.PhysicalCritPercent: 5,
	})
}

func (shaman *Shaman) earthElementalStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		power := ownerStats[stats.AttackPower] * 0.65

		return stats.Stats{
			stats.Stamina:     ownerStats[stats.Stamina],
			stats.AttackPower: power * EarthElementalSpellPowerScaling,
		}
	}
}
