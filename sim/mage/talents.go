package mage

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"math"
	"time"

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
		Kind:       core.SpellMod_ThreatMultiplier_Pct,
	})
}

func (mage *Mage) registerArcaneFocus() {
	if mage.Talents.ArcaneFocus == 0 {
		return
	}

	mage.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexArcane] += genRanks.ArcaneFocus.ValueAt(mage.Talents.ArcaneFocus)
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
		FloatValue: genRanks.ArcaneImpact.ValueAt(mage.Talents.ArcaneImpact),
		Kind:       core.SpellMod_BonusCrit_Percent,
	})
}

func (mage *Mage) registerArcaneMeditation() {
	if mage.Talents.ArcaneMeditation == 0 {
		return
	}

	mage.PseudoStats.SpiritRegenRateCasting += float64(mage.Talents.ArcaneMeditation) * 0.1
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
		FloatValue: genRanks.ArcaneInstability.Effect(shared.A_MOD_SPELL_CRIT_CHANCE_SCHOOL, 126).ValueAt(mage.Talents.ArcaneInstability),
		Kind:       core.SpellMod_BonusCrit_Percent,
	})

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll,
		FloatValue: genRanks.ArcaneInstability.Effect(shared.A_MOD_DAMAGE_PERCENT_DONE, 126).FractionAt(mage.Talents.ArcaneInstability),
		Kind:       core.SpellMod_DamageDone_Pct,
	})

}

func (mage *Mage) registerEmpoweredArcaneMissiles() {
	if mage.Talents.EmpoweredArcaneMissiles == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellArcaneMissilesTick,
		FloatValue: genRanks.EmpoweredArcaneMissiles.Effect(shared.A_ADD_FLAT_MODIFIER, shared.SPELLMOD_BONUS_MULTIPLIER).FractionAt(mage.Talents.EmpoweredArcaneMissiles),
		Kind:       core.SpellMod_BonusCoeffecient_Flat,
	})

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellArcaneMissilesCast,
		FloatValue: genRanks.EmpoweredArcaneMissiles.Effect(shared.A_ADD_PCT_MODIFIER, shared.SPELLMOD_COST).FractionAt(mage.Talents.EmpoweredArcaneMissiles),
		Kind:       core.SpellMod_PowerCost_Pct_Add,
	})
}

func (mage *Mage) registerSpellPower() {
	if mage.Talents.SpellPower == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll,
		FloatValue: genRanks.SpellPower.FractionAt(mage.Talents.SpellPower),
		Kind:       core.SpellMod_CritMultiplier_Flat,
	})
}

func (mage *Mage) registerMindMastery() {
	if mage.Talents.MindMastery == 0 {
		return
	}

	mage.AddStatDependency(stats.Intellect, stats.SpellDamage, genRanks.MindMastery.FractionAt(mage.Talents.MindMastery))
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
	igniteDamageMultiplier := float64(mage.Talents.Ignite) * .08

	igniteSpell := mage.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 12846},
		SpellSchool:      core.SpellSchoolFire,
		ProcMask:         core.ProcMaskSpellDamage,
		ClassSpellMask:   MageSpellIgnite,
		Flags:            core.SpellFlagIgnoreModifiers | core.SpellFlagNoSpellMods | core.SpellFlagNoOnCastComplete | core.SpellFlagIgnoreResists | core.SpellFlagProc,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:     "Ignite",
				Tag:       "IgniteDot",
				MaxStacks: math.MaxInt32,
			},
			NumberOfTicks: 2,
			TickLength:    2 * time.Second,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.Spell.CalcAndDealPeriodicDamage(sim, target, dot.SnapshotBaseDamage, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Dot(target).Apply(sim)
		},
	})

	refreshIgnite := func(sim *core.Simulation, target *core.Unit, damagePerTick float64) {
		dot := igniteSpell.Dot(target)
		igniteSpell.Cast(sim, target)
		dot.SnapshotBaseDamage = damagePerTick
		dot.Aura.SetStacks(sim, int32(dot.SnapshotBaseDamage))
	}

	procTrigger := core.ProcTrigger{
		Name:               "Ignite Talent",
		CanProcFromProcs:   true, // 11119, 11120, 12846-12848 carry the bit.
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskSpellDamage,
		ClassSpellMask:     FireSpellIgnitable,
		Outcome:            core.OutcomeCrit,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			target := result.Target
			dot := igniteSpell.Dot(target)
			outstandingDamage := dot.OutstandingDmg()
			if dot.RemainingTicks() <= 0 {
				outstandingDamage = 0
			}

			newDamage := result.Damage * igniteDamageMultiplier
			totalDamage := outstandingDamage + newDamage
			damagePerTick := totalDamage / float64(dot.BaseTickCount)

			refreshIgnite(sim, target, damagePerTick)
		},
	}
	igniteSpell.Unit.MakeProcTriggerAura(procTrigger)
	mage.Ignite = igniteSpell

	// This is needed because we want to listen for the spell "cast" event that refreshes the Dot
	mage.Ignite.Flags ^= core.SpellFlagNoOnCastComplete
}

