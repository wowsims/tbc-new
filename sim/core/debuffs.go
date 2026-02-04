package core

import (
	"strconv"
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// applyRaidDebuffEffects applies all raid-level debuffs based on the provided Debuffs proto.
func applyDebuffEffects(target *Unit, targetIdx int, debuffs *proto.Debuffs, raid *proto.Raid) {

	if debuffs.BloodFrenzy {
		MakePermanent(BloodFrenzyAura(target))
	}

	if debuffs.CurseOfElements != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(CurseOfElementsAura(target, IsImproved(debuffs.CurseOfElements)))
	}

	if debuffs.CurseOfRecklessness {
		MakePermanent(CurseOfRecklessnessAura(target))
	}

	if debuffs.DemoralizingRoar != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(DemoralizingRoarAura(target, IsImproved(debuffs.DemoralizingRoar)))
	}

	if debuffs.DemoralizingShout != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(DemoralizingRoarAura(target, IsImproved(debuffs.DemoralizingRoar)))
	}

	if debuffs.ExposeArmor != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(ExposeArmorAura(target, 5, 2))
	}

	if debuffs.ExposeWeaknessUptime > 0.0 {
		MakePermanent(ExposeWeaknessAura(target, debuffs.ExposeWeaknessUptime, debuffs.ExposeWeaknessHunterAgility))
	}

	if debuffs.FaerieFire != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(FaerieFireAura(target, IsImproved(debuffs.FaerieFire)))
	}

	if debuffs.HemorrhageUptime > 0.0 {
		MakePermanent(HemorrhageAura(target, debuffs.HemorrhageUptime))
	}

	if debuffs.GiftOfArthas {
		MakePermanent(GiftOfArthasAura(target))
	}

	if debuffs.HuntersMark != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(HuntersMarkAura(target, IsImproved(debuffs.HuntersMark)))
	}

	if debuffs.ImprovedScorch {
		MakePermanent(ImprovedScorchAura(target, 5))
	}

	if debuffs.ImprovedSealOfTheCrusader {
		MakePermanent(ImprovedSealOfTheCrusaderAura(target))
	}

	if debuffs.InsectSwarm {
		MakePermanent((InsectSwarmAura(target)))
	}

	if debuffs.IsbUptime > 0.0 {
		MakePermanent(ImprovedShadowBoltAura(target, debuffs.IsbUptime, 5))
	}

	if debuffs.JudgementOfLight {
		MakePermanent(JudgementOfLightAura(target))
	}

	if debuffs.JudgementOfWisdom {
		MakePermanent(JudgementOfWisdomAura(target))
	}

	if debuffs.Mangle {
		MakePermanent(MangleAura(target))
	}

	if debuffs.Misery {
		MakePermanent(MiseryAura(target))
	}

	if debuffs.ScorpidSting {
		MakePermanent(ScorpidStingAura(target))
	}

	if debuffs.Screech {
		MakePermanent(ScreechAura(target))
	}

	if debuffs.ShadowEmbrace {
		MakePermanent(ShadowEmbraceAura(target))
	}

	if debuffs.ShadowWeaving {
		MakePermanent(ShadowWeavingAura(target))
	}

	if debuffs.SunderArmor {
		MakePermanent(SunderArmorAura(target))
	}

	if debuffs.WintersChill {
		MakePermanent(WintersChillAura(target, 5))
	}
}

// Physical anmd Armor Related Debuffs
func BloodFrenzyAura(target *Unit) *Aura {
	return damageTakenDebuff(target, "Blood Frenzy", 29859, []stats.SchoolIndex{stats.SchoolIndexPhysical}, 1.04, time.Second*21)
}

// Damage Taken Debuffs
func CurseOfElementsAura(target *Unit, improved bool) *Aura {
	multiplier := 1.10
	if improved {
		multiplier += 0.03
	}

	return damageTakenDebuff(target, "Curse of Elements", 27228,
		[]stats.SchoolIndex{
			stats.SchoolIndexArcane,
			stats.SchoolIndexFire,
			stats.SchoolIndexFrost,
			stats.SchoolIndexShadow,
		},
		multiplier,
		time.Minute*5,
	)
}

func CurseOfRecklessnessAura(target *Unit) *Aura {
	return statsDebuff(target, "Curse of Recklesness", 27226, stats.Stats{stats.Armor: -800, stats.AttackPower: 135})
}

