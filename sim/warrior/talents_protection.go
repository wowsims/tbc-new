package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (war *Warrior) registerProtectionTalents() {
	// Tier 1
	// Improved Bloodrage implemented in bloodrage.go
	war.registerTacticalMastery()
	war.registerAnticipation()

	// Tier 2
	war.registerShieldSpecialization()
	war.registerToughness()

	// Tier 3
	war.registerLastStand()
	war.registerImprovedShieldBlock()
	// Improved Revenge not implemented
	war.registerDefiance()

	// Tier 4
	war.registerImprovedSunderArmor()
	// Improved Disarm not implemented
	// Improved Taunt not implemented

	// Tier 5
	war.registerImprovedShieldWall()
	// Concussion Blow not implemented
	// Improved Shield Bash not implemented

	// Tier 6
	war.registerShieldMastery()
	war.registerOneHandedWeaponSpecialization()

	// Tier 7
	war.registerImprovedDefensiveStance()
	war.registerShieldSlam()
	war.registerFocusedRage()

	// Tier 8
	war.registerVitality()

	// Tier 9
	war.registerDevastate()
}

func (war *Warrior) registerTacticalMastery() {
	if war.Talents.TacticalMastery == 0 {
		return
	}

	// Retained rage when swapping stances implemented in stances.go
	war.OnSpellRegistered(func(spell *core.Spell) {
		if !spell.Matches(SpellMaskDefensiveStance) {
			return
		}

		spell.RelatedSelfBuff.
			AttachSpellMod(core.SpellModConfig{
				ClassMask:  SpellMaskMortalStrike | SpellMaskBloodthirst,
				Kind:       core.SpellMod_ThreatMultiplier_Pct,
				FloatValue: 0.21 * float64(war.Talents.TacticalMastery),
			})
	})
}

func (war *Warrior) registerDefiance() {
	if war.Talents.Defiance == 0 {
		return
	}

	war.AddStat(stats.ExpertiseRating, 2*float64(war.Talents.Defiance)*core.ExpertisePerQuarterPercentReduction)

	war.OnSpellRegistered(func(spell *core.Spell) {
		if !spell.Matches(SpellMaskDefensiveStance) {
			return
		}

		spell.RelatedSelfBuff.
			AttachSpellMod(core.SpellModConfig{
				ClassMask:  SpellMaskMortalStrike | SpellMaskBloodthirst,
				Kind:       core.SpellMod_ThreatMultiplier_Pct,
				FloatValue: 0.05 * float64(war.Talents.Defiance),
			})
	})
}

func (war *Warrior) registerAnticipation() {
	if war.Talents.Anticipation == 0 {
		return
	}

	war.AddStat(stats.DefenseRating, 4*float64(war.Talents.Anticipation)*core.DefenseRatingPerDefenseLevel)

	war.OnSpellRegistered(func(spell *core.Spell) {
		if !spell.Matches(SpellMaskDefensiveStance) {
			return
		}

		spell.RelatedSelfBuff.
			AttachSpellMod(core.SpellModConfig{
				ClassMask:  SpellMaskMortalStrike | SpellMaskBloodthirst,
				Kind:       core.SpellMod_ThreatMultiplier_Pct,
				FloatValue: 0.05 * float64(war.Talents.Defiance),
			})
	})
}

func (war *Warrior) registerShieldSpecialization() {
	if war.Talents.ShieldSpecialization == 0 {
		return
	}

	war.AddStat(stats.BlockPercent, 0.01*float64(war.Talents.ShieldSpecialization))

	rageMetrics := war.NewRageMetrics(core.ActionID{SpellID: 23602})

	war.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Shield Specialization",
		ProcChance:         0.2 * float64(war.Talents.ShieldSpecialization),
		TriggerImmediately: true,
		Outcome:            core.OutcomeBlock,
		Callback:           core.CallbackOnSpellHitTaken,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			war.AddRage(sim, 1, rageMetrics)
		},
	})
}

func (war *Warrior) registerToughness() {
	if war.Talents.Toughness == 0 {
		return
	}

	war.MultiplyStat(stats.Armor, 1+0.02*float64(war.Talents.Toughness))
}

func (war *Warrior) registerLastStand() {
	if !war.Talents.LastStand {
		return
	}

	actionID := core.ActionID{SpellID: 12975}
	healthMetrics := war.NewHealthMetrics(actionID)

	var bonusHealth float64
	aura := war.RegisterAura(core.Aura{
		Label:    "Last Stand",
		ActionID: actionID,
		Duration: time.Second * 20,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			bonusHealth = war.MaxHealth() * 0.3
			war.AddStatsDynamic(sim, stats.Stats{stats.Health: bonusHealth})
			war.GainHealth(sim, bonusHealth, healthMetrics)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			war.AddStatsDynamic(sim, stats.Stats{stats.Health: -bonusHealth})
		},
	})

	spell := war.RegisterSpell(core.SpellConfig{
		ActionID: actionID,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Minute * 8,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			aura.Activate(sim)
		},

		RelatedSelfBuff: aura,
	})

	war.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeSurvival,
	})
}

func (war *Warrior) registerImprovedShieldBlock() {
	if !war.Talents.ImprovedShieldBlock {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskShieldBlock,
		Kind:      core.SpellMod_BuffDuration_Flat,
		TimeValue: time.Second * 1,
	})

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskShieldBlock,
		Kind:      core.SpellMod_Custom,
		ApplyCustom: func(mod *core.SpellMod, spell *core.Spell) {
			spell.RelatedSelfBuff.MaxStacks += 1
		},
		RemoveCustom: func(mod *core.SpellMod, spell *core.Spell) {
			spell.RelatedSelfBuff.MaxStacks -= 1
		},
	})
}

