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
	// warlock.applyShadowEmbrace()
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
	warlock.applyDevastation()
	warlock.applyImprovedFirebolt()
	warlock.applyImprovedLashOfPain()
	warlock.applyDestructiveReach()
	warlock.applyImprovedSearingPain()
	warlock.applyRuin()
	warlock.applyEmberstorm()
	warlock.applyBacklash()
	warlock.registerConflagrate()
	warlock.applyShadowAndFlame()
}

/*
Affliction
Skipping the following (for now)
- Soul Siphon
- Improved Life Tap -> included in lifetap.go
- Empowered Corruption -> included in corruption.go
- Siphon Life -> implemented in siphon_life.go
- Fel Concentration
- Grim Reach
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
		FloatValue: 0.02 * float64(warlock.Talents.Suppression),
		ClassMask:  WarlockAfflictionSpells,
	})
}

func (warlock *Warlock) applyImprovedCorruption() {
	if warlock.Talents.Suppression == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Millisecond * 400 * time.Duration(warlock.Talents.ImprovedCorruption),
		ClassMask: WarlockAfflictionSpells,
	})
}

func (warlock *Warlock) registerAmplifyCurse() {
	if !warlock.Talents.AmplifyCurse {
		return
	}

	actionID := core.ActionID{SpellID: 18288}
	warlock.AmplifyCurseAura = warlock.RegisterAura(core.Aura{
		Label:    "Amplify Curse",
		ActionID: actionID,
		Duration: time.Second * 30,
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
		FloatValue: 1 * (0.05 * float64(warlock.Talents.ImprovedCurseOfAgony)),
		ClassMask:  WarlockSpellCurseOfAgony,
	})
}

func (warlock *Warlock) applyNightfall() {
	if warlock.Talents.Nightfall == 0 {
		return
	}

	warlock.NightfallProcAura = warlock.RegisterAura(core.Aura{
		Label:    "Nightfall Shadow Trance",
		ActionID: core.ActionID{SpellID: 17941},
		Duration: time.Second * 10,
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// Check for an instant cast shadowbolt to disable aura
			if spell != warlock.ShadowBolt || spell.CurCast.CastTime != 0 {
				return
			}
			aura.Deactivate(sim)
		},
	})

	warlock.RegisterAura(core.Aura{
		Label:    "Nightfall",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnPeriodicDamageDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, spellEffect *core.SpellResult) {
			if spell != warlock.Corruption && spell != warlock.DrainLife {
				return
			}
			if sim.RandomFloat("nightfall") > 0.04 {
				return
			}
			warlock.NightfallProcAura.Activate(sim)
		},
	})

}

func (warlock *Warlock) applyEmpoweredCorruption() {
	if warlock.Talents.ImprovedCorruption == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCoeffecient_Flat,
		FloatValue: ((0.12 * float64(warlock.Talents.EmpoweredCorruption)) / 6),
		ClassMask:  WarlockSpellCorruption,
	})
}

// func (warlock *Warlock) applyShadowEmbrace() {
// 	if warlock.Talents.ShadowEmbrace == 0 {
// 		return
// 	}

// 	warlock.RegisterAura(core.Aura{
// 		Label:    "Shadow Embrace Talent",
// 		Duration: core.NeverExpires,
// 		OnReset: func(aura *core.Aura, sim *core.Simulation) {
// 			aura.Activate(sim)
// 		},
// 		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, spellEffect *core.SpellResult) {
// 			if !spellEffect.Landed() {
// 				return
// 			}

// 			if spell == warlock.Corruption || spell == warlock.SiphonLife || spell == warlock.CurseOfAgony || spell.SameAction(warlock.Seed.ActionID) {
// 				core.ShadowEmbrace(spellEffect.Target, warlock.Talents.ShadowEmbrace, spell.Dot(spellEffect.Target).Duration).Activate(sim)
// 			}
// 		},
// 	})
// }

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
Skipping so many for now
*/
func (warlock *Warlock) appyImprovedImp() {
	if warlock.Talents.ImprovedImp == 0 {
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
	warlock.ActivePet.MultiplyStat(stats.Mana, 1+(0.05)*float64(warlock.Talents.FelIntellect))
}

func (warlock *Warlock) applyImprovedSayaad() {
	if warlock.Talents.ImprovedSayaad == 0 {
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
	warlock.ActivePet.MultiplyStat(stats.Health, 1+(0.05)*float64(warlock.Talents.FelStamina))

}

func (warlock *Warlock) applyUnholyPower() {
	if warlock.Talents.UnholyPower == 0 {
		return
	}

	warlock.ActivePet.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= 1.0 + 0.04*float64(warlock.Talents.UnholyPower)
	warlock.Imp.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.04 * float64(warlock.Talents.UnholyPower),
		ClassMask:  WarlockSpellImpFireBolt,
	})
}

func (warlock *Warlock) applyDemonicSacrifice() {
	if !warlock.Talents.DemonicSacrifice {
		return
	}

	switch warlock.Options.Summon {
	case proto.WarlockOptions_Succubus:
		warlock.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] *= 1.15
	case proto.WarlockOptions_Imp:
		warlock.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexFire] *= 1.15
	case proto.WarlockOptions_Felguard:
		warlock.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] *= 1.10
		warlock.AddStat(stats.MP5, warlock.GetStats()[stats.Intellect]*1.25)
	case proto.WarlockOptions_Felhunter:
		warlock.AddStat(stats.MP5, warlock.GetStats()[stats.Intellect]*1.6667)
	}
}

