package database

import (
	"regexp"
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
	"github.com/wowsims/tbc/tools/database/dbc"
)

// Allows you to ignore certain Spell Effects that the Sim does not support.
// This prevents them from being added to the missing effects list to prevent confusion towards users.
// Empty array means ignore all effects of that type, otherwise it will be ignored
// based on EffectMiscValue_0
var IgnoreSpellEffectByAuraType = map[dbc.EffectAuraType][]int{
	dbc.A_MOD_MECHANIC_RESISTANCE: {},
	dbc.A_MOD_STEALTH:             {},
	dbc.A_MOD_STEALTH_DETECT:      {},
	dbc.A_MOD_STEALTH_LEVEL:       {},
	dbc.A_MOD_DECREASE_SPEED:      {},
	dbc.A_MOD_INVISIBILITY:        {},
	dbc.A_MOD_INVISIBILITY_DETECT: {},
	dbc.A_MOD_SKILL: {
		356, // Fishing Skill
		393, // Skinning Skill
	},
	dbc.A_MOD_INCREASE_MOUNTED_SPEED:        {},
	dbc.A_MOD_MOUNTED_SPEED_ALWAYS:          {},
	dbc.A_MOD_MOUNTED_SPEED_NOT_STACK:       {},
	dbc.A_MOD_INCREASE_MOUNTED_FLIGHT_SPEED: {},
	dbc.A_MOD_MOUNTED_FLIGHT_SPEED_ALWAYS:   {},
	dbc.A_TRANSFORM:                         {},
	dbc.A_MECHANIC_IMMUNITY:                 {},
	dbc.A_TRACK_CREATURES:                   {},
	dbc.A_TRACK_RESOURCES:                   {},
	dbc.A_FAR_SIGHT:                         {},
}

var IgnoreSpellEffectBySpellEffectType = map[dbc.SpellEffectType][]int{
	dbc.E_CREATE_ITEM:    {},
	dbc.E_SUMMON:         {},
	dbc.E_TELEPORT_UNITS: {},
}

// Spells that are flavour rather than mechanics, keyed by spell ID so that reporting them as
// missing effects can be suppressed without hiding anything real.
//
// Keyed per spell on purpose. The obvious shortcut - treat an effect made only of A_DUMMY auras
// as noise, which is what MoP does - does not hold in TBC, where A_DUMMY is how a whole class of
// real item effects is encoded: every relic, idol, libram and totem bonus, plus Zandalarian Hero
// Medallion, Thick Obsidian Breastplate and Ashtongue Talisman of Zeal. That rule removed 48
// entries here and only a handful of them were noise.
var IgnoreMissingEffectBySpellID = map[int]string{
	16372: "Seal of Ascension - no tooltip and no mechanic",
	43873: "Headless Horseman Laugh - holiday flavour aura",
}

var OtherItemIdsToFetch = []string{}
var ConsumableOverrides = []*proto.Consumable{
	{Id: 23334, CooldownDuration: int32(time.Hour.Seconds())}, // Cracked Power Core
	{Id: 23381, CooldownDuration: int32(time.Hour.Seconds())}, // Chipped Power Core
}
var ItemOverrides = []*proto.UIItem{
	{Id: 32649, Phase: 3},  // Medallion of Karabor
	{Id: 32658, Phase: 2},  // Badge of Tenacity
	{Id: 32757, Phase: 3},  // Blessed Medallion of Karabor
	{Id: 278774, Phase: 2}, // Cloak of the Frigid Winds (ilvl 128)
	{Id: 278819, Phase: 2}, // The Frost Lord's War Cloak (ilvl 128)
	{Id: 278823, Phase: 2}, // Icebound Cloak (ilvl 128)
	{Id: 278827, Phase: 2}, // Amulet of Bitter Hatred (ilvl 128)
	{Id: 278833, Phase: 2}, // Choker of the Arctic Flow (ilvl 128)
	{Id: 278838, Phase: 2}, // Amulet of Glacial Tranquility (ilvl 128)
	{Id: 278847, Phase: 2}, // Hailstone Pendant (ilvl 128)
	{Id: 278953, Phase: 2}, // Frostscythe of Lord Ahune (ilvl 128)
	{Id: 279240, Phase: 2}, // Shroud of Winter's Chill (ilvl 128)

	{Id: 34665, Phase: 5},
	{Id: 34666, Phase: 5},
	{Id: 34667, Phase: 5},
	{Id: 34670, Phase: 5},
	{Id: 34671, Phase: 5},
	{Id: 34672, Phase: 5},
	{Id: 34673, Phase: 5},
	{Id: 34674, Phase: 5},
	{Id: 34675, Phase: 5},
	{Id: 34676, Phase: 5},
	{Id: 34677, Phase: 5},
	{Id: 34678, Phase: 5},
	{Id: 34679, Phase: 5},
	{Id: 34680, Phase: 5},
}

