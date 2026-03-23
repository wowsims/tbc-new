package core

import (
	"fmt"
	"slices"
	"time"

	googleProto "google.golang.org/protobuf/proto"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type BuffConfig struct {
	Label             string
	ActionID          ActionID
	Duration          time.Duration
	Stats             []StatConfig
	ExclusiveCategory string
}

type StatConfig struct {
	Stat             stats.Stat
	Amount           float64
	IsMultiplicative bool
}

func makeMultiplierBuff(aura *Aura, stat stats.Stat, value float64) {
	dep := aura.Unit.NewDynamicMultiplyStat(stat, value)
	aura.ApplyOnGain(func(aura *Aura, sim *Simulation) {
		aura.Unit.EnableBuildPhaseStatDep(sim, dep)
	}).ApplyOnExpire(func(aura *Aura, sim *Simulation) {
		aura.Unit.DisableBuildPhaseStatDep(sim, dep)
	})
}

func makeFlatStatBuff(aura *Aura, stat stats.Stat, value float64) {
	aura.ApplyOnGain(func(aura *Aura, sim *Simulation) {
		aura.Unit.AddStatDynamic(sim, stat, value)
	}).ApplyOnExpire(func(aura *Aura, sim *Simulation) {
		aura.Unit.AddStatDynamic(sim, stat, -value)
	})
}

func registerStatEffect(aura *Aura, config []StatConfig) {
	for _, statConfig := range config {
		if statConfig.IsMultiplicative {
			makeMultiplierBuff(aura, statConfig.Stat, statConfig.Amount)
		} else {
			makeFlatStatBuff(aura, statConfig.Stat, statConfig.Amount)
		}
	}
}

func makeExclusiveMultiplierBuff(aura *Aura, stat stats.Stat, value float64, exclusiveCategory string) {
	dep := aura.Unit.NewDynamicMultiplyStat(stat, value)
	aura.NewExclusiveEffect(exclusiveCategory+stat.StatName()+"Mul", false, ExclusiveEffect{
		Priority: value,
		OnGain: func(ee *ExclusiveEffect, s *Simulation) {
			ee.Aura.Unit.EnableBuildPhaseStatDep(s, dep)
		},
		OnExpire: func(ee *ExclusiveEffect, s *Simulation) {
			ee.Aura.Unit.DisableBuildPhaseStatDep(s, dep)
		},
	})
}

func makeExclusiveFlatStatBuff(aura *Aura, stat stats.Stat, value float64, exclusiveCategory string) {
	aura.NewExclusiveEffect(exclusiveCategory+stat.StatName()+"Add", false, ExclusiveEffect{
		Priority: value,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stat, value)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stat, -value)
		},
	})
}

func registerExlusiveEffects(aura *Aura, config []StatConfig, exclusiveCategory string) {
	for _, statConfig := range config {
		if statConfig.IsMultiplicative {
			makeExclusiveMultiplierBuff(aura, statConfig.Stat, statConfig.Amount, exclusiveCategory)
		} else {
			makeExclusiveFlatStatBuff(aura, statConfig.Stat, statConfig.Amount, exclusiveCategory)
		}
	}
}

func makeStatBuff(char *Character, config BuffConfig) *Aura {
	if config.Label == "" {
		panic("Buff without label.")
	}

	if ActionID.IsEmptyAction(config.ActionID) {
		panic("Buff without ActionID")
	}

	if config.ActionID.Tag == 0 {
		config.ActionID = config.ActionID.WithTag(-1)
	}

	baseAura := char.GetOrRegisterAura(Aura{
		Label:      config.Label,
		ActionID:   config.ActionID,
		Duration:   TernaryDuration(config.Duration > 0, config.Duration, NeverExpires),
		BuildPhase: CharacterBuildPhaseBuffs,
	})

	if config.ExclusiveCategory != "" {
		registerExlusiveEffects(baseAura, config.Stats, config.ExclusiveCategory)
	} else {
		registerStatEffect(baseAura, config.Stats)
	}
	return baseAura
}

// Applies buffs that affect individual players.
func applyBuffEffects(agent Agent, raidBuffs *proto.RaidBuffs, partyBuffs *proto.PartyBuffs, individual *proto.IndividualBuffs) {
	char := agent.GetCharacter()

	// Raid Buffs
	if raidBuffs.ArcaneBrilliance {
		MakePermanent(ArcaneBrillianceAura(char))
	}

	if raidBuffs.DivineSpirit != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(DivineSpiritAura(char, IsImproved(raidBuffs.DivineSpirit)))
	}

	if raidBuffs.GiftOfTheWild != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(GiftOfTheWildAura(char, IsImproved(raidBuffs.GiftOfTheWild)))
	}

	if raidBuffs.PowerWordFortitude != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(PowerWordFortitudeAura(char, IsImproved(raidBuffs.PowerWordFortitude)))
	}

	if raidBuffs.ShadowProtection {
		MakePermanent(ShadowProtectionAura(char))
	}

	if raidBuffs.Bloodlust {
		registerBloodlustCD(char)
	}

	// Party Buffs
	if partyBuffs.AtieshDruid > 0 {
		MakePermanent(AtieshAura(char, proto.Class_ClassDruid, float64(partyBuffs.AtieshDruid)))
	}

	if partyBuffs.AtieshMage > 0 {
		MakePermanent(AtieshAura(char, proto.Class_ClassMage, float64(partyBuffs.AtieshMage)))
	}

	if partyBuffs.AtieshPriest > 0 {
		MakePermanent(AtieshAura(char, proto.Class_ClassPriest, float64(partyBuffs.AtieshPriest)))
	}

	if partyBuffs.AtieshWarlock > 0 {
		MakePermanent(AtieshAura(char, proto.Class_ClassWarlock, float64(partyBuffs.AtieshWarlock)))
	}

	if partyBuffs.BattleShout != proto.TristateEffect_TristateEffectMissing {
		boomingVoicePoints := int32(5)
		aura := BattleShoutAura(
			char,
			false,
			boomingVoicePoints,
			TernaryFloat64(IsImproved(partyBuffs.BattleShout), 1.25, 1.0),
			partyBuffs.BsSolarianSapphire,
			false,
		)

		ApplyFixedUptimeAura(aura, 1, aura.Duration+1, -1)
	}

	if partyBuffs.BloodPact != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(BloodPactAura(char, IsImproved(partyBuffs.BloodPact)))
	}

	if partyBuffs.BraidedEterniumChain {
		MakePermanent(BraidedEterniumChainAura(char))
	}

	if partyBuffs.ChainOfTheTwilightOwl {
		MakePermanent(ChainOfTheTwilightOwlAura(char))
	}

	if partyBuffs.CommandingShout != proto.TristateEffect_TristateEffectMissing {
		boomingVoicePoints := int32(5)

		aura := MakePermanent(CommandingShoutAura(
			char,
			false,
			boomingVoicePoints,
			TernaryFloat64(IsImproved(partyBuffs.CommandingShout), 1.25, 1.0),
			false,
		))

		ApplyFixedUptimeAura(aura, 1, aura.Duration, -1)
	}

	if partyBuffs.DevotionAura != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(DevotionAuraBuff(char, IsImproved(partyBuffs.DevotionAura)))
	}

	if partyBuffs.DraeneiRacialCaster {
		DraneiRacialAura(char, true)
	}

	if partyBuffs.DraeneiRacialMelee {
		DraneiRacialAura(char, false)
	}

	if partyBuffs.EyeOfTheNight {
		MakePermanent(EyeOfTheNightAura(char))
	}

	if partyBuffs.FerociousInspiration > 0 {
		MakePermanent(FerociousInspiration(char, partyBuffs.FerociousInspiration))
	}

	if partyBuffs.GraceOfAirTotem != proto.TristateEffect_TristateEffectMissing {
		GraceOfAirTotemAura(char, IsImproved(partyBuffs.GraceOfAirTotem), partyBuffs.TotemTwisting)
	}

	if partyBuffs.JadePendantOfBlasting {
		MakePermanent(JadePendantOfBlastingAura(char))
	}

	if partyBuffs.LeaderOfThePack != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(LeaderOfThePackAura(char, IsImproved(partyBuffs.LeaderOfThePack)))
	}

	if partyBuffs.ManaSpringTotem != proto.TristateEffect_TristateEffectMissing {
		ManaSpringTotemAura(char, IsImproved(partyBuffs.ManaSpringTotem))
	}

	if partyBuffs.ManaTideTotems > 0 {
		registerManaTideTotemCD(char, partyBuffs.ManaTideTotems)
	}

	if partyBuffs.MoonkinAura != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(MoonkinAuraBuff(char, IsImproved(partyBuffs.MoonkinAura)))
	}

	if partyBuffs.RetributionAura != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(RetributionAuraBuff(char, IsImproved(partyBuffs.RetributionAura), 5))
	}

	if partyBuffs.SanctityAura != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(SanctityAuraBuff(char, IsImproved(partyBuffs.SanctityAura)))
	}

	if partyBuffs.StrengthOfEarthTotem != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(StrengthOfEarthTotemAura(char, TernaryInt32(IsImproved(partyBuffs.StrengthOfEarthTotem), 2, 0), partyBuffs.SoeEnhancement_2Pt4))
	}

	if partyBuffs.TotemOfWrath > 0 {
		MakePermanent(TotemOfWrathAura(char, partyBuffs.TotemOfWrath))
	}

	if partyBuffs.TranquilAirTotem {
		MakePermanent(TranquilAirTotemAura(char))
	}

	if partyBuffs.TrueshotAura {
		MakePermanent(TrueShotAuraBuff(char))
	}

	if partyBuffs.WindfuryTotem != proto.TristateEffect_TristateEffectMissing {
		WindfuryTotemAura(char, IsImproved(partyBuffs.WindfuryTotem))
	}

	if partyBuffs.WrathOfAirTotem != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(WrathOfAirTotemAura(char, IsImproved(partyBuffs.WrathOfAirTotem)))
	}
	if partyBuffs.Drums > 0 {
		DrumsBuff(char, partyBuffs.Drums)
	}

	// Individual Buffs
	if individual.BlessingOfKings {
		MakePermanent(BlessingOfKingsAura(char))
	}

	if individual.BlessingOfMight != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(BlessingOfMightAura(char, IsImproved(individual.BlessingOfMight)))
	}

	if individual.BlessingOfSalvation {
		MakePermanent(BlessingOfSalvationAura(char))
	}

	if individual.BlessingOfSanctuary {
		MakePermanent(BlessingOfSanctuaryAura(char))
	}

	if individual.BlessingOfWisdom != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(BlessingOfWisdomAura(char, IsImproved(individual.BlessingOfWisdom)))
	}

	if individual.Innervates > 0 {
		registerInnervateCD(char, individual.Innervates)
	}

	if individual.PowerInfusions > 0 {
		registerPowerInfusionCD(char, individual.PowerInfusions)
	}

	if individual.ShadowPriestDps > 0 {
		MakePermanent(ShadowPriestDPSManaAura(char, float64(individual.ShadowPriestDps)))
	}

	if individual.UnleashedRage {
		MakePermanent(UnleashedRageAura(char, -1, 5))
	}

}