func (warlock *Warlock) applyMasterDemonologist() {
	if warlock.Talents.MasterDemonologist == 0 {
		return
	}

	switch warlock.Options.Summon {
	case proto.WarlockOptions_Imp:
		warlock.PseudoStats.ThreatMultiplier *= 1.0 - (0.04 * float64(warlock.Talents.MasterDemonologist))
	case proto.WarlockOptions_Succubus:
		warlock.PseudoStats.DamageDealtMultiplier *= 1.0 + 0.02*float64(warlock.Talents.MasterDemonologist)
	case proto.WarlockOptions_Felguard:
		warlock.PseudoStats.DamageDealtMultiplier *= 1.0 + 0.01*float64(warlock.Talents.MasterDemonologist)
	case proto.WarlockOptions_Voidwalker:
		warlock.PseudoStats.BonusPhysicalDamageTaken *= 1.0 - (0.02 * float64(warlock.Talents.MasterDemonologist))
	}
}

func (warlock *Warlock) applyDemonicKnowledge() {
	if warlock.Talents.DemonicKnowledge == 0 {
		return
	}

	petStats := warlock.ActivePet.GetStats()
	warlock.AddStat(stats.SpellDamage, (0.04*float64(warlock.Talents.DemonicKnowledge))*(petStats[stats.Stamina]+petStats[stats.Intellect]))
}

func (warlock *Warlock) applySoulLink() {
	if !warlock.Talents.SoulLink {
		return
	}

	// Add if/while pet is alive
	warlock.PseudoStats.DamageTakenMultiplier *= 0.80
	warlock.PseudoStats.DamageDealtMultiplier *= 1.05
}

func (warlock *Warlock) applyDemonicTactics() {
	if warlock.Talents.DemonicTactics == 0 {
		return
	}

	warlock.AddStat(stats.SpellCritPercent, 0.01*float64(warlock.Talents.DemonicTactics))
}

/*
Destruction
Skip for now:
 - Improved shadowbolt - included in shadowbolt.go
 - ImprovedImmolate - include in immolate.go
*/

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
	if warlock.Talents.Cataclysm == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: -(time.Millisecond * 100) * time.Duration(warlock.Talents.Bane),
		ClassMask: WarlockSpellShadowBolt | WarlockSpellImmolate,
	})

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: -(time.Millisecond * 400) * time.Duration(warlock.Talents.Bane),
		ClassMask: WarlockSpellSoulFire,
	})
}

func (warlock *Warlock) applyDevastation() {
	if warlock.Talents.Devastation == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: 5.0,
		ClassMask:  WarlockDestructionSpells,
	})
}

func (warlock *Warlock) applyImprovedFirebolt() {
	if warlock.Talents.ImprovedFirebolt == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: -(time.Millisecond * 250) * time.Duration(warlock.Talents.ImprovedFirebolt),
		ClassMask: WarlockSpellImpFireBolt,
	})
}

func (warlock *Warlock) applyImprovedLashOfPain() {
	if warlock.Talents.ImprovedLashOfPain == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_Cooldown_Flat,
		TimeValue: -(time.Second * 3) * time.Duration(warlock.Talents.ImprovedFirebolt),
		ClassMask: WarlockSpellSuccubusLashOfPain,
	})
}

func (warlock *Warlock) applyDestructiveReach() {
	if warlock.Talents.DestructiveReach == 0 {
		return
	}

	warlock.PseudoStats.ThreatMultiplier *= 1.0 - (0.5 * float64(warlock.Talents.DestructiveReach))
}

func (warlock *Warlock) applyImprovedSearingPain() {
	if warlock.Talents.ImprovedSearingPain == 0 {
		return
	}
	var critBonus = 0
	switch warlock.Talents.ImprovedSearingPain {
	case 1:
		critBonus = 4
	case 2:
		critBonus = 7
	case 10:
		critBonus = 10
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: float64(critBonus),
		ClassMask:  WarlockSpellSearingPain,
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
		FloatValue: 0.01 * float64(warlock.Talents.Emberstorm),
		ClassMask:  WarlockFireDamage,
	})

	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_CastTime_Pct,
		FloatValue: 0.02 * float64(warlock.Talents.Emberstorm),
		ClassMask:  WarlockSpellImmolate,
	})
}

func (warlock *Warlock) applyBacklash() {
	if warlock.Talents.Backlash == 0 {
		return
	}

	warlock.AddStat(stats.SpellCritPercent, float64(warlock.Talents.Backlash))
}

// ToDo
func (warlock *Warlock) applySoulLeech() {
	if warlock.Talents.SoulLeech == 0 {
		return
	}

	warlock.AddStaticMod(core.SpellModConfig{
		Kind: core.SpellMod_Custom,
		ApplyCustom: func(mod *core.SpellMod, spell *core.Spell) {

		},
		RemoveCustom: func(mod *core.SpellMod, spell *core.Spell) {

		},
		ClassMask: WarlockSpellShadowBolt | WarlockSpellShadowBurn | WarlockSpellSoulFire |
			WarlockSpellIncinerate | WarlockSpellSearingPain | WarlockSpellConflagrate,
		FloatValue: 2.0,
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
