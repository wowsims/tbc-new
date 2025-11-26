package survival

import (
	"testing"

	"github.com/wowsims/tbc/sim/common" // imported to get item effects included.
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func init() {
	RegisterSurvivalHunter()
	common.RegisterAllEffects()
}

func TestSurvival(t *testing.T) {
	var talentSets []core.TalentsCombo
	talentSets = core.GenerateTalentVariationsForRows(SurvivalTalents, SurvivalDefaultGlyphs, []int{4, 5})

	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassHunter,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceWorgen},

			GearSet:         core.GetGearSet("../../../ui/hunter/survival/gear_sets", "p2"),
			Talents:         SurvivalTalents,
			OtherTalentSets: talentSets,
			Glyphs:          SurvivalDefaultGlyphs,
			Consumables:     FullConsumesSpec,
			SpecOptions:     core.SpecOptionsCombo{Label: "Basic", SpecOptions: PlayerOptionsBasic},
			Rotation:        core.GetAplRotation("../../../ui/hunter/survival/apls", "sv"),
			Profession1:     proto.Profession_Engineering,
			Profession2:     proto.Profession_Tailoring,

			ItemFilter:       ItemFilter,
			StartingDistance: 24,
		},
	}))
}

var ItemFilter = core.ItemFilter{
	ArmorType: proto.ArmorType_ArmorTypeMail,
	RangedWeaponTypes: []proto.RangedWeaponType{
		proto.RangedWeaponType_RangedWeaponTypeBow,
		proto.RangedWeaponType_RangedWeaponTypeCrossbow,
		proto.RangedWeaponType_RangedWeaponTypeGun,
	},
}

var FullConsumesSpec = &proto.ConsumesSpec{
	FlaskId:  76084, // Flask of Spring Blossoms
	FoodId:   74648, // Sea Mist Rice Noodles
	PotId:    76089, // Virmen's Bite
	PrepotId: 76089, // Virmen's Bite
}

var SurvivalTalents = "312111"
var SurvivalDefaultGlyphs = &proto.Glyphs{
	Major1: int32(proto.HunterMajorGlyph_GlyphOfLiberation),
	Major2: int32(proto.HunterMajorGlyph_GlyphOfAnimalBond),
	Major3: int32(proto.HunterMajorGlyph_GlyphOfDeterrence),
}

var PlayerOptionsBasic = &proto.Player_SurvivalHunter{
	SurvivalHunter: &proto.SurvivalHunter{
		Options: &proto.SurvivalHunter_Options{
			ClassOptions: &proto.HunterOptions{
				PetType:           proto.HunterOptions_Tallstrider,
				PetUptime:         1,
				UseHuntersMark:    true,
				GlaiveTossSuccess: 0.8,
			},
		},
	},
}
