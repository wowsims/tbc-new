package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (warlock *Warlock) applyAfflictionTalents() {
	warlock.applySuppression()
	warlock.applyImprovedCorruption()
	warlock.registerAmplifyCurse()
	warlock.applyImprovedCurseOfAgony()
	warlock.applyNightfall()
	warlock.applyEmpoweredCorruption()
	warlock.applyShadowEmbrace()
	warlock.applyShadowMastery()
	warlock.applyContagion()
	warlock.applyUnstableAffliction()
}

func (warlock *Warlock) applyDemonologyTalents() {
	warlock.appyImprovedImp()
	warlock.applyDemonicEmbrace()
	warlock.applyFelIntellect()
	warlock.applyFelStamina()
	warlock.applyImprovedSayaad()
	warlock.applyUnholyPower()
	warlock.applyDemonicSacrifice()
	warlock.applyMasterDemonologist()
	warlock.applySoulLink()
	warlock.applyDemonicKnowledge()
	warlock.applyDemonicTactics()

}

func (warlock *Warlock) applyDestructionTalents() {
	warlock.applyCataclysm()
	warlock.applyBane()
	warlock.applyImprovedFirebolt()
	warlock.applyImprovedLashOfPain()
	warlock.applyDevastation()
	warlock.applyShadowburn()
	warlock.applyImprovedShadowBolt()
	warlock.applyDestructiveReach()
	warlock.applyImprovedSearingPain()
	warlock.applyImprovedImmolate()
	warlock.applyRuin()
	warlock.applyEmberstorm()
	warlock.applyBacklash()
	warlock.applyConflagrate()
	warlock.applySoulLeech()
	warlock.applyShadowAndFlame()
	warlock.applyShadowfury()
}

/*
Affliction
Skipping the following (for now)
- Soul Siphon -> included in drain_life.go
- Improved Life Tap -> included in lifetap.go
- Empowered Corruption -> included in corruption.go
- Siphon Life -> implemented in siphon_life.go
- Fel Concentration
- Grim Reach
- Shadow Embrace -> implemented in corruption.go, curseOfAgony.go, siphon_life.go, and seed.go
- Curse of Weakness
- Curse of Exhaustion
- Dark Pact
- Improved Howl of Terror
*/
func (warlock *Warlock) applySuppression() {
	if warlock.Talents.Suppression == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusHit_Percent,
		FloatValue: 2.0 * float64(warlock.Talents.Suppression),
		ClassMask:  WarlockAfflictionSpells,
	})
}

func (warlock *Warlock) applyImprovedCorruption() {
	if warlock.Talents.Suppression == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Millisecond * -400 * time.Duration(warlock.Talents.ImprovedCorruption),
		ClassMask: WarlockAfflictionSpells,
	})
}

func (warlock *Warlock) registerAmplifyCurse() {
	if !warlock.Talents.AmplifyCurse {
		return
	}

	actionID := core.ActionID{SpellID: 18288}
	warlock.AmplifyCurseAura = warlock.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Amplify Curse Aura",
		ActionID:       actionID,
		ClassSpellMask: WarlockSpellCurseOfAgony,
		Callback:       core.CallbackOnApplyEffects,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			spell.DamageMultiplier *= 1.5
		},
	})

	warlock.AmplifyCurse = warlock.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: time.Minute * 3,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			warlock.AmplifyCurseAura.Activate(sim)
		},
		RelatedSelfBuff: warlock.AmplifyCurseAura,
	})

	warlock.AddMajorCooldown(core.MajorCooldown{
		Spell: warlock.AmplifyCurse,
		Type:  core.CooldownTypeDPS,
	})
}

func (warlock *Warlock) applyImprovedCurseOfAgony() {
	if warlock.Talents.ImprovedCurseOfAgony == 0 {
		return
	}

	//This is a flat X% dot dmg buff, technically incorrect, fix later
	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DotDamageDone_Pct,
		FloatValue: 0.05 * float64(warlock.Talents.ImprovedCurseOfAgony),
		ClassMask:  WarlockSpellCurseOfAgony,
	})
}