func (mage *Mage) registerImprovedFireBlast() {
	if mage.Talents.ImprovedFireBlast == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask: MageSpellFireBlast,
		TimeValue: time.Millisecond * time.Duration(-500*float64(mage.Talents.ImprovedFireBlast)),
		Kind:      core.SpellMod_Cooldown_Flat,
	})
}

func (mage *Mage) registerIncineration() {
	if mage.Talents.Incineration == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFireBlast | MageSpellScorch,
		FloatValue: genRanks.Incineration.ValueAt(mage.Talents.Incineration),
		Kind:       core.SpellMod_BonusCrit_Percent,
	})
}

func (mage *Mage) registerImprovedFlamestrike() {
	if mage.Talents.ImprovedFlamestrike == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFlamestrike,
		FloatValue: genRanks.ImprovedFlamestrike.FractionAt(mage.Talents.ImprovedFlamestrike),
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
		Kind:       core.SpellMod_ThreatMultiplier_Pct,
	})
}

func (mage *Mage) registerMasterOfElements() {
	if mage.Talents.MasterOfElements == 0 {
		return
	}

	refundCoeff := genRanks.MasterOfElements.FractionAt(mage.Talents.MasterOfElements)
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
		ClassMask:  MageSpellsAll,
		FloatValue: genRanks.PlayingWithFire.Effect(shared.A_MOD_DAMAGE_PERCENT_DONE, 126).FractionAt(mage.Talents.PlayingWithFire),
		Kind:       core.SpellMod_DamageDone_Pct,
	})
}

func (mage *Mage) registerCriticalMass() {
	if mage.Talents.CriticalMass == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		SpellFlag:  core.SpellFlag(core.SpellSchoolFire),
		FloatValue: genRanks.CriticalMass.ValueAt(mage.Talents.CriticalMass),
		Kind:       core.SpellMod_BonusCrit_Percent,
	})
}

func (mage *Mage) registerFirePower() {
	if mage.Talents.FirePower == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFire,
		FloatValue: genRanks.FirePower.Effect(shared.A_ADD_PCT_MODIFIER, shared.SPELLMOD_DAMAGE).FractionAt(mage.Talents.FirePower),
		Kind:       core.SpellMod_DamageDone_Flat,
	})
}

func (mage *Mage) registerPyromaniac() {
	if mage.Talents.Pyromaniac == 0 {
		return
	}

	points := float64(mage.Talents.Pyromaniac)

	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFire,
		FloatValue: 1 * points,
		Kind:       core.SpellMod_BonusCrit_Percent,
	})

	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFire,
		FloatValue: -0.01 * points,
		Kind:       core.SpellMod_PowerCost_Pct,
	})
}

func (mage *Mage) registerMoltenFury() {
	if mage.Talents.MoltenFury == 0 {
		return
	}

	moltenFury := mage.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: .1 * float64(mage.Talents.MoltenFury),
		ClassMask:  MageSpellsAll,
	})

	mage.RegisterResetEffect(func(sim *core.Simulation) {
		moltenFury.Deactivate()
		sim.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
			if isExecute == 20 {
				moltenFury.Activate()
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
		FloatValue: genRanks.EmpoweredFireball.FractionAt(mage.Talents.EmpoweredFireball),
		Kind:       core.SpellMod_BonusCoeffecient_Flat,
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
	percent := genRanks.ElementalPrecision.Effect(shared.A_ADD_FLAT_MODIFIER, shared.SPELLMOD_RESIST_MISS_CHANCE).ValueAt(mage.Talents.ElementalPrecision)
	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFrostfire,
		FloatValue: -percent / 100,
		Kind:       core.SpellMod_PowerCost_Pct,
	})

	// Bug: Gives 2% hit per point instead of 1% to frost spells.
	// https://www.warcraftlogs.com/reports/kwd3V8MA9FgrRYhf/?boss=-3&difficulty=0&type=damage-done&source=1&target=2
	mage.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexFrost] += (percent * 2)
	mage.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexFire] += percent
}

