package marksmanship

import (
	"testing"

	"github.com/wowsims/tbc/sim/common" // imported to get item effects included.
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func init() {
	RegisterMarksmanshipHunter()
	common.RegisterAllEffects()
}

func TestMarksmanship(t *testing.T) {
	var talentSets []core.TalentsCombo
	talentSets = core.GenerateTalentVariationsForRows(MarksmanshipTalents, MarksmanshipDefaultGlyphs, []int{4, 5})

	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassHunter,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceWorgen},

			GearSet:         core.GetGearSet("../../../ui/hunter/marksmanship/gear_sets", "p2"),
			Talents:         MarksmanshipTalents,
			OtherTalentSets: talentSets,
			Glyphs:          MarksmanshipDefaultGlyphs,
			Consumables:     FullConsumesSpec,
			SpecOptions:     core.SpecOptionsCombo{Label: "Basic", SpecOptions: PlayerOptionsBasic},
			Rotation:        core.GetAplRotation("../../../ui/hunter/marksmanship/apls", "mm"),
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

var MarksmanshipTalents = "312111"
var MarksmanshipDefaultGlyphs = &proto.Glyphs{
	Major1: int32(proto.HunterMajorGlyph_GlyphOfAimedShot),
	Major2: int32(proto.HunterMajorGlyph_GlyphOfAnimalBond),
	Major3: int32(proto.HunterMajorGlyph_GlyphOfDeterrence),
}

var PlayerOptionsBasic = &proto.Player_MarksmanshipHunter{
	MarksmanshipHunter: &proto.MarksmanshipHunter{
		Options: &proto.MarksmanshipHunter_Options{
			ClassOptions: &proto.HunterOptions{
				PetType:           proto.HunterOptions_Tallstrider,
				PetUptime:         1,
				UseHuntersMark:    true,
				GlaiveTossSuccess: 0.8,
			},
		},
	},
}
