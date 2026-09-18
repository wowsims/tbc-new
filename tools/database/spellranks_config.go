package database

import "strings"

// Which spell families get generated rank tables, and for which class.
//
// One line per family, anchored on the max-rank spell ID that is already registered in the sim. The
// anchor is identity - the ladder is rebuilt from that spell's name and skill lines, and the generator
// hard-fails if the anchor does not come back as the highest rank it found.
//
// SkillLineAbility.SupercedesSpell is 0 for every caster spell in this build, so the chain cannot be
// walked; name plus "Rank N" plus the class bit is what there is.
type RankFamily struct {
	Var      string
	Class    string
	ClassBit int
	Name     string
	Anchor   int32

	// Ranks the sim deliberately does not register. Emitted anyway - the table is the full ladder and
	// the spec picks from it with .Ranks(...) - but listed here so a reader knows the omission is a
	// choice rather than a resolver failure.
	Unregistered []int32

	// A rank whose spell ID the resolver cannot pick on its own. Seal of Righteousness rank 1 is the
	// case this exists for: 20154 and 21084 are both rank 1, same effect shape, same 20 mana.
	Pin map[int32]int32
}

// The struct field a spell reads its table through: genConsecrationRanks -> genRanks.Consecration.
func (f RankFamily) Field() string {
	return strings.TrimSuffix(strings.TrimPrefix(f.Var, "gen"), "Ranks")
}

const (
	ClassBitPaladin = 2
	ClassBitPriest  = 16
	ClassBitShaman  = 64
	ClassBitMage    = 128
	ClassBitDruid   = 1024
)

var SpellRankConfigs = []RankFamily{
	{Var: "genConsecrationRanks", Class: "paladin", ClassBit: ClassBitPaladin, Name: "Consecration", Anchor: 27173},
	{Var: "genHammerOfWrathRanks", Class: "paladin", ClassBit: ClassBitPaladin, Name: "Hammer of Wrath", Anchor: 27180},
	{Var: "genHolyWrathRanks", Class: "paladin", ClassBit: ClassBitPaladin, Name: "Holy Wrath", Anchor: 27139},
	{Var: "genExorcismRanks", Class: "paladin", ClassBit: ClassBitPaladin, Name: "Exorcism", Anchor: 27138},
	{Var: "genHolyLightRanks", Class: "paladin", ClassBit: ClassBitPaladin, Name: "Holy Light", Anchor: 27136},
	{Var: "genFlashOfLightRanks", Class: "paladin", ClassBit: ClassBitPaladin, Name: "Flash of Light", Anchor: 27137},
	{Var: "genLayOnHandsRanks", Class: "paladin", ClassBit: ClassBitPaladin, Name: "Lay on Hands", Anchor: 27154},
	{Var: "genHolyShieldRanks", Class: "paladin", ClassBit: ClassBitPaladin, Name: "Holy Shield", Anchor: 27179},
	{Var: "genHolyShockRanks", Class: "paladin", ClassBit: ClassBitPaladin, Name: "Holy Shock", Anchor: 33072},
	{Var: "genAvengersShieldRanks", Class: "paladin", ClassBit: ClassBitPaladin, Name: "Avenger's Shield", Anchor: 32700},

	{Var: "genMindBlastRanks", Class: "priest", ClassBit: ClassBitPriest, Name: "Mind Blast", Anchor: 25375},
	{Var: "genMindFlayRanks", Class: "priest", ClassBit: ClassBitPriest, Name: "Mind Flay", Anchor: 25387},
	{Var: "genShadowWordPainRanks", Class: "priest", ClassBit: ClassBitPriest, Name: "Shadow Word: Pain", Anchor: 25368},
	{Var: "genShadowWordDeathRanks", Class: "priest", ClassBit: ClassBitPriest, Name: "Shadow Word: Death", Anchor: 32996},
	{Var: "genSmiteRanks", Class: "priest", ClassBit: ClassBitPriest, Name: "Smite", Anchor: 25364},
	{Var: "genDevouringPlagueRanks", Class: "priest", ClassBit: ClassBitPriest, Name: "Devouring Plague", Anchor: 25467},
	{Var: "genHolyNovaRanks", Class: "priest", ClassBit: ClassBitPriest, Name: "Holy Nova", Anchor: 25331},
	{Var: "genStarshardsRanks", Class: "priest", ClassBit: ClassBitPriest, Name: "Starshards", Anchor: 25446},
	{Var: "genVampiricTouchRanks", Class: "priest", ClassBit: ClassBitPriest, Name: "Vampiric Touch", Anchor: 34917},

	{Var: "genLightningBoltRanks", Class: "shaman", ClassBit: ClassBitShaman, Name: "Lightning Bolt", Anchor: 25449},
	{Var: "genChainLightningRanks", Class: "shaman", ClassBit: ClassBitShaman, Name: "Chain Lightning", Anchor: 25442},

	{Var: "genFlamestrikeRanks", Class: "mage", ClassBit: ClassBitMage, Name: "Flamestrike", Anchor: 27086,
		Unregistered: []int32{1, 2, 3, 4, 5}},

	{Var: "genStarfireRanks", Class: "druid", ClassBit: ClassBitDruid, Name: "Starfire", Anchor: 26986,
		Unregistered: []int32{1, 2, 3, 4, 5, 7}},
}
