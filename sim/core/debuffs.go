package core

import (
	"math/rand"
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// applyRaidDebuffEffects applies all raid-level debuffs based on the provided Debuffs proto.
func applyDebuffEffects(target *Unit, targetIdx int, debuffs *proto.Debuffs, raid *proto.Raid) {

	if debuffs.BloodFrenzy {

		MakePermanent(BloodFrenzy(target))
	}

	if debuffs.CurseOfElements != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(CurseOfElements(target, IsImproved(debuffs.CurseOfElements)))
	}

	if debuffs.CurseOfRecklessness {
		MakePermanent(CurseOfRecklessness(target))
	}

	if debuffs.DemoralizingRoar != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(DemoralizingRoar(target, IsImproved(debuffs.DemoralizingRoar)))
	}

	if debuffs.DemoralizingShout != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(DemoralizingRoar(target, IsImproved(debuffs.DemoralizingRoar)))
	}

	if debuffs.ExposeArmor != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(ExposeArmor(target, IsImproved(debuffs.ExposeArmor)))
	}

	if debuffs.ExposeWeaknessUptime > 0.0 {
		MakePermanent(ExposeWeakness(target, debuffs.ExposeWeaknessUptime, debuffs.ExposeWeaknessHunterAgility))
	}

	if debuffs.FaerieFire != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(FaerieFire(target, IsImproved(debuffs.FaerieFire)))
	}

	if debuffs.HemorrhageUptime > 0.0 {
		MakePermanent(Hemorrhage(target, debuffs.HemorrhageUptime))
	}

	if debuffs.GiftOfArthas {
		MakePermanent(GiftOfArthas(target))
	}

	if debuffs.HuntersMark != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(HuntersMark(target, IsImproved(debuffs.HuntersMark)))
	}

	if debuffs.ImprovedScorch {
		MakePermanent(ImprovedScorch(target))
	}

	if debuffs.ImprovedSealOfTheCrusader {
		MakePermanent(ImprovedSealOfTheCrusader(target))
	}

	if debuffs.InsectSwarm {
		MakePermanent((InsectSwarm(target)))
	}

	if debuffs.IsbUptime > 0.0 {
		MakePermanent(ImprovedShadowBolt(target, debuffs.IsbUptime))
	}

	if debuffs.JudgementOfLight {
		MakePermanent(JudgementOfLight(target))
	}

	if debuffs.JudgementOfWisdom {
		MakePermanent(JudgementOfWisdom(target))
	}

	if debuffs.Mangle {
		MakePermanent(Mangle(target))
	}

	if debuffs.Misery {
		MakePermanent(Misery(target))
	}

	if debuffs.ScorpidSting {
		MakePermanent(ScorpidSting(target))
	}

	if debuffs.Screech {
		MakePermanent(Screech(target))
	}

	if debuffs.ShadowEmbrace {
		MakePermanent(ShadowEmbrace(target))
	}

	if debuffs.ShadowWeaving {
		MakePermanent(ShadowWeaving(target))
	}

	if debuffs.SunderArmor {
		MakePermanent(SunderArmor(target))
	}

	if debuffs.WintersChill {
		MakePermanent(WintersChill(target))
	}
}

// Physical anmd Armor Related Debuffs
func BloodFrenzy(target *Unit) *Aura {
	return damageTakenDebuff(target, "Blood Frenzy", 29859, []stats.SchoolIndex{stats.SchoolIndexPhysical}, 1.04, time.Second*21)
}