func (mage *Mage) registerIceShards() {
	if mage.Talents.IceShards == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFrost,
		FloatValue: genRanks.IceShards.FractionAt(mage.Talents.IceShards),
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
		FloatValue: genRanks.PiercingIce.Effect(shared.A_ADD_PCT_MODIFIER, shared.SPELLMOD_DAMAGE).FractionAt(mage.Talents.PiercingIce),
		Kind:       core.SpellMod_DamageDone_Flat,
	})
}

func (mage *Mage) registerFrostChanneling() {
	if mage.Talents.FrostChanneling == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFrost,
		FloatValue: -.05 * float64(mage.Talents.FrostChanneling),
		Kind:       core.SpellMod_PowerCost_Pct_Add,
	})

	threatMod := []float64{.04, .07, .1}
	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFrost,
		FloatValue: -threatMod[mage.Talents.FrostChanneling-1],
		Kind:       core.SpellMod_ThreatMultiplier_Pct,
	})
}

func (mage *Mage) registerImprovedConeOfCold() {
	if mage.Talents.ImprovedConeOfCold == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellConeOfCold,
		FloatValue: .15 + (.10 * (float64(mage.Talents.ImprovedConeOfCold) - 1)),
		Kind:       core.SpellMod_DamageDone_Flat,
	})
}

func (mage *Mage) registerIceFloes() {
	if mage.Talents.IceFloes == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellColdSnap | MageSpellConeOfCold | MageSpellIceBarrier | MageSpellIceBlock,
		FloatValue: genRanks.IceFloes.MultiplierAt(mage.Talents.IceFloes),
		Kind:       core.SpellMod_Cooldown_Multiplier,
	})
}

func (mage *Mage) registerWinterChill() {
	if mage.Talents.WintersChill == 0 {
		return
	}

	procChance := 0.20 * float64(mage.Talents.WintersChill)

	wcAuras := mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.WintersChillAura(target, 0)
	})

	mage.Env.RegisterPreFinalizeEffect(func() {
		for _, spell := range mage.GetSpellsMatchingSchool(core.SpellSchoolFrost) {
			spell.RelatedAuraArrays.Append(wcAuras)
		}
	})

	mage.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Winters Chill Talent",
		Callback:       core.CallbackOnSpellHitDealt,
		Outcome:        core.OutcomeLanded,
		ClassSpellMask: MageSpellFrost,
		ProcChance:     procChance,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura := wcAuras.Get(result.Target)
			aura.Activate(sim)
			aura.AddStack(sim)
		},
	})
}

func (mage *Mage) registerArcticWinds() {
	if mage.Talents.ArcticWinds == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFrost,
		FloatValue: genRanks.ArcticWinds.Effect(shared.A_MOD_DAMAGE_PERCENT_DONE, 16).FractionAt(mage.Talents.ArcticWinds),
		Kind:       core.SpellMod_DamageDone_Pct,
	})
}

func (mage *Mage) registerEmpoweredFrostbolt() {
	if mage.Talents.EmpoweredFrostbolt == 0 {
		return
	}

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFireball,
		FloatValue: genRanks.EmpoweredFrostbolt.Effect(shared.A_ADD_FLAT_MODIFIER, shared.SPELLMOD_BONUS_MULTIPLIER).FractionAt(mage.Talents.EmpoweredFrostbolt),
		Kind:       core.SpellMod_BonusCoeffecient_Flat,
	})

	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFrostbolt,
		FloatValue: genRanks.EmpoweredFrostbolt.Effect(shared.A_ADD_FLAT_MODIFIER, shared.SPELLMOD_CRITICAL_CHANCE).FractionAt(mage.Talents.EmpoweredFrostbolt),
		Kind:       core.SpellMod_BonusCrit_Percent,
	})
}
