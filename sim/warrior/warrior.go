package warrior

import (
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

var TalentTreeSizes = [3]int{23, 21, 22}

type WarriorInputs struct {
	StanceSnapshot bool
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
		SpellMaskBloodthirst | SpellMaskMortalStrike | SpellMaskIntercept | SpellMaskDevastate

	SpellMaskDamageSpells = SpellMaskDirectDamageSpells | SpellMaskDeepWounds | SpellMaskRend
)

const EnrageTag = "EnrageEffect"

type Warrior struct {
	core.Character

	ClassSpellScaling float64

	Talents *proto.WarriorTalents

	WarriorInputs

	// Current state
	Stance              Stance
	CriticalBlockChance []float64 // Can be gained as non-prot via certain talents and spells
	PrePullChargeGain   float64
	ChargeRageGain      float64

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

	sharedShoutsCD *core.Timer

	RendAura            *core.Aura
	DeepWoundsAura      *core.Aura
	SweepingStrikesAura *core.Aura
	EnrageAura          *core.Aura
	BerserkerRageAura   *core.Aura
	LastStandAura       *core.Aura
	VictoryRushAura     *core.Aura
	ShieldBarrierAura   *core.DamageAbsorptionAura

	SkullBannerAura         *core.Aura
	DemoralizingBannerAuras core.AuraArray

	RallyingCryAuras       core.AuraArray
	DemoralizingShoutAuras core.AuraArray
	SunderArmorAuras       core.AuraArray
	ThunderClapAuras       core.AuraArray
	WeakenedArmorAuras     core.AuraArray

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

	warrior.registerCharge()
	warrior.registerIntercept()
	warrior.registerPummel()
	warrior.registerHamstring()

	warrior.registerSunderArmor()
	warrior.registerHeroicStrike()
	warrior.registerCleave()
	warrior.registerExecute()
	warrior.registerOverpower()
	warrior.registerShieldBlock()
	warrior.registerShieldBash()

	warrior.registerStances()
	warrior.registerShouts()
	warrior.registerPassives()

	// warrior.registerRallyingCry()
	// warrior.registerExecuteSpell()
	// warrior.registerHeroicStrikeSpell()
	// warrior.registerCleaveSpell()
	// warrior.registerRecklessness()
	// warrior.registerVictoryRush()
	// warrior.registerShieldWall()
	// warrior.registerSunderArmor()
	// warrior.registerHamstring()
	// warrior.registerThunderClap()
	// warrior.registerWhirlwind()
	// warrior.registerCharge()
	// warrior.registerPummel()
}

func (warrior *Warrior) registerPassives() {
	// warrior.registerEnrage()
	// warrior.registerDeepWounds()
	// warrior.registerBloodAndThunder()
}

func (warrior *Warrior) Reset(_ *core.Simulation) {
	warrior.Stance = StanceNone
	warrior.ChargeRageGain = 15.0
}

func (warrior *Warrior) OnEncounterStart(sim *core.Simulation) {
	warrior.PrePullChargeGain = 0
}

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
	warrior.ChargeRageGain = 15.0

	return warrior
}

// Agent is a generic way to access underlying warrior on any of the agents.
type WarriorAgent interface {
	GetWarrior() *Warrior
}
