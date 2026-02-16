package core

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
)

var DefaultSimTestOptions = &proto.SimOptions{
	Iterations: 20,
	IsTest:     true,
	Debug:      false,
	RandomSeed: 101,
}
var StatWeightsDefaultSimTestOptions = &proto.SimOptions{
	Iterations: 300,
	IsTest:     true,
	Debug:      false,
	RandomSeed: 101,
}
var AverageDefaultSimTestOptions = &proto.SimOptions{
	Iterations: 2000,
	IsTest:     true,
	Debug:      false,
	RandomSeed: 101,
}

const ShortDuration = 60
const LongDuration = 180

func FreshDefaultTargetConfig() *proto.Target {
	return &proto.Target{
		Level: CharacterLevel + 3,
		Stats: stats.Stats{
			stats.Armor:       7684,
			stats.AttackPower: 320,
			stats.BlockValue:  54,
		}.ToProtoArray(),
		MobType: proto.MobType_MobTypeMechanical,

		SwingSpeed:    2,
		MinBaseDamage: 4192.05,
		ParryHaste:    true,
		DamageSpread:  0.4,
	}
}

var DefaultTargetProto = FreshDefaultTargetConfig()

var FullRaidBuffs = &proto.RaidBuffs{
	Bloodlust:          true,
	ArcaneBrilliance:   true,
	PowerWordFortitude: proto.TristateEffect_TristateEffectImproved,
	DivineSpirit:       proto.TristateEffect_TristateEffectImproved,
	GiftOfTheWild:      proto.TristateEffect_TristateEffectImproved,
	Thorns:             proto.TristateEffect_TristateEffectImproved,
	ShadowProtection:   true,
}

var FullPartyBuffs = &proto.PartyBuffs{
	FerociousInspiration:  1,
	BloodPact:             proto.TristateEffect_TristateEffectImproved,
	MoonkinAura:           proto.TristateEffect_TristateEffectImproved,
	LeaderOfThePack:       proto.TristateEffect_TristateEffectImproved,
	SanctityAura:          proto.TristateEffect_TristateEffectImproved,
	DevotionAura:          proto.TristateEffect_TristateEffectImproved,
	RetributionAura:       proto.TristateEffect_TristateEffectImproved,
	TrueshotAura:          true,
	DraeneiRacialMelee:    true,
	DraeneiRacialCaster:   true,
	AtieshDruid:           1,
	AtieshMage:            1,
	AtieshPriest:          1,
	AtieshWarlock:         1,
	BraidedEterniumChain:  true,
	EyeOfTheNight:         true,
	ChainOfTheTwilightOwl: true,
	JadePendantOfBlasting: true,

	ManaSpringTotem:      proto.TristateEffect_TristateEffectImproved,
	ManaTideTotems:       1,
	TotemOfWrath:         1,
	WrathOfAirTotem:      proto.TristateEffect_TristateEffectImproved,
	GraceOfAirTotem:      proto.TristateEffect_TristateEffectImproved,
	StrengthOfEarthTotem: proto.TristateEffect_TristateEffectImproved,
	TranquilAirTotem:     true,
	WindfuryTotem:        proto.TristateEffect_TristateEffectImproved,

	BattleShout:     proto.TristateEffect_TristateEffectImproved,
	CommandingShout: proto.TristateEffect_TristateEffectImproved,
}

var FullIndividualBuffs = &proto.IndividualBuffs{
	BlessingOfKings:     true,
	BlessingOfSanctuary: true,
	BlessingOfWisdom:    proto.TristateEffect_TristateEffectImproved,
	BlessingOfMight:     proto.TristateEffect_TristateEffectImproved,
	UnleashedRage:       true,
}

var FullDebuffs = &proto.Debuffs{
	JudgementOfWisdom:         true,
	JudgementOfLight:          true,
	ImprovedSealOfTheCrusader: true,
	Misery:                    true,
	CurseOfElements:           proto.TristateEffect_TristateEffectImproved,
	IsbUptime:                 1,
	ShadowWeaving:             true,
	ImprovedScorch:            true,
	WintersChill:              true,
	BloodFrenzy:               true,
	GiftOfArthas:              true,
	Mangle:                    true,
	ExposeArmor:               proto.TristateEffect_TristateEffectImproved,
	FaerieFire:                proto.TristateEffect_TristateEffectImproved,
	SunderArmor:               true,
	CurseOfRecklessness:       true,
	HuntersMark:               proto.TristateEffect_TristateEffectImproved,
	DemoralizingRoar:          proto.TristateEffect_TristateEffectImproved,
	// DemoralizingShout:         proto.TristateEffect_TristateEffectImproved,
	ThunderClap:      proto.TristateEffect_TristateEffectImproved,
	InsectSwarm:      true,
	ScorpidSting:     true,
	ShadowEmbrace:    true,
	Screech:          true,
	HemorrhageUptime: 1,
}

