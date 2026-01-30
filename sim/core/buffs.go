package core

import (
	"slices"
	"time"

	googleProto "google.golang.org/protobuf/proto"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type BuffConfig struct {
	Label    string
	ActionID ActionID
	Stats    []StatConfig
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

func makeStatBuff(char *Character, config BuffConfig) *Aura {
	if config.Label == "" {
		panic("Buff without label.")
	}

	if ActionID.IsEmptyAction(config.ActionID) {
		panic("Buff without ActionID")
	}

	baseAura := MakePermanent(char.GetOrRegisterAura(Aura{
		Label:      config.Label,
		ActionID:   config.ActionID,
		BuildPhase: CharacterBuildPhaseBuffs,
	}))

	registerStatEffect(baseAura, config.Stats)
	return baseAura
}

// Applies buffs that affect individual players.
func applyBuffEffects(agent Agent, raidBuffs *proto.RaidBuffs, partyBuffs *proto.PartyBuffs, individual *proto.IndividualBuffs) {
	char := agent.GetCharacter()

	// Raid Buffs
	if raidBuffs.ArcaneBrilliance {
		ArcaneBrillianceAura(char)
	}

	if raidBuffs.DivineSpirit != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(DivineSpiritAura(char, IsImproved(raidBuffs.DivineSpirit)))
	}

	if raidBuffs.GiftOfTheWild != proto.TristateEffect_TristateEffectMissing {
		GiftOfTheWildAura(char, IsImproved(raidBuffs.GiftOfTheWild))
	}

	if raidBuffs.PowerWordFortitude != proto.TristateEffect_TristateEffectMissing {
		PowerWordFortitudeAura(char, IsImproved(raidBuffs.PowerWordFortitude))
	}

	if raidBuffs.ShadowProtection {
		ShadowProtectionAura(char)
	}

	// Party Buffs
	if partyBuffs.AtieshDruid > 0 {
		AtieshAura(char, proto.Class_ClassDruid.Enum(), float64(partyBuffs.AtieshDruid))
	}

	if partyBuffs.AtieshMage > 0 {
		AtieshAura(char, proto.Class_ClassMage.Enum(), float64(partyBuffs.AtieshMage))
	}

	if partyBuffs.AtieshPriest > 0 {
		AtieshAura(char, proto.Class_ClassPriest.Enum(), float64(partyBuffs.AtieshPriest))
	}

	if partyBuffs.AtieshWarlock > 0 {
		AtieshAura(char, proto.Class_ClassWarlock.Enum(), float64(partyBuffs.AtieshWarlock))
	}

	if partyBuffs.BattleShout != proto.TristateEffect_TristateEffectMissing {
		BattleShoutAura(char, IsImproved(partyBuffs.BattleShout), partyBuffs.BsSolarianSapphire)
	}

	if partyBuffs.BloodPact != proto.TristateEffect_TristateEffectMissing {
		BloodPactAura(char, IsImproved(partyBuffs.BloodPact))
	}

	if partyBuffs.BraidedEterniumChain {
		BraidedEterniumChainAura(char)
	}

	if partyBuffs.ChainOfTheTwilightOwl {
		ChainOfTheTwilightOwlAura(char)
	}

	if partyBuffs.CommandingShout != proto.TristateEffect_TristateEffectMissing {
		CommandingShoutAura(char, IsImproved(partyBuffs.CommandingShout))
	}

	if partyBuffs.DevotionAura != proto.TristateEffect_TristateEffectMissing {
		DevotionAuraBuff(char, IsImproved(partyBuffs.DevotionAura))
	}

	if partyBuffs.DraeneiRacialCaster {
		DraneiRacialAura(char, true)
	}

	if partyBuffs.DraeneiRacialMelee {
		DraneiRacialAura(char, false)
	}

	if partyBuffs.EyeOfTheNight {
		EyeOfTheNightAura(char)
	}

	if partyBuffs.FerociousInspiration > 0 {
		FerociousInspiration(char, partyBuffs.FerociousInspiration)
	}

	if partyBuffs.GraceOfAirTotem != proto.TristateEffect_TristateEffectMissing {
		GraceOfAirTotemAura(char, IsImproved(partyBuffs.GraceOfAirTotem))
	}

	if partyBuffs.JadePendantOfBlasting {
		JadePendantOfBlastingAura(char)
	}

	if partyBuffs.LeaderOfThePack != proto.TristateEffect_TristateEffectMissing {
		LeaderOfThePackAura(char, IsImproved(partyBuffs.LeaderOfThePack))
	}

	if partyBuffs.ManaSpringTotem != proto.TristateEffect_TristateEffectMissing {
		ManaSpringTotemAura(char, IsImproved(partyBuffs.ManaSpringTotem))
	}

	if partyBuffs.MoonkinAura != proto.TristateEffect_TristateEffectMissing {
		MoonkinAuraBuff(char, IsImproved(partyBuffs.MoonkinAura))
	}

	if partyBuffs.RetributionAura != proto.TristateEffect_TristateEffectMissing {
		RetributionAuraBuff(char, IsImproved(partyBuffs.RetributionAura), 5)
	}

	if partyBuffs.SanctityAura != proto.TristateEffect_TristateEffectMissing {
		SanctityAuraBuff(char, IsImproved(partyBuffs.SanctityAura))
	}

	if partyBuffs.StrengthOfEarthTotem != proto.StrengthOfEarthType_None {
		StrengthOfEarthTotemAura(char, partyBuffs.StrengthOfEarthTotem.Enum())
	}

	if partyBuffs.TotemOfWrath > 0 {
		TotemOfWrathAura(char, partyBuffs.TotemOfWrath)
	}

	if partyBuffs.TranquilAirTotem {
		TranquilAirTotemAura(char)
	}

	if partyBuffs.TrueshotAura {
		TrueShotAuraBuff(char)
	}

	if partyBuffs.WindfuryTotemRank > 0 && char.AutoAttacks.anyEnabled() {
		WindfuryTotemAura(char, partyBuffs.WindfuryTotemIwt)
	}

	if partyBuffs.WrathOfAirTotem != proto.TristateEffect_TristateEffectMissing {
		WrathOfAirTotemAura(char, IsImproved(partyBuffs.WrathOfAirTotem))
	}

	// Individual Buffs
	if individual.BlessingOfKings {
		BlessingOfKingsAura(char)
	}

	if individual.BlessingOfMight != proto.TristateEffect_TristateEffectMissing {
		BlessingOfMightAura(char, IsImproved(individual.BlessingOfMight))
	}

	if individual.BlessingOfSalvation {
		BlessingOfSalvationAura(char)
	}

	if individual.BlessingOfSanctuary {
		BlessingOfSanctuaryAura(char)
	}

	if individual.BlessingOfWisdom != proto.TristateEffect_TristateEffectMissing {
		BlessingOfWisdomAura(char, IsImproved(individual.BlessingOfWisdom))
	}

	if individual.Innervates > 0 {
		registerInnervateCD(char, individual.Innervates)
	}

	if individual.PowerInfusions > 0 {
		registerPowerInfusionCD(char, individual.PowerInfusions)
	}

	if individual.ShadowPriestDps > 0 {
		ShadowPriestDPSManaAura(char, float64(individual.ShadowPriestDps))
	}

	if individual.UnleashedRage {
		UnleashedRageAura(char)
	}

}

///////////////////////////////////////////////////////////////////////////
//							Raid Buffs
///////////////////////////////////////////////////////////////////////////

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

	dsSDStatDep := char.NewDynamicStatDependency(stats.Spirit, stats.SpellDamage, 1.1)
	dsHPStatDep := char.NewDynamicStatDependency(stats.Spirit, stats.HealingPower, 1.1)

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

///////////////////////////////////////////////////////////////////////////
//							Party Buffs
///////////////////////////////////////////////////////////////////////////

func BattleShoutAura(char *Character, improved bool, sapphire bool) *Aura {
	apBuff := 306.0
	if improved {
		apBuff *= 1.25
	}

	if sapphire {
		apBuff += 70
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Battle Shout",
		ActionID: ActionID{SpellID: 2048},
		Stats: []StatConfig{
			{stats.AttackPower, apBuff, false},
		},
	})
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

func CommandingShoutAura(char *Character, improved bool) *Aura {
	hpBuff := 1080.0
	if improved {
		hpBuff *= 1.25
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Commanding Shout",
		ActionID: ActionID{SpellID: 469},
		Stats: []StatConfig{
			{stats.Health, hpBuff, false},
		},
	})
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
		{stats.RangedCritPercent, 5, false},
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

	return char.RegisterAura(Aura{
		Label:    "Retribution Aura",
		ActionID: actionID,
		Duration: NeverExpires,
		OnReset: func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if result.Landed() && spell.SpellSchool == SpellSchoolPhysical {
				procSpell.Cast(sim, spell.Unit)
			}
		},
	})
}