func (war *Warrior) registerImprovedSunderArmor() {
	if war.Talents.ImprovedSunderArmor == 0 {
		return
	}

	// Retained rage when swapping stances implemented in stances.go
	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskSunderArmor | SpellMaskDevastate,
		Kind:      core.SpellMod_PowerCost_Flat,
		IntValue:  -war.Talents.TacticalMastery,
	})
}

func (war *Warrior) registerImprovedShieldWall() {
	if war.Talents.ImprovedShieldWall == 0 {
		return
	}

	duration := []time.Duration{0, 3, 5}[war.Talents.ImprovedShieldWall]

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskShieldWall,
		Kind:      core.SpellMod_BuffDuration_Flat,
		TimeValue: time.Second * duration,
	})
}

func (war *Warrior) registerShieldMastery() {
	if war.Talents.ShieldMastery == 0 {
		return
	}

	war.PseudoStats.BlockValueMultiplier *= 1 + 0.1*float64(war.Talents.ShieldMastery)
}

func (war *Warrior) registerOneHandedWeaponSpecialization() {
	if war.Talents.OneHandedWeaponSpecialization == 0 {
		return
	}

	weaponMod := war.AddDynamicMod(core.SpellModConfig{
		School:     core.SpellSchoolPhysical,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: 0.02 * float64(war.Talents.OneHandedWeaponSpecialization),
	})

	hasOneHandEquipped := func() bool {
		mh := war.GetMHWeapon()
		if mh != nil && mh.HandType == proto.HandType_HandTypeOneHand {
			return true
		}
		oh := war.GetOHWeapon()
		if oh != nil && oh.HandType == proto.HandType_HandTypeOneHand {
			return true
		}
		return false
	}

	if hasOneHandEquipped() {
		weaponMod.Activate()
	}

	war.RegisterItemSwapCallback(core.AllMeleeWeaponSlots(), func(sim *core.Simulation, slot proto.ItemSlot) {
		if hasOneHandEquipped() {
			weaponMod.Activate()
		} else {
			weaponMod.Deactivate()
		}
	})
}

func (war *Warrior) registerImprovedDefensiveStance() {
	impDefStanceMultiplier := 1 - 0.02*float64(war.Talents.ImprovedDefensiveStance)

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: SpellMaskDefensiveStance,
		Kind:      core.SpellMod_Custom,
		ApplyCustom: func(mod *core.SpellMod, spell *core.Spell) {
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexArcane] *= impDefStanceMultiplier
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire] *= impDefStanceMultiplier
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFrost] *= impDefStanceMultiplier
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexHoly] *= impDefStanceMultiplier
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexNature] *= impDefStanceMultiplier
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow] *= impDefStanceMultiplier
		},
		RemoveCustom: func(mod *core.SpellMod, spell *core.Spell) {
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexArcane] /= impDefStanceMultiplier
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire] /= impDefStanceMultiplier
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFrost] /= impDefStanceMultiplier
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexHoly] /= impDefStanceMultiplier
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexNature] /= impDefStanceMultiplier
			war.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow] /= impDefStanceMultiplier
		},
	})
}

func (war *Warrior) registerShieldSlam() {
	if !war.Talents.ShieldSlam {
		return
	}

	war.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 23922},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		MaxRange:    core.MaxMeleeRange,

		RageCost: core.RageCostOptions{
			Cost:   20,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Second * 6,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.PseudoStats.CanBlock
		},

		DamageMultiplier: 1,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,
		FlatThreatBonus:  305,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(381, 399) + war.BlockDamageReduction()
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}

func (war *Warrior) registerFocusedRage() {
	if war.Talents.FocusedRage == 0 {
		return
	}

	war.AddStaticMod(core.SpellModConfig{
		ClassMask: WarriorSpellsAll ^ (SpellMaskRampage | SpellMaskDeathWish | SpellMaskBattleShout | SpellMaskCommandingShout),
		Kind:      core.SpellMod_PowerCost_Flat,
		IntValue:  -war.Talents.FocusedRage,
	})
}

func (war *Warrior) registerVitality() {
	if war.Talents.Vitality == 0 {
		return
	}

	war.MultiplyStat(stats.Stamina, 1+0.01*float64(war.Talents.Vitality))
	war.MultiplyStat(stats.Strength, 1+0.02*float64(war.Talents.Vitality))
}

func (war *Warrior) registerDevastate() {
	if !war.Talents.Devastate {
		return
	}

	war.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 20243},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		MaxRange:    core.MaxMeleeRange,

		RageCost: core.RageCostOptions{
			Cost:   15,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Second * 6,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return war.PseudoStats.CanBlock
		},

		DamageMultiplier: 1,
		CritMultiplier:   war.DefaultMeleeCritMultiplier(),
		ThreatMultiplier: 1,
		FlatThreatBonus:  301.5 + 100,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := war.MHWeaponDamage(sim, spell.MeleeAttackPower())*0.5 + spell.BonusDamage()

			sunderStacks := war.SunderArmorAuras.Get(target).GetStacks()
			sunderDamage := core.TernaryFloat64(war.CanApplySunderAura(target), float64(sunderStacks)*war.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower()), 0)
			result := spell.CalcAndDealDamage(sim, target, baseDamage+sunderDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if result.Landed() {
				war.TryApplySunderArmorEffect(sim, target)
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
