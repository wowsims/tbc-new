package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

var TalentTreeSizes = [3]int{23, 21, 22}

type WarriorInputs struct {
	DefaultShout  proto.WarriorShout
	DefaultStance proto.WarriorStance

	StartingRage          float64
	QueueDelay            int32
	StanceSnapshot        bool
	HasBsSolarianSapphire bool
	HasBsT2               bool
}

const (
	SpellFlagBleed = core.SpellFlagAgentReserved1
)

const (
	SpellMaskNone int64 = 0
	// Abilities that don't cost rage and aren't attacks
	SpellMaskBattleShout int64 = 1 << iota
	SpellMaskCommandingShout
	SpellMaskBerserkerRage
	SpellMaskRallyingCry
	SpellMaskRecklessness
	SpellMaskDeathWish
	SpellMaskRetaliation
	SpellMaskRetaliationHit
	SpellMaskRampage
	SpellMaskShieldWall
	SpellMaskLastStand
	SpellMaskCharge
	SpellMaskIntercept
	SpellMaskDemoralizingShout

	// Stances
	SpellMaskBattleStance
	SpellMaskBerserkerStance
	SpellMaskDefensiveStance

	// Special attacks
	SpellMaskRend
	SpellMaskDeepWounds
	SpellMaskSweepingStrikes
	SpellMaskSweepingStrikesHit
	SpellMaskSweepingStrikesNormalizedHit
	SpellMaskCleave
	SpellMaskDevastate
	SpellMaskExecute
	SpellMaskHeroicStrike
	SpellMaskOverpower
	SpellMaskRevenge
	SpellMaskSlam
	SpellMaskSweepingSlam
	SpellMaskSunderArmor
	SpellMaskThunderClap
	SpellMaskWhirlwind
	SpellMaskWhirlwindOh
	SpellMaskShieldSlam
	SpellMaskShieldBash
	SpellMaskBloodthirst
	SpellMaskMortalStrike
	SpellMaskWildStrike
	SpellMaskShieldBlock
	SpellMaskHamstring
	SpellMaskPummel

	WarriorSpellLast
	WarriorSpellsAll = WarriorSpellLast<<1 - 1

	SpellMaskShouts             = SpellMaskCommandingShout | SpellMaskBattleShout | SpellMaskDemoralizingShout
	SpellMaskDirectDamageSpells = SpellMaskSweepingStrikesHit | SpellMaskSweepingStrikesNormalizedHit |
		SpellMaskCleave | SpellMaskExecute | SpellMaskHeroicStrike | SpellMaskOverpower |
		SpellMaskRevenge | SpellMaskSlam | SpellMaskSweepingSlam | SpellMaskShieldBash | SpellMaskSunderArmor |
		SpellMaskThunderClap | SpellMaskWhirlwind | SpellMaskWhirlwindOh | SpellMaskShieldSlam |
		SpellMaskBloodthirst | SpellMaskMortalStrike | SpellMaskIntercept | SpellMaskDevastate | SpellMaskRetaliationHit

	SpellMaskDamageSpells = SpellMaskDirectDamageSpells | SpellMaskDeepWounds | SpellMaskRend
)

const EnrageTag = "EnrageEffect"

type Warrior struct {
	core.Character

	ClassSpellScaling float64

	Talents *proto.WarriorTalents

	WarriorInputs

	// Current state
	Stance                Stance
	ChargeRageGain        float64
	BerserkerRageRageGain float64

	BattleShout       *core.Spell
	CommandingShout   *core.Spell
	DemoralizingShout *core.Spell
	BattleStance      *core.Spell
	DefensiveStance   *core.Spell
	BerserkerStance   *core.Spell

	Rend         *core.Spell
	DeepWounds   *core.Spell
	MortalStrike *core.Spell

	HeroicStrike       *core.Spell
	Cleave             *core.Spell
	curQueueAura       *core.Aura
	curQueuedAutoSpell *core.Spell

	sharedMCD        *core.Timer // Recklessness, Shield Wall & Retaliation
	sharedShoutsCD   *core.Timer
	queuedRealismICD *core.Cooldown

	EnrageAura *core.Aura

	SkullBannerAura         *core.Aura
	DemoralizingBannerAuras core.AuraArray

	DemoralizingShoutAuras core.AuraArray
	SunderArmorAuras       core.AuraArray

	// Set bonuses
	T6Tank2P *core.Aura
}