func SanctityAuraBuff(char *Character, improved bool) *Aura {
	aura := char.GetOrRegisterAura(Aura{
		Label:    "Sanctity Aura",
		ActionID: ActionID{SpellID: 20218},
	}).AttachMultiplicativePseudoStatBuff(&char.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexHoly], 1.1)

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

func UnleashedRageAura(char *Character) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:    "Unleashed Rage",
		ActionID: ActionID{SpellID: 30809},
		Stats: []StatConfig{
			{stats.AttackPower, 1.1, true},
		},
	})
}

// //////////////////////////
//
//	Totems
//
// //////////////////////////
func GraceOfAirTotemAura(char *Character, improved bool) *Aura {
	agiBuff := 77.0
	if improved {
		agiBuff *= 1.15
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Grace of Air Totem",
		ActionID: ActionID{SpellID: 25359},
		Stats: []StatConfig{
			{stats.Agility, agiBuff, false},
		},
	})
}

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
	})
}

func StrengthOfEarthTotemAura(char *Character, totem *proto.StrengthOfEarthType) *Aura {
	strBuff := 86.0

	switch totem {
	case proto.StrengthOfEarthType_CycloneBonus.Enum(),
		proto.StrengthOfEarthType_EnhancingTotems.Enum():
		strBuff = 98
	case proto.StrengthOfEarthType_EnhancingAndCyclone.Enum():
		strBuff = 112
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Strength of Earth Totem",
		ActionID: ActionID{SpellID: 25528},
		Stats: []StatConfig{
			{stats.Strength, strBuff, false},
		},
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
		Duration: time.Minute * 2,
	}).AttachMultiplicativePseudoStatBuff(&char.PseudoStats.ThreatMultiplier, 0.8)
}

