package core

import (
	"testing"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

var allGemColors = []proto.GemColor{
	proto.GemColor_GemColorMeta,
	proto.GemColor_GemColorRed,
	proto.GemColor_GemColorBlue,
	proto.GemColor_GemColorYellow,
	proto.GemColor_GemColorGreen,
	proto.GemColor_GemColorOrange,
	proto.GemColor_GemColorPurple,
	proto.GemColor_GemColorPrismatic,
}

// Which gem colors count towards a socket bonus, written out by hand rather than derived from
// ColorIntersects so that this stays an independent statement of the rules: a socket accepts its own
// color plus the two hybrids that contain it, a meta socket accepts only meta, and prismatic matches
// everything in both directions.
var socketBonusMatches = map[proto.GemColor][]proto.GemColor{
	proto.GemColor_GemColorMeta: {proto.GemColor_GemColorMeta, proto.GemColor_GemColorPrismatic},
	proto.GemColor_GemColorRed: {
		proto.GemColor_GemColorRed, proto.GemColor_GemColorOrange, proto.GemColor_GemColorPurple, proto.GemColor_GemColorPrismatic,
	},
	proto.GemColor_GemColorBlue: {
		proto.GemColor_GemColorBlue, proto.GemColor_GemColorGreen, proto.GemColor_GemColorPurple, proto.GemColor_GemColorPrismatic,
	},
	proto.GemColor_GemColorYellow: {
		proto.GemColor_GemColorYellow, proto.GemColor_GemColorGreen, proto.GemColor_GemColorOrange, proto.GemColor_GemColorPrismatic,
	},
	proto.GemColor_GemColorGreen: {
		proto.GemColor_GemColorGreen, proto.GemColor_GemColorBlue, proto.GemColor_GemColorYellow, proto.GemColor_GemColorPrismatic,
	},
	proto.GemColor_GemColorOrange: {
		proto.GemColor_GemColorOrange, proto.GemColor_GemColorRed, proto.GemColor_GemColorYellow, proto.GemColor_GemColorPrismatic,
	},
	proto.GemColor_GemColorPurple: {
		proto.GemColor_GemColorPurple, proto.GemColor_GemColorRed, proto.GemColor_GemColorBlue, proto.GemColor_GemColorPrismatic,
	},
	proto.GemColor_GemColorPrismatic: allGemColors,
}

func socketBonusExpected(socketColor, gemColor proto.GemColor) bool {
	for _, c := range socketBonusMatches[socketColor] {
		if c == gemColor {
			return true
		}
	}
	return false
}

// One socket, one gem, a socket bonus on the item. Returns the total of the one stat everything
// contributes to, so callers can decode which parts were counted.
const (
	gemStatValue    = 160.0
	socketBonusStat = 120.0
)

func itemWithOneSocket(socketColor, gemColor proto.GemColor, disabled bool) Item {
	gemStats := stats.Stats{}
	gemStats[stats.Agility] = gemStatValue
	bonus := stats.Stats{}
	bonus[stats.Agility] = socketBonusStat

	return Item{
		ID:          1,
		GemSockets:  []proto.GemColor{socketColor},
		Gems:        []Gem{{ID: 1, Color: gemColor, Stats: gemStats, Disabled: disabled}},
		SocketBonus: bonus,
	}
}

// Every socket colour against every gem colour, so a change to the matching rules cannot slip past
// on the colours nobody happens to have written a case for.
func TestSocketBonusForEveryColorCombination(t *testing.T) {
	for _, socketColor := range allGemColors {
		for _, gemColor := range allGemColors {
			got := ItemEquipmentGemAndEnchantStats(itemWithOneSocket(socketColor, gemColor, false))[stats.Agility]

			want := gemStatValue
			if socketBonusExpected(socketColor, gemColor) {
				want += socketBonusStat
			}

			if got != want {
				t.Errorf("%v socket with a %v gem: expected %v agility, got %v", socketColor, gemColor, want, got)
			}
		}
	}
}

// An empty socket earns no bonus. The one exception is a prismatic socket: ColorIntersects treats
// prismatic as matching anything on either side, so an empty one still counts. That is pre-existing
// behaviour and is left alone here -- no TBC item has a prismatic socket, so it cannot be reached.
func TestEmptySocketNeverEarnsSocketBonus(t *testing.T) {
	for _, socketColor := range allGemColors {
		item := itemWithOneSocket(socketColor, proto.GemColor_GemColorUnknown, false)
		item.Gems = []Gem{{}}

		want := 0.0
		if socketColor == proto.GemColor_GemColorPrismatic {
			want = socketBonusStat
		}

		if got := ItemEquipmentGemAndEnchantStats(item)[stats.Agility]; got != want {
			t.Errorf("%v socket left empty: expected %v agility, got %v", socketColor, want, got)
		}
	}
}

// Guards the claim made above, so that if a prismatic socket ever appears on a TBC item the exception
// in TestEmptySocketNeverEarnsSocketBonus has to be revisited rather than silently mattering.
func TestNoTBCItemHasAPrismaticSocket(t *testing.T) {
	if len(ItemsByID) == 0 {
		t.Skip("no item database loaded; run with -tags with_db")
	}

	for _, item := range ItemsByID {
		for _, socketColor := range item.GemSockets {
			if socketColor == proto.GemColor_GemColorPrismatic {
				t.Fatalf("%q has a prismatic socket; empty-prismatic-socket handling now matters", item.Name)
			}
		}
	}
}

// A disabled gem contributes no stats but keeps its colour, so the socket bonus is unaffected. Only
// meta gems are disabled today (when their requirements are not met), but the rule is colour-agnostic
// and is checked that way.
func TestDisabledGemKeepsSocketBonusButLosesItsStats(t *testing.T) {
	for _, socketColor := range allGemColors {
		for _, gemColor := range allGemColors {
			got := ItemEquipmentGemAndEnchantStats(itemWithOneSocket(socketColor, gemColor, true))[stats.Agility]

			want := 0.0
			if socketBonusExpected(socketColor, gemColor) {
				want = socketBonusStat
			}

			if got != want {
				t.Errorf("%v socket with a disabled %v gem: expected %v agility, got %v", socketColor, gemColor, want, got)
			}
		}
	}
}

// The colour rules above are synthetic. This runs the same invariant over every gem actually in the
// database, so a gem whose colour is mis-parsed on import is caught too.
func TestSocketBonusForEveryGemInDatabase(t *testing.T) {
	if len(GemsByID) == 0 {
		t.Skip("no gem database loaded; run with -tags with_db")
	}

	for _, gem := range GemsByID {
		for _, socketColor := range allGemColors {
			item := Item{
				ID:          1,
				GemSockets:  []proto.GemColor{socketColor},
				Gems:        []Gem{gem},
				SocketBonus: stats.Stats{},
			}
			item.SocketBonus[stats.Agility] = socketBonusStat

			want := gem.Stats[stats.Agility]
			if socketBonusExpected(socketColor, gem.Color) {
				want += socketBonusStat
			}

			if got := ItemEquipmentGemAndEnchantStats(item)[stats.Agility]; got != want {
				t.Fatalf("%q (%v) in a %v socket: expected %v agility, got %v", gem.Name, gem.Color, socketColor, want, got)
			}
		}
	}
}

// Every meta gem in the database, disabled: no stats of its own, but the head socket bonus survives.
// This is the bug this change exists to fix.
func TestEveryDisabledMetaGemInDatabaseKeepsSocketBonus(t *testing.T) {
	if len(GemsByID) == 0 {
		t.Skip("no gem database loaded; run with -tags with_db")
	}

	checked := 0
	for _, gem := range GemsByID {
		if gem.Color != proto.GemColor_GemColorMeta {
			continue
		}
		checked++

		disabled := gem
		disabled.Disabled = true
		item := Item{
			ID:          1,
			GemSockets:  []proto.GemColor{proto.GemColor_GemColorMeta},
			Gems:        []Gem{disabled},
			SocketBonus: stats.Stats{},
		}
		item.SocketBonus[stats.Agility] = socketBonusStat

		if got := ItemEquipmentGemAndEnchantStats(item)[stats.Agility]; got != socketBonusStat {
			t.Fatalf("disabled %q: expected the %v socket bonus and nothing else, got %v", gem.Name, socketBonusStat, got)
		}
	}

	if checked == 0 {
		t.Fatal("no meta gems found in the database")
	}
	t.Logf("checked %d meta gems", checked)
}

// Equipment is round-tripped back through NewItem in places, so the disabled flag has to survive
// Item -> ItemSpec proto -> Item or the meta gem silently reactivates mid-run.
func TestDisabledMetaGemSurvivesProtoRoundTrip(t *testing.T) {
	const testItemID, testMetaGemID, testRedGemID = 990001, 990002, 990003

	addToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{{
			Id:             testItemID,
			Name:           "Test Helm",
			GemSockets:     []proto.GemColor{proto.GemColor_GemColorMeta, proto.GemColor_GemColorRed},
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
		}},
		Gems: []*proto.SimGem{
			{Id: testMetaGemID, Name: "Test Meta", Color: proto.GemColor_GemColorMeta},
			{Id: testRedGemID, Name: "Test Red", Color: proto.GemColor_GemColorRed},
		},
	})

	item := NewItem(ItemSpec{ID: testItemID, Gems: []int32{testMetaGemID, testRedGemID}, MetaGemDisabled: true})
	if !item.Gems[0].Disabled {
		t.Fatal("meta gem should be disabled after NewItem")
	}
	if item.Gems[0].Color != proto.GemColor_GemColorMeta {
		t.Fatal("disabled meta gem must keep its color so the socket bonus still matches")
	}
	if item.Gems[1].Disabled {
		t.Fatal("only the meta gem should be disabled")
	}

	roundTripped := NewItem(ProtoToEquipmentSpec(&proto.EquipmentSpec{Items: []*proto.ItemSpec{item.ToItemSpecProto()}})[0])
	if !roundTripped.Gems[0].Disabled {
		t.Fatal("meta gem reactivated after a proto round-trip")
	}
}

