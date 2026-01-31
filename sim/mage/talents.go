package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (mage *Mage) ApplyTalents() {

	//------- ARCANE --------
	mage.registerArcaneSubtlety()
	mage.registerArcaneFocus()
	//mage.registerImprovedArcaneMissiles()

	// mage.registerWandSpecialization()
	// mage.registerMagicAbsorption()
	mage.registerArcaneConcentration()

	// mage.registerMagicAttunement()
	mage.registerArcaneImpact()
	//mage.registerArcaneFortitude()

	// mage.registerImprovedManaShield()
	// mage.registerImprovedCounterspell()
	mage.registerArcaneMeditation()

	//mage.registerImprovedBlink()
	mage.registerArcaneMind()

	// mage.registerPrismaticCloak()
	mage.registerArcaneInstability()

	mage.registerEmpoweredArcaneMissiles()
	mage.registerSpellPower()

	mage.registerMindMastery()

	//-------  FIRE  --------
	mage.registerImprovedFireball()
	// mage.registerImpact()

	mage.registerIgnite()
	// mage.registerFlameThrowing()
	mage.registerImprovedFireBlast()

	mage.registerIncineration()
	mage.registerImprovedFlamestrike()
	mage.registerBurningSoul()

	// mage.registerMoltenShields()
	mage.registerMasterOfElements()

	mage.registerPlayingWithFire()
	mage.registerCriticalMass()

	// mage.registerBlazingSpeed()
	mage.registerFirePower()

	mage.registerPyromaniac()
	mage.registerMoltenFury()

	mage.registerEmpoweredFireball()

	//------- FROST --------
	// mage.registerFrostWarding()
	mage.registerImprovedFrostbolt()
	mage.registerElementalPrecision()

	mage.registerIceShards()
	// mage.registerFrostbite()
	mage.registerImprovedFrostNova()
	// mage.registerPermafrost()

	mage.registerPiercingIce()
	// mage.registerImprovedBlizzard()

	// mage.registerArcticReach()
	mage.registerFrostChanneling()
	// mage.registerShatter()

	mage.registerImprovedConeOfCold()

	mage.registerIceFloes()
	mage.registerWinterChill()

	mage.registerArcticWinds()

	mage.registerEmpoweredFrostbolt()
}

func (mage *Mage) registerArcaneSubtlety() {
	if mage.Talents.ArcaneSubtlety == 0 {
		return
	}

	//all spells resist 5 & arcance spells threat 20% per rank
	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolArcane,
		FloatValue: -.20 * float64(mage.Talents.ArcaneSubtlety),
		Kind:       core.SpellMod_ThreatMultiplier_Flat,
	})
}

func (mage *Mage) registerArcaneFocus() {
	if mage.Talents.ArcaneFocus == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolArcane,
		FloatValue: 2 * float64(mage.Talents.ArcaneFocus),
		Kind:       core.SpellMod_BonusHit_Percent,
	})
}

func (mage *Mage) registerArcaneConcentration() {
	if mage.Talents.ArcaneConcentration == 0 {
		return
	}

	bonusCrit := float64(mage.Talents.ArcanePotency) * 10 * core.SpellCritRatingPerCritPercent
	var proccedAt time.Duration
	var proccedSpell *core.Spell

	mage.ClearCasting = mage.RegisterAura(core.Aura{
		Label:    "Clearcasting",
		ActionID: core.ActionID{SpellID: 12536},
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			mage.AddStatDynamic(sim, stats.SpellCritRating, bonusCrit)
			aura.Unit.PseudoStats.SpellCostPercentModifier -= 100
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			mage.AddStatDynamic(sim, stats.SpellCritRating, -bonusCrit)
			aura.Unit.PseudoStats.SpellCostPercentModifier += 100
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.ClassSpellMask&MageSpellsAllDamaging == 0 {
				return
			}

			if spell.DefaultCast.Cost == 0 {
				return
			}

			if proccedAt == sim.CurrentTime && proccedSpell == spell {
				// Means this is another hit from the same cast that procced CC.
				return
			}

			aura.Deactivate(sim)
		},
	})

	mage.RegisterAura(core.Aura{
		Label:    "Arcane Concentration",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ClassSpellMask&MageSpellsAllDamaging == 0 {
				return
			}

			if !result.Landed() {
				return
			}

			procChance := 0.02 * float64(mage.Talents.ArcaneConcentration)
			if sim.Proc(procChance, "Arcane Concentration") {
				proccedAt = sim.CurrentTime
				proccedSpell = spell
				mage.ClearCasting.Activate(sim)
			}
		},
	})
}

