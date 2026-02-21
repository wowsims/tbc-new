package warlock

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

var TalentTreeSizes = [3]int{21, 22, 21}

type Warlock struct {
	core.Character
	Talents *proto.WarlockTalents
	Options *proto.WarlockOptions

	// Base Spells
	Corruption  *core.Spell
	DrainLife   *core.Spell
	Hellfire    *core.Spell
	Immolate    *core.Spell
	Incinerate  *core.Spell
	SearingPain *core.Spell
	Seed        *core.Spell
	ShadowBolt  *core.Spell
	Soulfire    *core.Spell

	LifeTap *core.Spell

	// Curses
	CurseOfAgony         *core.Spell
	CurseOfDoom          *core.Spell
	CurseOfElements      *core.Spell
	CurseOfElementsAuras core.AuraArray
	CurseOfRecklessness  *core.Spell
	CurseOfTongues       *core.Spell

	// Talent Tree Spells
	AmplifyCurse       *core.Spell
	Conflagrate        *core.Spell
	Shadowburn         *core.Spell
	SiphonLife         *core.Spell
	UnstableAffliction *core.Spell

	// Auras
	AmplifyCurseAura       *core.Aura
	NightfallProcAura      *core.Aura
	ImpShadowboltAura      *core.Aura
	ShadowEmbraceAura      *core.Aura
	DemonicKnowledgeAura   *core.Aura
	MasterDemonologistAura *core.Aura

	// Pets
	ActivePet  *WarlockPet
	Felhunter  *WarlockPet
	Felguard   *WarlockPet
	Imp        *WarlockPet
	Succubus   *WarlockPet
	Voidwalker *WarlockPet

	// Armors
	FelArmor   *core.Aura
	DemonArmor *core.Aura

	serviceTimer *core.Timer
}

func (warlock *Warlock) GetCharacter() *core.Character {
	return &warlock.Character
}

func (warlock *Warlock) GetWarlock() *Warlock {
	return warlock
}

func RegisterWarlock() {
	core.RegisterAgentFactory(
		proto.Player_Warlock{},
		proto.Spec_SpecWarlock,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewWarlock(character, options, options.GetWarlock().Options.ClassOptions)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Warlock)
			if !ok {
				panic("Invalid spec value for Warlock!")
			}
			player.Spec = playerSpec
		},
	)
}

func (warlock *Warlock) ApplyTalents() {
	warlock.applyAfflictionTalents()
	warlock.applyDemonologyTalents()
	warlock.applyDestructionTalents()
}

func (warlock *Warlock) Initialize() {

	// Curses
	warlock.registerCurseOfElements()
	warlock.registerCurseOfDoom()
	warlock.registerCurseOfAgony()

	warlock.registerCorruption()
	warlock.registerSeed()
	warlock.registerDrainLife()
	warlock.registerHellfire()
	warlock.registerImmolate()
	warlock.registerIncinerate()
	warlock.registerLifeTap()
	warlock.registerShadowBolt()
	warlock.registerSearingPain()
	warlock.registerSiphonLifeSpell()
	warlock.registerSoulfire()

	warlock.registerArmors()

	warlock.PseudoStats.SelfHealingMultiplier = 1.0
	// doomguardInfernalTimer := warlock.NewTimer()
	// warlock.registerSummonDoomguard(doomguardInfernalTimer)
	// warlock.registerSummonInfernal(doomguardInfernalTimer)

}

func (warlock *Warlock) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {

}

func (warlock *Warlock) AddPartyBuffs(partyBuffs *proto.PartyBuffs) {

}

func (warlock *Warlock) Reset(sim *core.Simulation) {
}

func (warlock *Warlock) OnEncounterStart(sim *core.Simulation) {}

func NewWarlock(character *core.Character, options *proto.Player, warlockOptions *proto.WarlockOptions) *Warlock {
	warlock := &Warlock{
		Character: *character,
		Talents:   &proto.WarlockTalents{},
		Options:   warlockOptions,
	}
	core.FillTalentsProto(warlock.Talents.ProtoReflect(), options.TalentsString, TalentTreeSizes)
	warlock.EnableManaBar()
	warlock.AddStatDependency(stats.Strength, stats.AttackPower, 1)

	// warlock.Infernal = warlock.NewInfernalPet()
	// warlock.Doomguard = warlock.NewDoomguardPet()

	// warlock.serviceTimer = character.NewTimer()

	if !warlock.Options.SacrificeSummon {
		warlock.registerPets()
	}

	// warlock.registerGrimoireOfService()

	return warlock
}

func (warlock *Warlock) AfflictionCount(target *core.Unit) float64 {
	return float64(len(target.GetAurasWithTag("Affliction")))
}

// Agent is a generic way to access underlying warlock on any of the agents.
type WarlockAgent interface {
	GetWarlock() *Warlock
}