func WindfuryTotemAura(char *Character, iwtTalentPoints int32) *Aura {
	buffActionID := ActionID{SpellID: 25587}
	apBonus := 445.0
	apBonus *= 1 + 0.15*float64(iwtTalentPoints)

	var charges int32
	icd := Cooldown{
		Timer:    char.NewTimer(),
		Duration: 1,
	}

	wfBuffAura := char.NewTemporaryStatsAuraWrapped("Windfury Buff", buffActionID, stats.Stats{stats.AttackPower: apBonus}, time.Millisecond*1500, func(config *Aura) {
		config.OnSpellHitDealt = func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			// *Special Case* Windfury should not proc on Seal of Command
			if spell.ActionID.SpellID == 20424 {
				return
			}
			if !spell.ProcMask.Matches(ProcMaskMeleeWhiteHit) || spell.ProcMask.Matches(ProcMaskMeleeSpecial) {
				return
			}
			charges--
			if charges == 0 {
				aura.Deactivate(sim)
			}
		}
	})
	const procChance = 0.2
	var wfSpell *Spell

	return char.GetOrRegisterAura(Aura{
		Label:    "Windfury Totem",
		ActionID: ActionID{SpellID: 25587},
		OnInit: func(aura *Aura, sim *Simulation) {
			wfSpell = char.GetOrRegisterSpell(SpellConfig{
				ActionID:    buffActionID, // temporary buff ("Windfury Attack") spell id
				SpellSchool: SpellSchoolPhysical,
				Flags:       SpellFlagMeleeMetrics | SpellFlagNoOnCastComplete,

				ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
					wfSwing := char.AutoAttacks.MHAuto()
					wfSwing.BonusSpellDamage = 445
					wfSwing.Cast(sim, target)
				},
			})
		},
		OnReset: func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			// *Special Case* Windfury should not proc on Seal of Command
			if spell.ActionID.SpellID == 20424 {
				return
			}
			if !result.Landed() || !spell.ProcMask.Matches(ProcMaskMeleeMHAuto) {
				return
			}

			if wfBuffAura.IsActive() {
				return
			}
			if !icd.IsReady(sim) {
				// Checking for WF buff aura isn't quite enough now that we refactored auras.
				// TODO: Clean this up to remove the need for an instant ICD.
				return
			}

			if sim.RandomFloat("Windfury Totem") > procChance {
				return
			}

			// TODO: the current proc system adds auras after cast and damage, in game they're added after cast
			startCharges := int32(2)
			if !spell.ProcMask.Matches(ProcMaskMeleeMHSpecial) {
				startCharges--
			}
			charges = startCharges
			wfBuffAura.Activate(sim)
			icd.Use(sim)

			aura.Unit.AutoAttacks.MaybeReplaceMHSwing(sim, wfSpell).Cast(sim, result.Target)
		},
	})
}

