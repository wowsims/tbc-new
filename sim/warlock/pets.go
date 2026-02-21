package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type WarlockPet struct {
	core.Pet

	AutoCastAbilities []*core.Spell
	MinMana           float64 // The minimum amount of energy needed to the AI casts a spell
	ManaIntRatio      float64
}

var petBaseStats = map[proto.WarlockOptions_Summon]*stats.Stats{
	proto.WarlockOptions_Imp: {
		stats.Mana:        2988,
		stats.Stamina:     101,
		stats.Strength:    153, //fix these later
		stats.Agility:     108, //fix these later
		stats.Intellect:   327,
		stats.Spirit:      263,
		stats.AttackPower: 135,
		stats.MP5:         123,
	},
	proto.WarlockOptions_Voidwalker: {
		stats.Stamina:     280,
		stats.Strength:    153,
		stats.Agility:     108,
		stats.Intellect:   133,
		stats.Spirit:      122,
		stats.AttackPower: 286,
		stats.MP5:         48,
	},
	proto.WarlockOptions_Succubus: {
		stats.Mana:        3862,
		stats.Stamina:     280,
		stats.Strength:    154,
		stats.Agility:     108,
		stats.Intellect:   133,
		stats.Spirit:      122,
		stats.AttackPower: 286,
		stats.MP5:         48,
	},
	proto.WarlockOptions_Felhunter: {},
	proto.WarlockOptions_Felguard: {
		stats.Stamina:     280,
		stats.Mana:        3862,
		stats.Strength:    153,
		stats.Agility:     108,
		stats.Intellect:   133,
		stats.Spirit:      122,
		stats.AttackPower: 286,
		stats.MP5:         48,
	},
}

func (warlock *Warlock) SimplePetStatInheritanceWithScale() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		const resistScale = 0.4
		const baseStatScale = 0.3

		return stats.Stats{
			stats.Stamina:          ownerStats[stats.Stamina] * 0.3,
			stats.Intellect:        ownerStats[stats.Intellect] * 0.3,
			stats.Armor:            ownerStats[stats.Armor] * 0.35,
			stats.SpellPenetration: ownerStats[stats.SpellPenetration], // not 100% on this one
			stats.SpellDamage:      max(ownerStats[stats.ShadowDamage], ownerStats[stats.FireDamage]) * 0.15,
			stats.AttackPower:      max(ownerStats[stats.ShadowDamage], ownerStats[stats.FireDamage]) * 0.57,
			stats.ArcaneResistance: ownerStats[stats.ArcaneResistance] * resistScale,
			stats.FireResistance:   ownerStats[stats.FireResistance] * resistScale,
			stats.FrostResistance:  ownerStats[stats.FrostResistance] * resistScale,
			stats.NatureResistance: ownerStats[stats.NatureResistance] * resistScale,
			stats.ShadowResistance: ownerStats[stats.ShadowResistance] * resistScale,
		}
	}
}

func AutoAttackConfig(min float64, max float64) *core.AutoAttackOptions {
	return &core.AutoAttackOptions{
		MainHand: core.Weapon{
			BaseDamageMin:  float64(min),
			BaseDamageMax:  float64(max),
			SwingSpeed:     2.0,
			CritMultiplier: 2,
		},
		AutoSwingMelee: true,
	}
}