func (warrior *Warrior) GetCharacter() *core.Character {
	return &warrior.Character
}

func (warrior *Warrior) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
}

func (warrior *Warrior) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (warrior *Warrior) Initialize() {

	warrior.registerRecklessness()
	warrior.registerShieldWall()
	warrior.registerRetaliation()

	warrior.registerBerserkerRage()
	warrior.registerBloodrage()
	warrior.registerCharge()
	warrior.registerIntercept()
	warrior.registerPummel()
	warrior.registerHamstring()

	warrior.registerRend()
	warrior.registerSunderArmor()
	warrior.registerHeroicStrike()
	warrior.registerCleave()
	warrior.registerOverpower()
	warrior.registerSlam()
	warrior.registerWhirlwind()
	warrior.registerExecute()
	warrior.registerThunderClap()
	warrior.registerRevenge()
	warrior.registerShieldBlock()
	warrior.registerShieldBash()

	warrior.registerStances()
	warrior.registerShouts()
}

func (warrior *Warrior) Reset(_ *core.Simulation) {
	warrior.curQueueAura = nil
	warrior.curQueuedAutoSpell = nil

	warrior.Stance = StanceNone
	warrior.ChargeRageGain = 15
	warrior.BerserkerRageRageGain = 0
}

func (warrior *Warrior) OnEncounterStart(sim *core.Simulation) {}

func (war *Warrior) GetHandType() proto.HandType {
	mh := war.GetMHWeapon()

	if mh != nil && (mh.HandType == proto.HandType_HandTypeTwoHand) {
		return proto.HandType_HandTypeTwoHand
	}

	return proto.HandType_HandTypeOneHand
}

func NewWarrior(character *core.Character, options *proto.WarriorOptions, talents string, inputs WarriorInputs) *Warrior {
	warrior := &Warrior{
		Character:     *character,
		Talents:       &proto.WarriorTalents{},
		WarriorInputs: inputs,
	}
	core.FillTalentsProto(warrior.Talents.ProtoReflect(), talents, TalentTreeSizes)

	warrior.EnableRageBar(core.RageBarOptions{
		MaxRage:            100,
		BaseRageMultiplier: 1,
		StartingRage:       inputs.StartingRage,
	})

	warrior.EnableAutoAttacks(warrior, core.AutoAttackOptions{
		MainHand:       warrior.WeaponFromMainHand(warrior.DefaultMeleeCritMultiplier()),
		OffHand:        warrior.WeaponFromOffHand(warrior.DefaultMeleeCritMultiplier()),
		AutoSwingMelee: true,
		ReplaceMHSwing: warrior.TryHSOrCleave,
	})

	warrior.PseudoStats.CanParry = true

	warrior.AddStatDependency(stats.Strength, stats.AttackPower, 2)
	warrior.AddStatDependency(stats.Strength, stats.BlockValue, 1/20.0)
	warrior.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])
	warrior.AddStatDependency(stats.Agility, stats.DodgeRating, 1/30.0*core.DodgeRatingPerDodgePercent)
	warrior.AddStatDependency(stats.BonusArmor, stats.Armor, 1)

	warrior.sharedShoutsCD = warrior.NewTimer()
	warrior.sharedMCD = warrior.NewTimer()
	warrior.ChargeRageGain = 15
	warrior.BerserkerRageRageGain = 0
	// The sim often re-enables heroic strike in an unrealistic amount of time.
	// This can cause an unrealistic immediate double-hit around wild strikes procs
	warrior.queuedRealismICD = &core.Cooldown{
		Timer:    warrior.NewTimer(),
		Duration: time.Millisecond * time.Duration(warrior.WarriorInputs.QueueDelay),
	}

	return warrior
}

// Agent is a generic way to access underlying warrior on any of the agents.
type WarriorAgent interface {
	GetWarrior() *Warrior
}