// Damage Taken Debuffs
func CurseOfElements(target *Unit, improved bool) *Aura {
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

func CurseOfRecklessness(target *Unit) *Aura {
	return statsDebuff(target, "Curse of Recklesness", 27226, stats.Stats{stats.Armor: -8000, stats.AttackPower: 135})
}

func DemoralizingRoar(target *Unit, improved bool) *Aura {
	apReduction := 248.0
	if improved {
		apReduction *= 1.4
	}

	return statsDebuff(target, "Demoralizing Roar", 26998, stats.Stats{stats.AttackPower: apReduction})
}

func DemoralizingShout(target *Unit, improved bool) *Aura {
	apReduction := 300.0
	if improved {
		apReduction *= 1.4
	}

	return statsDebuff(target, "Demoralizing Shout", 25203, stats.Stats{stats.AttackPower: apReduction})
}

func ExposeArmor(target *Unit, improved bool) *Aura {
	eaValue := 2050.0
	if improved {
		eaValue *= 1.50
	}
	return statsDebuff(target, "Expose Armor", 26866, stats.Stats{stats.Armor: eaValue})
}

func ExposeWeakness(target *Unit, uptime float64, hunterAgility float64) *Aura {
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

func FaerieFire(target *Unit, improved bool) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    "Faerie Fire",
		ActionID: ActionID{SpellID: 26993},
		Duration: time.Second * 40,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.AddStatsDynamic(sim, stats.Stats{stats.Armor: -6100})
			aura.Unit.PseudoStats.ReducedPhysicalHitTakenChance -= 3.0
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.AddStatsDynamic(sim, stats.Stats{stats.Armor: 6100})
			aura.Unit.PseudoStats.ReducedPhysicalHitTakenChance += 3.0
		},
	})
}

func GiftOfArthas(target *Unit) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    "Gift of Arthas",
		ActionID: ActionID{SpellID: 11374},
		Duration: time.Minute * 3,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.BonusPhysicalDamageTaken += 8
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.BonusPhysicalDamageTaken -= 8
		},
	})
}

func Hemorrhage(target *Unit, uptime float64) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:     "Hemorrhage",
		ActionID:  ActionID{SpellID: 33876},
		Duration:  time.Second * 15,
		MaxStacks: 15,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.stacks = 15
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.stacks = 15
		},
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if !spell.SpellSchool.Matches(SpellSchoolPhysical) {
				return
			}

			if aura.stacks <= 0 {
				aura.Deactivate(sim)
				return
			}

			result.Damage += 42 * uptime
			spell.DealDamage(sim, result)

			aura.stacks--
		},
	})
}

func HuntersMark(target *Unit, improved bool) *Aura {
	maxBonus := 440.0

	return target.GetOrRegisterAura(Aura{
		Label:    "HuntersMark",
		Tag:      "HuntersMark",
		ActionID: ActionID{SpellID: 14325},
		Duration: NeverExpires,
		OnGain: func(aura *Aura, sim *Simulation) {
			if improved {
				aura.Unit.AddStat(stats.AttackPower, maxBonus)
			}
			aura.Unit.AddStat(stats.RangedAttackPower, maxBonus)
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			if improved {
				aura.Unit.AddStat(stats.AttackPower, -maxBonus)
			}
			aura.Unit.AddStat(stats.RangedAttackPower, -maxBonus)
		},
	})
}

func ImprovedScorch(target *Unit) *Aura {
	return damageTakenDebuff(target, "Improved Scorch", 12873, []stats.SchoolIndex{stats.SchoolIndexFire}, 1.15, time.Second*30)
}

func ImprovedSealOfTheCrusader(target *Unit) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    "Improved Seal of the Crusader",
		ActionID: ActionID{SpellID: 20337},
		Duration: time.Second * 60,
		OnGain: func(aura *Aura, sim *Simulation) {
			target.PseudoStats.ReducedCritTakenChance -= 3.0
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			target.PseudoStats.ReducedCritTakenChance += 3.0
		},
	})

}

func ImprovedShadowBolt(target *Unit, uptime float64) *Aura {
	multiplier := 1.2 * uptime

	return target.GetOrRegisterAura(Aura{
		Label:     "ImprovedShadowBolt",
		ActionID:  ActionID{SpellID: 17803},
		Duration:  time.Second * 12,
		MaxStacks: 4,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow] *= multiplier
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow] /= multiplier
		},
	})
}

func InsectSwarm(target *Unit) *Aura {
	return statsDebuff(target, "Insect Swarm", 27013, stats.Stats{
		stats.AllPhysHitRating: 0.98,
		stats.SpellHitPercent:  0.98,
	})
}

