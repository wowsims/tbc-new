package druid

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type DruidForm uint8

const (
	Humanoid DruidForm = 1 << iota
	Bear
	Cat
	Moonkin
	Tree
	Any = Humanoid | Bear | Cat | Moonkin | Tree
)

// Converts from 0.009327 to 0.0085
const AnimalSpiritRegenSuppression = 0.911337

// Thick Hide contribution handled separately in talents code for cleanliness
// and UI stats display.
const BaseBearArmorMulti = 5.0

func (form DruidForm) Matches(other DruidForm) bool {
	return (form & other) != 0
}

// func (druid *Druid) GetForm() DruidForm {
// 	return druid.form
// }

func (druid *Druid) InForm(form DruidForm) bool {
	return druid.form.Matches(form)
}

func (druid *Druid) ClearForm(sim *core.Simulation) {
	if druid.InForm(Cat) {
		druid.CatFormAura.Deactivate(sim)
	} else if druid.InForm(Bear) {
		druid.BearFormAura.Deactivate(sim)
	} else if druid.InForm(Moonkin) {
		druid.MoonkinFormAura.Deactivate(sim)
	}

	druid.form = Humanoid
	druid.SetCurrentPowerBar(core.ManaBar)
}

func (druid *Druid) GetCatWeapon() core.Weapon {
	unscaledWeapon := druid.WeaponFromMainHand(0)
	return core.Weapon{
		BaseDamageMin:        unscaledWeapon.BaseDamageMin / unscaledWeapon.SwingSpeed,
		BaseDamageMax:        unscaledWeapon.BaseDamageMax / unscaledWeapon.SwingSpeed,
		SwingSpeed:           1.0,
		NormalizedSwingSpeed: 1.0,
		CritMultiplier:       druid.FeralCritMultiplier(),
		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
		MaxRange:             core.MaxMeleeRange,
	}
}

func (druid *Druid) GetBearWeapon() core.Weapon {
	unscaledWeapon := druid.WeaponFromMainHand(0)
	return core.Weapon{
		BaseDamageMin:        unscaledWeapon.BaseDamageMin / unscaledWeapon.SwingSpeed * 2.5,
		BaseDamageMax:        unscaledWeapon.BaseDamageMax / unscaledWeapon.SwingSpeed * 2.5,
		SwingSpeed:           2.5,
		NormalizedSwingSpeed: 2.5,
		CritMultiplier:       druid.FeralCritMultiplier(),
		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
		MaxRange:             core.MaxMeleeRange,
	}
}