func (warlock *Warlock) applyNightfall() {
	if warlock.Talents.Nightfall == 0 {
		return
	}

	isbMod := warlock.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_CastTime_Pct,
		FloatValue: 1.0,
		ClassMask:  WarlockSpellShadowBolt,
	})

	warlock.NightfallProcAura = warlock.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Nightfall",
		ClassSpellMask: WarlockSpellCorruption | WarlockSpellDrainLife,
		Callback:       core.CallbackOnPeriodicDamageDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !sim.Proc(0.02*float64(warlock.Talents.Nightfall), "nightfall") {
				return
			}
			warlock.NightfallProcAura.Activate(sim)
			isbMod.Activate()
		},
	})

	warlock.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Nightfall Shadow Trance",
		ClassSpellMask: WarlockSpellShadowBolt,
		Callback:       core.CallbackOnCastComplete,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.CurCast.CastTime != 0 {
				return
			}

			isbMod.Deactivate()
			warlock.NightfallProcAura.Deactivate(sim)
		},
	})

}

func (warlock *Warlock) applyEmpoweredCorruption() {
	if warlock.Talents.ImprovedCorruption == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DotBonusCoeffecient_Flat,
		FloatValue: (0.12 * float64(warlock.Talents.EmpoweredCorruption)) / 6,
		ClassMask:  WarlockSpellCorruption,
	})
}

func (warlock *Warlock) applyShadowEmbrace() {
	if warlock.Talents.ShadowEmbrace == 0 {
		return
	}

	warlock.ShadowEmbraceAura = core.ShadowEmbraceAura(warlock.CurrentTarget, warlock.Talents.ShadowEmbrace)
	warlock.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Shadow Embrace Trigger" + warlock.Label,
		Callback:       core.CallbackOnSpellHitDealt,
		Outcome:        core.OutcomeLanded,
		ProcChance:     1,
		ClassSpellMask: WarlockShadowEmbraceSpells,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			warlock.ShadowEmbraceAura.Activate(sim)
		},
	})

}

func (warlock *Warlock) applyShadowMastery() {
	if warlock.Talents.ShadowMastery == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.02 * float64(warlock.Talents.ShadowMastery),
		ClassMask:  WarlockShadowDamage,
	})
}

func (warlock *Warlock) applyContagion() {
	if warlock.Talents.Contagion == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.01 * float64(warlock.Talents.Contagion),
		ClassMask:  WarlockContagionSpells,
	})
}

func (warlock *Warlock) applyUnstableAffliction() {
	if warlock.Talents.UnstableAffliction {
		warlock.registerUnstableAffliction()
	}
}

/*
Demonology
Skipping the following:
  - Improved Healthstone
  - Improved Health Funnel
  - Improved Voidwalker
  - Fel Domination
  - Demonic Aegis -> implemented in armors.go
  - Mana Feed -> applied in lifetap.go
*/
func (warlock *Warlock) appyImprovedImp() {
	if warlock.Talents.ImprovedImp == 0 || warlock.Options.SacrificeSummon {
		return
	}

	warlock.Imp.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.1 * float64(warlock.Talents.ImprovedImp),
		ClassMask:  WarlockSpellImpFireBolt,
	})
}

func (warlock *Warlock) applyDemonicEmbrace() {
	if warlock.Talents.DemonicEmbrace == 0 {
		return
	}

	warlock.MultiplyStat(stats.Stamina, 1.0+(0.03)*float64(warlock.Talents.DemonicEmbrace))
	warlock.MultiplyStat(stats.Spirit, 1.0-(0.03)*float64(warlock.Talents.DemonicEmbrace))
}

func (warlock *Warlock) applyFelIntellect() {
	if warlock.Talents.FelIntellect == 0 {
		return
	}

	warlock.MultiplyStat(stats.Mana, 1.0+(0.01)*float64(warlock.Talents.FelIntellect))
	for _, pet := range warlock.Pets {
		pet.MultiplyStat(stats.Mana, 1+(0.05)*float64(warlock.Talents.FelIntellect))
	}

}

