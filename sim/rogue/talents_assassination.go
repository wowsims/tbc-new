package rogue

import (
	"slices"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (rogue *Rogue) registerAssassinationTalents() {
	// Tier 1
	rogue.registerImprovedEviscerate()
	// Remorseless Attacks not implemented
	rogue.registerMalice()

	// Tier 2
	// Ruthless implemented in ApplyFinisher
	rogue.registerMurder()
	rogue.registerPuncturingWounds()

	// Tier 3
	// Relentless Strikes implemented in ApplyFinisher
	rogue.registerImprovedExposeArmor()
	rogue.registerLethality()

	// Tier 4
	rogue.registerVilePoisons()
	// Improved Poisons implemented in poisons.go

	// Tier 5
	// Fleet Footed NYI
	rogue.registerColdBlood()
	// Improved Kidney NYI
	// Quick Recovery implemented in individual finisher EnergyCostOptions

	// Tier 6
	rogue.registerSealFate()
	rogue.registerMasterPoisoner()

	// Tier 7
	// Vigor implemented in rogue.go
	// Deadened Nerves NYI

	// Tier 8
	rogue.registerFindWeakness()

	// Tier 9
	rogue.registerMutilate()
}

func (rogue *Rogue) registerImprovedEviscerate() {
	if rogue.Talents.ImprovedEviscerate == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		ClassMask:  RogueSpellEviscerate,
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: .05 * float64(rogue.Talents.ImprovedEviscerate),
	})
}

func (rogue *Rogue) registerMalice() {
	if rogue.Talents.Malice == 0 {
		return
	}

	rogue.AddStat(stats.AllPhysCritRating, float64(rogue.Talents.Malice)*core.PhysicalCritRatingPerCritPercent)
}

func (rogue *Rogue) registerMurder() {
	if rogue.Talents.Murder == 0 {
		return
	}
	var multiplier float64 = 1.0 + (0.01 * float64(rogue.Talents.Murder))
	rogue.Env.RegisterPostFinalizeEffect(func() {
		for _, at := range rogue.AttackTables {
			if slices.Contains([]proto.MobType{proto.MobType_MobTypeHumanoid, proto.MobType_MobTypeGiant, proto.MobType_MobTypeBeast, proto.MobType_MobTypeDragonkin}, at.Defender.MobType) {
				at.DamageDealtMultiplier *= multiplier
				at.CritMultiplier *= multiplier
			}
		}
	})
}

func (rogue *Rogue) registerPuncturingWounds() {
	if rogue.Talents.PuncturingWounds == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ClassMask:  RogueSpellBackstab,
		FloatValue: 10.0 * float64(rogue.Talents.PuncturingWounds),
	})
	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ClassMask:  RogueSpellMutilateHit,
		FloatValue: 5.0 * float64(rogue.Talents.PuncturingWounds),
	})
}

func (rogue *Rogue) registerImprovedExposeArmor() {
	if rogue.Talents.ImprovedExposeArmor == 0 {
		return
	}

	rogue.ExposeArmorModifier = 1.5
}

func (rogue *Rogue) registerLethality() {
	if rogue.Talents.Lethality == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_CritMultiplier_Flat,
		ClassMask:  RogueSpellLethality,
		FloatValue: 0.06 * float64(rogue.Talents.Lethality),
	})
}

func (rogue *Rogue) registerVilePoisons() {
	if rogue.Talents.VilePoisons == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellPoisons,
		FloatValue: 0.04 * float64(rogue.Talents.VilePoisons),
	})
}

func (rogue *Rogue) registerColdBlood() {
	if !rogue.Talents.ColdBlood {
		return
	}

	cbAura := rogue.GetOrRegisterAura(core.Aura{
		Label:    "Cold Blood",
		ActionID: core.ActionID{SpellID: 14177},
		Duration: core.NeverExpires,

		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ClassSpellMask&RogueSpellCanCrit != 0 {
				aura.Deactivate(sim)
			}
		},
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ClassMask:  RogueSpellCanCrit,
		FloatValue: 100.0,
	})

	rogue.ColdBlood = rogue.GetOrRegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 14177},

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Minute * 3,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			cbAura.Activate(sim)
		},
	})

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.ColdBlood,
		Type:  core.CooldownTypeDPS,
	})
}

