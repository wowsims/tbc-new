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
		MakePermanent(BloodFrenzyAura(target, 2))
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
		MakePermanent(DemoralizingShoutAura(target, 5, TernaryInt32(IsImproved(debuffs.DemoralizingShout), 5, 0)))
	}

	if debuffs.ExposeWeaknessUptime > 0.0 {
		ExposeWeaknessAura(target, debuffs.ExposeWeaknessUptime, debuffs.ExposeWeaknessHunterAgility)
	}

	if debuffs.FaerieFire != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(FaerieFireAura(target, IsImproved(debuffs.FaerieFire)))
	}

	if debuffs.HemorrhageUptime > 0.0 {
		HemorrhageAura(target, debuffs.HemorrhageUptime)
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
		ImprovedShadowBoltAura(target, debuffs.IsbUptime, 5)
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

	if debuffs.ExposeArmor != proto.TristateEffect_TristateEffectMissing {
		aura := MakePermanent(ExposeArmorAura(target, func() int32 { return 5 }, TernaryInt32(debuffs.ExposeArmor == 2, 2, 0)))

		ScheduledMajorArmorAura(aura, PeriodicActionOptions{
			Period:   time.Second * 3,
			NumTicks: 1,
			OnAction: func(sim *Simulation) {
				aura.Activate(sim)
			},
		}, raid)
	}

	if debuffs.SunderArmor {
		aura := MakePermanent(SunderArmorAura(target))

		ScheduledMajorArmorAura(aura, PeriodicActionOptions{
			Period:          GCDDefault,
			NumTicks:        5,
			TickImmediately: true,
			Priority:        ActionPriorityDOT, // High prio so it comes before actual warrior sunders.
			OnAction: func(sim *Simulation) {
				aura.Activate(sim)
				if aura.IsActive() {
					aura.AddStack(sim)
				}
			},
		}, raid)
	}

	if debuffs.WintersChill {
		MakePermanent(WintersChillAura(target, 5))
	}

	if debuffs.ThunderClap != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(ThunderClapAura(target, GetTristateValueInt32(debuffs.ThunderClap, 0, 3)))
	}
}

func ScheduledMajorArmorAura(aura *Aura, options PeriodicActionOptions, raid *proto.Raid) {
	aura.OnReset = func(aura *Aura, sim *Simulation) {
		aura.Duration = NeverExpires
		StartPeriodicAction(sim, options)
	}
}

// Physical and Armor Related Debuffs
func BloodFrenzyAura(target *Unit, points int32) *Aura {
	return damageTakenDebuff(target,
		"Blood Frenzy",
		29859,
		[]stats.SchoolIndex{stats.SchoolIndexPhysical},
		1+0.02*float64(points),
		NeverExpires,
	)
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
	return statsDebuff(target, "Curse of Recklesness", 27226, stats.Stats{stats.Armor: -800, stats.AttackPower: 135}, time.Minute*2)
}

func DemoralizingRoarAura(target *Unit, improved bool) *Aura {
	apReduction := 248.0
	if improved {
		apReduction *= 1.4
	}

	return statsDebuff(target, "Demoralizing Roar", 26998, stats.Stats{stats.AttackPower: apReduction}, time.Second*30)
}

func DemoralizingShoutAura(target *Unit, boomingVoicePoints int32, improvedDemoShoutPoints int32) *Aura {
	apReduction := 300.0 * (1 + 0.1*float64(improvedDemoShoutPoints))
	duration := time.Duration(float64(time.Second*30) * (1 + 0.1*float64(boomingVoicePoints)))

	return statsDebuff(target, "Demoralizing Shout", 25203, stats.Stats{stats.AttackPower: apReduction}, duration)
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

func ExposeWeaknessAura(target *Unit, uptime float64, hunterAgility float64) *Aura {
	apBonus := hunterAgility * 0.25
	stats := stats.Stats{stats.AttackPower: apBonus, stats.RangedAttackPower: apBonus}
	var character *Character
	for _, party := range target.Env.Raid.Parties {
		for _, agent := range party.Players {
			c := agent.GetCharacter()
			if c.Type == PlayerUnit {
				character = c
				break
			}
		}
	}

	hasAura := target.HasAura("Expose Weakness")
	aura := target.GetOrRegisterAura(Aura{
		Label:    "Expose Weakness",
		Tag:      "ExposeWeakness",
		ActionID: ActionID{SpellID: 34503},
		Duration: time.Second * 7,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.AddStatsDynamic(sim, stats)
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			character.AddStatsDynamic(sim, stats.Invert())
		},
	})

	if !hasAura {
		ApplyFixedUptimeAura(aura, uptime, aura.Duration, 1)
	}

	return aura

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
	var effect *ExclusiveEffect
	aura := target.GetOrRegisterAura(Aura{
		Label:    "Gift of Arthas",
		ActionID: ActionID{SpellID: 11374},
		Duration: time.Minute * 3,
		OnGain: func(aura *Aura, sim *Simulation) {
			effect.SetPriority(sim, 8)
		},
	})

	effect = aura.NewExclusiveEffect("GiftOfArthasAura", true, ExclusiveEffect{
		Priority: 0,
		OnGain: func(ee *ExclusiveEffect, s *Simulation) {
			ee.Aura.Unit.PseudoStats.BonusPhysicalDamageTaken += ee.Priority
		},
		OnExpire: func(ee *ExclusiveEffect, s *Simulation) {
			ee.Aura.Unit.PseudoStats.BonusPhysicalDamageTaken -= ee.Priority
		},
	})

	return aura
}