func (warlock *Warlock) applyImprovedSayaad() {
	if warlock.Talents.ImprovedSayaad == 0 || !warlock.Options.SacrificeSummon {
		return
	}

	//This might not actually increase the damage, find a source to prove this
	warlock.Succubus.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.1 * float64(warlock.Talents.ImprovedSayaad),
		ClassMask:  WarlockSpellSuccubusLashOfPain,
	})
}

func (warlock *Warlock) applyFelStamina() {
	if warlock.Talents.FelStamina == 0 {
		return
	}

	warlock.MultiplyStat(stats.Health, 1.0+0.01*float64(warlock.Talents.FelStamina))
	for _, pet := range warlock.Pets {
		pet.MultiplyStat(stats.Health, 1+(0.05)*float64(warlock.Talents.FelStamina))
	}

}

func (warlock *Warlock) applyUnholyPower() {
	if warlock.Talents.UnholyPower == 0 || warlock.Options.SacrificeSummon {
		return
	}

	for _, pet := range warlock.Pets {
		if pet != &warlock.Imp.Pet {
			pet.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= 1.0 + 0.04*float64(warlock.Talents.UnholyPower)
		}
	}

	warlock.Imp.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.04 * float64(warlock.Talents.UnholyPower),
		ClassMask:  WarlockSpellImpFireBolt,
	})
}

func (warlock *Warlock) applyDemonicSacrifice() {
	if !warlock.Talents.DemonicSacrifice || warlock.Options.SacrificeSummon == false {
		return
	}

	switch warlock.Options.Summon {
	case proto.WarlockOptions_Succubus:
		warlock.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] *= 1.15
	case proto.WarlockOptions_Imp:
		warlock.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexFire] *= 1.15
	case proto.WarlockOptions_Felguard:
		warlock.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] *= 1.10
		warlock.AddStat(stats.MP5, warlock.GetStat(stats.Intellect)*1.25)
	case proto.WarlockOptions_Felhunter:
		warlock.AddStat(stats.MP5, warlock.GetStat(stats.Intellect)*1.6667)
	}
}

func (warlock *Warlock) applyMasterDemonologist() {
	if warlock.Talents.MasterDemonologist == 0 || warlock.Options.SacrificeSummon == true {
		return
	}
	points := float64(warlock.Talents.MasterDemonologist)

	switch warlock.Options.Summon {
	case proto.WarlockOptions_Imp:
		warlock.MasterDemonologistAura = warlock.NewTemporaryStatsAura("Master Demonologist", core.ActionID{SpellID: (23825 + int32(points))}, stats.Stats{}, core.NeverExpires).Aura
		warlock.MasterDemonologistAura.AttachMultiplicativePseudoStatBuff(&warlock.PseudoStats.ThreatMultiplier, 1.0-0.04*points)
	case proto.WarlockOptions_Succubus:
		warlock.MasterDemonologistAura = warlock.NewTemporaryStatsAura("Master Demonologist", core.ActionID{SpellID: (23832 + int32(points))}, stats.Stats{}, core.NeverExpires).Aura
		warlock.MasterDemonologistAura.AttachMultiplicativePseudoStatBuff(&warlock.PseudoStats.DamageDealtMultiplier, 1.0+0.02*points)
	case proto.WarlockOptions_Felguard:
		resistsBonus := 0.10 * points * 70
		warlock.MasterDemonologistAura = warlock.NewTemporaryStatsAura("Master Demonologist", core.ActionID{SpellID: (35701 + int32(points))}, stats.Stats{
			stats.ArcaneResistance: resistsBonus,
			stats.FireResistance:   resistsBonus,
			stats.FrostResistance:  resistsBonus,
			stats.NatureResistance: resistsBonus,
			stats.ShadowResistance: resistsBonus,
		}, core.NeverExpires).Aura
		warlock.MasterDemonologistAura.AttachMultiplicativePseudoStatBuff(&warlock.PseudoStats.DamageDealtMultiplier, 1.0+0.01*points)
	case proto.WarlockOptions_Voidwalker:
		warlock.PseudoStats.BonusPhysicalDamageTaken *= 1.0 - 0.02*points
		warlock.MasterDemonologistAura = warlock.NewTemporaryStatsAura("Master Demonologist", core.ActionID{SpellID: (23840 + int32(points))}, stats.Stats{}, core.NeverExpires).Aura
		warlock.MasterDemonologistAura.AttachMultiplicativePseudoStatBuff(&warlock.PseudoStats.BonusPhysicalDamageTaken, 1.0-0.02*points)
	case proto.WarlockOptions_Felhunter:
		resistsBonus := 0.20 * points * 70
		warlock.MasterDemonologistAura = warlock.NewTemporaryStatsAura("Master Demonologist", core.ActionID{SpellID: (23836 + int32(points))}, stats.Stats{}, core.NeverExpires).Aura
		warlock.MasterDemonologistAura.AttachStatsBuff(stats.Stats{
			stats.ArcaneResistance: resistsBonus,
			stats.FireResistance:   resistsBonus,
			stats.FrostResistance:  resistsBonus,
			stats.NatureResistance: resistsBonus,
			stats.ShadowResistance: resistsBonus,
		})
	}
}

