package warlock

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

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
	Seeds       []*core.Spell
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
	AmplifyCurseAura  *core.Aura
	NightfallProcAura *core.Aura
	ImpShadowboltAura *core.Aura

	// Pets
	ActivePet  *WarlockPet
	Felhunter  *WarlockPet
	Felguard   *WarlockPet
	Imp        *WarlockPet
	Succubus   *WarlockPet
	Voidwalker *WarlockPet

	Doomguard *DoomguardPet
	Infernal  *InfernalPet

	serviceTimer *core.Timer
}

func (warlock *Warlock) GetCharacter() *core.Character {
	return &warlock.Character
}

func (warlock *Warlock) GetWarlock() *Warlock {
	return warlock
}

func (warlock *Warlock) ApplyTalents() {
	warlock.registerHarvestLife()
	warlock.registerArchimondesDarkness()
	warlock.registerKilJaedensCunning()
	warlock.registerMannarothsFury()
	warlock.registerGrimoireOfSupremacy()
	warlock.registerGrimoireOfSacrifice()
}

func (warlock *Warlock) Initialize() {

	warlock.registerCurseOfElements()
	doomguardInfernalTimer := warlock.NewTimer()
	warlock.registerSummonDoomguard(doomguardInfernalTimer)
	warlock.registerSummonInfernal(doomguardInfernalTimer)
	warlock.registerLifeTap()

	// Fel Armor 10% Stamina
	core.MakePermanent(
		warlock.RegisterAura(core.Aura{
			Label:    "Fel Armor",
			ActionID: core.ActionID{SpellID: 104938},
		}))
	warlock.MultiplyStat(stats.Stamina, 1.1)
	warlock.MultiplyStat(stats.Health, 1.1)

	// 5% int passive
	warlock.MultiplyStat(stats.Intellect, 1.05)
}

func (warlock *Warlock) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {

}

func (warlock *Warlock) AddPartyBuffs(partyBuffs *proto.PartyBuffs) {

}

func (warlock *Warlock) Reset(sim *core.Simulation) {
}

func (warlock *Warlock) OnEncounterStart(_ *core.Simulation) {
}

func NewWarlock(character *core.Character, options *proto.Player, warlockOptions *proto.WarlockOptions) *Warlock {
	warlock := &Warlock{
		Character: *character,
		Talents:   &proto.WarlockTalents{},
		Options:   warlockOptions,
	}
	core.FillTalentsProto(warlock.Talents.ProtoReflect(), options.TalentsString)
	warlock.EnableManaBar()
	warlock.AddStatDependency(stats.Strength, stats.AttackPower, 1)

	warlock.Infernal = warlock.NewInfernalPet()
	warlock.Doomguard = warlock.NewDoomguardPet()

	warlock.serviceTimer = character.NewTimer()
	warlock.registerPets()
	warlock.registerGrimoireOfService()

	return warlock
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
	WarlockSpellAgony
	WarlockSpellDrainLife
	WarlockSpellSeedOfCorruption
	WarlockSpellSeedOfCorruptionExposion
	WarlockSpellHellfire
	WarlockSpellImmolationAura
	WarlockSpellSearingPain
	WarlockSpellSummonDoomguard
	WarlockSpellDoomguardDoomBolt
	WarlockSpellSummonFelguard
	WarlockSpellFelGuardLegionStrike
	WarlockSpellFelGuardFelstorm
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
	WarlockSpellAll int64 = 1<<iota - 1

	WarlockShadowDamage = WarlockSpellCorruption | WarlockSpellUnstableAffliction | WarlockSpellDrainLife | WarlockSpellAgony |
		WarlockSpellShadowBolt | WarlockSpellSeedOfCorruptionExposion | WarlockSpellShadowBurn

	WarlockPeriodicShadowDamage = WarlockSpellCorruption | WarlockSpellUnstableAffliction |
		WarlockSpellDrainLife | WarlockSpellAgony

	WarlockFireDamage = WarlockSpellConflagrate | WarlockSpellImmolate | WarlockSpellIncinerate | WarlockSpellSoulFire |
		WarlockSpellSearingPain | WarlockSpellImmolateDot | WarlockSpellShadowBurn

	WarlockDoT = WarlockSpellCorruption | WarlockSpellUnstableAffliction |
		WarlockSpellDrainLife | WarlockSpellAgony | WarlockSpellImmolateDot

	WarlockSummonSpells = WarlockSpellSummonImp | WarlockSpellSummonSuccubus | WarlockSpellSummonFelhunter |
		WarlockSpellSummonFelguard

	WarlockAllSummons = WarlockSummonSpells | WarlockSpellSummonInfernal | WarlockSpellSummonDoomguard

	WarlockCurses = WarlockSpellCurseOfAgony | WarlockSpellCurseOfDoom | WarlockSpellCurseOfElements |
		WarlockSpellCurseOfRecklessness | WarlockSpellCurseOfTongues | WarlockSpellCurseOfWeakness

	WarlockAfflictionSpells = WarlockSpellCorruption | WarlockSpellCurseOfAgony | WarlockSpellCurseOfDoom | WarlockSpellCurseOfRecklessness | WarlockSpellCurseOfElements |
		WarlockSpellCurseOfTongues | WarlockSpellCurseOfWeakness | WarlockSpellDrainLife |
		WarlockSpellSeedOfCorruption

	WarlockDemonologySpells = WarlockAllSummons

	WarlockDestructionSpells = WarlockSpellHellfire | WarlockSpellImmolate | WarlockSpellIncinerate | WarlockSpellRainOfFire | WarlockSpellSearingPain |
		WarlockSpellShadowBolt | WarlockSpellSoulFire
)

// Called to handle custom resources
type WarlockSpellCastedCallback func(resultList core.SpellResultSlice, spell *core.Spell, sim *core.Simulation)