///////////////////////////////////////////////////////////////////////////
//							Raid Buffs
///////////////////////////////////////////////////////////////////////////

func ThornsAura(char *Character, points int32) *Aura {
	actionID := ActionID{SpellID: 26992}

	procSpell := char.RegisterSpell(SpellConfig{
		ActionID:    actionID,
		SpellSchool: SpellSchoolNature,
		Flags:       SpellFlagBinary,
		ProcMask:    ProcMaskEmpty,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			baseDamage := 25 * (1 + 0.25*float64(points))
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeAlwaysHit)
		},
	})

	return char.MakeProcTriggerAura(ProcTrigger{
		Name:     "Thorns",
		ActionID: actionID,
		Duration: time.Minute * 10,
		Outcome:  OutcomeLanded,
		Callback: CallbackOnSpellHitTaken,
		Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
			if spell.SpellSchool.Matches(SpellSchoolPhysical) {
				procSpell.Cast(sim, spell.Unit)
			}
		},
	}).AttachMultiplicativePseudoStatBuff(&char.PseudoStats.BonusPhysicalDamageTaken, -80)
}

func ArcaneBrillianceAura(char *Character) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:    "Arcane Brilliance",
		ActionID: ActionID{SpellID: 27127},
		Stats: []StatConfig{
			{stats.Intellect, 40, false},
		},
	})
}

func DivineSpiritAura(char *Character, improved bool) *Aura {
	spiritBuff := stats.Stats{stats.Spirit: 50}

	dsSDStatDep := char.NewDynamicStatDependency(stats.Spirit, stats.SpellDamage, 0.1)
	dsHPStatDep := char.NewDynamicStatDependency(stats.Spirit, stats.HealingPower, 0.1)

	return char.GetOrRegisterAura(Aura{
		Label:      "Divine Spirit Buff",
		ActionID:   ActionID{SpellID: 25312},
		Duration:   time.Minute * 30,
		BuildPhase: CharacterBuildPhaseBuffs,

		OnGain: func(aura *Aura, sim *Simulation) {
			char.AddStatsDynamic(sim, spiritBuff)
			if improved {
				char.EnableBuildPhaseStatDep(sim, dsSDStatDep)
				char.EnableBuildPhaseStatDep(sim, dsHPStatDep)
			}
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			char.AddStatsDynamic(sim, spiritBuff.Invert())
			if improved {
				char.DisableBuildPhaseStatDep(sim, dsSDStatDep)
				char.DisableBuildPhaseStatDep(sim, dsHPStatDep)
			}
		},
	})
}

func GiftOfTheWildAura(char *Character, improved bool) *Aura {
	mod := 1.0
	if improved {
		mod = 1.35
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Gift of the Wild",
		ActionID: ActionID{SpellID: 26991},
		Stats: []StatConfig{
			{stats.Armor, 340 * mod, false},
			{stats.Stamina, 14 * mod, false},
			{stats.Strength, 14 * mod, false},
			{stats.Agility, 14 * mod, false},
			{stats.Intellect, 14 * mod, false},
			{stats.Spirit, 14 * mod, false},
			{stats.ArcaneResistance, 25 * mod, false},
			{stats.FireResistance, 25 * mod, false},
			{stats.FrostResistance, 25 * mod, false},
			{stats.NatureResistance, 25 * mod, false},
			{stats.ShadowResistance, 25 * mod, false},
		},
	})
}

func PowerWordFortitudeAura(char *Character, improved bool) *Aura {
	stat := 79.0
	if improved {
		stat *= 1.3
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Power Word: Fortitude",
		ActionID: ActionID{SpellID: 25389},
		Stats: []StatConfig{
			{stats.Stamina, stat, false},
		},
	})
}

func ShadowProtectionAura(char *Character) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:    "Shadow Protection",
		ActionID: ActionID{SpellID: 10958},
		Stats: []StatConfig{
			{stats.ShadowResistance, 60, false},
		},
	})
}

// /////////////////////////////////////////////////////////////////////////
//
//	Party Buffs
//
// /////////////////////////////////////////////////////////////////////////
var BattleShoutCategory = "BattleShout"