func DemoralizingRoarAura(target *Unit, improved bool) *Aura {
	apReduction := 248.0
	if improved {
		apReduction *= 1.4
	}

	return statsDebuff(target, "Demoralizing Roar", 26998, stats.Stats{stats.AttackPower: apReduction})
}

func DemoralizingShoutAura(target *Unit, improved bool) *Aura {
	apReduction := 300.0
	if improved {
		apReduction *= 1.4
	}

	return statsDebuff(target, "Demoralizing Shout", 25203, stats.Stats{stats.AttackPower: apReduction})
}

func SlowAura(target *Unit) *Aura {
	return castSlowReductionAura(target, "Slow", 31589, 1.5, time.Second*15)
}

func castSlowReductionAura(target *Unit, label string, spellID int32, multiplier float64, duration time.Duration) *Aura {
	aura := target.GetOrRegisterAura(Aura{Label: label, ActionID: ActionID{SpellID: spellID}, Duration: duration})
	aura.NewExclusiveEffect("CastSpdReduction", false, ExclusiveEffect{
		Priority: multiplier,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, 1/multiplier)
			ee.Aura.Unit.MultiplyRangedSpeed(sim, 1/multiplier)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, multiplier)
			ee.Aura.Unit.MultiplyRangedSpeed(sim, multiplier)
		},
	})
	return aura
}

func ExposeArmorAura(target *Unit, comboPoints int32, talentPoints int32) *Aura {
	eaValue := 410.0 * float64(comboPoints)
	eaValue *= 1.0 + 0.25*float64(talentPoints)
	return statsDebuff(target, "Expose Armor", 26866, stats.Stats{stats.Armor: -eaValue})
}

func ExposeWeaknessAura(target *Unit, uptime float64, hunterAgility float64) *Aura {
	apBonus := hunterAgility * 0.25 * uptime

	return target.GetOrRegisterAura(Aura{
		Label:    "Expose Weakness",
		Tag:      "ExposeWeakness",
		ActionID: ActionID{SpellID: 34503},
		Duration: time.Second * 7,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.AddStat(stats.AttackPower, apBonus)
			aura.Unit.AddStat(stats.RangedAttackPower, apBonus)
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.AddStat(stats.AttackPower, -apBonus)
			aura.Unit.AddStat(stats.RangedAttackPower, -apBonus)
		},
	})

}

func FaerieFireAura(target *Unit, improved bool) *Aura {
	aura := target.GetOrRegisterAura(Aura{
		Label:    "Faerie Fire",
		ActionID: ActionID{SpellID: 26993},
		Duration: time.Second * 40,
	}).AttachStatBuff(stats.Armor, -610)

	if improved {
		aura.AttachAdditivePseudoStatBuff(&target.PseudoStats.ReducedPhysicalHitTakenChance, -3)
	}

	return aura
}

func GiftOfArthasAura(target *Unit) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    "Gift of Arthas",
		ActionID: ActionID{SpellID: 11374},
		Duration: time.Minute * 3,
	}).AttachAdditivePseudoStatBuff(&target.PseudoStats.BonusPhysicalDamageTaken, 8)
}

func HemorrhageAura(target *Unit, uptime float64) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    "Hemorrhage",
		ActionID: ActionID{SpellID: 33876},
		Duration: time.Second * 15,
	}).AttachAdditivePseudoStatBuff(&target.PseudoStats.BonusPhysicalDamageTaken, 42*uptime)
}

func HuntersMarkAura(target *Unit, improved bool) *Aura {
	maxBonus := 440.0

	return target.GetOrRegisterAura(Aura{
		Label:    "HuntersMark",
		Tag:      "HuntersMark",
		ActionID: ActionID{SpellID: 14325},
		Duration: NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			for _, unit := range sim.AllUnits {
				if unit.Type == PlayerUnit || unit.Type == PetUnit {
					if improved {
						unit.PseudoStats.BonusAttackPower += maxBonus
					}
					unit.PseudoStats.BonusRangedAttackPower += maxBonus
				}
			}

		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			for _, unit := range sim.AllUnits {
				if unit.Type == PlayerUnit || unit.Type == PetUnit {
					if improved {
						unit.PseudoStats.BonusAttackPower -= maxBonus
					}
					unit.PseudoStats.BonusRangedAttackPower -= maxBonus
				}
			}
		},
	})
}

