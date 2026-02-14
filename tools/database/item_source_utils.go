package database

import (
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/tbc/sim/core/proto"
)

func InferPhase(item *proto.UIItem) int32 {
	ilvl := item.ScalingOptions[int32(0)].Ilvl
	name := item.Name

	// PvE Sets
	if item.SetId > 0 {
		if ilvl == 120 {
			return 1
		}
		if ilvl == 133 {
			return 2
		}
		if ilvl == 146 {
			return 3
		}
		if ilvl == 154 {
			return 5
		}
	}

	//AtlasLoot‐style source checks
	for _, src := range item.Sources {
		if craft := src.GetCrafted(); craft != nil {
			if strings.Contains(item.Name, "Figurine") && ilvl == 125 {
				return 5
			}
			if ilvl <= 127 {
				return 1
			}
			if ilvl == 146 || ilvl == 136 { // T5 + Vortex BoP Crafts
				return 2
			}
			if ilvl >= 128 && ilvl <= 141 { // T6 Crafts
				return 3
			}
			if ilvl == 159 { // SWP Crafts
				return 5
			}
		}
		if drop := src.GetDrop(); drop != nil {
			if slices.Contains([]int32{3457, 3923, 3836}, drop.ZoneId) { // Kara, Gruul, Mag
				return 1
			}
			if slices.Contains([]int32{3845, 3607}, drop.ZoneId) { // TK, SSC
				return 2
			}
			if slices.Contains([]int32{3606, 3959}, drop.ZoneId) { // MH, BT
				return 3
			}
			if slices.Contains([]int32{3805}, drop.ZoneId) { // ZA
				return 4
			}
			if slices.Contains([]int32{4075, 4131}, drop.ZoneId) { // SWP, MGT
				return 5
			}
			if ilvl <= 117 {
				return 1
			}
		}
	}

	// PvP Sets
	if item.Quality == proto.ItemQuality_ItemQualityEpic && ilvl > 115 {
		switch {
		case strings.Contains(name, "Merciless Gladiator"),
			strings.Contains(name, "Veteran's"):
			return 2
		case strings.Contains(name, "Vengeful Gladiator"),
			strings.Contains(name, "Vindicator's"):
			return 3
		case strings.Contains(name, "Brutal Gladiator"),
			strings.Contains(name, "Guardian's"):
			return 5
		case strings.Contains(name, "Gladiator's"),
			strings.Contains(name, "Marshal's"),
			strings.Contains(name, "General's"),
			strings.Contains(name, "Sergeant's"):
			return 1
		}
	}

	if ilvl <= 117 || (ilvl <= 120 && item.Quality == proto.ItemQuality_ItemQualityUncommon) { // Catch-all for Pre-TBC, Outlands Questing, Random Green, and Heroic Dungeon Gear
		return 1
	}

	if (ilvl == 120 || ilvl == 125) && item.Quality == proto.ItemQuality_ItemQualityEpic { // P1 World Boss, Mag Head Rings
		return 1
	}

	if strings.Contains(item.Name, "Violet Signet") { // Kara Rep Rings
		return 1
	}

	if ilvl == 138 { // TK Quest Necks
		return 2
	}

	if strings.Contains(item.Name, "Band of Eternity") || strings.Contains(item.Name, "Band of the Eternal") { // Hyjal Rep Rings
		return 3
	}

	if ilvl == 128 || ilvl == 136 || ilvl == 133 || ilvl == 132 { // ZA Badge Gear
		return 4
	}

	if ilvl >= 159 || ilvl == 141 || ilvl == 146 || ilvl == 135 || item.Id == 34407 { // SWP Mote Turn-ins, SWP Badge Armor, SWP Badge Weapons, The 2 Ring, Tranquil Moonlight Wraps (SWP Mote item but lower ilvl???)
		return 5
	}

	println("Uncategorized Item: " + item.Name + " with Ilvl: " + strconv.FormatInt(int64(ilvl), 10))
	return 0
}