func GetBattleShoutValue(boomingVoicePoints int32, commandingPresenceMultiplier float64, hasSolarianSapphire bool, hasT2 bool, isPrepull bool) float64 {
	baseApBuff := 306.0
	apBuff := baseApBuff
	if isPrepull {
		if hasSolarianSapphire {
			apBuff += 70
		}
		if hasT2 {
			apBuff += 30
		}
	}
	return apBuff * commandingPresenceMultiplier
}

func BattleShoutAura(char *Character, isPlayer bool, boomingVoicePoints int32, commandingPresenceMultiplier float64, hasSolarianSapphire bool, hasT2 bool) *Aura {
	prepullApBuff := GetBattleShoutValue(boomingVoicePoints, commandingPresenceMultiplier, hasSolarianSapphire, hasT2, true)
	apBuff := GetBattleShoutValue(boomingVoicePoints, commandingPresenceMultiplier, hasSolarianSapphire, hasT2, false)

	var ee *ExclusiveEffect
	aura := char.GetOrRegisterAura(Aura{
		Label:      fmt.Sprintf("Battle Shout (%s)", Ternary(isPlayer, "Player", "External")),
		ActionID:   ActionID{SpellID: 2048}.WithTag(TernaryInt32(isPlayer, 0, 1)),
		Duration:   time.Duration(float64(time.Minute*2) * (1 + 0.1*float64(boomingVoicePoints))),
		BuildPhase: CharacterBuildPhaseBuffs,
		OnGain: func(aura *Aura, sim *Simulation) {
			ee.SetPriority(sim, TernaryFloat64(sim.CurrentTime > 0, apBuff, prepullApBuff))
		},
	})

	ee = aura.NewExclusiveEffect(BattleShoutCategory, true, ExclusiveEffect{
		Priority: 0,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.AttackPower, ee.Priority)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.AttackPower, -ee.Priority)
		},
	})

	return aura
}

func BloodPactAura(char *Character, improved bool) *Aura {
	stamBuff := 70.0
	if improved {
		stamBuff *= 1.3
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Blood Pact",
		ActionID: ActionID{SpellID: 27268},
		Stats: []StatConfig{
			{stats.Stamina, stamBuff, false},
		},
	})
}

var CommandingShoutCategory = "CommandingShout"

func GetCommandingShoutValue(boomingVoicePoints int32, commandingPresenceMultiplier float64, hasT6Tank2P bool, isPrepull bool) float64 {
	baseHpBuff := 1080.0
	if isPrepull {
		if hasT6Tank2P {
			baseHpBuff += 170
		}
	}
	return baseHpBuff * commandingPresenceMultiplier
}

func CommandingShoutAura(char *Character, isPlayer bool, boomingVoicePoints int32, commandingPresenceMultiplier float64, hasT6Tank2P bool) *Aura {
	prepullHpBuff := GetCommandingShoutValue(boomingVoicePoints, commandingPresenceMultiplier, hasT6Tank2P, true)
	hpBuff := GetCommandingShoutValue(boomingVoicePoints, commandingPresenceMultiplier, hasT6Tank2P, false)

	var ee *ExclusiveEffect
	aura := char.GetOrRegisterAura(Aura{
		Label:      fmt.Sprintf("Commanding Shout (%s)", Ternary(isPlayer, "Player", "External")),
		ActionID:   ActionID{SpellID: 469}.WithTag(TernaryInt32(isPlayer, 0, 1)),
		Duration:   time.Duration(float64(time.Minute*2) * (1 + 0.1*float64(boomingVoicePoints))),
		BuildPhase: CharacterBuildPhaseBuffs,
		OnGain: func(aura *Aura, sim *Simulation) {
			ee.SetPriority(sim, TernaryFloat64(sim.CurrentTime > 0, hpBuff, prepullHpBuff))
		},
	})

	ee = aura.NewExclusiveEffect(CommandingShoutCategory, true, ExclusiveEffect{
		Priority: 0,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.Health, ee.Priority)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.Health, -ee.Priority)
		},
	})

	return aura
}

func DevotionAuraBuff(char *Character, improved bool) *Aura {
	armorBuff := 861.0
	if improved {
		armorBuff *= 1.40
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Devotion Aura",
		ActionID: ActionID{SpellID: 27149},
		Stats: []StatConfig{
			{stats.Armor, armorBuff, false},
		},
	})
}

func FerociousInspiration(char *Character, count int32) *Aura {
	dmgBuff := 0.03 * float64(count)

	return char.GetOrRegisterAura(Aura{
		Label:    "Ferocious Inspiration",
		ActionID: ActionID{SpellID: 34460},
		Duration: time.Second * 10,
	}).AttachMultiplicativePseudoStatBuff(&char.PseudoStats.DamageDealtMultiplier, 1+dmgBuff)
}

func LeaderOfThePackAura(char *Character, improved bool) *Aura {
	statsConfig := []StatConfig{
		{stats.PhysicalCritPercent, 5, false},
	}

	if improved {
		statsConfig = append(statsConfig, StatConfig{stats.MeleeCritRating, 20, false})
		statsConfig = append(statsConfig, StatConfig{stats.SpellCritRating, 20, false})
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Leader of the Pack",
		ActionID: ActionID{SpellID: 17007},
		Stats:    statsConfig,
	})
}

func MoonkinAuraBuff(char *Character, improved bool) *Aura {
	statsConfig := []StatConfig{
		{stats.SpellCritPercent, 5, false},
	}
	if improved {
		statsConfig = append(statsConfig, StatConfig{stats.SpellCritRating, 20, false})
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Moonkin Aura",
		ActionID: ActionID{SpellID: 24907},
		Stats:    statsConfig,
	})
}

func RetributionAuraBuff(char *Character, improved bool, points int32) *Aura {
	actionID := ActionID{SpellID: 27150}

	procSpell := char.RegisterSpell(SpellConfig{
		ActionID:    actionID,
		SpellSchool: SpellSchoolHoly,
		Flags:       SpellFlagBinary,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			baseDamage := 26 * (1 + 0.25*float64(points))
			if improved {
				baseDamage *= 1.50
			}
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeAlwaysHit)
			spell.DealDamage(sim, result)
		},
	})

	return MakePermanent(char.RegisterAura(Aura{
		Label:    "Retribution Aura",
		ActionID: actionID,
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if result.Landed() && spell.SpellSchool == SpellSchoolPhysical {
				procSpell.Cast(sim, spell.Unit)
			}
		},
	}))
}

func SanctityAuraBuff(char *Character, improved bool) *Aura {
	aura := MakePermanent(char.GetOrRegisterAura(Aura{
		Label:    "Sanctity Aura",
		ActionID: ActionID{SpellID: 20218},
	}).AttachMultiplicativePseudoStatBuff(&char.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexHoly], 1.1))

	if improved {
		aura.AttachMultiplicativePseudoStatBuff(&char.PseudoStats.DamageDealtMultiplier, 1.02)
	}

	return aura
}

func TrueShotAuraBuff(char *Character) *Aura {
	apBuff := 125.0

	return makeStatBuff(char, BuffConfig{
		Label:    "Trueshot Aura",
		ActionID: ActionID{SpellID: 27066},
		Stats: []StatConfig{
			{stats.RangedAttackPower, apBuff, false},
			{stats.AttackPower, apBuff, false},
		},
	})
}

var UnleashedRageCategory = "UnleashedRage"

