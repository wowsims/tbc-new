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

	Talents *proto.PaladinTalents

	Forbearance *core.Aura

	PreviousSeal      *core.Aura
	PreviousJudgement *core.Spell
	PreviousSealSpell *core.Spell
	CurrentSeal       *core.Aura
	CurrentJudgement  *core.Spell
	CurrentSealSpell  *core.Spell

	// Timers for spells with multiple ranks
	consecrationTimer   *core.Timer
	hammerOfWrathTimer  *core.Timer
	holyShieldTimer     *core.Timer
	holyShockTimer      *core.Timer
	holyWrathTimer      *core.Timer
	exorcismTimer       *core.Timer
	avengersShieldTimer *core.Timer

	JudgementAuras []core.AuraArray

	T6_4pcAura *core.Aura
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
	paladin.RegisterSpiritualAttunement()
}

func (paladin *Paladin) registerSpells() {
	// Core abilities
	paladin.registerJudgement()
	ConsecrationRankMap.RegisterAll(paladin.registerConsecration)
	HammerOfWrathRankMap.RegisterAll(paladin.registerHammerOfWrath)
	HolyWrathRankMap.RegisterAll(paladin.registerHolyWrath)
	ExorcismRankMap.RegisterAll(paladin.registerExorcism)
	paladin.registerAvengingWrath()
	paladin.registerRighteousFury()

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
	paladin.CurrentSeal = nil
	paladin.PreviousSeal = nil
}

func (paladin *Paladin) OnEncounterStart(sim *core.Simulation) {
}

func NewPaladin(character *core.Character, talentsStr string, options *proto.PaladinOptions) *Paladin {
	paladin := &Paladin{
		Character: *character,
		Talents:   &proto.PaladinTalents{},
	}

	core.FillTalentsProto(paladin.Talents.ProtoReflect(), talentsStr, TalentTreeSizes)

	paladin.PseudoStats.CanParry = true
	paladin.PseudoStats.BaseDodgeChance += 0.0065
	paladin.PseudoStats.BaseParryChance += 0.05
	paladin.PseudoStats.BaseBlockChance += 0.05

	paladin.EnableManaBar()

	paladin.EnableAutoAttacks(paladin, core.AutoAttackOptions{
		MainHand:       paladin.WeaponFromMainHand(paladin.DefaultMeleeCritMultiplier()),
		AutoSwingMelee: true,
	})

	paladin.AddStatDependency(stats.Strength, stats.AttackPower, 2)
	paladin.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])
	paladin.AddStatDependency(stats.Intellect, stats.SpellCritPercent, core.CritPerIntMaxLevel[character.Class])
	paladin.AddStatDependency(stats.Agility, stats.DodgeRating, 1/25.0*core.DodgeRatingPerDodgePercent)
	paladin.AddStatDependency(stats.BonusArmor, stats.Armor, 1)

	return paladin
}
