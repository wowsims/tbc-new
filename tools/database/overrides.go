package database

import (
	"regexp"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

var OtherItemIdsToFetch = []string{}
var ConsumableOverrides = []*proto.Consumable{}
var ItemOverrides = []*proto.UIItem{}

// Keep these sorted by item ID.
var ItemAllowList = map[int32]struct{}{
	2140: {},
	2505: {},
}

// Keep these sorted by item ID.
var ItemDenyList = map[int32]struct{}{
	17782: {}, // talisman of the binding shard
	17783: {}, // talisman of the binding fragment
	17802: {}, // Deprecated version of Thunderfury
	18582: {},
	18583: {},
	18584: {},
	22736: {},
	24265: {},
	32384: {},
	32421: {},
	32422: {},
	32482: {},
	32824: {}, // Trashbringer
	33482: {},
	33350: {},
	34576: {}, // Battlemaster's Cruelty
	34577: {}, // Battlemaster's Depreavity
	34578: {}, // Battlemaster's Determination
	34579: {}, // Battlemaster's Audacity
	34580: {}, // Battlemaster's Perseverence
}

// Item icons to include in the DB, so they don't need to be separately loaded in the UI.
var ExtraItemIcons = []int32{
	// Pet foods
	33874,

	// Demonic Rune
	12662,

	// Food IDs
	27655,
	27657,
	27658,
	27664,
	33052,
	33825,
	33872,
	34753,
	34754,
	34756,
	34758,
	34767,
	34769,

	// Flask IDs
	13512,
	22851,
	22853,
	22854,
	22861,
	22866,
	33208,

	// Elixer IDs
	13452,
	13454,
	22824,
	22827,
	22831,
	22833,
	22834,
	22835,
	22840,
	28103,
	28104,
	31679,
	32062,
	32067,
	32068,

	// Potions / In Battle Consumes
	13442,
	20520,
	22105,
	22788,
	22828,
	22832,
	22837,
	22838,
	22839,
	22849,
	31677,

	// Thistle Tea
	7676,

	// Scrolls
	27498,
	27499,
	27500,
	27501,
	27502,
	27503,

	// Greater Drums
	185848,
	185850,
	185852,
}

// Item Ids of consumables to allow
var ConsumableAllowList = []int32{
	7676,  // Thisle Tea
	13442, // Migty Rage Potion
	20520, // Dark Rune
}
var ConsumableDenyList = []int32{}

// Raid buffs / debuffs
var SharedSpellsIcons = []int32{
	// Revitalize, Rejuv, WG
	26982,

	// Registered CD's
	10060,
	16190,
	29166,
	53530,
	33206,
	2825,

	17051,

	25898,
	25899,

	20140,
	8071,
	16293,

	14767,

	8075,

	20045,

	30808,
	19506,

	31869,
	31583,
	34460,

	12861,
	18696,

	20245,
	5675,
	16206,

	17007,
	34300,
	29801,

	8512,
	29193,

	31878,

	24907,

	3738,
	8227,

	31025,
	31035,
	6562,
	31033,
	16840,

	// Raid Debuffs
	8647,

	770,
	33602,
	702,
	18180,

	26016,
	12879,
	16862,

	30706,
	20337,

	12666,

	3043,
	29859,

	17800,
	17803,
	12873,
	28593,

	33198,
	1490,

	20271,

	11374,
	15235,

	27013,

	30708,
}

// If any of these match the item name, don't include it.
var DenyListNameRegexes = []*regexp.Regexp{
	regexp.MustCompile(`30 Epic`),
	regexp.MustCompile(`130 Epic`),
	regexp.MustCompile(`63 Blue`),
	regexp.MustCompile(`63 Green`),
	regexp.MustCompile(`66 Epic`),
	regexp.MustCompile(`90 Epic`),
	regexp.MustCompile(`90 Green`),
	regexp.MustCompile(`Boots 1`),
	regexp.MustCompile(`Boots 2`),
	regexp.MustCompile(`Boots 3`),
	regexp.MustCompile(`Bracer 1`),
	regexp.MustCompile(`Bracer 2`),
	regexp.MustCompile(`Bracer 3`),
	regexp.MustCompile(`DB\d`),
	regexp.MustCompile(`DEPRECATED`),
	regexp.MustCompile(`OLD`),
	regexp.MustCompile(`Deprecated`),
	regexp.MustCompile(`Deprecated: Keanna`),
	regexp.MustCompile(`Indalamar`),
	regexp.MustCompile(`Monster -`),
	regexp.MustCompile(`NEW`),
	regexp.MustCompile(`PH`),
	regexp.MustCompile(`QR XXXX`),
	regexp.MustCompile(`TEST`),
	regexp.MustCompile(`Test`),
	regexp.MustCompile(`Enchant Template`),
	regexp.MustCompile(`Arcane Amalgamation`),
	regexp.MustCompile(`Deleted`),
	regexp.MustCompile(`DELETED`),
	regexp.MustCompile(`zOLD`),
	regexp.MustCompile(`Archaic Spell`),
	regexp.MustCompile(`Well Repaired`),
	regexp.MustCompile(`Boss X`),
	regexp.MustCompile(`Adventurine`),
	regexp.MustCompile(`Sardonyx`),
	regexp.MustCompile(`Zyanite`),
	regexp.MustCompile(`zzold`),
	regexp.MustCompile(`Tom's`),
}

// Allows manual overriding for Gem fields in case WowHead is wrong.
var GemOverrides = []*proto.UIGem{
	{Id: 33131, Stats: stats.Stats{stats.AttackPower: 32, stats.RangedAttackPower: 32}.ToProtoArray()},
}
var GemAllowList = map[int32]struct{}{
	//22459: {}, // Void Sphere
	//36766: {}, // Bright Dragon's Eye
	//36767: {}, // Solid Dragon's Eye
}
var EnchantDenyListSpells = map[int32]struct{}{
	141168: {},
	141973: {},
	142173: {},
	142175: {},
	141170: {},
	141974: {},
	142177: {},
	141868: {},
	141984: {},
	141177: {},
	141981: {},
	141176: {},
	141978: {},
	141173: {},
	141975: {},
	141862: {},
	141983: {},
	141175: {},
	141977: {},
}
var EnchantDenyListItems = map[int32]struct{}{
	87583: {},
	89717: {},
	79061: {},
}
var GemDenyList = map[int32]struct{}{
	// pvp non-unique gems not in game currently.
	32735: {},
	33132: {},
	33137: {},
	33138: {},
	33139: {},
	33141: {},
	33142: {},
	35489: {},
	38545: {},
	38546: {},
	38547: {},
	38548: {},
	38549: {},
	38550: {},
}

var EnchantDenyList = map[int32]struct{}{
	3269: {}, // Truesilver Fishing Line
	3289: {}, // Skybreaker Whip/Riding Crop
	3315: {}, // Carrot on a Stick
	4671: {}, // Kyle's Test Enchantment
	4687: {}, // Enchant Weapon - Ninja (TEST VERSION)
	4717: {}, // Enchant Weapon - Pandamonium (DNT)
	5029: {}, // Custom - Jaina - Crackling Lightning
	5110: {}, // Lightweave Embroidery - Junk
}

var EnchantAllowList = []int32{
	368,  // Enchant Cloak - Greater Agility
	684,  // Enchant Gloves - Major Strength
	1593, // Bracer 24 AP
	2564, // Weapon 15 Agi
	2647, // Enchant Bracer - Brawn
}

// Note: EffectId is required for all enchants, because they are
// used by various importers/exporters
var EnchantOverrides = []*proto.UIEnchant{}
