package shaman

import (
	"math"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type FireElemental struct {
	core.Pet

	FireBlast *core.Spell
	FireNova  *core.Spell
	Immolate  *core.Spell

	shamanOwner *Shaman
}

var FireElementalSpellPowerScaling = 0.36

func (shaman *Shaman) NewFireElemental() *FireElemental {
	fireElemental := &FireElemental{
		Pet: core.NewPet(core.PetConfig{
			Name:                            "Greater Fire Elemental",
			Owner:                           &shaman.Character,
			BaseStats:                       shaman.fireElementalBaseStats(),
			NonHitExpStatInheritance:        shaman.fireElementalStatInheritance(),
			EnabledOnStart:                  false,
			IsGuardian:                      true,
			HasDynamicCastSpeedInheritance:  true,
			HasDynamicMeleeSpeedInheritance: true,
		}),
		shamanOwner: shaman,
	}
	baseMeleeDamage := 0.0
	fireElemental.EnableManaBar()
	fireElemental.EnableAutoAttacks(fireElemental, core.AutoAttackOptions{
		MainHand: core.Weapon{
			BaseDamageMin:  baseMeleeDamage,
			BaseDamageMax:  baseMeleeDamage,
			SwingSpeed:     1.4,
			CritMultiplier: fireElemental.DefaultMeleeCritMultiplier(),
			SpellSchool:    core.SpellSchoolFire,
		},
		AutoSwingMelee: true,
	})
	// Need to randomize in enable because the first auto at 0 happens before the randomization (because of prepull)
	fireElemental.AutoAttacks.RandomMeleeOffset = false
	fireElemental.AutoAttacks.MHConfig().ProcMask |= core.ProcMaskSpellDamage
	fireElemental.AutoAttacks.MHConfig().Flags |= SpellFlagShamanSpell
	fireElemental.AutoAttacks.MHConfig().ClassSpellMask |= SpellMaskFireElementalMelee

	fireElemental.OnPetEnable = fireElemental.enable()
	fireElemental.OnPetDisable = fireElemental.disable

	shaman.AddPet(fireElemental)

	return fireElemental
}

func (fireElemental *FireElemental) enable() func(*core.Simulation) {
	return func(sim *core.Simulation) {
		fireElemental.AutoAttacks.RandomizeMeleeTiming(sim)
	}
}

func (fireElemental *FireElemental) disable(sim *core.Simulation) {
}

func (fireElemental *FireElemental) GetPet() *core.Pet {
	return &fireElemental.Pet
}

func (fireElemental *FireElemental) Initialize() {

	fireElemental.registerFireBlast()
	fireElemental.registerFireNova()
	fireElemental.registerImmolate()
}

func (fireElemental *FireElemental) Reset(_ *core.Simulation) {
}

func (fireElemental *FireElemental) OnEncounterStart(_ *core.Simulation) {
}

func (fireElemental *FireElemental) ExecuteCustomRotation(sim *core.Simulation) {
	/*
		Fire Blast on CD, Fire nova on CD when 2+ targets, Immolate on CD if not up on a target
	*/
	target := fireElemental.CurrentTarget

	if len(sim.Encounter.ActiveTargetUnits) > 2 {
		fireElemental.TryCast(sim, target, fireElemental.FireNova)
	}
	fireElemental.FireBlast.Cast(sim, target)

	if !fireElemental.GCD.IsReady(sim) {
		return
	}

	minCd := min(fireElemental.FireBlast.CD.ReadyAt(), fireElemental.FireNova.CD.ReadyAt(), fireElemental.Immolate.CD.ReadyAt())
	fireElemental.ExtendGCDUntil(sim, max(minCd, fireElemental.AutoAttacks.NextAttackAt()))

}

func (fireElemental *FireElemental) TryCast(sim *core.Simulation, target *core.Unit, spell *core.Spell) bool {
	if !spell.Cast(sim, target) {
		return false
	}
	// all spell casts reset the elemental's swing timer
	fireElemental.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+spell.CurCast.CastTime)
	return true
}

func (shaman *Shaman) fireElementalBaseStats() stats.Stats {
	return stats.Stats{
		stats.Mana:    9916,
		stats.Stamina: 7843,
	}
}

func (shaman *Shaman) fireElementalStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		ownerSpellCritPercent := ownerStats[stats.SpellCritPercent]
		ownerPhysicalCritPercent := ownerStats[stats.PhysicalCritPercent]
		ownerHasteRating := ownerStats[stats.SpellHasteRating]
		critPercent := core.TernaryFloat64(math.Abs(ownerPhysicalCritPercent) > math.Abs(ownerSpellCritPercent), ownerPhysicalCritPercent, ownerSpellCritPercent)

		power := core.TernaryFloat64(shaman.Spec == proto.Spec_SpecEnhancementShaman, ownerStats[stats.AttackPower]*0.65, ownerStats[stats.SpellDamage])

		return stats.Stats{
			stats.Stamina:     ownerStats[stats.Stamina] * 0.75,
			stats.SpellDamage: power * FireElementalSpellPowerScaling,

			stats.SpellCritPercent:    critPercent,
			stats.PhysicalCritPercent: critPercent,
			stats.SpellHasteRating:    ownerHasteRating,
		}
	}
}