func ImprovedScorchAura(target *Unit, startingStacks int32) *Aura {
	fireBonus := 0.03

	dynamicMods := make(map[int32]*SpellMod, len(target.Env.AllUnits))

	for _, unit := range target.Env.AllUnits {
		if unit.Type == PlayerUnit || unit.Type == PetUnit {
			dynamicMods[unit.UnitIndex] = unit.AddDynamicMod(SpellModConfig{
				Kind:       SpellMod_DamageDone_Pct,
				FloatValue: 0,
				School:     SpellSchoolFire,
			})
		}
	}

	return target.GetOrRegisterAura(Aura{
		Label:     "Improved Scorch",
		ActionID:  ActionID{SpellID: 12873},
		Duration:  time.Second * 30,
		MaxStacks: 5,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.SetStacks(sim, startingStacks)
		},
		OnStacksChange: func(aura *Aura, sim *Simulation, oldStacks int32, newStacks int32) {
			for _, unit := range sim.AllUnits {
				if unit.Type == PlayerUnit || unit.Type == PetUnit {
					dynamicMods[unit.UnitIndex].Activate()
					dynamicMods[unit.UnitIndex].UpdateFloatValue(fireBonus * float64(newStacks))
				}
			}
		},
	})

}

func ImprovedSealOfTheCrusaderAura(target *Unit) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    "Improved Seal of the Crusader",
		ActionID: ActionID{SpellID: 20337},
		Duration: time.Second * 60,
	}).AttachAdditivePseudoStatBuff(&target.PseudoStats.ReducedCritTakenChance, -3)

}

func ImprovedShadowBoltAura(target *Unit, uptime float64, points int32) *Aura {
	bonus := 0.04 * float64(points)
	multiplier := 1 + bonus
	if uptime != 0.0 {
		multiplier = 1 + bonus*uptime
	}

	config := Aura{
		Label:     "ImprovedShadowBolt-" + strconv.Itoa(int(points)),
		Tag:       "ImprovedShadowBolt",
		ActionID:  ActionID{SpellID: 17803},
		Duration:  time.Second * 12,
		MaxStacks: 4,
	}

	if uptime == 0 {
		config.OnSpellHitTaken = func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if !spell.SpellSchool.Matches(SpellSchoolShadow) || !result.Landed() || result.Damage == 0 || !spell.ProcMask.Matches(ProcMaskSpellDamage) {
				return
			}
			aura.RemoveStack(sim)
		}
	}

	return target.GetOrRegisterAura(config).
		AttachMultiplicativePseudoStatBuff(&target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow], multiplier)
}

func InsectSwarmAura(target *Unit) *Aura {
	return statsDebuff(target, "Insect Swarm", 27013, stats.Stats{
		stats.AllPhysHitRating: 0.98,
		stats.SpellHitPercent:  0.98,
	})
}

func JudgementOfLightAura(target *Unit) *Aura {
	actionId := ActionID{SpellID: 27163}
	healthMetrics := target.NewHealthMetrics(actionId)

	return target.GetOrRegisterAura(Aura{
		Label:    "Judgement of Light",
		ActionID: actionId,
		Duration: time.Second * 20,
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {

			if !spell.ProcMask.Matches(ProcMaskMelee) || !result.Landed() {
				return
			}

			if spell.ActionID.SameAction(ActionID{SpellID: 35395}) {
				aura.Refresh(sim)
			}

			if sim.Proc(0.5, "Judgement of Light - Heal") {
				aura.Unit.GainHealth(sim, 95.0, healthMetrics)
			}
		},
	})
}

func JudgementOfWisdomAura(target *Unit) *Aura {
	actionId := ActionID{SpellID: 27167}

	return target.GetOrRegisterAura(Aura{
		Label:    "Judgement of Wisdom",
		ActionID: actionId,
		Duration: time.Second * 20,
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if spell.ProcMask.Matches(ProcMaskEmpty) {
				return // Phantom spells (Romulo's, Lightning Capacitor, etc) don't proc JoW.
			}

			// Melee claim that wisdom can proc on misses.
			if !spell.ProcMask.Matches(ProcMaskMeleeOrRanged) && !result.Landed() {
				return
			}

			unit := spell.Unit
			if unit.HasManaBar() {
				if unit.JowManaMetrics == nil {
					unit.JowManaMetrics = unit.NewManaMetrics(actionId)
				}
				unit.AddMana(sim, 121.0, unit.JowManaMetrics)
			}

			if spell.ActionID.SameAction(ActionID{SpellID: 35395}) {
				aura.Refresh(sim)
			}
		},
	})
}

func MangleAura(target *Unit) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    "Mangle",
		ActionID: ActionID{SpellID: 33876},
		Duration: time.Second * 15,
	}).AttachMultiplicativePseudoStatBuff(&target.PseudoStats.PeriodicPhysicalDamageTakenMultiplier, 1.3)
}