// Keep these sorted by item ID.
var ItemAllowList = map[int32]struct{}{
	1168:   {}, // Skullflame Shield
	2140:   {},
	2505:   {},
	8345:   {}, // Wolfshead Helm
	11815:  {}, // Hand of Justice
	18168:  {}, // Force Reactive Disk
	19337:  {}, // The Black Book
	32649:  {}, // Medallion of Karabor (ItemClass "Quest" in game files)
	32757:  {}, // Blessed Medallion of Karabor (ItemClass "Quest" in game files)
	186071: {}, // Communal Totem of Lightning
	186073: {}, // Communal Totem of the Storm
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
	23363: {}, // Titanic Breastplate
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

	// Ahune items with the wrong ilvl
	35494:  {}, // Shroud of Winter's Chill (ilvl 110)
	35495:  {}, // The Frost Lord's War Cloak (ilvl 110)
	35496:  {}, // Icebound Cloak (ilvl 110)
	35497:  {}, // Cloak of the Frigid Winds (ilvl 110)
	35507:  {}, // Amulet of Bitter Hatred (ilvl 110)
	35508:  {}, // Choker of the Arctic Flow (ilvl 110)
	35509:  {}, // Amulet of Glacial Tranquility (ilvl 110)
	35511:  {}, // Hailstone Pendant (ilvl 110)
	35514:  {}, // Frostscythe of Lord Ahune (ilvl 110)
	278752: {}, // Cloak of the Frigid Winds (ilvl 115)
	278775: {}, // Cloak of the Frigid Winds (ilvl 141)
	278807: {}, // Cloak of the Frigid Winds (ilvl 154)
	278817: {}, // The Frost Lord's War Cloak (ilvl 115)
	278820: {}, // The Frost Lord's War Cloak (ilvl 141)
	278821: {}, // The Frost Lord's War Cloak (ilvl 154)
	278822: {}, // Icebound Cloak (ilvl 115)
	278824: {}, // Icebound Cloak (ilvl 141)
	278825: {}, // Icebound Cloak (ilvl 154)
	278826: {}, // Amulet of Bitter Hatred (ilvl 115)
	278828: {}, // Amulet of Bitter Hatred (ilvl 141)
	278829: {}, // Amulet of Bitter Hatred (ilvl 154)
	278832: {}, // Choker of the Arctic Flow (ilvl 115)
	278834: {}, // Choker of the Arctic Flow (ilvl 141)
	278835: {}, // Choker of the Arctic Flow (ilvl 154)
	278837: {}, // Amulet of Glacial Tranquility (ilvl 115)
	278839: {}, // Amulet of Glacial Tranquility (ilvl 141)
	278840: {}, // Amulet of Glacial Tranquility (ilvl 154)
	278846: {}, // Hailstone Pendant (ilvl 115)
	278949: {}, // Hailstone Pendant (ilvl 141)
	278950: {}, // Hailstone Pendant (ilvl 154)
	278952: {}, // Frostscythe of Lord Ahune (ilvl 115)
	278954: {}, // Frostscythe of Lord Ahune (ilvl 141)
	278955: {}, // Frostscythe of Lord Ahune (ilvl 154)
	279239: {}, // Shroud of Winter's Chill (ilvl 115)
	279241: {}, // Shroud of Winter's Chill (ilvl 141)
	279242: {}, // Shroud of Winter's Chill (ilvl 154)

	// TBC - Brewfest - Old items
	// NOTE: Enable the correct ilvl once the holiday comes around
	281726: {}, // Dark Iron Smoking Pipe (ilvl 115)
	281734: {}, // Dark Iron Smoking Pipe (ilvl 128)
	281735: {}, // Dark Iron Smoking Pipe (ilvl 141)
	281736: {}, // Dark Iron Smoking Pipe (ilvl 154)
	281737: {}, // Empty Mug of Direbrew (ilvl 115)
	281738: {}, // Empty Mug of Direbrew (ilvl 128)
	281739: {}, // Empty Mug of Direbrew (ilvl 141)
	281740: {}, // Empty Mug of Direbrew (ilvl 154)
	281741: {}, // Coren's Lucky Coin (ilvl 115)
	281742: {}, // Coren's Lucky Coin (ilvl 128)
	281743: {}, // Coren's Lucky Coin (ilvl 141)
	281744: {}, // Coren's Lucky Coin (ilvl 154)
	281745: {}, // Direbrew Hops (ilvl 115)
	281747: {}, // Direbrew Hops (ilvl 128)
	281748: {}, // Direbrew Hops (ilvl 141)
	281749: {}, // Direbrew Hops (ilvl 154)
	281892: {}, // Balebrew Charm (ilvl 128)
	281893: {}, // Balebrew Charm (ilvl 141)
	281894: {}, // Brightbrew Charm (ilvl 128)
	281895: {}, // Brightbrew Charm (ilvl 141)
	281897: {}, // Brightbrew Charm (ilvl 154)
	281898: {}, // Brightbrew Charm (ilvl 115)
	281899: {}, // Balebrew Charm (ilvl 154)
	281900: {}, // Balebrew Charm (ilvl 115)
	281901: {}, // Direbrew's Shanker (ilvl 115)
	281902: {}, // Direbrew's Shanker (ilvl 128)
	281903: {}, // Direbrew's Shanker (ilvl 141)
	281904: {}, // Direbrew's Shanker (ilvl 154)

	// TBC - Hallows End - Old items
	// NOTE: Enable the correct ilvl once the holiday comes around
	281905: {}, // Ring of Ghoulish Delight (ilvl 115)
	281906: {}, // Ring of Ghoulish Delight (ilvl 128)
	281907: {}, // Ring of Ghoulish Delight (ilvl 141)
	281908: {}, // Ring of Ghoulish Delight (ilvl 154)
	281909: {}, // The Horseman's Signet Ring (ilvl 115)
	281910: {}, // The Horseman's Signet Ring (ilvl 128)
	281911: {}, // The Horseman's Signet Ring (ilvl 141)
	281912: {}, // The Horseman's Signet Ring (ilvl 154)
	281913: {}, // Witches Band (ilvl 115)
	281914: {}, // Witches Band (ilvl 128)
	281915: {}, // Witches Band (ilvl 141)
	281916: {}, // Witches Band (ilvl 154)
	281917: {}, // The Horseman's Helm (ilvl 115)
	281918: {}, // The Horseman's Helm (ilvl 128)
	281919: {}, // The Horseman's Helm (ilvl 141)
	281920: {}, // The Horseman's Helm (ilvl 154)
	281921: {}, // The Horseman's Blade (ilvl 115)
	281922: {}, // The Horseman's Blade (ilvl 128)
	281923: {}, // The Horseman's Blade (ilvl 141)
	281924: {}, // The Horseman's Blade (ilvl 154)
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
	9224,
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
	9088,  // Gift of Arthas
	9155,  // Arcane Elixir
	9224,  // Elixir of Demonslaying
	13442, // Migty Rage Potion
	13452, // Elixir of the Mongoose
	13454, // Greater Arcane Elixir
	12662, // Demonic Rune
	22105, // Master Healthstone
	22788, // Flamecap
	22797, // Nightmare Seed
	23334, // Cracked Power Core
	23381, // Chipped Power Core
}
var ConsumableDenyList = []int32{
	32762, // Rulkster's Brain Juice
	32902, // Bottled Nethergon Energy
}

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
	// Raid buffs, debuffs and shared consumables.
	603,    // Curse of Doom
	688,    // Summon Imp
	691,    // Summon Felhunter
	697,    // Summon Voidwalker
	704,    // Curse of Recklessness
	706,    // Demon Armor
	712,    // Summon Succubus
	6117,   // Mage Armor
	7302,   // Ice Armor
	8024,   // Flametongue Weapon
	8033,   // Frostbrand Weapon
	8232,   // Windfury Weapon
	10538,  // Fire Resistance Totem
	12470,  // Fire Nova
	13339,  // Fire Blast
	13376,  // Fire Shield
	13889,  // Minor Speed
	14325,  // Hunter's Mark
	17768,  // Wolfshead Helm
	18803,  // Focus
	19615,  // Frenzy Effect
	20574,  // Axe Specialization
	20575,  // Command
	20576,  // Command
	20594,  // Stoneform
	20595,  // Gun Specialization
	20597,  // Sword Specialization
	20864,  // Mace Specialization
	23110,  // Dash
	23563,  // Enhanced Battle Shout
	25076,  // Cobra Reflexes
	25312,  // Divine Spirit
	25389,  // Power Word: Fortitude
	25433,  // Shadow Protection
	25485,  // Rockbiter Weapon
	25489,  // Flametongue Weapon
	25500,  // Frostbrand Weapon
	25505,  // Windfury Weapon
	25560,  // Frost Resistance Totem
	25574,  // Nature Resistance Totem
	25607,  // Jade Pendant of Blasting
	25894,  // Greater Blessing of Wisdom
	25895,  // Greater Blessing of Salvation
	26654,  // Sweeping Strikes
	26991,  // Gift of the Wild
	26992,  // Thorns
	27045,  // Aspect of the Wild
	27050,  // Bite
	27051,  // Screech
	27141,  // Greater Blessing of Might
	27143,  // Greater Blessing of Wisdom
	27163,  // Judgement of Light
	27169,  // Greater Blessing of Sanctuary
	27187,  // Deadly Poison VII
	27267,  // Firebolt
	27268,  // Blood Pact
	27270,  // Torment
	27274,  // Lash of Pain
	28093,  // Lightning Speed
	28142,  // Power of the Guardian
	28143,  // Power of the Guardian
	28878,  // Inspiring Presence
	28880,  // Gift of the Naaru
	29414,  // Haste
	30168,  // Shadow Cage
	30223,  // Cleave
	30576,  // Quake
	30616,  // Blast Nova
	30619,  // Cleave
	32176,  // Stormstrike
	32850,  // Demonic Frenzy
	33697,  // Blood Fury
	33876,  // Mangle (Cat)
	34026,  // Kill Command
	34027,  // Kill Command
	34258,  // Justice
	34260,  // Justice
	34471,  // The Beast Within
	34775,  // Haste
	35081,  // Band of the Eternal Champion
	35084,  // Band of the Eternal Sage
	35298,  // Gore
	35476,  // Drums of Battle
	37174,  // Perceived Weakness
	37186,  // Increased Judgement of Crusader
	37212,  // Improved Wrath of Air Totem
	37223,  // Improved Strength of Earth
	37444,  // Arcane Madness
	37445,  // Mana Surge
	37658,  // Electrical Charge
	37661,  // Lightning Bolt
	38390,  // Improved Aspect of the Viper
	38392,  // Improved Steady Shot
	38393,  // Improved Shadow Bolt and Incinerate
	38398,  // Reduced Cleave Cost
	38421,  // Improved Spiritual Attunement
	38422,  // Improved Consecration
	38436,  // Improved Lightning Bolt
	38437,  // Totemic Mastery
	38447,  // Improved Mangle
	38927,  // Fel Ache
	39374,  // Prayer of Shadow Protection
	39445,  // Vengeance
	40293,  // Siphon Essence
	40407,  // Illidan Tank Shield
	40477,  // Forceful Strike
	41434,  // The Twin Blades of Azzinoth
	41435,  // The Twin Blades of Azzinoth
	42084,  // Fury of the Crashing Waves
	44949,  // Whirlwind
	369770, // Tinnitus
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
	regexp.MustCompile(`Stabilized Eternium Scope`),
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
	369,  // Enchant Bracer - Major Intellect
	684,  // Enchant Gloves - Major Strength
	963,  // Enchant Weapon - Major Striking
	1593, // Bracer 24 AP
	1594, // Gloves 26 AP
	1900, // Enchant Weapon - Crusader
	2564, // Weapon 15 Agi
	2583, // Presence of Might
	2588, // Presence of Sight
	2647, // Enchant Bracer - Brawn
	2659, // Enchant Chest - Exceptional Health
}

// Note: EffectId is required for all enchants, because they are
// used by various importers/exporters
// Note: ItemId, SpellId and Name are part of the enchant DB key, so they must
// match the generated enchant or the override is added as a separate entry.
var EnchantOverrides = []*proto.UIEnchant{
	// The head/leg resistance kits grant their stats through an equip spell that no
	// longer exists in the client data, so they have to be filled in by hand.
	{EffectId: 2681, ItemId: 22635, SpellId: 28162, Name: "Savage Guard", Stats: stats.Stats{stats.NatureResistance: 10}.ToProtoArray()},
	{EffectId: 2682, ItemId: 22636, SpellId: 28164, Name: "Ice Guard", Stats: stats.Stats{stats.FrostResistance: 10}.ToProtoArray()},
	{EffectId: 2683, ItemId: 22638, SpellId: 28166, Name: "Shadow Guard", Stats: stats.Stats{stats.ShadowResistance: 10}.ToProtoArray()},
}
