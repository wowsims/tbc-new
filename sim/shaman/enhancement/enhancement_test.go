package enhancement

import (
	"testing"

	"github.com/wowsims/tbc/sim/common" // imported to get item effects included.
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

func init() {
	RegisterEnhancementShaman()
	common.RegisterAllEffects()
}

func TestEnhancement(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassShaman,
			Race:       proto.Race_RaceTroll,
			OtherRaces: []proto.Race{proto.Race_RaceOrc, proto.Race_RaceDraenei},
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: &proto.Player_EnhancementShaman{
				EnhancementShaman: &proto.EnhancementShaman{
					Options: &proto.EnhancementShaman_Options{
						SyncType:    proto.ShamanSyncType_Auto,
						ImbueOh:     proto.ShamanImbue_WindfuryWeapon,
						ImbueOhSwap: proto.ShamanImbue_WindfuryWeapon,
						ClassOptions: &proto.ShamanOptions{
							ImbueMh:        proto.ShamanImbue_WindfuryWeapon,
							ImbueMhSwap:    proto.ShamanImbue_WindfuryWeapon,
							ShieldProcrate: 0.0,
						},
					},
				},
			}},
			GearSet: core.GetGearSet("../../../ui/shaman/enhancement/gear_sets", "p1"),
			Talents: "03-500502210501133531151-50005301",
			OtherTalentSets: []core.TalentsCombo{
				{
					Label:   "Sub-Restoration ILS",
					Talents: "03-500503210500133531151-50005301",
				},
				{
					Label:   "Sub-Elemental",
					Talents: "250031501-500503210500133531151",
				},
			},
			Rotation: core.GetAplRotation("../../../ui/shaman/enhancement/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeOffHand,
				},
				ArmorType: proto.ArmorType_ArmorTypeMail,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeTotem,
				},
			},
		},
	}))
}