func UnleashedRageAura(char *Character, casterIdx int32, points int32) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:    "Unleashed Rage",
		Duration: time.Second * 10,
		ActionID: ActionID{SpellID: 30809}.WithTag(casterIdx),
		Stats: []StatConfig{
			{stats.AttackPower, 1 + 0.02*float64(points), true},
		},
		ExclusiveCategory: UnleashedRageCategory,
	})
}

// //////////////////////////
//
//	Totems
//
// //////////////////////////
var GraceOfAirTotemCategory = "GraceOfAirTotem"

func GraceOfAirTotemAura(char *Character, improved bool, wfActive bool) *Aura {
	agiBuff := 77.0
	if improved {
		agiBuff *= 1.15
	}

	duration := NeverExpires
	if wfActive {
		duration = time.Second * 9
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Grace of Air Totem",
		ActionID: ActionID{SpellID: 25359},
		Stats: []StatConfig{
			{stats.Agility, agiBuff, false},
		},
		Duration:          duration,
		ExclusiveCategory: GraceOfAirTotemCategory,
	}).ApplyOnReset(func(aura *Aura, sim *Simulation) {
		if wfActive {
			StartPeriodicAction(sim, PeriodicActionOptions{
				Period:   time.Second * 10,
				Priority: ActionPriorityAuto,
				OnAction: func(sim *Simulation) {
					aura.Activate(sim)
				},
			})
		} else {
			aura.Activate(sim)
		}
	})
}

var ManaSpringTotemCategory = "ManaSpringTotem"

func ManaSpringTotemAura(char *Character, improved bool) *Aura {
	mp5Buff := 50.0
	if improved {
		mp5Buff *= 1.25
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Mana Spring Totem",
		ActionID: ActionID{SpellID: 25570},
		Stats: []StatConfig{
			{stats.MP5, mp5Buff, false},
		},
		ExclusiveCategory: ManaSpringTotemCategory,
	})
}

const (
	StrengthOfEarthTotemCategory      = "StrengthOfEarthTotem"
	StrengthOfEarthTotemBaseValue     = 86.0
	StrengthOfEarthTotemImprovedValue = 12.0
)

var StrengthOfEarthMultipliers = []float64{1, 1.08, 1.15}

func StrengthOfEarthTotemValue(enhancingTotemsPoints int32, hasEnh2pT4 bool) float64 {
	return (86.0 + TernaryFloat64(hasEnh2pT4, StrengthOfEarthTotemImprovedValue, 0)) * StrengthOfEarthMultipliers[enhancingTotemsPoints]
}

func StrengthOfEarthTotemAura(char *Character, enhancingTotemsPoints int32, hasEnh2pT4 bool) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:    "Strength of Earth Totem",
		ActionID: ActionID{SpellID: 25528},
		Stats: []StatConfig{
			{stats.Strength, StrengthOfEarthTotemValue(enhancingTotemsPoints, hasEnh2pT4), false},
		},
		ExclusiveCategory: StrengthOfEarthTotemCategory,
	})
}

func TotemOfWrathAura(char *Character, count int32) *Aura {
	modValue := 3.0 * float64(count)

	return makeStatBuff(char, BuffConfig{
		Label:    "Totem of Wrath",
		ActionID: ActionID{SpellID: 30706},
		Stats: []StatConfig{
			{stats.SpellCritPercent, modValue, false},
			{stats.SpellHitPercent, modValue, false},
		},
	})
}

func TranquilAirTotemAura(char *Character) *Aura {
	return char.GetOrRegisterAura(Aura{
		Label:    "Tranquil Air Totem",
		ActionID: ActionID{SpellID: 25909},
	}).AttachMultiplicativePseudoStatBuff(&char.PseudoStats.ThreatMultiplier, 0.8)
}

var WindfuryTotemCategory = "WindfuryTotem"

func WindfuryTotemAura(char *Character, isImpoved bool) *Aura {
	apBonus := 445.0
	if isImpoved {
		apBonus *= 1.3
	}
	// Chance on MH Auto Attack to instantly attack with another AA with apBonus.
	// AP bonus lingers until 2 auto attacks are performed.
	// If procced from a normal auto, this consumes the buff almost instantly (server tick rate applies)
	// If procced from a "Next Auto" special (HS/Cleave), this can result in the aura lasting the entire 1.5s duration.

	wfProcAura := char.NewTemporaryStatsAura("Windfury Totem Proc", ActionID{SpellID: 25584}, stats.Stats{stats.AttackPower: apBonus}, time.Millisecond*1500)
	wfProcAura.MaxStacks = 2
	wfProcAura.AttachProcTrigger(ProcTrigger{
		Name:     "Windfury Attack",
		Callback: CallbackOnSpellHitDealt,
		ProcMask: ProcMaskMeleeMHAuto | ProcMaskMeleeOHAuto,
		// TriggerImmediately ommited for improved UI clarity (the timeline tick would be near invisible for MHAuto procs)
		Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
			if wfProcAura.IsActive() && !spell.ProcMask.Matches(ProcMaskMeleeSpecial) {
				wfProcAura.RemoveStack(sim)
				if wfProcAura.GetStacks() == 0 {
					wfProcAura.Deactivate(sim)
				}
			}
		},
	})

	var windfurySpell *Spell
	wfProcTrigger := char.MakeProcTriggerAura(ProcTrigger{
		Name:               "Windfury Totem Trigger",
		MetricsActionID:    ActionID{SpellID: 25580, Tag: -1},
		ProcChance:         0.2,
		Duration:           NeverExpires,
		Outcome:            OutcomeLanded,
		Callback:           CallbackOnSpellHitDealt,
		ProcMask:           ProcMaskMeleeMHAuto,
		ICD:                time.Millisecond * 1500,
		TriggerImmediately: true,
		Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
			wfProcAura.Activate(sim)
			if spell.ProcMask == ProcMaskMeleeMHAuto {
				wfProcAura.SetStacks(sim, 1)
			} else {
				wfProcAura.SetStacks(sim, 2)
			}
			char.AutoAttacks.MaybeReplaceMHSwing(sim, windfurySpell).Cast(sim, result.Target)
		},
	})

	wfAura := char.GetOrRegisterAura(Aura{
		Label:    "Windfury Totem",
		ActionID: ActionID{SpellID: 25587, Tag: -1},
		Duration: time.Second * 10,
	}).ApplyOnInit(func(aura *Aura, sim *Simulation) {
		config := *char.AutoAttacks.MHConfig()
		config.ActionID = config.ActionID.WithTag(25584)
		windfurySpell = char.GetOrRegisterSpell(config)
	}).ApplyOnReset(func(aura *Aura, sim *Simulation) {
		aura.Activate(sim)
		StartPeriodicAction(sim, PeriodicActionOptions{
			Period:   time.Second * 5,
			Priority: ActionPriorityAuto,
			OnAction: func(sim *Simulation) {
				aura.Activate(sim)
			},
		})
	})

	wfAura.NewExclusiveEffect(WindfuryTotemCategory, false, ExclusiveEffect{
		Priority: apBonus,
		OnGain: func(_ *ExclusiveEffect, sim *Simulation) {
			wfProcTrigger.Activate(sim)
		},
		OnExpire: func(_ *ExclusiveEffect, sim *Simulation) {
			wfProcTrigger.Deactivate(sim)
			wfAura.Deactivate(sim)
		},
	})

	char.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *Simulation, slot proto.ItemSlot) {
		wfAura.Deactivate(sim)
	})

	return wfProcTrigger
}

const (
	WrathOfAirTotemCategory      = "WrathOfAirTotem"
	WrathOfAirTotemBaseValue     = 101.0
	WrathOfAirTotemImprovedValue = 20.0
)

func WrathOfAirTotemValue(improved bool) float64 {
	return WrathOfAirTotemBaseValue + TernaryFloat64(improved, WrathOfAirTotemImprovedValue, 0)
}

