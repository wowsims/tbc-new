package hunter

import (
	"testing"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

// The same table as ui/sim/player/debuff_stats.test.ts: a hunter who takes Expose Weakness is
// credited their own agility for the debuff on the character sheet.
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
			if got := core.CharacterSheetExposeWeaknessAgility(debuffs, tc.player, tc.characterAgility); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
