package shaman

import (
	"math"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type EarthElemental struct {
	core.Pet

	shamanOwner *Shaman
}

var EarthElementalSpellPowerScaling = 1.3 // Estimated from beta testing

func (shaman *Shaman) NewEarthElemental() *EarthElemental {
	earthElemental := &EarthElemental{
		Pet: core.NewPet(core.PetConfig{
			Name:                            "Greater Earth Elemental",
			Owner:                           &shaman.Character,
			BaseStats:                       shaman.earthElementalBaseStats(),
			NonHitExpStatInheritance:        shaman.earthElementalStatInheritance(),
			EnabledOnStart:                  false,
			IsGuardian:                      true,
			HasDynamicMeleeSpeedInheritance: true,
			HasDynamicCastSpeedInheritance:  true,
		}),
		shamanOwner: shaman,
	}
	baseMeleeDamage := 0.0
	earthElemental.EnableAutoAttacks(earthElemental, core.AutoAttackOptions{
		MainHand: core.Weapon{
			BaseDamageMin:  baseMeleeDamage,
			BaseDamageMax:  baseMeleeDamage,
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
	return stats.Stats{
		stats.Stamina: 10457,
	}
}

func (shaman *Shaman) earthElementalStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		ownerSpellCritPercent := ownerStats[stats.SpellCritPercent]
		ownerPhysicalCritPercent := ownerStats[stats.PhysicalCritPercent]
		ownerHasteRating := ownerStats[stats.SpellHasteRating]
		critPercent := core.TernaryFloat64(math.Abs(ownerPhysicalCritPercent) > math.Abs(ownerSpellCritPercent), ownerPhysicalCritPercent, ownerSpellCritPercent)

		power := core.TernaryFloat64(shaman.Spec == proto.Spec_SpecEnhancementShaman, ownerStats[stats.AttackPower]*0.65, ownerStats[stats.SpellDamage])

		return stats.Stats{
			stats.Stamina:     ownerStats[stats.Stamina],
			stats.AttackPower: power * EarthElementalSpellPowerScaling,

			stats.SpellCritPercent:    critPercent,
			stats.PhysicalCritPercent: critPercent,
			stats.SpellHasteRating:    ownerHasteRating,
		}
	}
}