func (druid *Druid) RegisterCatFormAura() {
	actionID := core.ActionID{SpellID: 768}
	energyMetrics := druid.NewEnergyMetrics(actionID)

	furorProcChance := 0.2 * float64(druid.Talents.Furor)
	wolfsheadEquipped := druid.HasItemEquipped(8345, []proto.ItemSlot{proto.ItemSlot_ItemSlotHead})

	// In Cat Form each point of Agility gives 1 AP.
	agiApDep := druid.NewDynamicStatDependency(stats.Agility, stats.AttackPower, 1)
	// In Cat Form each point of Strength gives 2 AP (vs 1 AP in humanoid form).
	// The static dep in druid.go provides 1 AP/Str always; this dynamic dep adds the extra 1 AP/Str.
	strApDep := druid.NewDynamicStatDependency(stats.Strength, stats.AttackPower, 1)
	// Feral Attack Power (weapon/item feral-specific AP) converts 1:1 to AP in Cat Form.
	feralApDep := druid.NewDynamicStatDependency(stats.FeralAttackPower, stats.AttackPower, 1)

	// Talent: Heart of the Wild — +2% AP per rank while in Cat form.
	var hotWCatApDep *stats.StatDependency
	if druid.Talents.HeartOfTheWild > 0 {
		hotWCatApDep = druid.NewDynamicMultiplyStat(stats.AttackPower, 1+0.02*float64(druid.Talents.HeartOfTheWild))
	}

	clawWeapon := druid.GetCatWeapon()

	statBonus := stats.Stats{
		stats.AttackPower: 2 * float64(core.CharacterLevel),
	}

	druid.CatFormAura = druid.RegisterAura(core.Aura{
		Label:      "Cat Form",
		ActionID:   actionID,
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(druid.StartingForm.Matches(Cat), core.CharacterBuildPhaseBase, core.CharacterBuildPhaseNone),
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			if !druid.Env.MeasuringStats && druid.form != Humanoid {
				druid.ClearForm(sim)
			}
			druid.form = Cat
			druid.SetCurrentPowerBar(core.EnergyBar)

			druid.PseudoStats.ThreatMultiplier *= 0.71
			druid.PseudoStats.SpiritRegenMultiplier *= AnimalSpiritRegenSuppression

			druid.AddStatsDynamic(sim, statBonus)
			druid.EnableBuildPhaseStatDep(sim, agiApDep)
			druid.EnableBuildPhaseStatDep(sim, strApDep)
			druid.EnableBuildPhaseStatDep(sim, feralApDep)
			if hotWCatApDep != nil {
				druid.EnableBuildPhaseStatDep(sim, hotWCatApDep)
			}

			if !druid.Env.MeasuringStats {
				druid.AutoAttacks.SetMH(clawWeapon)
				druid.AutoAttacks.EnableAutoSwing(sim)
				druid.UpdateManaRegenRates()

				if sim.CurrentTime > 0 {
					if cur := druid.CurrentEnergy(); cur > 0 {
						//Resets energy to 0 when entering cat form
						druid.SpendEnergy(sim, cur, energyMetrics)
					}
					// Wolfshead Helm: +20 energy on shift into Cat.
					energyGain := core.TernaryFloat64(wolfsheadEquipped, 20.0, 0.0)
					// Furor: 20% chance per rank (rank 5 = 100%) to gain 40 energy on shift.
					if furorProcChance == 1 || (furorProcChance > 0 && sim.RandomFloat("Furor") < furorProcChance) {
						energyGain += 40.0
					}
					if energyGain > 0 {
						druid.AddEnergy(sim, energyGain, energyMetrics)
					}
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.form = Humanoid

			druid.PseudoStats.ThreatMultiplier /= 0.71
			druid.PseudoStats.SpiritRegenMultiplier /= AnimalSpiritRegenSuppression

			druid.AddStatsDynamic(sim, statBonus.Invert())
			druid.DisableBuildPhaseStatDep(sim, agiApDep)
			druid.DisableBuildPhaseStatDep(sim, strApDep)
			druid.DisableBuildPhaseStatDep(sim, feralApDep)
			if hotWCatApDep != nil {
				druid.DisableBuildPhaseStatDep(sim, hotWCatApDep)
			}

			if druid.TigersFuryAura != nil {
				druid.TigersFuryAura.Deactivate(sim)
			}

			if !druid.Env.MeasuringStats {
				druid.AutoAttacks.SetMH(druid.WeaponFromMainHand(druid.DefaultMeleeCritMultiplier()))
				druid.AutoAttacks.EnableAutoSwing(sim)
				druid.UpdateManaRegenRates()
			}
		},
	})

	druid.CatFormAura.NewPassiveMovementSpeedEffect(0.25)
}

func (druid *Druid) registerCatFormSpell() {
	druid.CatForm = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 768},
		ClassSpellMask: DruidSpellCatForm,
		Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 35,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			if druid.CatFormAura.IsActive() {
				druid.CatFormAura.Deactivate(sim)
			}
			druid.CatFormAura.Activate(sim)
		},
	})
}