func JudgementOfLight(target *Unit) *Aura {
	actionId := ActionID{SpellID: 27163}

	return target.GetOrRegisterAura(Aura{
		Label:    "Judgement of Light",
		ActionID: actionId,
		Duration: time.Second * 20,
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			var healthMetrics *ResourceMetrics
			healthMetrics = aura.Unit.NewHealthMetrics(actionId)
			if !spell.ProcMask.Matches(ProcMaskMelee) || !result.Landed() {
				return
			}

			if spell.ActionID.SpellID == 35395 {
				aura.Refresh(sim)
			}

			if rand.Float64() < 50.0 {
				aura.Unit.GainHealth(sim, 95.0, healthMetrics)
			}
		},
	})
}

func JudgementOfWisdom(target *Unit) *Aura {
	actionId := ActionID{SpellID: 27167}

	return target.GetOrRegisterAura(Aura{
		Label:    "Judgement of Light",
		ActionID: actionId,
		Duration: time.Second * 20,
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			var manaMetrics *ResourceMetrics
			manaMetrics = aura.Unit.NewManaMetrics(actionId)
			if !spell.ProcMask.Matches(ProcMaskMelee) || !result.Landed() {
				return
			}

			if spell.ActionID.SpellID == 35395 {
				aura.Refresh(sim)
			}

			if rand.Float64() < 50.0 {
				aura.Unit.AddMana(sim, 121.0, manaMetrics)
			}
		},
	})
}

func Mangle(target *Unit) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    "Mangle",
		ActionID: ActionID{SpellID: 33876},
		Duration: time.Second * 15,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.PeriodicPhysicalDamageTakenMultiplier *= 1.3
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.PeriodicPhysicalDamageTakenMultiplier /= 1.3
		},
	})
}

func Misery(target *Unit) *Aura {
	return damageTakenDebuff(target, "Misery", 33195, []stats.SchoolIndex{
		stats.SchoolIndexArcane,
		stats.SchoolIndexFire,
		stats.SchoolIndexFrost,
		stats.SchoolIndexHoly,
		stats.SchoolIndexNature,
		stats.SchoolIndexShadow,
	}, 1.05, time.Minute*1)
}

func ScorpidSting(target *Unit) *Aura {
	return statsDebuff(target, "Scorpid Sting", 3043, stats.Stats{stats.AllPhysHitRating: -5.0})
}

func Screech(target *Unit) *Aura {
	return statsDebuff(target, "Screech", 27051, stats.Stats{stats.AttackPower: -210})
}

func ShadowEmbrace(target *Unit) *Aura {
	return damageDealtDebuff(target, "Shadow Embrace", 32394, []stats.SchoolIndex{stats.SchoolIndexPhysical}, 0.95, NeverExpires)
}

func ShadowWeaving(target *Unit) *Aura {
	return damageTakenDebuff(target, "Shadow Weaving", 15334, []stats.SchoolIndex{stats.SchoolIndexShadow}, 1.10, time.Second*15)
}

func Stormstrike(target *Unit, uptime float64) *Aura {
	multiplier := 1.20
	if uptime != 0 {
		multiplier *= uptime
	}
	return damageTakenDebuff(target, "Stormstrike", 17364, []stats.SchoolIndex{stats.SchoolIndexNature}, multiplier, time.Second*12)
}

func SunderArmor(target *Unit) *Aura {
	return statsDebuff(target, "Sunder Amor", 25225, stats.Stats{stats.Armor: 2600})
}

func WintersChill(target *Unit) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    "Winter's Chill",
		ActionID: ActionID{SpellID: 28595},
		Duration: time.Second * 15,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.AddStaticMod(SpellModConfig{
				Kind:       SpellMod_BonusCrit_Percent,
				FloatValue: 10.0,
				School:     SpellSchoolFire,
			})
		},
	})
}

func damageTakenDebuff(target *Unit, label string, spellID int32, schools []stats.SchoolIndex, multiplier float64, duration time.Duration) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    label,
		ActionID: ActionID{SpellID: spellID},
		Duration: time.Second * 30,
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
		Duration: time.Second * 30,

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

		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.AddStatsDynamic(sim, stats)
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.AddStatsDynamic(sim, stats.Multiply(-1))
		},
	})
}
