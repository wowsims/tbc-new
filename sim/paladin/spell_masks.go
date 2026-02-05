package paladin

// TBC Paladin Spell Masks
// These are used for spell identification, talent modifiers, and proc triggers

const (
	SpellMaskNone int64 = 0

	// Core Abilities
	SpellMaskJudgement int64 = 1 << iota
	SpellMaskConsecration
	SpellMaskExorcism
	SpellMaskHolyLight
	SpellMaskFlashOfLight
	SpellMaskLayOnHands
	SpellMaskHammerOfJustice
	SpellMaskCleanse
	SpellMaskDivineShield
	SpellMaskDivineProtection
	SpellMaskBlessingOfProtection
	SpellMaskHammerOfWrath

	// Seals
	SpellMaskSealOfRighteousness
	SpellMaskSealOfCommand
	SpellMaskSealOfLight
	SpellMaskSealOfWisdom
	SpellMaskSealOfJustice
	SpellMaskSealOfTheCrusader
	SpellMaskSealOfBlood
	SpellMaskSealOfVengeance

	// Judgement Effects (different from the Judgement spell itself)
	SpellMaskJudgementOfRighteousness
	SpellMaskJudgementOfCommand
	SpellMaskJudgementOfLight
	SpellMaskJudgementOfWisdom
	SpellMaskJudgementOfJustice
	SpellMaskJudgementOfTheCrusader
	SpellMaskJudgementOfBlood
	SpellMaskJudgementOfVengeance

	// Auras
	SpellMaskDevotionAura
	SpellMaskRetributionAura
	SpellMaskConcentrationAura
	SpellMaskFireResistanceAura
	SpellMaskFrostResistanceAura
	SpellMaskShadowResistanceAura
	SpellMaskSanctityAura

	// Blessings
	SpellMaskBlessingOfMight
	SpellMaskBlessingOfWisdom
	SpellMaskBlessingOfKings
	SpellMaskBlessingOfSalvation
	SpellMaskBlessingOfSanctuary

	// Talent Abilities
	SpellMaskDivineFavor
	SpellMaskDivineIllumination
	SpellMaskHolyShock
	SpellMaskHolyShield
	SpellMaskAvengersShield
	SpellMaskCrusaderStrike
	SpellMaskRepentance
)

// Composite masks
const (
	SpellMaskAllSeals = SpellMaskSealOfRighteousness |
		SpellMaskSealOfCommand |
		SpellMaskSealOfLight |
		SpellMaskSealOfWisdom |
		SpellMaskSealOfJustice |
		SpellMaskSealOfTheCrusader |
		SpellMaskSealOfBlood |
		SpellMaskSealOfVengeance

	SpellMaskAllJudgements = SpellMaskJudgementOfRighteousness |
		SpellMaskJudgementOfCommand |
		SpellMaskJudgementOfLight |
		SpellMaskJudgementOfWisdom |
		SpellMaskJudgementOfJustice |
		SpellMaskJudgementOfTheCrusader |
		SpellMaskJudgementOfBlood |
		SpellMaskJudgementOfVengeance

	SpellMaskAllAuras = SpellMaskDevotionAura |
		SpellMaskRetributionAura |
		SpellMaskConcentrationAura |
		SpellMaskFireResistanceAura |
		SpellMaskFrostResistanceAura |
		SpellMaskShadowResistanceAura |
		SpellMaskSanctityAura

	SpellMaskAllBlessings = SpellMaskBlessingOfMight |
		SpellMaskBlessingOfWisdom |
		SpellMaskBlessingOfKings |
		SpellMaskBlessingOfSalvation |
		SpellMaskBlessingOfSanctuary

	SpellMaskHealingSpells = SpellMaskHolyLight |
		SpellMaskFlashOfLight |
		SpellMaskLayOnHands |
		SpellMaskHolyShock

	// Spells that can trigger Seal of Command
	SpellMaskCanTriggerSealOfCommand = SpellMaskCrusaderStrike |
		SpellMaskJudgement
)