func (warlock *Warlock) applySoulLink() {
	if !warlock.Talents.SoulLink {
		return
	}

	// Add if/while pet is alive
	warlock.PseudoStats.DamageTakenMultiplier *= 0.80
	warlock.PseudoStats.DamageDealtMultiplier *= 1.05

	for _, pet := range warlock.Pets {
		pet.PseudoStats.DamageDealtMultiplier *= 1.05
	}
}

func (warlock *Warlock) applyDemonicKnowledge() {
	if warlock.Talents.DemonicKnowledge == 0 {
		return
	}
	bonus := (0.04 * float64(warlock.Talents.DemonicKnowledge)) * (warlock.ActivePet.GetStat(stats.Stamina) + warlock.ActivePet.GetStat(stats.Intellect))
	warlock.DemonicKnowledgeAura = warlock.NewTemporaryStatsAura("Demonic Knowledge", core.ActionID{SpellID: 35693}, stats.Stats{stats.SpellDamage: bonus}, core.NeverExpires).Aura
}

func (warlock *Warlock) applyDemonicTactics() {
	if warlock.Talents.DemonicTactics == 0 {
		return
	}
	points := float64(warlock.Talents.DemonicTactics)
	warlock.AddStat(stats.SpellCritPercent, points)
	warlock.AddStat(stats.PhysicalCritPercent, points)
	warlock.AddStat(stats.RangedCritPercent, points)

	for _, pet := range warlock.Pets {
		pet.AddStat(stats.SpellCritPercent, points)
		pet.AddStat(stats.PhysicalCritPercent, points)
		pet.AddStat(stats.RangedCritPercent, points)
	}
}

/*
Destruction
Skipped Talents:
  - Aftermath
*/
func (warlock *Warlock) applyImprovedShadowBolt() {
	if warlock.Talents.ImprovedShadowBolt == 0 {
		return
	}
	warlock.ImpShadowboltAura = core.ImprovedShadowBoltAura(warlock.CurrentTarget, 0, warlock.Talents.ImprovedShadowBolt)
}

func (warlock *Warlock) applyCataclysm() {
	if warlock.Talents.Cataclysm == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: 1.0 - 0.01*float64(warlock.Talents.Cataclysm),
		ClassMask:  WarlockDestructionSpells,
	})
}

func (warlock *Warlock) applyBane() {
	if warlock.Talents.Bane == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Millisecond * time.Duration(-100*warlock.Talents.Bane),
		ClassMask: WarlockSpellShadowBolt | WarlockSpellImmolate,
	})

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Millisecond * time.Duration(-400*warlock.Talents.Bane),
		ClassMask: WarlockSpellSoulFire,
	})
}

