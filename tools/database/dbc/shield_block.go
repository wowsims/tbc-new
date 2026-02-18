package dbc

import (
	"bufio"
	_ "embed"
	"strconv"
	"strings"

	"github.com/wowsims/tbc/sim/core/proto"
)

//go:embed GameTables/ShieldBlockRegular.txt
var shieldBlockFile string

type ShieldBlock struct {
	Level  int
	Values map[proto.ItemQuality]float64
}

func (dbc *DBC) LoadShieldBlockValues() error {
	scanner := bufio.NewScanner(strings.NewReader(shieldBlockFile))
	scanner.Scan() // Skip first line

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		level, err := strconv.Atoi(parts[0])

		if err != nil {
			continue // consider handling or logging this situation
		}

		shieldBlock := ShieldBlock{
			Level: level,
			Values: map[proto.ItemQuality]float64{
				proto.ItemQuality_ItemQualityJunk:      parseScalingValue(parts[1]),
				proto.ItemQuality_ItemQualityCommon:    parseScalingValue(parts[2]),
				proto.ItemQuality_ItemQualityUncommon:  parseScalingValue(parts[3]),
				proto.ItemQuality_ItemQualityRare:      parseScalingValue(parts[4]),
				proto.ItemQuality_ItemQualityEpic:      parseScalingValue(parts[5]),
				proto.ItemQuality_ItemQualityLegendary: parseScalingValue(parts[6]),
				proto.ItemQuality_ItemQualityArtifact:  parseScalingValue(parts[7]),
			},
		}
		dbc.ShieldBlockValues[level] = shieldBlock
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (dbc *DBC) ShieldBlockValue(class proto.ItemQuality, level int) float64 {
	if scaling, ok := dbc.ShieldBlockValues[level]; ok {
		if value, ok := scaling.Values[class]; ok {
			return value
		}
	}
	return 0.0 // return a default or error value if not found
}

func parseShieldBlockValue(value string) float64 {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0 // consider how to handle or log this error properly
	}
	return v
}
