package reforgeoptimizer

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

type metaGemConstraint struct {
	minRed              int
	minYellow           int
	minBlue             int
	compareColorGreater proto.GemColor
	compareColorLesser  proto.GemColor
}

var tbcMetaGemConstraintsByID = map[int32]metaGemConstraint{
	25899: {minRed: 2, minYellow: 2, minBlue: 2},
	34220: {minBlue: 2},
	25890: {minRed: 2, minYellow: 2, minBlue: 2},
	35503: {minRed: 3},
	35501: {minYellow: 1, minBlue: 2},
	32641: {minYellow: 3},
	25901: {minRed: 2, minYellow: 2, minBlue: 2},
	25896: {minBlue: 3},
	32409: {minRed: 2, minYellow: 2, minBlue: 2},
	25894: {minRed: 1, minYellow: 2},
	28557: {minRed: 1, minYellow: 2},
	28556: {minRed: 1, minYellow: 2},
	25898: {minBlue: 5},
	32410: {minRed: 2, minYellow: 2, minBlue: 2},
	25897: {compareColorGreater: proto.GemColor_GemColorRed, compareColorLesser: proto.GemColor_GemColorBlue},
	25895: {compareColorGreater: proto.GemColor_GemColorRed, compareColorLesser: proto.GemColor_GemColorYellow},
	25893: {compareColorGreater: proto.GemColor_GemColorBlue, compareColorLesser: proto.GemColor_GemColorYellow},
	32640: {compareColorGreater: proto.GemColor_GemColorBlue, compareColorLesser: proto.GemColor_GemColorYellow},
}

func equippedMetaGemConstraint(equipment core.Equipment) (metaGemConstraint, bool) {
	for _, item := range equipment {
		for _, gem := range item.Gems {
			if gem.ID == 0 || gem.Color != proto.GemColor_GemColorMeta {
				continue
			}
			constraint, ok := tbcMetaGemConstraintsByID[gem.ID]
			return constraint, ok
		}
	}
	return metaGemConstraint{}, false
}

func metaGemColorCounts(equipment core.Equipment) map[proto.GemColor]int {
	counts := make(map[proto.GemColor]int)
	for _, item := range equipment {
		for _, gem := range item.Gems {
			if gem.ID == 0 || gem.Color == proto.GemColor_GemColorMeta {
				continue
			}
			red, yellow, blue := metaGemActivationColorContribution(gem.Color)
			if red != 0 {
				counts[proto.GemColor_GemColorRed] += red
			}
			if yellow != 0 {
				counts[proto.GemColor_GemColorYellow] += yellow
			}
			if blue != 0 {
				counts[proto.GemColor_GemColorBlue] += blue
			}
		}
	}
	return counts
}

func metaGemActivationColorContribution(gemColor proto.GemColor) (red int, yellow int, blue int) {
	switch gemColor {
	case proto.GemColor_GemColorRed:
		return 1, 0, 0
	case proto.GemColor_GemColorYellow:
		return 0, 1, 0
	case proto.GemColor_GemColorBlue:
		return 0, 0, 1
	case proto.GemColor_GemColorOrange:
		return 1, 1, 0
	case proto.GemColor_GemColorGreen:
		return 0, 1, 1
	case proto.GemColor_GemColorPurple:
		return 1, 0, 1
	default:
		return 0, 0, 0
	}
}

func metaGemCountForColor(counts map[proto.GemColor]int, color proto.GemColor) int {
	if counts == nil {
		return 0
	}
	return counts[color]
}