func NewDefaultTarget() *proto.Target {
	return DefaultTargetProto // seems to be read-only
}

func MakeDefaultEncounterCombos() []EncounterCombo {
	var DefaultTarget = NewDefaultTarget()

	multipleTargets := make([]*proto.Target, 21)
	for i := range multipleTargets {
		if i != 10 {
			multipleTargets[i] = DefaultTarget
		} else {
			disabledTarget := FreshDefaultTargetConfig()
			disabledTarget.DisabledAtStart = true
			multipleTargets[i] = disabledTarget
		}
	}

	return []EncounterCombo{
		{
			Label: "ShortSingleTarget",
			Encounter: &proto.Encounter{
				Duration:             ShortDuration,
				ExecuteProportion_20: 0.2,
				ExecuteProportion_25: 0.25,
				ExecuteProportion_35: 0.35,
				ExecuteProportion_45: 0.45,
				ExecuteProportion_90: 0.90,
				Targets: []*proto.Target{
					DefaultTarget,
				},
			},
		},
		{
			Label: "LongSingleTarget",
			Encounter: &proto.Encounter{
				Duration:             LongDuration,
				ExecuteProportion_20: 0.2,
				ExecuteProportion_25: 0.25,
				ExecuteProportion_35: 0.35,
				ExecuteProportion_45: 0.45,
				ExecuteProportion_90: 0.90,
				Targets: []*proto.Target{
					DefaultTarget,
				},
			},
		},
		{
			Label: "LongMultiTarget",
			Encounter: &proto.Encounter{
				Duration:             LongDuration,
				ExecuteProportion_20: 0.2,
				ExecuteProportion_25: 0.25,
				ExecuteProportion_35: 0.35,
				ExecuteProportion_45: 0.45,
				ExecuteProportion_90: 0.90,
				Targets:              multipleTargets,
			},
		},
	}
}

func MakeSingleTargetEncounter(variation float64) *proto.Encounter {
	return &proto.Encounter{
		Duration:             LongDuration,
		DurationVariation:    variation,
		ExecuteProportion_20: 0.2,
		ExecuteProportion_25: 0.25,
		ExecuteProportion_35: 0.35,
		ExecuteProportion_45: 0.45,
		ExecuteProportion_90: 0.90,
		Targets: []*proto.Target{
			NewDefaultTarget(),
		},
	}
}

func RaidSimTest(label string, t *testing.T, rsr *proto.RaidSimRequest, expectedDps float64) {
	result := RunRaidSim(rsr)
	if result.Error != nil {
		t.Fatalf("Sim failed with error: %s", result.Error.Message)
	}
	tolerance := 0.5
	if result.RaidMetrics.Dps.Avg < expectedDps-tolerance || result.RaidMetrics.Dps.Avg > expectedDps+tolerance {
		// Automatically print output if we had debugging enabled.
		if rsr.SimOptions.Debug {
			log.Printf("LOGS:\n%s\n", result.Logs)
		}
		t.Fatalf("%s failed: expected %0f dps from sim but was %0f", label, expectedDps, result.RaidMetrics.Dps.Avg)
	}
}

func RaidBenchmark(b *testing.B, rsr *proto.RaidSimRequest) {
	rsr.Encounter.Duration = LongDuration
	rsr.SimOptions.Iterations = 1

	// Set to false because IsTest adds a lot of computation.
	rsr.SimOptions.IsTest = false

	for i := 0; i < b.N; i++ {
		result := RunRaidSim(rsr)
		if result.Error != nil {
			b.Fatalf("RaidBenchmark() at iteration %d failed: %v", i, result.Error.Message)
		}
	}
}

func GetAplRotation(dir string, file string) RotationCombo {
	filePath := dir + "/" + file + ".apl.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to load apl json file: %s, %s", filePath, err)
	}

	return RotationCombo{Label: file, Rotation: APLRotationFromJsonString(string(data))}
}

func GetGearSet(dir string, file string) GearSetCombo {
	filePath := dir + "/" + file + ".gear.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to load gear json file: %s, %s", filePath, err)
	}

	return GearSetCombo{Label: file, GearSet: EquipmentSpecFromJsonString(string(data))}
}

func GetItemSwapGearSet(dir string, file string) ItemSwapSetCombo {
	filePath := dir + "/" + file + ".gear.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to load gear json file: %s, %s", filePath, err)
	}

	return ItemSwapSetCombo{Label: file, ItemSwap: ItemSwapFromJsonString(string(data))}
}