func HemorrhageAura(target *Unit, uptime float64) *Aura {
	hasAura := target.HasAura("Hemorrhage")
	aura := target.GetOrRegisterAura(Aura{
		Label:    "Hemorrhage",
		ActionID: ActionID{SpellID: 33876},
		Duration: time.Second * 15,
	})

	if !hasAura {
		aura.AttachAdditivePseudoStatBuff(&target.PseudoStats.BonusPhysicalDamageTaken, 42)
		ApplyFixedUptimeAura(aura, uptime, aura.Duration, 1)
	}

	return aura
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

	hasAura := target.HasAura(config.Label)
	aura := target.GetOrRegisterAura(config)
	if !hasAura {
		aura.AttachMultiplicativePseudoStatBuff(&target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow], multiplier)
		ApplyFixedUptimeAura(aura, uptime, aura.Duration, 1)
	}

	return aura
}

func InsectSwarmAura(target *Unit) *Aura {
	return statsDebuff(
		target,
		"Insect Swarm",
		27013,
		stats.Stats{
			stats.PhysicalHitPercent: -2,
			stats.SpellHitPercent:    -2,
		},
		time.Second*12,
	)
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
				spell.Unit.GainHealth(sim, 95.0, healthMetrics)
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
	return statsDebuff(target, "Scorpid Sting", 3043, stats.Stats{stats.PhysicalHitPercent: -5.0}, time.Second*20)
}

func ScreechAura(target *Unit) *Aura {
	return statsDebuff(target, "Screech", 27051, stats.Stats{stats.AttackPower: -210}, time.Second*4)
}

func ShadowEmbraceAura(target *Unit) *Aura {
	return damageDealtDebuff(target, "Shadow Embrace", 32394, []stats.SchoolIndex{stats.SchoolIndexPhysical}, 0.95, NeverExpires)
}

func ShadowWeavingAura(target *Unit) *Aura {
	return damageTakenDebuff(target, "Shadow Weaving", 15334, []stats.SchoolIndex{stats.SchoolIndexShadow}, 1.10, time.Second*15)
}

func StormstrikeAura(target *Unit, uptime float64) *Aura {
	multiplier := 1.20
	hasAura := target.HasAura("Stormstrike")
	aura := damageTakenDebuff(target, "Stormstrike", 17364, []stats.SchoolIndex{stats.SchoolIndexNature}, multiplier, time.Second*12)

	if !hasAura {
		ApplyFixedUptimeAura(aura, uptime, aura.Duration, 1)
	}

	return aura
}

var MajorArmorReductionEffectCategory = "MajorArmorReduction"

func ExposeArmorAura(target *Unit, getComboPoints func() int32, talents int32) *Aura {

	var effect *ExclusiveEffect
	aura := target.GetOrRegisterAura(Aura{
		Label:    "Expose Armor",
		ActionID: ActionID{SpellID: 26866},
		Duration: time.Second * 30,
		OnGain: func(aura *Aura, sim *Simulation) {
			eaValue := 410.0 * float64(getComboPoints())
			eaValue *= 1.0 + 0.25*float64(talents)
			effect.SetPriority(sim, eaValue)
		},
	})

	effect = aura.NewExclusiveEffect(MajorArmorReductionEffectCategory, true, ExclusiveEffect{
		Priority: 0,
		OnGain: func(ee *ExclusiveEffect, s *Simulation) {
			ee.Aura.Unit.stats[stats.Armor] -= ee.Priority
		},
		OnExpire: func(ee *ExclusiveEffect, s *Simulation) {
			ee.Aura.Unit.stats[stats.Armor] += ee.Priority
		},
	})

	return aura

}

func SunderArmorAura(target *Unit) *Aura {
	var effect *ExclusiveEffect
	aura := target.GetOrRegisterAura(Aura{
		Label:     "Sunder Armor",
		ActionID:  ActionID{SpellID: 25225},
		Duration:  time.Second * 30,
		MaxStacks: 5,
		OnStacksChange: func(aura *Aura, sim *Simulation, oldStacks int32, newStacks int32) {
			effect.SetPriority(sim, -520*float64(newStacks))
		},
	})

	effect = aura.NewExclusiveEffect(MajorArmorReductionEffectCategory, true, ExclusiveEffect{
		Priority: 0,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.stats[stats.Armor] += ee.Priority
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.stats[stats.Armor] -= ee.Priority
		},
	})

	return aura
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

func ThunderClapAura(target *Unit, points int32) *Aura {
	aura := target.GetOrRegisterAura(Aura{
		Label:    "ThunderClap-" + strconv.Itoa(int(points)),
		ActionID: ActionID{SpellID: 6343},
		Duration: time.Second * 30,
	})
	AtkSpeedReductionEffect(aura, []float64{1.1, 1.14, 1.17, 1.2}[points])
	return aura
}

func AtkSpeedReductionEffect(aura *Aura, speedMultiplier float64) *ExclusiveEffect {
	return aura.NewExclusiveEffect("AtkSpdReduction", false, ExclusiveEffect{
		Priority: speedMultiplier,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyAttackSpeed(sim, 1/speedMultiplier)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyAttackSpeed(sim, speedMultiplier)
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
				target.PseudoStats.SchoolDamageDealtMultiplier[school] *= multiplier
			}
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			for _, school := range schools {
				target.PseudoStats.SchoolDamageDealtMultiplier[school] /= multiplier
			}
		},
	})
}

func statsDebuff(target *Unit, label string, spellID int32, stats stats.Stats, duration time.Duration) *Aura {
	if duration == 0 {
		duration = time.Second * 30
	}

	return target.GetOrRegisterAura(Aura{
		Label:    label,
		ActionID: ActionID{SpellID: spellID},
		Duration: duration,
	}).AttachStatsBuff(stats)
}