func (warlock *Warlock) makePet(
	name string,
	enabledOnStart bool,
	baseStats stats.Stats,
	aaOptions *core.AutoAttackOptions,
	statInheritance core.PetStatInheritance,
	isGuardian bool,
) *WarlockPet {
	pet := &WarlockPet{
		Pet: core.NewPet(core.PetConfig{
			Name:                     name,
			Owner:                    &warlock.Character,
			BaseStats:                baseStats,
			NonHitExpStatInheritance: statInheritance,
			EnabledOnStart:           enabledOnStart,
			IsGuardian:               isGuardian,
		}),
	}

	// set pet class for proper scaling values
	if enabledOnStart {
		warlock.ActivePet = pet
		pet.OnPetEnable = func(sim *core.Simulation) {
			if warlock.Talents.DemonicKnowledge > 0 {
				if warlock.DemonicKnowledgeAura.IsActive() {
					warlock.DemonicKnowledgeAura.Deactivate(sim)
				}
				warlock.DemonicKnowledgeAura.Activate(sim)
			}
			if warlock.Talents.MasterDemonologist > 0 {
				if warlock.MasterDemonologistAura.IsActive() {
					warlock.MasterDemonologistAura.Deactivate(sim)
				}
				warlock.MasterDemonologistAura.Activate(sim)
			}
		}
		warlock.RegisterResetEffect(func(sim *core.Simulation) {
			warlock.ActivePet = pet
		})
	}

	warlock.setPetOptions(pet, aaOptions)

	return pet
}

func (warlock *Warlock) setPetOptions(petAgent core.PetAgent, aaOptions *core.AutoAttackOptions) {
	pet := petAgent.GetPet()
	if aaOptions != nil {
		pet.EnableAutoAttacks(petAgent, *aaOptions)
	}

	pet.EnableManaBar()
	warlock.AddPet(petAgent)
}

func (warlock *Warlock) registerPets() {
	warlock.Imp = warlock.registerImp()
	warlock.Succubus = warlock.registerSuccubus()
	warlock.Felhunter = warlock.registerFelHunter()
	warlock.Voidwalker = warlock.registerVoidWalker()
}

func (warlock *Warlock) registerImp() *WarlockPet {
	name := proto.WarlockOptions_Summon_name[int32(proto.WarlockOptions_Imp)]
	enabledOnStart := proto.WarlockOptions_Imp == warlock.Options.Summon
	return warlock.registerImpWithName(name, enabledOnStart, false)
}

func (warlock *Warlock) registerImpWithName(name string, enabledOnStart bool, isGuardian bool) *WarlockPet {
	pet := warlock.RegisterPet(proto.WarlockOptions_Imp, 0, 0, name, enabledOnStart, isGuardian)
	pet.registerFireboltSpell()
	pet.MinMana = 145
	return pet
}

func (warlock *Warlock) registerFelHunter() *WarlockPet {
	name := proto.WarlockOptions_Summon_name[int32(proto.WarlockOptions_Felhunter)]
	enabledOnStart := proto.WarlockOptions_Felhunter == warlock.Options.Summon
	return warlock.registerFelHunterWithName(name, enabledOnStart, false)
}

func (warlock *Warlock) registerFelHunterWithName(name string, enabledOnStart bool, isGuardian bool) *WarlockPet {
	pet := warlock.RegisterPet(proto.WarlockOptions_Felhunter, 2, 3.5, name, enabledOnStart, isGuardian)
	//add felhunter ability
	pet.MinMana = 130
	return pet
}

func (warlock *Warlock) registerVoidWalker() *WarlockPet {
	name := proto.WarlockOptions_Summon_name[int32(proto.WarlockOptions_Voidwalker)]
	enabledOnStart := proto.WarlockOptions_Voidwalker == warlock.Options.Summon
	return warlock.registerVoidWalkerWithName(name, enabledOnStart, false)
}

func (warlock *Warlock) registerVoidWalkerWithName(name string, enabledOnStart bool, isGuardian bool) *WarlockPet {
	pet := warlock.RegisterPet(proto.WarlockOptions_Voidwalker, 2, 3.5, name, enabledOnStart, isGuardian)
	pet.registerTormentSpell()
	pet.MinMana = 120
	return pet
}

func (warlock *Warlock) registerSuccubus() *WarlockPet {
	name := proto.WarlockOptions_Summon_name[int32(proto.WarlockOptions_Succubus)]
	enabledOnStart := proto.WarlockOptions_Succubus == warlock.Options.Summon
	return warlock.registerSuccubusWithName(name, enabledOnStart, false)
}

func (warlock *Warlock) registerSuccubusWithName(name string, enabledOnStart bool, isGuardian bool) *WarlockPet {
	pet := warlock.RegisterPet(proto.WarlockOptions_Succubus, 173, 216, name, enabledOnStart, isGuardian)
	pet.registerLashOfPainSpell()
	pet.MinMana = 190
	return pet
}