func WrathOfAirTotemAura(char *Character, improved bool) *Aura {
	buff := WrathOfAirTotemValue(improved)

	return makeStatBuff(char, BuffConfig{
		Label:    "Wrath of Air Totem",
		ActionID: ActionID{SpellID: 3738},
		Stats: []StatConfig{
			{stats.SpellDamage, buff, false},
			{stats.HealingPower, buff, false},
		},
		ExclusiveCategory: WrathOfAirTotemCategory,
	})
}

////////////////////////////
//	Item Buffs
////////////////////////////

func AtieshAura(char *Character, class proto.Class, numStaves float64) *Aura {
	switch class {
	case proto.Class_ClassDruid:
		return makeStatBuff(char, BuffConfig{
			Label:    "Power of the Guardian - Druid",
			ActionID: ActionID{SpellID: 28145},
			Stats: []StatConfig{
				{stats.MP5, 11 * numStaves, false},
			},
		})
	case proto.Class_ClassMage:
		return makeStatBuff(char, BuffConfig{
			Label:    "Power of the Guardian - Mage",
			ActionID: ActionID{SpellID: 28142},
			Stats: []StatConfig{
				{stats.SpellCritRating, 28 * numStaves, false},
			},
		})
	case proto.Class_ClassPriest:
		return makeStatBuff(char, BuffConfig{
			Label:    "Power of the Guardian - Priest",
			ActionID: ActionID{SpellID: 28144},
			Stats: []StatConfig{
				{stats.HealingPower, 62 * numStaves, false},
			},
		})
	default: // Use warlock as default to satisfy compiler
		return makeStatBuff(char, BuffConfig{
			Label:    "Power of the Guardian - Warlock",
			ActionID: ActionID{SpellID: 28143},
			Stats: []StatConfig{
				{stats.SpellDamage, 33 * numStaves, false},
				{stats.HealingPower, 33 * numStaves, false},
			},
		})
	}

}

func BraidedEterniumChainAura(char *Character) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:             "Braided Eternium Chain",
		ActionID:          ActionID{SpellID: 31025},
		ExclusiveCategory: "Braided Eternium Chain",
		Stats: []StatConfig{
			{stats.MeleeCritRating, 28, false},
		},
	})
}

func ChainOfTheTwilightOwlAura(char *Character) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:             "Chain of the Twilight Owl",
		ActionID:          ActionID{SpellID: 31035},
		ExclusiveCategory: "Chain of the Twilight Owl",
		Stats: []StatConfig{
			{stats.SpellCritPercent, 2, false},
		},
	})
}

func DraneiRacialAura(char *Character, caster bool) *Aura {
	alliance := []proto.Race{
		proto.Race_RaceDraenei,
		proto.Race_RaceDwarf,
		proto.Race_RaceGnome,
		proto.Race_RaceHuman,
		proto.Race_RaceNightElf,
	}
	if !slices.Contains(alliance, char.Race) {
		return nil
	}
	var aura *Aura
	if caster {
		aura = makeStatBuff(char, BuffConfig{
			Label:    "Inspiring Presence",
			ActionID: ActionID{SpellID: 28878},
			Stats: []StatConfig{
				{stats.SpellHitPercent, 1, false},
			},
			ExclusiveCategory: "Inspiring Presence",
		})
	} else {
		aura = makeStatBuff(char, BuffConfig{
			Label:    "Heroic Presence",
			ActionID: ActionID{SpellID: 6562},
			Stats: []StatConfig{
				{stats.PhysicalHitPercent, 1, false},
			},
			ExclusiveCategory: "Heroic Presence",
		})
	}

	return MakePermanent(aura)
}

func EyeOfTheNightAura(char *Character) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:             "Eye of the Night",
		ActionID:          ActionID{SpellID: 31033},
		ExclusiveCategory: "Eye of the Night",
		Stats: []StatConfig{
			{stats.SpellDamage, 34, false},
		},
	})
}

func JadePendantOfBlastingAura(char *Character) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:             "Jade Pendant of Blasting",
		ActionID:          ActionID{SpellID: 25607},
		ExclusiveCategory: "Jade Pendant of Blasting",
		Stats: []StatConfig{
			{stats.SpellDamage, 15, false},
		},
	})
}

const TinnitusAuraLabel = "Tinnitus"

func drumsSpellConfig(character *Character, drum proto.Drums, isExternal bool) SpellConfig {
	var drumLabel string
	var drumStats stats.Stats
	var duration time.Duration
	var actionID ActionID
	switch drum {
	case proto.Drums_GreaterDrumsOfBattle, proto.Drums_LesserDrumsOfBattle:
		drumLabel = "Drums of Battle"
		drumStats = stats.Stats{stats.MeleeHasteRating: 80, stats.SpellHasteRating: 80}
		duration = time.Second * 30
		actionID = ActionID{SpellID: 35476}
	case proto.Drums_GreaterDrumsOfWar, proto.Drums_LesserDrumsOfWar:
		drumLabel = "Drums of War"
		drumStats = stats.Stats{stats.AttackPower: 60, stats.RangedAttackPower: 60, stats.SpellDamage: 30}
		duration = time.Second * 30
		actionID = ActionID{SpellID: 35475}
	case proto.Drums_GreaterDrumsOfRestoration, proto.Drums_LesserDrumsOfRestoration:
		drumLabel = "Drums of Restoration"
		drumStats = stats.Stats{stats.MP5: 200}
		duration = time.Second * 15
		actionID = ActionID{SpellID: 35478}
	}

	if isExternal {
		actionID = actionID.WithTag(-1)
		drumLabel = drumLabel + " (External)"
	}

	aura := character.NewTemporaryStatsAura(drumLabel, actionID, drumStats, duration)

	tinnitus := character.GetOrRegisterAura(Aura{
		Label:    TinnitusAuraLabel,
		ActionID: ActionID{SpellID: 369770},
		Duration: time.Minute * 2,
	})

	aura.ApplyOnGain(func(_ *Aura, sim *Simulation) {
		tinnitus.Activate(sim)
	})

	spellConfig := SpellConfig{
		ActionID: actionID,
		Flags:    SpellFlagNoOnCastComplete,
		ProcMask: ProcMaskEmpty,
		ExtraCastCondition: func(sim *Simulation, target *Unit) bool {
			if !character.HasActiveAura(TinnitusAuraLabel) {
				return true
			}
			return false
		},
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			if !character.HasActiveAura(TinnitusAuraLabel) {
				aura.Activate(sim)
			}
		},

		RelatedSelfBuff: aura.Aura,
	}

	return spellConfig
}

func DrumsBuff(character *Character, drum proto.Drums) {
	config := drumsSpellConfig(character, drum, true)
	config.Cast = CastConfig{
		CD: Cooldown{
			Timer:    character.NewTimer(),
			Duration: time.Minute * 2,
		},
	}
	spell := character.RegisterSpell(config)

	character.AddMajorCooldown(MajorCooldown{
		Spell:    spell,
		Type:     CooldownTypeDPS,
		Priority: CooldownPriorityDrums,
	})
}

///////////////////////////////////////////////////////////////////////////
//							Individual Buffs
///////////////////////////////////////////////////////////////////////////

func AmplifyMagicAura(char *Character, improved bool) *Aura {
	baseMod := 120.0
	if improved {
		baseMod *= 1.50
	}
	return char.GetOrRegisterAura(Aura{
		Label:    "Amplify Magic",
		ActionID: ActionID{SpellID: 33946},
		Duration: time.Minute * 10,

		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.BonusHealingTaken += baseMod * 2
			aura.Unit.PseudoStats.BonusPhysicalDamageTaken += baseMod
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.BonusHealingTaken -= baseMod * 2
			aura.Unit.PseudoStats.BonusPhysicalDamageTaken -= baseMod
		},
	})
}

