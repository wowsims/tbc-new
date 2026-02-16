package paladin

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

var TalentTreeSizes = [3]int{20, 22, 22}

const JudgementAuraTag = "JudgementAura"

type Paladin struct {
	core.Character

	Seal    proto.PaladinSeal
	Talents *proto.PaladinTalents

	Forbearance *core.Aura

	PreviousSeal      *core.Aura
	PreviousJudgement *core.Spell
	CurrentSeal       *core.Aura
	CurrentJudgement  *core.Spell

	// Shared spells
	Judgement         *core.Spell
	Consecrations     []*core.Spell
	Exorcisms         []*core.Spell
	HammerOfWraths    []*core.Spell
	HolyWraths        []*core.Spell
	HolyLights        []*core.Spell
	FlashOfLights     []*core.Spell
	LayOnHands        []*core.Spell
	AvengingWrath     *core.Spell
	AvengingWrathAura *core.Aura

	// Seal Auras
	SealOfRighteousnessAuras []*core.Aura
	SealOfCommandAuras       []*core.Aura
	SealOfLightAuras         []*core.Aura
	SealOfWisdomAuras        []*core.Aura
	SealOfJusticeAuras       []*core.Aura
	SealOfTheCrusaderAuras   []*core.Aura
	SealOfBloodAuras         []*core.Aura
	SealOfVengeanceAuras     []*core.Aura

	// Seals
	SealOfRighteousness []*core.Spell
	SealOfCommand       []*core.Spell
	SealOfLight         []*core.Spell
	SealOfWisdom        []*core.Spell
	SealOfJustice       []*core.Spell
	SealOfTheCrusader   []*core.Spell
	SealOfBlood         []*core.Spell
	SealOfVengeance     []*core.Spell

	// Seal Judgements
	SealOfRighteousnessJudgements []*core.Spell
	SealOfCommandJudgements       []*core.Spell
	SealOfLightJudgements         []*core.Spell
	SealOfWisdomJudgements        []*core.Spell
	SealOfJusticeJudgements       []*core.Spell
	SealOfTheCrusaderJudgements   []*core.Spell
	SealOfBloodJudgements         []*core.Spell
	SealOfVengeanceJudgements     []*core.Spell

	// Talent-specific auras and spells
	DivineFavorAura         *core.Aura
	DivineIlluminationSpell *core.Spell
	DivineIlluminationAura  *core.Aura
	SanctityAura            *core.Aura
	HolyShields             []*core.Spell
	HolyShieldAuras         []*core.Aura
	AvengersShields         []*core.Spell
	CrusaderStrike          *core.Spell
	Repentance              *core.Spell
	HolyShocks              []*core.Spell
}

// Implemented by each Paladin spec.
type PaladinAgent interface {
	GetPaladin() *Paladin
}

func (paladin *Paladin) GetCharacter() *core.Character {
	return &paladin.Character
}

func (paladin *Paladin) GetPaladin() *Paladin {
	return paladin
}

func (paladin *Paladin) AddRaidBuffs(_ *proto.RaidBuffs) {
}

func (paladin *Paladin) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (paladin *Paladin) Initialize() {
	paladin.registerSpells()
}

func (paladin *Paladin) registerSpells() {
	// Core abilities
	paladin.registerJudgement()
	paladin.registerConsecration()
	paladin.registerHammerOfWrath()
	paladin.registerHolyWrath()
	paladin.registerExorcism()
	paladin.registerAvengingWrath()

	paladin.registerForbearance()

	// Seals
	paladin.registerSeals()

	// Auras
	paladin.registerAuras()

	// // Blessings
	// paladin.registerBlessings()

	// Healing spells
	paladin.registerHealingSpells()
}

func (paladin *Paladin) Reset(sim *core.Simulation) {
}

func (paladin *Paladin) OnEncounterStart(sim *core.Simulation) {
}

func NewPaladin(character *core.Character, talentsStr string, options *proto.PaladinOptions) *Paladin {
	paladin := &Paladin{
		Character: *character,
		Talents:   &proto.PaladinTalents{},
		Seal:      options.Seal,
	}

	core.FillTalentsProto(paladin.Talents.ProtoReflect(), talentsStr, TalentTreeSizes)

	paladin.PseudoStats.CanParry = true

	paladin.EnableManaBar()

	paladin.EnableAutoAttacks(paladin, core.AutoAttackOptions{
		MainHand:       paladin.WeaponFromMainHand(paladin.DefaultMeleeCritMultiplier()),
		AutoSwingMelee: true,
	})

	// TBC stat conversions
	// 1 Strength = 2 Attack Power
	paladin.AddStatDependency(stats.Strength, stats.AttackPower, 2)

	// Crit from Agility and Intellect
	paladin.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])
	paladin.AddStatDependency(stats.Intellect, stats.SpellCritPercent, core.CritPerIntMaxLevel[character.Class])

	// Dodge from Agility
	paladin.AddStatDependency(stats.Agility, stats.DodgeRating, 1/25.0*core.DodgeRatingPerDodgePercent)

	// Bonus Armor and Armor are treated identically for Paladins
	paladin.AddStatDependency(stats.BonusArmor, stats.Armor, 1)

	return paladin
}


func (paladin *Paladin) DefaultMeleeCritMultiplier() float64 {
	return paladin.Character.DefaultMeleeCritMultiplier()
}

func (paladin *Paladin) DefaultSpellCritMultiplier() float64 {
	return paladin.Character.DefaultSpellCritMultiplier()
}

func (paladin *Paladin) DefaultHealingCritMultiplier() float64 {
	return paladin.Character.DefaultHealingCritMultiplier()
}