// The request path is *proto.EquipmentSpec -> ProtoToEquipment -> Character.Equipment, so check the
// flag survives that seam and that the head item really does keep its socket bonus.
func TestDisabledMetaGemFromRequestKeepsSocketBonus(t *testing.T) {
	const testItemID, testMetaGemID, testRedGemID = 990011, 990012, 990013

	socketBonus := make([]float64, len(stats.Stats{}))
	socketBonus[stats.Agility] = socketBonusStat
	metaStats := make([]float64, len(stats.Stats{}))
	metaStats[stats.Agility] = 216
	redStats := make([]float64, len(stats.Stats{}))
	redStats[stats.Agility] = gemStatValue

	addToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{{
			Id:             testItemID,
			Name:           "Test Helm 2",
			Type:           proto.ItemType_ItemTypeHead,
			GemSockets:     []proto.GemColor{proto.GemColor_GemColorMeta, proto.GemColor_GemColorRed},
			SocketBonus:    socketBonus,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
		}},
		Gems: []*proto.SimGem{
			{Id: testMetaGemID, Name: "Test Meta 2", Color: proto.GemColor_GemColorMeta, Stats: metaStats},
			{Id: testRedGemID, Name: "Test Red 2", Color: proto.GemColor_GemColorRed, Stats: redStats},
		},
	})

	equipmentFor := func(metaDisabled bool) Equipment {
		return ProtoToEquipment(&proto.EquipmentSpec{Items: []*proto.ItemSpec{{
			Id:              testItemID,
			Gems:            []int32{testMetaGemID, testRedGemID},
			MetaGemDisabled: metaDisabled,
		}}})
	}

	active := equipmentFor(false)
	if got := ItemEquipmentGemAndEnchantStats(*active.Head())[stats.Agility]; got != 216+gemStatValue+socketBonusStat {
		t.Fatalf("active meta gem: expected %v agility, got %v", 216+gemStatValue+socketBonusStat, got)
	}

	inactive := equipmentFor(true)
	if !inactive.Head().Gems[0].Disabled {
		t.Fatal("meta_gem_disabled did not survive the proto -> Equipment conversion")
	}
	// Meta gem stats gone, head socket bonus intact -- the bug this whole change is about.
	if got := ItemEquipmentGemAndEnchantStats(*inactive.Head())[stats.Agility]; got != gemStatValue+socketBonusStat {
		t.Fatalf("disabled meta gem: expected %v agility, got %v", gemStatValue+socketBonusStat, got)
	}
}