func (mage *Mage) registerArcaneImpact() {
	if mage.Talents.ArcaneImpact == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellArcaneBlast | MageSpellArcaneExplosion,
		FloatValue: 2 * float64(mage.Talents.ArcaneImpact),
		Kind:       core.SpellMod_BonusCrit_Percent,
	})
}

func (mage *Mage) registerArcaneMeditation() {
	if mage.Talents.ArcaneMeditation == 0 {
		return
	}

	mage.PseudoStats.SpiritRegenRateCombat += float64(mage.Talents.ArcaneMeditation) * 0.1
	mage.UpdateManaRegenRates()
}

func (mage *Mage) registerArcaneMind() {
	if mage.Talents.ArcaneMind == 0 {
		return
	}

	mage.MultiplyStat(stats.Intellect, 1+(float64(mage.Talents.ArcaneMind)*.03))
}

func (mage *Mage) registerArcaneInstability() {
	if mage.Talents.ArcaneInstability == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll,
		FloatValue: 1 * float64(mage.Talents.ArcaneInstability),
		Kind:       core.SpellMod_BonusCrit_Percent,
	})

	mage.MultiplyStat(stats.SpellDamage, 1+(.01*float64(mage.Talents.ArcaneInstability)))
}

func (mage *Mage) registerEmpoweredArcaneMissiles() {
	if mage.Talents.EmpoweredArcaneMissiles == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellArcaneMissilesTick,
		FloatValue: .15 * float64(mage.Talents.EmpoweredArcaneMissiles),
		Kind:       core.SpellMod_DamageDone_Pct,
	})

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellArcaneMissilesCast,
		FloatValue: .02 * float64(mage.Talents.EmpoweredArcaneMissiles),
		Kind:       core.SpellMod_PowerCost_Pct,
	})
}

func (mage *Mage) registerSpellPower() {
	if mage.Talents.SpellPower == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll,
		FloatValue: .25 * float64(mage.Talents.SpellPower),
		Kind:       core.SpellMod_CritMultiplier_Flat,
	})
}

func (mage *Mage) registerMindMastery() {
	if mage.Talents.MindMastery == 0 {
		return
	}

	mage.AddStatDependency(stats.Intellect, stats.SpellDamage, .05*float64(mage.Talents.MindMastery))
}

// ------ FIRE TALENTS ------

func (mage *Mage) registerImprovedFireball() {
	if mage.Talents.ImprovedFireball == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask: MageSpellFireball,
		TimeValue: time.Millisecond * time.Duration(-100*float64(mage.Talents.ImprovedFireball)),
		Kind:      core.SpellMod_CastTime_Flat,
	})
}

func (mage *Mage) registerIgnite() {
	if mage.Talents.Ignite == 0 {
		return
	}

	mage.Ignite = shared.RegisterIgniteEffect(&mage.Unit, shared.IgniteConfig{
		ActionID:       core.ActionID{SpellID: 12846},
		ClassSpellMask: MageSpellIgnite,
		DotAuraLabel:   "Ignite",
		DotAuraTag:     "IgniteDot",

		ProcTrigger: core.ProcTrigger{
			Name:           "Ignite Talent",
			Callback:       core.CallbackOnSpellHitDealt,
			ProcMask:       core.ProcMaskSpellDamage,
			ClassSpellMask: FireSpellIgnitable,
			Outcome:        core.OutcomeCrit,
		},

		DamageCalculator: func(result *core.SpellResult) float64 {
			return result.Damage * (float64(mage.Talents.Ignite) * .08)
		},
	})

	// This is needed because we want to listen for the spell "cast" event that refreshes the Dot
	mage.Ignite.Flags ^= core.SpellFlagNoOnCastComplete
}