func (warlock *Warlock) applyImprovedFirebolt() {
	if warlock.Talents.ImprovedFirebolt == 0 {
		return
	}

	warlock.ActivePet.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Millisecond * time.Duration(-250*warlock.Talents.ImprovedFirebolt),
		ClassMask: WarlockSpellImpFireBolt,
	})
}

func (warlock *Warlock) applyImprovedLashOfPain() {
	if warlock.Talents.ImprovedLashOfPain == 0 {
		return
	}

	warlock.ActivePet.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_Cooldown_Flat,
		TimeValue: time.Second * time.Duration(-3*warlock.Talents.ImprovedLashOfPain),
		ClassMask: WarlockSpellSuccubusLashOfPain,
	})
}

func (warlock *Warlock) applyDevastation() {
	if warlock.Talents.Devastation == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: float64(warlock.Talents.Devastation),
		ClassMask:  WarlockDestructionSpells,
	})
}

func (warlock *Warlock) applyShadowburn() {
	if !warlock.Talents.Shadowburn {
		return
	}

	warlock.registerShadowBurn()
}

func (warlock *Warlock) applyDestructiveReach() {
	if warlock.Talents.DestructiveReach == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_ThreatMultiplier_Pct,
		FloatValue: -0.05 * float64(warlock.Talents.DestructiveReach),
		ClassMask:  WarlockDestructionSpells,
	})

}

func (warlock *Warlock) applyImprovedSearingPain() {
	if warlock.Talents.ImprovedSearingPain == 0 {
		return
	}
	critBonus := []float64{0, 4, 7, 10}[warlock.Talents.ImprovedSearingPain]

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: critBonus,
		ClassMask:  WarlockSpellSearingPain,
	})
}

func (warlock *Warlock) applyImprovedImmolate() {
	if warlock.Talents.ImprovedImmolate == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.05 * float64(warlock.Talents.ImprovedImmolate),
		ClassMask:  WarlockSpellImmolate,
	})
}

func (warlock *Warlock) applyRuin() {
	if !warlock.Talents.Ruin {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_CritMultiplier_Flat,
		FloatValue: 1.0,
		ClassMask:  WarlockDestructionSpells,
	})
}

func (warlock *Warlock) applyEmberstorm() {
	if warlock.Talents.Emberstorm == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.02 * float64(warlock.Talents.Emberstorm),
		ClassMask:  WarlockFireDamage,
	})

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_CastTime_Pct,
		FloatValue: -0.02 * float64(warlock.Talents.Emberstorm),
		ClassMask:  WarlockSpellIncinerate,
	})
}

func (warlock *Warlock) applyBacklash() {
	if warlock.Talents.Backlash == 0 {
		return
	}

	warlock.AddStat(stats.SpellCritPercent, float64(warlock.Talents.Backlash))
}

func (warlock *Warlock) applyConflagrate() {
	if !warlock.Talents.Conflagrate {
		return
	}

	warlock.registerConflagrate()
}

func (warlock *Warlock) applySoulLeech() {
	if warlock.Talents.SoulLeech == 0 {
		return
	}
	healthMetric := warlock.NewHealthMetrics(core.ActionID{SpellID: 30296})
	warlock.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Soul Leech",
		ClassSpellMask: WarlockSoulLeechSpells,
		Callback:       core.CallbackOnSpellHitDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !sim.Proc(0.10*float64(warlock.Talents.SoulLeech), "Soul Leech") {
				return
			}

			warlock.GainHealth(sim, result.Damage*0.2*warlock.PseudoStats.SelfHealingMultiplier, healthMetric)
		},
	})
}

func (warlock *Warlock) applyShadowAndFlame() {
	if warlock.Talents.ShadowAndFlame == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCoeffecient_Flat,
		FloatValue: 0.04 * float64(warlock.Talents.ShadowAndFlame),
		ClassMask:  WarlockSpellShadowBolt | WarlockSpellIncinerate,
	})
}

func (warlock *Warlock) applyShadowfury() {
	if !warlock.Talents.Shadowfury {
		return
	}

	warlock.registerShadowfury()
}
