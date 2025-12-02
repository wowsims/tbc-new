package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

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
	if warlock.Talents.AmplifyCurse == false {
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

func (warlock *Warlock) applyNighfall() {
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
		OnPeriodicDamageDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, spellEffect *core.SpellEffect) {
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

func (warlock *Warlock) applyShadowEmbrace() {
	if warlock.Talents.ShadowEmbrace == 0 {
		return
	}

	var debuffAuras []*core.Aura
	for _, target := range warlock.Env.Encounter.Targets {
		debuffAuras = append(debuffAuras, core.ShadowEmbraceAura(&target.Unit, warlock.Talents.ShadowEmbrace))
	}

	warlock.RegisterAura(core.Aura{
		Label:    "Shadow Embrace Talent",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, spellEffect *core.SpellEffect) {
			if !spellEffect.Landed() {
				return
			}

			if spell == warlock.Corruption || spell == warlock.SiphonLife || spell == warlock.CurseOfAgony || spell.SameAction(warlock.Seeds[0].ActionID) {
				debuffAuras[spellEffect.Target.Index].Activate(sim)
			}
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
Skipping so many for now
*/
func (warlock *Warlock) applyDemonicEmbrace() {
	if warlock.Talents.DemonicEmbrace == 0 {
		return
	}

	warlock.AddStatDependency(stats.Stamina, stats.Stamina, (0.03)*float64(warlock.Talents.DemonicEmbrace))
	warlock.AddStatDependency(stats.Spirit, stats.Spirit, (0.03)*float64(warlock.Talents.DemonicEmbrace))
}

// TODO - Add pet part
func (warlock *Warlock) applyFelIntellect() {
	if warlock.Talents.FelIntellect == 0 {
		return
	}

	warlock.AddStatDependency(stats.Intellect, stats.Mana, 15*(0.01)*float64(warlock.Talents.FelIntellect))

}

func (warlock *Warlock) applyFelStamina() {
	if warlock.Talents.FelStamina == 0 {
		return
	}

	warlock.AddStatDependency(stats.Health, stats.Health, 1+0.01*float64(warlock.Talents.FelStamina))
}

// Placeholder for Unholy Power
//func (warlock *Warlock) applyUnholyPower() {}

//Placeholder for DSac
//func (warlock *Warlock) applyDemonicSacrifice(){}

//Placeholder for MasterDemonologist
//func (warlock *Warlock) applyMasterDemonologist(){}

//Placeholder for Demonic Knowledge
//func (warlock *Warlock) applyDemonicKnowledge(){}

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

//TODO - implement the pet talents
// func (warlock *Warlock) applyImprovedFirebolt(){}
// func (warlock *Warlock) applyImprovedLashOfPain(){}

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

func (warlock *Warlock) applySoulLeech() {
	if warlock.Talents.SoulLeech == 0 {
		return
	}

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
