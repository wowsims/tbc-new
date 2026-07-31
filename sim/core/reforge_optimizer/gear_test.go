package reforgeoptimizer

import (
	"testing"

	"github.com/wowsims/tbc/sim/core/proto"
)

func TestGemMatchesSocketSecondaryColors(t *testing.T) {
	testCases := []struct {
		name        string
		gemColor    proto.GemColor
		socketColor proto.GemColor
		want        bool
	}{
		{name: "orange matches red", gemColor: proto.GemColor_GemColorOrange, socketColor: proto.GemColor_GemColorRed, want: true},
		{name: "orange matches yellow", gemColor: proto.GemColor_GemColorOrange, socketColor: proto.GemColor_GemColorYellow, want: true},
		{name: "purple matches red", gemColor: proto.GemColor_GemColorPurple, socketColor: proto.GemColor_GemColorRed, want: true},
		{name: "purple matches blue", gemColor: proto.GemColor_GemColorPurple, socketColor: proto.GemColor_GemColorBlue, want: true},
		{name: "green matches yellow", gemColor: proto.GemColor_GemColorGreen, socketColor: proto.GemColor_GemColorYellow, want: true},
		{name: "green matches blue", gemColor: proto.GemColor_GemColorGreen, socketColor: proto.GemColor_GemColorBlue, want: true},
		{name: "orange does not match blue", gemColor: proto.GemColor_GemColorOrange, socketColor: proto.GemColor_GemColorBlue, want: false},
		{name: "red matches prismatic", gemColor: proto.GemColor_GemColorRed, socketColor: proto.GemColor_GemColorPrismatic, want: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := gemMatchesSocket(testCase.gemColor, testCase.socketColor); got != testCase.want {
				t.Fatalf("gemMatchesSocket(%s, %s) = %t, want %t", testCase.gemColor, testCase.socketColor, got, testCase.want)
			}
		})
	}
}