func DampenMagicAura(char *Character, improved bool) *Aura {
	baseMod := 120.0
	if improved {
		baseMod *= 1.50
	}
	return char.GetOrRegisterAura(Aura{
		Label:    "Amplify Magic",
		ActionID: ActionID{SpellID: 33946},
		Duration: time.Minute * 10,

		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.BonusHealingTaken -= baseMod * 2
			aura.Unit.PseudoStats.BonusSpellDamageTaken -= baseMod
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.BonusHealingTaken += baseMod * 2
			aura.Unit.PseudoStats.BonusSpellDamageTaken += baseMod
		},
	})
}

// //////////////////////////
//
//	Blessings
//
// //////////////////////////
func BlessingOfKingsAura(char *Character) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:    "Blessing of Kings",
		ActionID: ActionID{SpellID: 20217},
		Stats: []StatConfig{
			{stats.Agility, 1.1, true},
			{stats.Strength, 1.1, true},
			{stats.Stamina, 1.1, true},
			{stats.Intellect, 1.1, true},
			{stats.Spirit, 1.1, true},
		},
	})
}

// func BlessingOfLight(char *Character) *Aura {
// 	return char.GetOrRegisterAura(Aura{
// 		Label:    "Blessing of Light",
// 		ActionID: ActionID{SpellID: 27145},
// 		Duration: time.Minute * 30,

// 		OnApplyEffects: func(aura *Aura, sim *Simulation, target *Unit, spell *Spell) {
// 			if spell.ProcMask != ProcMaskSpellHealing {
// 				return
// 			}

// 			if spell.Unit.ownerClass != proto.Class_ClassPaladin {
// 				return
// 			}

// 			// Keep an eye on if this changes in paladin.go
// 			// FlashOfLight = 2
// 			// HolyLight = 3
// 			if spell.ClassSpellMask != 2 || spell.ClassSpellMask != 3 {
// 				return
// 			}

// 			if spell.ClassSpellMask == 2 {
// 				spell.BonusSpellDamage += 185
// 			} else {
// 				spell.BonusSpellDamage += 580
// 			}
// 		},
// 	})
// }

func BlessingOfMightAura(char *Character, improved bool) *Aura {
	apBuff := 220.0
	if improved {
		apBuff *= 1.2
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Blessing Of Might",
		ActionID: ActionID{SpellID: 27141},
		Stats: []StatConfig{
			{stats.AttackPower, apBuff, false},
			{stats.RangedAttackPower, apBuff, false},
		},
	})
}

func BlessingOfSalvationAura(char *Character) *Aura {
	return char.GetOrRegisterAura(Aura{
		Label:    "Blessing Of Salvation",
		ActionID: ActionID{SpellID: 25895},
		Duration: time.Minute * 30,
	}).AttachMultiplicativePseudoStatBuff(&char.PseudoStats.ThreatMultiplier, 0.7)
}

func BlessingOfSanctuaryAura(char *Character) *Aura {
	actionID := ActionID{SpellID: 27169}

	procSpell := char.RegisterSpell(SpellConfig{
		ActionID:    actionID,
		SpellSchool: SpellSchoolHoly,
		Flags:       SpellFlagBinary,
		ProcMask:    ProcMaskEmpty,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealDamage(sim, target, 46, spell.OutcomeAlwaysHit)
		},
	})

	return char.MakeProcTriggerAura(ProcTrigger{
		Name:     "Blessing of Sanctuary",
		ActionID: actionID,
		Duration: time.Minute * 10,
		Outcome:  OutcomeBlock,
		Callback: CallbackOnSpellHitTaken,
		Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
			procSpell.Cast(sim, spell.Unit)
		},
	}).AttachMultiplicativePseudoStatBuff(&char.PseudoStats.BonusPhysicalDamageTaken, -80)
}

func BlessingOfWisdomAura(char *Character, improved bool) *Aura {
	mp5Buff := 41.0
	if improved {
		mp5Buff *= 1.20
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Blessing of Wisdom",
		ActionID: ActionID{SpellID: 25894},
		Stats: []StatConfig{
			{stats.MP5, mp5Buff, false},
		},
	})
}

////////////////////////////
//  Individual Buffs
////////////////////////////

func ShadowPriestDPSManaAura(char *Character, dps float64) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:    "Vampiric Touch",
		ActionID: ActionID{SpellID: 34914},
		Stats: []StatConfig{
			{stats.MP5, dps * 0.25, false},
		},
	})
}

////////////////////////////
//  Cooldowns
////////////////////////////

var PowerInfusionAuraTag = "PowerInfusion"

const PowerInfusionDuration = time.Second * 15
const PowerInfusionCD = time.Minute * 3

func registerPowerInfusionCD(char *Character, numPowerInfusions int32) {
	if numPowerInfusions == 0 {
		return
	}

	piAura := PowerInfusionAura(char, -1)

	registerExternalConsecutiveCDApproximation(
		char,
		externalConsecutiveCDApproximation{
			ActionID:         ActionID{SpellID: 10060, Tag: -1},
			AuraTag:          PowerInfusionAuraTag,
			CooldownPriority: CooldownPriorityDefault,
			AuraDuration:     PowerInfusionDuration,
			AuraCD:           PowerInfusionCD,
			Type:             CooldownTypeDPS,

			ShouldActivate: func(sim *Simulation, character *Character) bool {
				// Haste portion doesn't stack with Bloodlust, so prefer to wait.
				return !character.HasActiveAuraWithTag(BloodlustAuraTag)
			},
			AddAura: func(sim *Simulation, character *Character) { piAura.Activate(sim) },
		},
		numPowerInfusions)
}

func PowerInfusionAura(char *Character, actionTag int32) *Aura {
	actionID := ActionID{SpellID: 10060, Tag: actionTag}

	aura := char.GetOrRegisterAura(Aura{
		Label:    "PowerInfusion-" + actionID.String(),
		Tag:      PowerInfusionAuraTag,
		ActionID: actionID,
		Duration: PowerInfusionDuration,
	})

	aura.NewExclusiveEffect("ManaCost", true, ExclusiveEffect{
		Priority: -20,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			if ee.Aura.Unit.HasManaBar() {
				ee.Aura.Unit.PseudoStats.SpellCostPercentModifier -= 20
			}
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			if ee.Aura.Unit.HasManaBar() {
				ee.Aura.Unit.PseudoStats.SpellCostPercentModifier += 20
			}
		},
	})
	multiplyCastSpeedEffect(aura, 1.2)
	return aura
}

func multiplyCastSpeedEffect(aura *Aura, multiplier float64) *ExclusiveEffect {
	return aura.NewExclusiveEffect("MultiplyCastSpeed", false, ExclusiveEffect{
		Priority: multiplier,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, multiplier)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, 1/multiplier)
		},
	})
}

var InnervateAuraTag = "Innervate"

const InnervateDuration = time.Second * 20
const InnervateCD = time.Minute * 6

func InnervateManaThreshold(character *Character) float64 {
	if character.Class == proto.Class_ClassMage {
		// Mages burn mana really fast so they need a higher threshold.
		return character.MaxMana() * 0.7
	} else {
		return 1000
	}
}

