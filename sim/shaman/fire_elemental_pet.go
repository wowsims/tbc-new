package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
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
			Name:            "Greater Fire Elemental",
			Owner:           &shaman.Character,
			BaseStats:       shaman.fireElementalBaseStats(),
			StatInheritance: shaman.fireElementalStatInheritance(),
			IsGuardian:      true,
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
			SwingSpeed:     2.0,
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
	if random >= .92 {
		fireElemental.TryCast(sim, target, fireElemental.FireBlast)
	} else if random >= .84 && random < 0.92 {
		fireElemental.TryCast(sim, target, fireElemental.FireNova)
	}

	if !fireElemental.GCD.IsReady(sim) {
		return
	}

	fireElemental.ExtendGCDUntil(sim, max(sim.CurrentTime+time.Second, fireElemental.AutoAttacks.NextAttackAt()))
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
	// Assuming warrior stats for now with 5% of each crit type.
	// Logs suggest at least the crit chances are probably correct
	// and damage value are looking reliable right now
	return core.ClassBaseStats[proto.Class_ClassWarrior].Add(stats.Stats{
		stats.Mana:                4910, // Confirmed in-game level 70
		stats.PhysicalCritPercent: 5,
		stats.SpellCritPercent:    5,
	})
}

func (shaman *Shaman) fireElementalStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		power := ownerStats[stats.SpellDamage] + ownerStats[stats.NatureDamage] -
			ownerStats[stats.AttackPower]*0.1*float64(shaman.Talents.MentalQuickness) // remove Spell Damage that comes from Mental Quickness

		return stats.Stats{
			stats.Stamina:     ownerStats[stats.Stamina] * 0.30,
			stats.Intellect:   ownerStats[stats.Intellect] * 0.30, // https://discord.com/channels/260297137554849794/1474479843428139101/1474888606454775983
			stats.SpellDamage: power,
		}
	}
}