func WrathOfAirTotemAura(char *Character, improved bool) *Aura {
	buff := 101.0
	if improved {
		buff += 20.0
	}

	return makeStatBuff(char, BuffConfig{
		Label:    "Wrath of Air Totem",
		ActionID: ActionID{SpellID: 3738},
		Stats: []StatConfig{
			{stats.SpellDamage, buff, false},
			{stats.HealingPower, buff, false},
		},
	})
}

////////////////////////////
//	Item Buffs
////////////////////////////

func AtieshAura(char *Character, class *proto.Class, numStaves float64) *Aura {
	switch class {
	case proto.Class_ClassDruid.Enum():
		return makeStatBuff(char, BuffConfig{
			Label:    "Power of the Guardian - Druid",
			ActionID: ActionID{SpellID: 28145},
			Stats: []StatConfig{
				{stats.MP5, 11 * numStaves, false},
			},
		})
	case proto.Class_ClassMage.Enum():
		return makeStatBuff(char, BuffConfig{
			Label:    "Power of the Guardian - Mage",
			ActionID: ActionID{SpellID: 28142},
			Stats: []StatConfig{
				{stats.SpellCritRating, 28 * numStaves, false},
			},
		})
	case proto.Class_ClassPriest.Enum():
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
		Label:    "Braided Eternium Chain",
		ActionID: ActionID{SpellID: 31025},
		Stats: []StatConfig{
			{stats.AllPhysCritRating, 28, false},
		},
	})
}

func ChainOfTheTwilightOwlAura(char *Character) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:    "Chain of the Twilight Owl",
		ActionID: ActionID{SpellID: 31035},
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

	if caster {
		return makeStatBuff(char, BuffConfig{
			Label:    "Inspiring Presence",
			ActionID: ActionID{SpellID: 28878},
			Stats: []StatConfig{
				{stats.SpellHitPercent, 1, false},
			},
		})
	} else {
		return makeStatBuff(char, BuffConfig{
			Label:    "Heroic Presence",
			ActionID: ActionID{SpellID: 6562},
			Stats: []StatConfig{
				{stats.MeleeHitRating, 1, false},
				{stats.RangedHitPercent, 1, false},
			},
		})
	}
}

func EyeOfTheNightAura(char *Character) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:    "Eye of the Night",
		ActionID: ActionID{SpellID: 31033},
		Stats: []StatConfig{
			{stats.SpellDamage, 33, false},
		},
	})
}