func registerInnervateCD(char *Character, numInnervates int32) {
	if numInnervates == 0 {
		return
	}

	innervateThreshold := 0.0
	expectedManaPerInnervate := 0.0
	var innervateAura *Aura

	char.Env.RegisterPostFinalizeEffect(func() {
		innervateThreshold = InnervateManaThreshold(char)
		expectedManaPerInnervate = char.SpiritManaRegenPerSecond() * 5 * 20
		innervateAura = InnervateAura(char, expectedManaPerInnervate, -1)
	})

	registerExternalConsecutiveCDApproximation(
		char,
		externalConsecutiveCDApproximation{
			ActionID:         ActionID{SpellID: 29166, Tag: -1},
			AuraTag:          InnervateAuraTag,
			CooldownPriority: CooldownPriorityDefault,
			AuraDuration:     InnervateDuration,
			AuraCD:           InnervateCD,
			Type:             CooldownTypeMana,
			ShouldActivate: func(sim *Simulation, character *Character) bool {
				// Only cast innervate when very low on mana, to make sure all other mana CDs are prioritized.
				if character.CurrentMana() > innervateThreshold {
					return false
				}
				return true
			},
			AddAura: func(sim *Simulation, character *Character) {
				innervateAura.Activate(sim)

				// newRemainingUsages := int(sim.GetRemainingDuration() / InnervateCD)
				// AddInnervateAura already accounts for 1 usage, which is why we subtract 1 less.
				// character.ExpectedBonusMana -= expectedManaPerInnervate * MaxFloat(0, float64(remainingInnervateUsages-newRemainingUsages-1))
				// remainingInnervateUsages = newRemainingUsages

			},
		},
		numInnervates)
}

func InnervateAura(character *Character, expectedBonusManaReduction float64, actionTag int32) *Aura {
	actionID := ActionID{SpellID: 29166, Tag: actionTag}
	manaMetrics := character.NewManaMetrics(actionID)
	return character.GetOrRegisterAura(Aura{
		Label:    "Innervate-" + actionID.String(),
		Tag:      InnervateAuraTag,
		ActionID: actionID,
		Duration: InnervateDuration,
		OnGain: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.ForceFullSpiritRegen = true
			character.PseudoStats.SpiritRegenMultiplier *= 5.0
			character.UpdateManaRegenRates()

			expectedBonusManaPerTick := expectedBonusManaReduction / 10
			StartPeriodicAction(sim, PeriodicActionOptions{
				Period:   InnervateDuration / 10,
				NumTicks: 10,
				OnAction: func(sim *Simulation) {
					manaMetrics.AddEvent(expectedBonusManaPerTick, expectedBonusManaPerTick)
				},
			})
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			character.PseudoStats.ForceFullSpiritRegen = false
			character.PseudoStats.SpiritRegenMultiplier /= 5.0
			character.UpdateManaRegenRates()
		},
	})
}

func InspirationAura(unit *Unit, points int32) *Aura {
	multiplier := 1 + []float64{0, .08, .16, .25}[points]

	armorDep := unit.NewDynamicMultiplyStat(stats.Armor, multiplier)

	return unit.GetOrRegisterAura(Aura{
		Label:    "Inspiration",
		ActionID: ActionID{SpellID: 15363},
		Duration: time.Second * 15,
	}).AttachStatDependency(armorDep)
}

func ApplyInspiration(character *Character, uptime float64) {
	if uptime <= 0 {
		return
	}
	uptime = min(1, uptime)

	inspirationAura := InspirationAura(&character.Unit, 3)

	ApplyFixedUptimeAura(inspirationAura, uptime, time.Millisecond*2500, 1)
}

// Applies buffs to pets.
func applyPetBuffEffects(petAgent PetAgent, raidBuffs *proto.RaidBuffs, partyBuffs *proto.PartyBuffs, individualBuffs *proto.IndividualBuffs) {
	// Summoned pets, like Mage Water Elemental, aren't around to receive raid buffs.
	if petAgent.GetPet().IsGuardian() {
		return
	}

	// We need to modify the buffs a bit because some things are applied to pets by
	// the owner during combat (Bloodlust) or don't make sense for a pet.
	raidBuffs = googleProto.Clone(raidBuffs).(*proto.RaidBuffs)
	raidBuffs.Bloodlust = false
	raidBuffs.Thorns = proto.TristateEffect_TristateEffectMissing

	partyBuffs = googleProto.Clone(partyBuffs).(*proto.PartyBuffs)
	// Pets can't get extra attacks, doh!
	partyBuffs.WindfuryTotem = proto.TristateEffect_TristateEffectMissing

	individualBuffs = googleProto.Clone(individualBuffs).(*proto.IndividualBuffs)
	individualBuffs.Innervates = 0
	individualBuffs.PowerInfusions = 0

	// Pets don't benefit from buffs that are ratings, e.g. crit rating or haste rating.
	partyBuffs.Drums = proto.Drums_DrumsUnknown
	partyBuffs.LeaderOfThePack = MinTristate(partyBuffs.LeaderOfThePack, proto.TristateEffect_TristateEffectRegular)
	partyBuffs.MoonkinAura = MinTristate(partyBuffs.MoonkinAura, proto.TristateEffect_TristateEffectRegular)
	partyBuffs.BraidedEterniumChain = false

	if !petAgent.GetPet().enabledOnStart {
		// Auras etc still apply, but not targeted buffs (usually)
		partyBuffs.ChainOfTheTwilightOwl = false
		partyBuffs.EyeOfTheNight = false
		partyBuffs.JadePendantOfBlasting = false

		// Strip targeted buffs that require presence at fight start
		raidBuffs.ArcaneBrilliance = false
		raidBuffs.DivineSpirit = proto.TristateEffect_TristateEffectMissing
		raidBuffs.GiftOfTheWild = proto.TristateEffect_TristateEffectMissing
		raidBuffs.PowerWordFortitude = proto.TristateEffect_TristateEffectMissing
		raidBuffs.ShadowProtection = false
		raidBuffs.Thorns = proto.TristateEffect_TristateEffectMissing
		individualBuffs.BlessingOfMight = proto.TristateEffect_TristateEffectMissing
		individualBuffs.BlessingOfKings = false
		individualBuffs.BlessingOfWisdom = proto.TristateEffect_TristateEffectMissing

		// Only individual buff that would apply is Unleashed Rage.
		unleashedRage := individualBuffs.UnleashedRage
		individualBuffs = &proto.IndividualBuffs{}
		individualBuffs.UnleashedRage = unleashedRage
	}

	applyBuffEffects(petAgent, raidBuffs, partyBuffs, individualBuffs)
}

// Used for approximating cooldowns applied by other players to you, such as
// bloodlust, innervate, power infusion, etc. This is specifically for buffs
// which can be consecutively applied multiple times to a single player.
type externalConsecutiveCDApproximation struct {
	ActionID         ActionID
	AuraTag          string
	CooldownPriority int32
	Type             CooldownType
	AuraDuration     time.Duration
	AuraCD           time.Duration

	// Callback for extra activation conditions.
	ShouldActivate CooldownActivationCondition

	// Applies the buff.
	AddAura           CooldownActivation
	RelatedSelfBuff   *Aura             // Used to attach the aura to the generic spell
	RelatedAuraArrays LabeledAuraArrays // Used to attach the aura to the generic spell
}