func GenerateTalentVariations(baseTalents string) []TalentsCombo {
	return GenerateTalentVariationsForRows(baseTalents, []int{0, 1, 2, 3, 4, 5})
}

func GenerateTalentVariationsForRows(baseTalents string, rowsToVary []int) []TalentsCombo {
	if len(baseTalents) != 6 {
		log.Fatalf("Expected 6-digit talent string, got: %s", baseTalents)
	}

	var combinations []TalentsCombo

	baseRunes := []rune(baseTalents)
	for _, row := range rowsToVary {
		if row < 0 || row >= 6 {
			log.Fatalf("Invalid row index: %d, must be between 0 and 5", row)
		}

		for choice := 1; choice <= 3; choice++ {
			if int(baseRunes[row]-'0') == choice {
				continue
			}

			variation := make([]rune, 6)
			copy(variation, baseRunes)
			variation[row] = rune('0' + choice)

			combinations = append(combinations, TalentsCombo{
				Label:   fmt.Sprintf("Row%d_Talent%d", row+1, choice),
				Talents: string(variation),
			})
		}
	}

	return combinations
}

func GetTestBuildFromJSON(class proto.Class, dir string, file string, itemFilter ItemFilter, epReferenceStat *proto.Stat, statsToWeigh *[]proto.Stat) CharacterSuiteConfig {
	filePath := dir + "/" + file + ".build.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to load gear json file: %s, %s", filePath, err)
	}

	simSettings := &proto.IndividualSimSettings{}
	if err := protojson.Unmarshal(data, simSettings); err != nil {
		panic(err)
	}

	config := CharacterSuiteConfig{
		Class:       class,
		Race:        simSettings.Player.Race,
		Profession1: simSettings.Player.Profession1,
		Profession2: simSettings.Player.Profession2,
		GearSet: GearSetCombo{
			Label:   file,
			GearSet: simSettings.Player.Equipment,
		},
		SpecOptions: SpecOptionsCombo{
			Label:       file,
			SpecOptions: getPlayerSpecOptions(simSettings.Player),
		},
		Talents: simSettings.Player.TalentsString,
		Rotation: RotationCombo{
			Label:    file,
			Rotation: simSettings.Player.Rotation,
		},
		Encounter: EncounterCombo{
			Label:     file,
			Encounter: simSettings.Encounter,
		},
		ItemSwapSet: ItemSwapSetCombo{
			Label:    file,
			ItemSwap: simSettings.Player.ItemSwap,
		},
		StartingDistance:   simSettings.Player.DistanceFromTarget,
		ReactionTimeMs:     simSettings.Player.ReactionTimeMs,
		ChannelClipDelayMs: simSettings.Player.ChannelClipDelayMs,

		Consumables:     simSettings.Player.Consumables,
		IndividualBuffs: simSettings.Player.Buffs,
		PartyBuffs:      simSettings.PartyBuffs,
		RaidBuffs:       simSettings.RaidBuffs,
		Debuffs:         simSettings.Debuffs,
		Cooldowns:       simSettings.Player.Cooldowns,

		InFrontOfTarget: simSettings.Player.InFrontOfTarget,
		TargetDummies:   simSettings.TargetDummies,
		HealingModel:    simSettings.Player.HealingModel,

		ItemFilter: itemFilter,
	}

	if simSettings.Tanks != nil {
		config.Tanks = simSettings.Tanks

		// Check if any of the tanks is the player.
		for _, tank := range simSettings.Tanks {
			if tank.Type == proto.UnitReference_Player {
				config.IsTank = true
				break
			}
		}
	}

	if epReferenceStat != nil {
		config.EPReferenceStat = *epReferenceStat
	}
	if statsToWeigh != nil {
		config.StatsToWeigh = *statsToWeigh
	}

	return config
}

func getPlayerSpecOptions(player *proto.Player) interface{} {
	if playerSpec, ok := player.Spec.(*proto.Player_BalanceDruid); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_FeralDruid); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_GuardianDruid); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_RestorationDruid); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_Hunter); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_Mage); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_HolyPaladin); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_ProtectionPaladin); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_RetributionPaladin); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_DisciplinePriest); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_HolyPriest); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_ShadowPriest); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_Rogue); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_ElementalShaman); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_EnhancementShaman); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_RestorationShaman); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_Warlock); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_DpsWarrior); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_ProtectionWarrior); ok {
		return playerSpec
	}

	panic("Unsupported spec provided to getPlayerSpecOptions. Please add a case for the spec.")
}