func (warlock *Warlock) RegisterPet(
	t proto.WarlockOptions_Summon,
	min float64,
	max float64,
	name string,
	enabledOnStart bool,
	isGuardian bool,
) *WarlockPet {
	baseStats, ok := petBaseStats[t]
	if !ok {
		panic("Undefined base stats for pet")
	}

	var attackOptions *core.AutoAttackOptions = nil
	if t > 1 {
		attackOptions = AutoAttackConfig(min, max)
	}

	inheritance := warlock.SimplePetStatInheritanceWithScale()
	return warlock.makePet(name, enabledOnStart, *baseStats, attackOptions, inheritance, isGuardian)
}

func (pet *WarlockPet) GetPet() *core.Pet {
	return &pet.Pet
}

func (pet *WarlockPet) Reset(_ *core.Simulation) {
}

func (pet *WarlockPet) OnEncounterStart(_ *core.Simulation) {
}

func (pet *WarlockPet) ExecuteCustomRotation(sim *core.Simulation) {
	waitUntil := time.Duration(1<<63 - 1)

	for _, spell := range pet.AutoCastAbilities {
		if spell.CanCast(sim, pet.CurrentTarget) && pet.CurrentMana() > pet.MinMana {
			spell.Cast(sim, pet.CurrentTarget)
			return
		}

		// calculate energy required
		cost := max(pet.MinMana, spell.Cost.GetCurrentCost())
		timeTillMana := max(0, (cost-pet.CurrentMana())/pet.ManaRegenPerSecondWhileCombat())
		waitUntil = min(waitUntil, time.Duration(float64(time.Second)*timeTillMana))
	}

	// for now average the delay out to 100 ms so we don't need to roll random every time
	pet.WaitUntil(sim, sim.CurrentTime+waitUntil+time.Millisecond*100)
}

var petActionFireBolt = core.ActionID{SpellID: 3110}

func (pet *WarlockPet) registerFireboltSpell() {
	pet.AutoCastAbilities = append(pet.AutoCastAbilities, pet.RegisterSpell(core.SpellConfig{
		ActionID:       petActionFireBolt,
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: WarlockSpellImpFireBolt,
		MissileSpeed:   16,

		ManaCost: core.ManaCostOptions{
			FlatCost: 145,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      time.Millisecond * 1500,
				CastTime: time.Millisecond * 2000,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   1.5,
		ThreatMultiplier: 1,
		BonusCoefficient: 0.571,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			dmgRoll := pet.CalcAndRollDamageRange(sim, 112, 127)
			result := spell.CalcDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}))
}

var petActionLashOfPain = core.ActionID{SpellID: 7814}

func (pet *WarlockPet) registerLashOfPainSpell() {
	pet.AutoCastAbilities = append(pet.AutoCastAbilities, pet.RegisterSpell(core.SpellConfig{
		ActionID:       petActionLashOfPain,
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: WarlockSpellSuccubusLashOfPain,
		ManaCost: core.ManaCostOptions{
			FlatCost: 190,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			CD: core.Cooldown{
				Timer:    pet.NewTimer(),
				Duration: 12 * time.Second,
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   1.5,
		ThreatMultiplier: 1,
		BonusCoefficient: 0.429,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, 123, spell.OutcomeMagicHitAndCrit)
			spell.DealDamage(sim, result)
		},
	}))

}

var petActionTorment = core.ActionID{SpellID: 27270}

func (pet *WarlockPet) registerTormentSpell() {
	pet.AutoCastAbilities = append(pet.AutoCastAbilities, pet.RegisterSpell(core.SpellConfig{
		ActionID:       petActionTorment,
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: WarlockSpellVoidwalkerTorment,
		ManaCost: core.ManaCostOptions{
			FlatCost: 130,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, 1000, spell.OutcomeMagicHitAndCrit)
			spell.DealDamage(sim, result)
		},
	}))
}
