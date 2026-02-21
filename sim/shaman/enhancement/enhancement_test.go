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

func TestStormstrikeCasts(t *testing.T) {
	gear := core.GetGearSet("../../../ui/shaman/enhancement/gear_sets", "preraid").GearSet
	rotation := core.APLRotationFromJsonString(`{"type":"TypeAPL","priorityList":[{"action":{"castSpell":{"spellId":{"spellId":17364}}}}]}`)

	player := core.WithSpec(&proto.Player{
		Name:          "EnhStormstrikeTest",
		Class:         proto.Class_ClassShaman,
		Race:          proto.Race_RaceTroll,
		Equipment:     gear,
		Rotation:      rotation,
		TalentsString: "0-0-0",
	}, &proto.Player_EnhancementShaman{
		EnhancementShaman: &proto.EnhancementShaman{
			Options: &proto.EnhancementShaman_Options{
				ClassOptions: &proto.ShamanOptions{
					Shield:      proto.ShamanShield_LightningShield,
					ImbueMh:     proto.ShamanImbue_WindfuryWeapon,
					ImbueMhSwap: proto.ShamanImbue_WindfuryWeapon,
				},
				SyncType:    proto.ShamanSyncType_Auto,
				ImbueOh:     proto.ShamanImbue_FlametongueWeapon,
				ImbueOhSwap: proto.ShamanImbue_FlametongueWeapon,
			},
		},
	})

	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = 20

	rsr := &proto.RaidSimRequest{
		Raid:      core.SinglePlayerRaidProto(player, nil, nil, nil),
		Encounter: encounter,
		SimOptions: &proto.SimOptions{
			Iterations:          1,
			IsTest:              true,
			RandomSeed:          101,
			Debug:               true,
			DebugFirstIteration: true,
		},
	}

	result := core.RunRaidSim(rsr)
	if result.Error != nil {
		t.Fatalf("Sim failed: %s", result.Error.Message)
	}
	logSimLogs(t, result)

	mainCasts := findCasts(result, 17364, 0)
	mhCasts := findCasts(result, 17364, 1)
	ohCasts := findCasts(result, 17364, 2)
	auraUptime := findTargetAuraUptime(result, 17364)
	manaSpend := findResourceMetrics(result, 17364, proto.ResourceType_ResourceTypeMana)

	if mainCasts == 0 {
		t.Fatalf("Expected Stormstrike casts > 0, got %d", mainCasts)
	}
	if mhCasts == 0 || ohCasts == 0 {
		t.Fatalf("Expected Stormstrike MH/OH hits > 0, got MH=%d OH=%d", mhCasts, ohCasts)
	}
	if mhCasts != ohCasts {
		t.Fatalf("Expected Stormstrike MH/OH casts to match, got MH=%d OH=%d", mhCasts, ohCasts)
	}
	if mhCasts > mainCasts || ohCasts > mainCasts {
		t.Fatalf("Expected Stormstrike MH/OH casts <= main casts, got main=%d MH=%d OH=%d", mainCasts, mhCasts, ohCasts)
	}
	if auraUptime <= 0 {
		t.Fatalf("Expected Stormstrike debuff uptime > 0, got %f", auraUptime)
	}
	if manaSpend == nil || manaSpend.Events == 0 || manaSpend.Gain >= 0 {
		t.Fatalf("Expected Stormstrike mana spend to be recorded, got %+v", manaSpend)
	}
}

func findCasts(result *proto.RaidSimResult, spellID int32, tag int32) int32 {
	if result == nil || len(result.RaidMetrics.Parties) == 0 || len(result.RaidMetrics.Parties[0].Players) == 0 {
		return 0
	}
	var total int32
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		if action.Id.GetSpellId() != spellID || action.Id.Tag != tag {
			continue
		}
		for _, targetMetrics := range action.Targets {
			total += targetMetrics.Casts
		}
	}
	return total
}

func logSimLogs(t *testing.T, result *proto.RaidSimResult) {
	if result == nil {
		t.Log("No sim result to log.")
		return
	}
	if result.Logs == "" {
		t.Log("No sim logs captured (Debug may be false or logs disabled).")
		return
	}
	t.Logf("LOGS:\n%s", result.Logs)
}

func findTargetAuraUptime(result *proto.RaidSimResult, spellID int32) float64 {
	if result == nil || result.EncounterMetrics == nil || len(result.EncounterMetrics.Targets) == 0 {
		return 0
	}
	for _, aura := range result.EncounterMetrics.Targets[0].Auras {
		if aura.Id.GetSpellId() == spellID {
			return aura.UptimeSecondsAvg
		}
	}
	return 0
}

func findResourceMetrics(result *proto.RaidSimResult, spellID int32, resourceType proto.ResourceType) *proto.ResourceMetrics {
	if result == nil || len(result.RaidMetrics.Parties) == 0 || len(result.RaidMetrics.Parties[0].Players) == 0 {
		return nil
	}
	for _, rm := range result.RaidMetrics.Parties[0].Players[0].Resources {
		if rm.Type == resourceType && rm.Id.GetSpellId() == spellID {
			return rm
		}
	}
	return nil
}