func (druid *Druid) RegisterBearFormAura() {
	actionID := core.ActionID{SpellID: 9634} // Dire Bear Form
	healthMetrics := druid.NewHealthMetrics(actionID)

	statBonus := stats.Stats{
		stats.AttackPower: 3 * float64(core.CharacterLevel),
	}

	strApDep := druid.NewDynamicStatDependency(stats.Strength, stats.AttackPower, 1)
	feralApDep := druid.NewDynamicStatDependency(stats.FeralAttackPower, stats.AttackPower, 1)
	stamDep := druid.NewDynamicMultiplyStat(stats.Stamina, 1.25)
	// Talent: Heart of the Wild — +4% Stamina per rank while in Bear form.
	var hotWBearStamDep *stats.StatDependency
	if druid.Talents.HeartOfTheWild > 0 {
		hotWBearStamDep = druid.NewDynamicMultiplyStat(stats.Stamina, 1+0.04*float64(druid.Talents.HeartOfTheWild))
	}

	clawWeapon := druid.GetBearWeapon()

	druid.BearFormAura = druid.RegisterAura(core.Aura{
		Label:      "Bear Form",
		ActionID:   actionID,
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(druid.StartingForm.Matches(Bear), core.CharacterBuildPhaseBase, core.CharacterBuildPhaseNone),
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			if !druid.Env.MeasuringStats && druid.form != Humanoid {
				druid.ClearForm(sim)
			}
			druid.form = Bear
			druid.SetCurrentPowerBar(core.RageBar)

			druid.PseudoStats.ThreatMultiplier *= 1.3
			druid.PseudoStats.SpiritRegenMultiplier *= AnimalSpiritRegenSuppression

			druid.AddStatsDynamic(sim, statBonus)
			druid.ApplyDynamicEquipScaling(sim, stats.Armor, BaseBearArmorMulti)
			druid.ApplyDynamicEquipScaling(sim, stats.BonusArmor, BaseBearArmorMulti)
			druid.EnableBuildPhaseStatDep(sim, strApDep)
			druid.EnableBuildPhaseStatDep(sim, feralApDep)

			// Preserve fraction of max health when shifting
			healthFrac := druid.CurrentHealth() / druid.MaxHealth()
			druid.EnableBuildPhaseStatDep(sim, stamDep)
			if hotWBearStamDep != nil {
				druid.EnableBuildPhaseStatDep(sim, hotWBearStamDep)
			}

			if !druid.Env.MeasuringStats {
				druid.GainHealth(sim, healthFrac*druid.MaxHealth()-druid.CurrentHealth(), healthMetrics)
				druid.AutoAttacks.SetMH(clawWeapon)
				druid.AutoAttacks.EnableAutoSwing(sim)
				druid.UpdateManaRegenRates()
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.form = Humanoid

			druid.PseudoStats.ThreatMultiplier /= 1.3
			druid.PseudoStats.SpiritRegenMultiplier /= AnimalSpiritRegenSuppression

			druid.AddStatsDynamic(sim, statBonus.Invert())
			druid.RemoveDynamicEquipScaling(sim, stats.Armor, BaseBearArmorMulti)
			druid.RemoveDynamicEquipScaling(sim, stats.BonusArmor, BaseBearArmorMulti)
			druid.DisableBuildPhaseStatDep(sim, strApDep)
			druid.DisableBuildPhaseStatDep(sim, feralApDep)

			healthFrac := druid.CurrentHealth() / druid.MaxHealth()
			druid.DisableBuildPhaseStatDep(sim, stamDep)
			if hotWBearStamDep != nil {
				druid.DisableBuildPhaseStatDep(sim, hotWBearStamDep)
			}

			if !druid.Env.MeasuringStats {
				druid.RemoveHealth(sim, druid.CurrentHealth()-healthFrac*druid.MaxHealth())
				druid.AutoAttacks.SetMH(druid.WeaponFromMainHand(druid.DefaultMeleeCritMultiplier()))
				druid.AutoAttacks.EnableAutoSwing(sim)
				druid.UpdateManaRegenRates()
			}
		},
	})
}

func (druid *Druid) registerBearFormSpell() {
	actionID := core.ActionID{SpellID: 9634} // Dire Bear Form
	rageMetrics := druid.NewRageMetrics(actionID)

	druid.BearForm = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID:       actionID,
		ClassSpellMask: DruidSpellBearForm,
		Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 35,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			rageDelta := 10.0 - druid.CurrentRage()
			if rageDelta > 0 {
				druid.AddRage(sim, rageDelta, rageMetrics)
			} else if rageDelta < 0 {
				druid.SpendRage(sim, -rageDelta, rageMetrics)
			}
			druid.BearFormAura.Activate(sim)
		},
	})
}

func (druid *Druid) RegisterMoonkinFormAura() {
	if !druid.Talents.MoonkinForm {
		return
	}

	druid.MoonkinFormAura = druid.RegisterAura(core.Aura{
		Label:      "Moonkin Form",
		ActionID:   core.ActionID{SpellID: 24858},
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(druid.StartingForm.Matches(Moonkin), core.CharacterBuildPhaseBase, core.CharacterBuildPhaseNone),
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			if !druid.Env.MeasuringStats && druid.form != Moonkin {
				druid.ClearForm(sim)
			}

			druid.ApplyDynamicEquipScaling(sim, stats.Armor, 4)

			druid.form = Moonkin
			druid.SetCurrentPowerBar(core.ManaBar)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.RemoveDynamicEquipScaling(sim, stats.Armor, 4)
			druid.form = Humanoid
		},
	})

	manaMetrics := druid.NewManaMetrics(core.ActionID{SpellID: 33926 /* Elune's Touch */})

	// Elune's Touch is assumed to have a PPM of 15.
	// Mana gained is 30% of melee attack power.

	druid.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Elune's Touch",
		DPM:                druid.NewStaticLegacyPPMManager(15, core.ProcMaskMeleeWhiteHit),
		RequireDamageDealt: true,
		Outcome:            core.OutcomeLanded,
		Callback:           core.CallbackOnSpellHitDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			druid.AddMana(sim, float64(stats.AttackPower)*0.3, manaMetrics)
		},
	})
}

func (druid *Druid) RegisterMoonkinFormSpell() {
	if !druid.Talents.MoonkinForm {
		return
	}

	druid.MoonkinForm = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID: core.ActionID{SpellID: 24858},
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 9.3,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.MoonkinFormAura.Activate(sim)
		},
	})
}