func JadePendantOfBlastingAura(char *Character) *Aura {
	return makeStatBuff(char, BuffConfig{
		Label:    "Jade Pendant of Blasting",
		ActionID: ActionID{SpellID: 25607},
		Stats: []StatConfig{
			{stats.SpellDamage, 15, false},
		},
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

		// ApplyEffects: ApplyEffectFuncDirectDamage(SpellEffect{
		// 	ProcMask:         ProcMaskEmpty,
		// 	DamageMultiplier: 1,
		// 	ThreatMultiplier: 1,

		// 	BaseDamage:     BaseDamageConfigFlat(46),
		// 	OutcomeApplier: character.OutcomeFuncMagicHitBinary(),
		// }),
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealDamage(sim, target, 46, spell.OutcomeAlwaysHit)
		},
	})

	return char.RegisterAura(Aura{
		Label:    "Blessing of Sanctuary",
		ActionID: actionID,
		Duration: NeverExpires,
		OnReset: func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
		},
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.BonusPhysicalDamageTaken -= 80
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.BonusPhysicalDamageTaken += 80
		},
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if result.Outcome.Matches(OutcomeBlock) {
				procSpell.Cast(sim, spell.Unit)
			}
		},
	})
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

	return char.GetOrRegisterAura(Aura{
		Label:    "PowerInfusion-" + actionID.String(),
		Tag:      PowerInfusionAuraTag,
		ActionID: actionID,
		Duration: PowerInfusionDuration,
		OnGain: func(aura *Aura, sim *Simulation) {
			if char.HasManaBar() {
				// TODO: Double-check this is how the calculation works.
				char.PseudoStats.SpellCostPercentModifier *= 80

			}
			if !char.HasActiveAuraWithTag(BloodlustAuraTag) {
				char.MultiplyCastSpeed(sim, 1.2)
			}
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			if char.HasManaBar() {
				char.PseudoStats.SpellCostPercentModifier /= 80
			}
			if !char.HasActiveAuraWithTag(BloodlustAuraTag) {
				char.MultiplyCastSpeed(sim, 1/1.2)
			}
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

// Applies buffs to pets.
func applyPetBuffEffects(petAgent PetAgent, raidBuffs *proto.RaidBuffs, partyBuffs *proto.PartyBuffs, individualBuffs *proto.IndividualBuffs) {
	// Summoned pets, like Mage Water Elemental, aren't around to receive raid buffs.
	if petAgent.GetPet().IsGuardian() {
		return
	}
	raidBuffs = googleProto.Clone(raidBuffs).(*proto.RaidBuffs)
	partyBuffs = googleProto.Clone(partyBuffs).(*proto.PartyBuffs)
	individualBuffs = googleProto.Clone(individualBuffs).(*proto.IndividualBuffs)

	//Todo: Only cancel the buffs that are supposed to be cancelled
	// Check beta when pets are better implemented?
	raidBuffs = &proto.RaidBuffs{}
	partyBuffs = &proto.PartyBuffs{}
	individualBuffs = &proto.IndividualBuffs{}

	if !petAgent.GetPet().enabledOnStart {
		// What do we do with permanent pets that are not enabled at start?
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

func registerBloodlustCD(agent Agent, spellID int32) {
	character := agent.GetCharacter()
	BloodlustActionID.SpellID = spellID
	bloodlustAura := BloodlustAura(character, -1)

	spell := character.RegisterSpell(SpellConfig{
		ActionID: bloodlustAura.ActionID,
		Flags:    SpellFlagNoOnCastComplete | SpellFlagNoMetrics | SpellFlagNoLogs,

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
	})

	character.AddMajorCooldown(MajorCooldown{
		Spell:    spell,
		Priority: CooldownPriorityBloodlust,
		Type:     CooldownTypeDPS,
		ShouldActivate: func(sim *Simulation, character *Character) bool {
			return !character.HasActiveAura(SatedAuraLabel)
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

	aura := character.GetOrRegisterAura(Aura{
		Label:    "Bloodlust-" + actionID.String(),
		Tag:      BloodlustAuraTag,
		ActionID: actionID,
		Duration: BloodlustDuration,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.MultiplyAttackSpeed(sim, 1.3)
			aura.Unit.MultiplyCastSpeed(sim, 1.3)
			sated.Activate(sim)
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.MultiplyAttackSpeed(sim, 1/1.3)
			aura.Unit.MultiplyCastSpeed(sim, 1/1.3)
		},
	})

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
