package hunter

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type HunterPet struct {
	core.Pet

	config PetConfig

	hunterOwner *Hunter

	BestialWrathAura *core.Aura

	KillCommand      *core.Spell
	primaryAbility   *core.Spell
	secondaryAbility *core.Spell
	Dash             *core.Spell

	uptimePercent float64
}

func (hunter *Hunter) NewHunterPet() *HunterPet {
	if hunter.Options.PetType == proto.HunterOptions_PetNone {
		return nil
	}

	if hunter.Options.PetUptime <= 0 {
		return nil
	}

	petConfig := DefaultPetConfigs[hunter.Options.PetType]
	conf := core.PetConfig{
		Name:  petConfig.Name,
		Owner: &hunter.Character,
		BaseStats: stats.Stats{
			stats.Agility:     127,
			stats.Strength:    162,
			stats.AttackPower: -20, // Apparently pets and warriors have a AP penalty.

			// Add 1.8% because pets aren't affected by that component of crit suppression.
			stats.MeleeCritRating: (1.1515 + 1.8) * core.PhysicalCritRatingPerCritPercent,
		},
		StatInheritance:       hunter.makeStatInheritance(),
		EnabledOnStart:        true,
		IsDynamic:             true,
		IsGuardian:            false,
		StartsAtOwnerDistance: true,
	}
	hp := &HunterPet{
		Pet:         core.NewPet(conf),
		config:      petConfig,
		hunterOwner: hunter,
	}

	hp.AddStatDependency(stats.Strength, stats.AttackPower, 2.0)
	hp.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[proto.Class_ClassWarrior])

	hp.EnableFocusBar(1.0 + 0.5*float64(hp.hunterOwner.Talents.BestialDiscipline))

	hp.EnableAutoAttacks(hp, core.AutoAttackOptions{
		MainHand: core.Weapon{
			BaseDamageMin:  42,
			BaseDamageMax:  68,
			CritMultiplier: 2,
			SwingSpeed:     2,
			MaxRange:       core.MaxMeleeRange,
		},
		AutoSwingMelee: true,
	})

	// Pet damage multiplier
	hp.AutoAttacks.MHConfig().DamageMultiplier *= petConfig.DamageMultiplier

	// Happiness
	hp.PseudoStats.DamageDealtMultiplier *= 1.25

	hunter.AddPet(hp)
	return hp
}

func (hp *HunterPet) ApplyTalents() {
	core.MakePermanent(hp.RegisterAura(core.Aura{
		Label:    "Cobra Reflexes",
		ActionID: core.ActionID{SpellID: 25076},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			hp.AutoAttacks.MHAuto().DamageMultiplier *= 0.85
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			hp.AutoAttacks.MHAuto().DamageMultiplier /= 0.85
		},
	}).AttachMultiplicativePseudoStatBuff(
		&hp.PseudoStats.MeleeSpeedMultiplier, 1.3,
	))
}

func (hp *HunterPet) GetPet() *core.Pet {
	return &hp.Pet
}

func (hp *HunterPet) Initialize() {
	hp.Pet.Initialize()
	cfg := DefaultPetConfigs[hp.hunterOwner.Options.PetType]

	if hp.hunterOwner.Options.PetSingleAbility {
		hp.primaryAbility = hp.NewPetAbility(cfg.SecondaryAbility)
		hp.config.RandomSelection = false
	} else {
		hp.primaryAbility = hp.NewPetAbility(cfg.PrimaryAbility)
		hp.secondaryAbility = hp.NewPetAbility(cfg.SecondaryAbility)
	}

	hp.registerKillCommandSpell()
	hp.registerDash()
}

func (hp *HunterPet) Reset(sim *core.Simulation) {
	hp.uptimePercent = min(1, max(0, hp.hunterOwner.Options.PetUptime))
}

func (hp *HunterPet) OnEncounterStart(_ *core.Simulation) {
}

func (hp *HunterPet) ExecuteCustomRotation(sim *core.Simulation) {
	if hp.DistanceFromTarget > core.MaxMeleeRange {
		if hp.Dash.CanCast(sim, hp.CurrentTarget) {
			hp.Dash.Cast(sim, hp.CurrentTarget)
		}

		if !hp.Moving {
			hp.MoveTo(core.MaxMeleeRange-1, sim)
		}

		return
	}

	percentRemaining := sim.GetRemainingDurationPercent()
	if percentRemaining < 1.0-hp.uptimePercent { // once fight is % completed, disable pet.
		hp.Disable(sim)
		return
	}

	target := hp.CurrentTarget

	if hp.config.RandomSelection {
		if sim.RandomFloat("Hunter Pet Ability") >= 0.5 {
			if hp.secondaryAbility.CanCast(sim, target) {
				hp.secondaryAbility.Cast(sim, target)
			} else if hp.primaryAbility.CanCast(sim, target) {
				hp.primaryAbility.Cast(sim, target)
			}

			return
		}
	}

	if hp.primaryAbility.CanCast(sim, target) {
		hp.primaryAbility.Cast(sim, target)
	} else if hp.secondaryAbility.CanCast(sim, target) {
		hp.secondaryAbility.Cast(sim, target)
	}
}

func (hunter *Hunter) makeStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		return stats.Stats{
			stats.Stamina:     ownerStats[stats.Stamina] * 0.3,
			stats.Armor:       ownerStats[stats.Armor] * 0.35,
			stats.AttackPower: ownerStats[stats.RangedAttackPower] * 0.22,
			stats.SpellDamage: ownerStats[stats.RangedAttackPower] * 0.128,
		}
	}
}

type PetConfig struct {
	Name string

	DamageMultiplier float64

	PrimaryAbility   PetAbilityType
	SecondaryAbility PetAbilityType

	RandomSelection bool
}

var DefaultPetConfigs = [...]PetConfig{
	proto.HunterOptions_PetNone:     {},
	proto.HunterOptions_Bat:         {Name: "Bat", DamageMultiplier: 1.07, PrimaryAbility: Bite, SecondaryAbility: Screech},
	proto.HunterOptions_Bear:        {Name: "Bear", DamageMultiplier: 0.91, PrimaryAbility: Bite, SecondaryAbility: Claw},
	proto.HunterOptions_Cat:         {Name: "Cat", DamageMultiplier: 1.1, PrimaryAbility: Bite, SecondaryAbility: Claw},
	proto.HunterOptions_Crab:        {Name: "Crab", DamageMultiplier: 0.95, PrimaryAbility: Claw},
	proto.HunterOptions_Owl:         {Name: "Owl", DamageMultiplier: 1.07, PrimaryAbility: Claw, SecondaryAbility: Screech, RandomSelection: true},
	proto.HunterOptions_Raptor:      {Name: "Raptor", DamageMultiplier: 1.1, PrimaryAbility: Bite, SecondaryAbility: Claw},
	proto.HunterOptions_Ravager:     {Name: "Ravager", DamageMultiplier: 1.1, PrimaryAbility: Bite, SecondaryAbility: Gore},
	proto.HunterOptions_WindSerpent: {Name: "Wind Serpent", DamageMultiplier: 1.07, PrimaryAbility: Bite, SecondaryAbility: LightningBreath},
	proto.HunterOptions_Dragonhawk:  {Name: "Dragonhawk", DamageMultiplier: 1.0, PrimaryAbility: Bite, SecondaryAbility: FireBreath},
}