func (mage *Mage) registerImprovedFireBlast() {
	if mage.Talents.ImprovedFireBlast == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFireBlast,
		FloatValue: -.05 * float64(mage.Talents.ImprovedFireBlast),
		Kind:       core.SpellMod_Cooldown_Flat,
	})
}

func (mage *Mage) registerIncineration() {
	if mage.Talents.Incineration == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFireBlast | MageSpellScorch,
		FloatValue: 2 * float64(mage.Talents.Incineration),
		Kind:       core.SpellMod_BonusCrit_Percent,
	})
}

func (mage *Mage) registerImprovedFlamestrike() {
	if mage.Talents.ImprovedFlamestrike == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFlamestrike,
		FloatValue: .05 * float64(mage.Talents.ImprovedFlamestrike),
		Kind:       core.SpellMod_BonusCrit_Percent,
	})
}

func (mage *Mage) registerBurningSoul() {
	if mage.Talents.BurningSoul == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFire,
		FloatValue: -.05 * float64(mage.Talents.BurningSoul),
		Kind:       core.SpellMod_ThreatMultiplier_Flat,
	})
}

func (mage *Mage) registerMasterOfElements() {
	if mage.Talents.MasterOfElements == 0 {
		return
	}

	refundCoeff := 0.1 * float64(mage.Talents.MasterOfElements)
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: 29076})

	mage.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Master of Elements",
		Duration:       core.NeverExpires,
		ClassSpellMask: MageSpellFire | MageSpellFrost,
		Outcome:        core.OutcomeCrit,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.CurCast.Cost == 0 {
				return
			}
			mage.AddMana(sim, spell.DefaultCast.Cost*refundCoeff, manaMetrics)
		},
	})
}

func (mage *Mage) registerPlayingWithFire() {
	if mage.Talents.PlayingWithFire == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ProcMask:   core.ProcMaskSpellDamage,
		FloatValue: .01 * float64(mage.Talents.PlayingWithFire),
		Kind:       core.SpellMod_DamageDone_Pct,
	})
}

func (mage *Mage) registerCriticalMass() {
	if mage.Talents.CriticalMass == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		SpellFlag:  core.SpellFlag(core.SpellSchoolFire),
		FloatValue: 2 * float64(mage.Talents.CriticalMass),
		Kind:       core.SpellMod_BonusCrit_Percent,
	})
}

func (mage *Mage) registerFirePower() {
	if mage.Talents.FirePower == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFire,
		FloatValue: .02 * float64(mage.Talents.FirePower),
		Kind:       core.SpellMod_DamageDone_Pct,
	})
}

func (mage *Mage) registerPyromaniac() {
	if mage.Talents.Pyromaniac == 0 {
		return
	}

	percent := 1 * float64(mage.Talents.Pyromaniac)
	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFire,
		FloatValue: percent,
		Kind:       core.SpellMod_BonusCrit_Percent,
	})

	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFire,
		FloatValue: -percent,
		Kind:       core.SpellMod_ThreatMultiplier_Flat,
	})
}

func (mage *Mage) registerMoltenFury() {
	if mage.Talents.MoltenFury == 0 {
		return
	}

	multiplier := .1 * float64(mage.Talents.MoltenFury)
	mage.RegisterResetEffect(func(sim *core.Simulation) {
		sim.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
			if isExecute == 20 {
				mage.PseudoStats.DamageDealtMultiplier *= multiplier
			}
		})
	})
}

func (mage *Mage) registerEmpoweredFireball() {
	if mage.Talents.EmpoweredFireball == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFireball,
		FloatValue: (.03 * float64(mage.Talents.EmpoweredFireball)) * mage.GetStat(stats.FireDamage),
		Kind:       core.SpellMod_BonusSpellDamage_Flat,
	})
}

// ------ FROST TALENTS ------

func (mage *Mage) registerImprovedFrostbolt() {
	if mage.Talents.ImprovedFrostbolt == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask: MageSpellFrostbolt,
		TimeValue: time.Millisecond * time.Duration(-100*float64(mage.Talents.ImprovedFrostbolt)),
		Kind:      core.SpellMod_CastTime_Flat,
	})
}