const (
	WarlockSpellFlagNone    int64 = 0
	WarlockSpellConflagrate int64 = 1 << iota
	WarlockSpellShadowBolt
	WarlockSpellImmolate
	WarlockSpellImmolateDot
	WarlockSpellIncinerate
	WarlockSpellSoulFire
	WarlockSpellShadowBurn
	WarlockSpellLifeTap
	WarlockSpellCorruption
	WarlockSpellUnstableAffliction
	WarlockSpellCurseOfAgony
	WarlockSpellCurseOfElements
	WarlockSpellDrainLife
	WarlockSpellSeedOfCorruption
	WarlockSpellSeedOfCorruptionExplosion
	WarlockSpellHellfire
	WarlockSpellImmolationAura
	WarlockSpellSearingPain
	WarlockSpellSummonDoomguard
	WarlockSpellDoomguardDoomBolt
	WarlockSpellSummonFelguard
	WarlockSpellSummonImp
	WarlockSpellImpFireBolt
	WarlockSpellSummonFelhunter
	WarlockSpellFelHunterShadowBite
	WarlockSpellSummonSuccubus
	WarlockSpellSuccubusLashOfPain
	WarlockSpellVoidwalkerTorment
	WarlockSpellSummonInfernal
	WarlockSpellRainOfFire
	WarlockSpellCurseOfDoom
	WarlockSpellCurseOfRecklessness
	WarlockSpellCurseOfWeakness
	WarlockSpellCurseOfTongues
	WarlockSpellSiphonLife
	WarlockSpellDrainSoul
	WarlockSpellShadowFury
	WarlockSpellShadowbolt2
	WarlockSpellAll int64 = 1<<iota - 1

	WarlockShadowDamage = WarlockSpellCorruption | WarlockSpellUnstableAffliction | WarlockSpellDrainLife | WarlockSpellCurseOfAgony |
		WarlockSpellShadowBolt | WarlockSpellSeedOfCorruptionExplosion | WarlockSpellSeedOfCorruption | WarlockSpellShadowBurn | WarlockSpellSiphonLife

	WarlockPeriodicShadowDamage = WarlockSpellCorruption | WarlockSpellUnstableAffliction |
		WarlockSpellDrainLife | WarlockSpellCurseOfAgony

	WarlockFireDamage = WarlockSpellConflagrate | WarlockSpellImmolate | WarlockSpellIncinerate | WarlockSpellSoulFire |
		WarlockSpellSearingPain | WarlockSpellImmolateDot | WarlockSpellShadowBurn

	WarlockDoT = WarlockSpellCorruption | WarlockSpellUnstableAffliction |
		WarlockSpellDrainLife | WarlockSpellCurseOfAgony | WarlockSpellImmolateDot

	WarlockSummonSpells = WarlockSpellSummonImp | WarlockSpellSummonSuccubus | WarlockSpellSummonFelhunter |
		WarlockSpellSummonFelguard

	WarlockAllSummons = WarlockSummonSpells | WarlockSpellSummonInfernal | WarlockSpellSummonDoomguard

	WarlockContagionSpells = WarlockSpellCurseOfAgony | WarlockSpellCorruption | WarlockSpellSeedOfCorruption | WarlockSpellSeedOfCorruptionExplosion

	WarlockShadowEmbraceSpells = WarlockSpellCorruption | WarlockSpellCurseOfAgony | WarlockSpellSiphonLife | WarlockSpellSeedOfCorruption

	WarlockCurses = WarlockSpellCurseOfAgony | WarlockSpellCurseOfDoom | WarlockSpellCurseOfElements |
		WarlockSpellCurseOfRecklessness | WarlockSpellCurseOfTongues | WarlockSpellCurseOfWeakness

	WarlockSoulLeechSpells = WarlockSpellShadowBolt | WarlockSpellShadowBurn | WarlockSpellSoulFire |
		WarlockSpellIncinerate | WarlockSpellSearingPain | WarlockSpellConflagrate

	WarlockAfflictionSpells = WarlockSpellCorruption | WarlockSpellCurseOfAgony | WarlockSpellCurseOfDoom | WarlockSpellCurseOfRecklessness | WarlockSpellCurseOfElements |
		WarlockSpellCurseOfTongues | WarlockSpellCurseOfWeakness | WarlockSpellDrainLife |
		WarlockSpellSeedOfCorruption

	WarlockDemonologySpells = WarlockAllSummons

	WarlockDestructionSpells = WarlockSpellHellfire | WarlockSpellImmolate | WarlockSpellIncinerate | WarlockSpellRainOfFire | WarlockSpellSearingPain |
		WarlockSpellShadowBolt | WarlockSpellSoulFire
)

// Called to handle custom resources
type WarlockSpellCastedCallback func(resultList core.SpellResultSlice, spell *core.Spell, sim *core.Simulation)
