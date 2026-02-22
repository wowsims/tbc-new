package shaman

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

type FireElemental struct {
	core.Pet

	FireBlast  *core.Spell
	FireNova   *core.Spell
	FireShield *core.Spell

	shamanOwner *Shaman
}

func (shaman *Shaman) NewFireElemental() *FireElemental {
	fireElemental := &FireElemental{
		Pet: core.NewPet(core.PetConfig{
			Name:                            "Greater Fire Elemental",
			Owner:                           &shaman.Character,
			BaseStats:                       shaman.fireElementalBaseStats(),
			NonHitExpStatInheritance:        shaman.fireElementalStatInheritance(),
			EnabledOnStart:                  false,
			IsGuardian:                      true,
			HasDynamicCastSpeedInheritance:  false,
			HasDynamicMeleeSpeedInheritance: false,
		}),
		shamanOwner: shaman,
	}
	baseMeleeDamage := 134.0
	fireElemental.EnableManaBar()
	fireElemental.AddStatDependency(stats.Intellect, stats.Mana, 15)
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
	fireElemental.AutoAttacks.MHConfig().BonusCoefficient = 0.412

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
	fireElemental.registerFireShield()
}

func (fireElemental *FireElemental) Reset(_ *core.Simulation) {
}

func (fireElemental *FireElemental) OnEncounterStart(_ *core.Simulation) {
}

func (fireElemental *FireElemental) ExecuteCustomRotation(sim *core.Simulation) {
	/*
		Fire Shield on CD, Fire Blast/Fire nova random
	*/
	target := fireElemental.CurrentTarget

	if !fireElemental.FireShield.AOEDot().IsActive() {
		fireElemental.TryCast(sim, target, fireElemental.FireShield)
	}
	random := sim.RandomFloat("Fire Elemental Pet Spell")
	if random >= .75 {
		fireElemental.TryCast(sim, target, fireElemental.FireBlast)
	} else if random >= .40 && random < 0.75 {
		fireElemental.TryCast(sim, target, fireElemental.FireNova)
	}

	if !fireElemental.GCD.IsReady(sim) {
		return
	}

	minCd := min(fireElemental.FireBlast.CD.ReadyAt(), fireElemental.FireNova.CD.ReadyAt())
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
		stats.Mana:    3130,
		stats.Stamina: 323,
	}
}

func (shaman *Shaman) fireElementalStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		ownerSpellCritPercent := ownerStats[stats.SpellCritPercent]
		ownerPhysicalCritPercent := ownerStats[stats.PhysicalCritPercent]
		ownerSpellHasteRating := ownerStats[stats.SpellHasteRating]
		ownerMeleeHasteRating := ownerStats[stats.MeleeHasteRating]

		power := ownerStats[stats.SpellDamage] + ownerStats[stats.NatureDamage] -
			ownerStats[stats.AttackPower]*0.1*float64(shaman.Talents.MentalQuickness) // remove Spell Damage that comes from Mental Quickness

		return stats.Stats{
			stats.Stamina:     ownerStats[stats.Stamina] * 0.30,
			stats.Intellect:   ownerStats[stats.Intellect] * 0.30, // https://discord.com/channels/260297137554849794/1474479843428139101/1474888606454775983
			stats.SpellDamage: power,

			stats.SpellCritPercent:    ownerSpellCritPercent,
			stats.PhysicalCritPercent: ownerPhysicalCritPercent,
			stats.SpellHasteRating:    ownerSpellHasteRating,
			stats.MeleeHasteRating:    ownerMeleeHasteRating,
		}
	}
}