func (mage *Mage) registerElementalPrecision() {
	if mage.Talents.ElementalPrecision == 0 {
		return
	}

	percent := 1 * float64(mage.Talents.ElementalPrecision)
	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFrostfire,
		FloatValue: -percent / 100,
		Kind:       core.SpellMod_PowerCost_Pct,
	})

	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFrostfire,
		FloatValue: percent,
		Kind:       core.SpellMod_BonusHit_Percent,
	})
}

func (mage *Mage) registerIceShards() {
	if mage.Talents.IceShards == 0 {
		return
	}
	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll,
		FloatValue: .2 * float64(mage.Talents.IceShards),
		Kind:       core.SpellMod_CritMultiplier_Flat,
	})

}

func (mage *Mage) registerImprovedFrostNova() {
	if mage.Talents.ImprovedFrostNova == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask: MageSpellFrostNova,
		TimeValue: time.Second * time.Duration(-2*mage.Talents.ImprovedFrostNova),
		Kind:      core.SpellMod_CastTime_Flat,
	})
}

func (mage *Mage) registerPiercingIce() {
	if mage.Talents.PiercingIce == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFrost,
		FloatValue: .02 * float64(mage.Talents.PiercingIce),
		Kind:       core.SpellMod_DamageDone_Pct,
	})
}

func (mage *Mage) registerFrostChanneling() {
	if mage.Talents.FrostChanneling == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFrost,
		FloatValue: -.05 * float64(mage.Talents.FrostChanneling),
		Kind:       core.SpellMod_PowerCost_Pct,
	})

	threatMod := []float64{.04, .07, .1}
	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFrost,
		FloatValue: -threatMod[mage.Talents.FrostChanneling-1],
		Kind:       core.SpellMod_ThreatMultiplier_Flat,
	})
}

func (mage *Mage) registerImprovedConeOfCold() {
	if mage.Talents.ImprovedConeOfCold == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellConeOfCold,
		FloatValue: .15 + (.10 * (float64(mage.Talents.ImprovedConeOfCold) - 1)),
		Kind:       core.SpellMod_DamageDone_Pct,
	})
}

func (mage *Mage) registerIceFloes() {
	if mage.Talents.IceFloes == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellColdSnap | MageSpellConeOfCold | MageSpellIceBarrier | MageSpellIceBlock,
		FloatValue: 1 - .1*float64(mage.Talents.IceFloes),
		Kind:       core.SpellMod_Cooldown_Multiplier,
	})
}

func (mage *Mage) registerWinterChill() {
	if mage.Talents.WintersChill == 0 {
		return
	}

	procChance := []float64{0, 0.33, 0.66, 1}[mage.Talents.WintersChill]

	wcAuras := mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.WintersChillAura(target, 0)
	})

	mage.Env.RegisterPreFinalizeEffect(func() {
		for _, spell := range mage.GetSpellsMatchingSchool(core.SpellSchoolFrost) {
			spell.RelatedAuraArrays.Append(wcAuras)
		}
	})

	mage.RegisterAura(core.Aura{
		Label:    "Winters Chill Talent",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.SpellSchool.Matches(core.SpellSchoolFrost) {
				return
			}

			if sim.Proc(procChance, "Winters Chill") {
				aura := wcAuras.Get(result.Target)
				aura.Activate(sim)
				if aura.IsActive() {
					aura.AddStack(sim)
				}
			}
		},
	})
}

func (mage *Mage) registerArcticWinds() {
	if mage.Talents.ArcticWinds == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFrost,
		FloatValue: .01 * float64(mage.Talents.ArcticWinds),
		Kind:       core.SpellMod_DamageDone_Pct,
	})
}

func (mage *Mage) registerEmpoweredFrostbolt() {
	if mage.Talents.EmpoweredFrostbolt == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFrostbolt,
		FloatValue: (.02 * float64(mage.Talents.EmpoweredFrostbolt)) * mage.GetStat(stats.FrostDamage),
		Kind:       core.SpellMod_BonusSpellDamage_Flat,
	})

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFrostbolt,
		FloatValue: .01 * float64(mage.Talents.EmpoweredFrostbolt),
		Kind:       core.SpellMod_BonusCrit_Percent,
	})
}