func MiseryAura(target *Unit) *Aura {
	return damageTakenDebuff(
		target,
		"Misery",
		33195,
		[]stats.SchoolIndex{
			stats.SchoolIndexArcane,
			stats.SchoolIndexFire,
			stats.SchoolIndexFrost,
			stats.SchoolIndexHoly,
			stats.SchoolIndexNature,
			stats.SchoolIndexShadow,
		},
		1.05,
		time.Minute*1,
	)
}

func ScorpidStingAura(target *Unit) *Aura {
	return statsDebuff(target, "Scorpid Sting", 3043, stats.Stats{stats.AllPhysHitRating: -5.0})
}

func ScreechAura(target *Unit) *Aura {
	return statsDebuff(target, "Screech", 27051, stats.Stats{stats.AttackPower: -210})
}

func ShadowEmbraceAura(target *Unit) *Aura {
	return damageDealtDebuff(target, "Shadow Embrace", 32394, []stats.SchoolIndex{stats.SchoolIndexPhysical}, 0.95, NeverExpires)
}

func ShadowWeavingAura(target *Unit) *Aura {
	return damageTakenDebuff(target, "Shadow Weaving", 15334, []stats.SchoolIndex{stats.SchoolIndexShadow}, 1.10, time.Second*15)
}

func StormstrikeAura(target *Unit, uptime float64) *Aura {
	multiplier := 1.20
	if uptime != 0 {
		multiplier *= uptime
	}
	return damageTakenDebuff(target, "Stormstrike", 17364, []stats.SchoolIndex{stats.SchoolIndexNature}, multiplier, time.Second*12)
}

func SunderArmorAura(target *Unit) *Aura {
	return statsDebuff(target, "Sunder Amor", 25225, stats.Stats{stats.Armor: -2600})
}

func WintersChillAura(target *Unit, startingStacks int32) *Aura {
	critBonus := 2.0

	dynamicMods := make(map[int32]*SpellMod, len(target.Env.AllUnits))

	for _, unit := range target.Env.AllUnits {
		if unit.Type == PlayerUnit || unit.Type == PetUnit {
			dynamicMods[unit.UnitIndex] = unit.AddDynamicMod(SpellModConfig{
				Kind:       SpellMod_BonusCrit_Percent,
				FloatValue: 0,
				School:     SpellSchoolFrost,
			})
		}
	}

	return target.GetOrRegisterAura(Aura{
		Label:     "Winter's Chill",
		ActionID:  ActionID{SpellID: 28595},
		Duration:  time.Second * 15,
		MaxStacks: 5,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.SetStacks(sim, startingStacks)
		},
		OnStacksChange: func(aura *Aura, sim *Simulation, oldStacks int32, newStacks int32) {
			for _, unit := range sim.AllUnits {
				if unit.Type == PlayerUnit || unit.Type == PetUnit {
					dynamicMods[unit.UnitIndex].Activate()
					dynamicMods[unit.UnitIndex].UpdateFloatValue(critBonus * float64(newStacks))
				}
			}
		},
	})
}

func damageTakenDebuff(target *Unit, label string, spellID int32, schools []stats.SchoolIndex, multiplier float64, duration time.Duration) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    label,
		ActionID: ActionID{SpellID: spellID},
		Duration: duration,
		OnGain: func(aura *Aura, sim *Simulation) {
			for _, school := range schools {
				target.PseudoStats.SchoolDamageTakenMultiplier[school] *= multiplier
			}
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			for _, school := range schools {
				target.PseudoStats.SchoolDamageDealtMultiplier[school] /= -multiplier
			}
		},
	})
}

func damageDealtDebuff(target *Unit, label string, spellID int32, schools []stats.SchoolIndex, multiplier float64, duration time.Duration) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    label,
		ActionID: ActionID{SpellID: spellID},
		Duration: duration,

		OnGain: func(aura *Aura, sim *Simulation) {
			for _, school := range schools {
				target.PseudoStats.SchoolDamageDealtMultiplier[school] *= 1.0 - multiplier
			}
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			for _, school := range schools {
				target.PseudoStats.SchoolDamageDealtMultiplier[school] /= 1.0 - multiplier
			}
		},
	})
}

func statsDebuff(target *Unit, label string, spellID int32, stats stats.Stats) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    label,
		ActionID: ActionID{SpellID: spellID},
		Duration: time.Second * 30,
	}).AttachStatsBuff(stats)
}