func (rogue *Rogue) registerSealFate() {
	if rogue.Talents.SealFate == 0 {
		return
	}

	sfMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14195})

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Seal Fate Trigger",
		ActionID:       core.ActionID{SpellID: 14195},
		ProcChance:     0.2 * float64(rogue.Talents.SealFate),
		Callback:       core.CallbackOnSpellHitDealt,
		Outcome:        core.OutcomeCrit,
		ClassSpellMask: RogueSpellLethality,
		ICD:            time.Millisecond * 500,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			rogue.AddComboPoints(sim, 1, sfMetrics)
		},
	})
}

func (rogue *Rogue) registerMasterPoisoner() {
	if rogue.Talents.MasterPoisoner == 0 {
		return
	}

	rogue.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusHit_Percent,
		ClassMask:  RogueSpellPoisons,
		FloatValue: 5.0 * float64(rogue.Talents.MasterPoisoner),
	})
}

func (rogue *Rogue) registerFindWeakness() {
	if rogue.Talents.FindWeakness == 0 {
		return
	}

	fwAura := rogue.GetOrRegisterAura(core.Aura{
		Label:    "Find Weakness",
		Duration: time.Second * 10,
		ActionID: core.ActionID{SpellID: 31242},
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellsAll,
		FloatValue: 0.2 * float64(rogue.Talents.FindWeakness),
	})

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Find Weakness Trigger",
		ActionID:       core.ActionID{SpellID: 31242},
		ProcChance:     1,
		Callback:       core.CallbackOnSpellHitDealt,
		Outcome:        core.OutcomeLanded,
		ClassSpellMask: RogueSpellFinisher,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			fwAura.Activate(sim)
		},
	})
}

const MutilateSpellID int32 = 34413

func (rogue *Rogue) registerMutilate() {
	if !rogue.Talents.Mutilate {
		return
	}

	rogue.MutilateMH = rogue.newMutilateHitSpell(true)
	rogue.MutilateOH = rogue.newMutilateHitSpell(false)

	rogue.Mutilate = rogue.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: MutilateSpellID, Tag: 0},
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: RogueSpellMutilate,

		EnergyCost: core.EnergyCostOptions{
			Cost:   60,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			if rogue.HasDagger(core.MainHand) && rogue.HasDagger(core.OffHand) {
				return true
			}
			return false
		},

		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit) // Miss/Dodge/Parry/Hit
			if result.Landed() {
				rogue.AddComboPoints(sim, 2, spell.ComboPointMetrics())
				rogue.MutilateOH.Cast(sim, target)
				rogue.MutilateMH.Cast(sim, target)
			} else {
				spell.IssueRefund(sim)
			}
			spell.DealOutcome(sim, result)
		},
	})
}

func (rogue *Rogue) newMutilateHitSpell(isMH bool) *core.Spell {
	actionID := core.ActionID{SpellID: MutilateSpellID, Tag: 1}
	procMask := core.ProcMaskMeleeMHSpecial
	if !isMH {
		actionID = core.ActionID{SpellID: MutilateSpellID, Tag: 2}
		procMask = core.ProcMaskMeleeOHSpecial
	}
	mutBaseDamage := 101.0

	return rogue.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       procMask,
		Flags:          core.SpellFlagMeleeMetrics,
		ClassSpellMask: RogueSpellMutilateHit,

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		CritMultiplier:           rogue.DefaultMeleeCritMultiplier(),
		ThreatMultiplier:         1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			var baseDamage float64
			if isMH {
				baseDamage = mutBaseDamage + spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			} else {
				baseDamage = mutBaseDamage + spell.Unit.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			}

			oldMultiplier := spell.DamageMultiplier
			if rogue.DeadlyPoison.Dot(target).IsActive() || rogue.WoundPoisonDebuffAuras.Get(target).IsActive() {
				spell.DamageMultiplier += 0.5
			}

			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialBlockAndCrit)
			spell.DamageMultiplier = oldMultiplier
		},
	})
}