// numSources is the number of other players assigned to apply the buff to this player.
// E.g. the number of other shaman in the group using bloodlust.
func registerExternalConsecutiveCDApproximation(char *Character, config externalConsecutiveCDApproximation, numSources int32) {
	if numSources == 0 {
		panic("Need at least 1 source!")
	}

	var nextExternalIndex int

	externalTimers := make([]*Timer, numSources)
	for i := 0; i < int(numSources); i++ {
		externalTimers[i] = char.NewTimer()
	}
	sharedTimer := char.NewTimer()

	spell := char.RegisterSpell(SpellConfig{
		ActionID: config.ActionID,
		Flags:    SpellFlagNoOnCastComplete | SpellFlagNoMetrics | SpellFlagNoLogs,

		Cast: CastConfig{
			CD: Cooldown{
				Timer:    sharedTimer,
				Duration: config.AuraDuration, // Assumes that multiple buffs are different sources.
			},
		},
		ExtraCastCondition: func(sim *Simulation, target *Unit) bool {
			if !externalTimers[nextExternalIndex].IsReady(sim) {
				return false
			}

			if char.HasActiveAuraWithTag(config.AuraTag) {
				return false
			}

			return true
		},

		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			config.AddAura(sim, char)
			externalTimers[nextExternalIndex].Set(sim.CurrentTime + config.AuraCD)

			nextExternalIndex = (nextExternalIndex + 1) % len(externalTimers)

			if externalTimers[nextExternalIndex].IsReady(sim) {
				sharedTimer.Set(sim.CurrentTime + config.AuraDuration)
			} else {
				sharedTimer.Set(sim.CurrentTime + externalTimers[nextExternalIndex].TimeToReady(sim))
			}
		},
		RelatedSelfBuff:   config.RelatedSelfBuff,
		RelatedAuraArrays: config.RelatedAuraArrays,
	})

	char.AddMajorCooldown(MajorCooldown{
		Spell:    spell,
		Priority: config.CooldownPriority,
		Type:     config.Type,

		ShouldActivate: config.ShouldActivate,
	})
}

var BloodlustActionID = ActionID{SpellID: 2825}

const SatedAuraLabel = "Sated"
const BloodlustAuraTag = "Bloodlust"
const BloodlustDuration = time.Second * 40
const BloodlustCD = time.Minute * 10

func registerBloodlustCD(character *Character) {
	bloodlustAura := BloodlustAura(character, -1)

	spell := character.RegisterSpell(SpellConfig{
		ActionID: bloodlustAura.ActionID,
		Flags:    SpellFlagAPL | SpellFlagNoOnCastComplete | SpellFlagNoMetrics | SpellFlagNoLogs,

		Cast: CastConfig{
			CD: Cooldown{
				Timer:    character.NewTimer(),
				Duration: BloodlustCD,
			},
		},

		ApplyEffects: func(sim *Simulation, target *Unit, _ *Spell) {
			if !target.HasActiveAura(SatedAuraLabel) {
				bloodlustAura.Activate(sim)
			}
		},

		RelatedSelfBuff: bloodlustAura,
	})

	character.AddMajorCooldown(MajorCooldown{
		Spell:    spell,
		Priority: CooldownPriorityBloodlust,
		Type:     CooldownTypeDPS,
		ShouldActivate: func(sim *Simulation, character *Character) bool {
			// Haste portion doesn't stack with Power Infusion, so prefer to wait.
			return !character.HasActiveAuraWithTag(PowerInfusionAuraTag) && !character.HasActiveAura(SatedAuraLabel)
		},
	})
}

func BloodlustAura(character *Character, actionTag int32) *Aura {
	actionID := BloodlustActionID.WithTag(actionTag)

	sated := character.GetOrRegisterAura(Aura{
		Label:    SatedAuraLabel,
		ActionID: ActionID{SpellID: 57724},
		Duration: time.Minute * 10,
	})

	for _, pet := range character.Pets {
		if !pet.IsGuardian() {
			BloodlustAura(&pet.Character, actionTag)
		}
	}

	aura := character.GetOrRegisterAura(Aura{
		Label:    "Bloodlust-" + actionID.String(),
		Tag:      BloodlustAuraTag,
		ActionID: actionID,
		Duration: BloodlustDuration,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.MultiplyAttackSpeed(sim, 1.3)
			for _, pet := range character.Pets {
				if pet.IsEnabled() && !pet.IsGuardian() {
					pet.GetAura(aura.Label).Activate(sim)
				}
			}
			sated.Activate(sim)
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.MultiplyAttackSpeed(sim, 1/1.3)
		},
	})
	multiplyCastSpeedEffect(aura, 1.3)
	return aura
}

var PainSuppressionAuraTag = "PainSuppression"

const PainSuppressionDuration = time.Second * 8
const PainSuppressionCD = time.Minute * 3

func registerPainSuppressionCD(char *Character, numPainSuppressions int32) {
	if numPainSuppressions == 0 {
		return
	}

	psAura := PainSuppressionAura(char, -1)

	registerExternalConsecutiveCDApproximation(
		char,
		externalConsecutiveCDApproximation{
			ActionID:         ActionID{SpellID: 33206, Tag: -1},
			AuraTag:          PainSuppressionAuraTag,
			CooldownPriority: CooldownPriorityDefault,
			RelatedSelfBuff:  psAura,
			AuraDuration:     PainSuppressionDuration,
			AuraCD:           PainSuppressionCD,
			Type:             CooldownTypeSurvival,

			ShouldActivate: func(sim *Simulation, character *Character) bool {
				return true
			},
			AddAura: func(sim *Simulation, character *Character) {
				psAura.Activate(sim)
			},
		},
		numPainSuppressions)
}

func PainSuppressionAura(character *Character, actionTag int32) *Aura {
	actionID := ActionID{SpellID: 33206, Tag: actionTag}

	return character.GetOrRegisterAura(Aura{
		Label:    "PainSuppression-" + actionID.String(),
		Tag:      PainSuppressionAuraTag,
		ActionID: actionID,
		Duration: PainSuppressionDuration,
	}).AttachMultiplicativePseudoStatBuff(&character.PseudoStats.DamageTakenMultiplier, 0.6)
}

var ManaTideTotemActionID = ActionID{SpellID: 16190}
var ManaTideTotemAuraTag = "ManaTideTotem"

const ManaTideTotemDuration = time.Second * 12
const ManaTideTotemCD = time.Minute * 5

func registerManaTideTotemCD(char *Character, numManaTideTotems int32) {
	if numManaTideTotems == 0 {
		return
	}

	initialDelay := time.Duration(0)
	var mttAura *Aura

	mttAura = ManaTideTotemAura(char, -1)

	char.Env.RegisterPostFinalizeEffect(func() {
		// Use first MTT at 60s, or halfway through the fight, whichever comes first.
		initialDelay = min(char.Env.BaseDuration/2, time.Second*60)
	})

	registerExternalConsecutiveCDApproximation(
		char,
		externalConsecutiveCDApproximation{
			ActionID:         ManaTideTotemActionID.WithTag(-1),
			AuraTag:          ManaTideTotemAuraTag,
			CooldownPriority: CooldownPriorityDefault,
			RelatedSelfBuff:  mttAura,
			AuraDuration:     ManaTideTotemDuration,
			AuraCD:           ManaTideTotemCD,
			Type:             CooldownTypeMana,
			ShouldActivate: func(sim *Simulation, character *Character) bool {
				// A normal resto shaman would wait to use MTT.
				return sim.CurrentTime >= initialDelay
			},
			AddAura: func(sim *Simulation, character *Character) {
				mttAura.Activate(sim)
			},
		},
		numManaTideTotems)
}

func ManaTideTotemAura(character *Character, actionTag int32) *Aura {
	actionID := ManaTideTotemActionID.WithTag(actionTag)
	dep := character.NewDynamicMultiplyStat(stats.Spirit, 2)
	return character.GetOrRegisterAura(Aura{
		Label:    "ManaTideTotem-" + actionID.String(),
		Tag:      ManaTideTotemAuraTag,
		ActionID: actionID,
		Duration: ManaTideTotemDuration,
	}).AttachStatDependency(dep)
}
