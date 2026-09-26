package core

import (
	"testing"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

// The same table as ui/sim/player/debuff_stats.test.ts: the Go and TypeScript copies of what the
// stats panel adds for the raid's debuffs must agree.
func TestCharacterSheetDebuffStats(t *testing.T) {
	pseudo := stats.UnitStatFromPseudoStat
	stat := stats.UnitStatFromStat
	for _, tc := range []struct {
		name    string
		debuffs *proto.Debuffs
		want    map[stats.UnitStat]float64
	}{
		{"none", &proto.Debuffs{}, map[stats.UnitStat]float64{}},
		{"improved faerie fire", &proto.Debuffs{FaerieFire: proto.TristateEffect_TristateEffectImproved}, map[stats.UnitStat]float64{
			pseudo(proto.PseudoStat_PseudoStatMeleeHitPercent): 3, pseudo(proto.PseudoStat_PseudoStatRangedHitPercent): 3,
		}},
		{"regular faerie fire adds nothing", &proto.Debuffs{FaerieFire: proto.TristateEffect_TristateEffectRegular}, map[stats.UnitStat]float64{}},
		{"improved seal of the crusader", &proto.Debuffs{ImprovedSealOfTheCrusader: proto.TristateEffect_TristateEffectRegular}, map[stats.UnitStat]float64{
			pseudo(proto.PseudoStat_PseudoStatMeleeCritPercent): 3, pseudo(proto.PseudoStat_PseudoStatRangedCritPercent): 3, pseudo(proto.PseudoStat_PseudoStatSpellCritPercent): 3,
		}},
		{"expose weakness", &proto.Debuffs{ExposeWeaknessUptime: 0.9, ExposeWeaknessHunterAgility: 800}, map[stats.UnitStat]float64{
			stat(stats.AttackPower): 200, stat(stats.RangedAttackPower): 200,
		}},
		{"expose weakness without agility adds nothing", &proto.Debuffs{ExposeWeaknessUptime: 0.9}, map[stats.UnitStat]float64{}},
		{"hunter's mark", &proto.Debuffs{HuntersMark: proto.TristateEffect_TristateEffectRegular}, map[stats.UnitStat]float64{
			stat(stats.RangedAttackPower): 440,
		}},
		{"improved hunter's mark", &proto.Debuffs{HuntersMark: proto.TristateEffect_TristateEffectImproved}, map[stats.UnitStat]float64{
			stat(stats.RangedAttackPower): 440, stat(stats.AttackPower): 110,
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := CharacterSheetDebuffStats(tc.debuffs, tc.debuffs.GetExposeWeaknessHunterAgility())
			for i := range got.Stats {
				if want := tc.want[stat(stats.Stat(i))]; got.Stats[i] != want {
					t.Fatalf("%s = %v, want %v", stats.Stat(i).StatName(), got.Stats[i], want)
				}
			}
			for i := range got.PseudoStats {
				if want := tc.want[pseudo(proto.PseudoStat(i))]; got.PseudoStats[i] != want {
					t.Fatalf("%s = %v, want %v", proto.PseudoStat(i), got.PseudoStats[i], want)
				}
			}
		})
	}
}

func TestWithCharacterSheetDebuffs(t *testing.T) {
	finalStats := &proto.UnitStats{Stats: make([]float64, stats.ProtoStatsLen), PseudoStats: make([]float64, stats.PseudoStatsLen)}
	finalStats.PseudoStats[proto.PseudoStat_PseudoStatSpellCritPercent] = 20
	sheet := WithCharacterSheetDebuffs(finalStats, &proto.Debuffs{ImprovedSealOfTheCrusader: proto.TristateEffect_TristateEffectImproved}, &proto.Player{})
	if got := sheet.PseudoStats[proto.PseudoStat_PseudoStatSpellCritPercent]; got != 23 {
		t.Fatalf("spell crit %v, want 23", got)
	}
	if finalStats.PseudoStats[proto.PseudoStat_PseudoStatSpellCritPercent] != 20 {
		t.Fatal("the final stats passed in must not be modified")
	}
}

// The same table as ui/sim/player/debuff_stats.test.ts.
func TestCharacterSheetExposeWeaknessAgility(t *testing.T) {
	// The Survival tree's 21st talent is Expose Weakness (proto field 62 = 21 + 20 + 21).
	const exposeWeaknessTalents = "--000000000000000000001"
	debuffs := &proto.Debuffs{ExposeWeaknessUptime: 1, ExposeWeaknessHunterAgility: 100}
	for _, tc := range []struct {
		name             string
		player           *proto.Player
		characterAgility float64
		want             float64
	}{
		{"a hunter with the talent is credited their own agility", &proto.Player{Class: proto.Class_ClassHunter, TalentsString: exposeWeaknessTalents}, 700, 700},
		{"a hunter without the talent is credited the configured agility", &proto.Player{Class: proto.Class_ClassHunter, TalentsString: "502-0550201205"}, 700, 100},
		{"another class is credited the configured agility", &proto.Player{Class: proto.Class_ClassMage, TalentsString: exposeWeaknessTalents}, 700, 100},
		{"unknown own agility falls back to the configured agility", &proto.Player{Class: proto.Class_ClassHunter, TalentsString: exposeWeaknessTalents}, 0, 100},
		{"a malformed talent string falls back to the configured agility", &proto.Player{Class: proto.Class_ClassHunter, TalentsString: "9-9-9-9-9-9-9-9-9-9-9-9-9"}, 700, 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CharacterSheetExposeWeaknessAgility(debuffs, tc.player, tc.characterAgility); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
