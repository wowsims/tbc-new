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
	SpellMaskShieldWall
	SpellMaskLastStand
	SpellMaskCharge
	SpellMaskDemoralizingShout

	// Special attacks
	SpellMaskSweepingStrikes
	SpellMaskSweepingStrikesHit
	SpellMaskSweepingStrikesNormalizedHit
	SpellMaskCleave
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
	SpellMaskBloodthirst
	SpellMaskMortalStrike
	SpellMaskWildStrike
	SpellMaskShieldBlock
	SpellMaskHamstring
	SpellMaskPummel

	// Talents
	SpellMaskImpendingVictory
	SpellMaskBladestorm
	SpellMaskBladestormMH
	SpellMaskBladestormOH

	SpellMaskShouts = SpellMaskCommandingShout | SpellMaskBattleShout
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

	HeroicStrikeCleaveCostMod *core.SpellMod

	BattleShout     *core.Spell
	CommandingShout *core.Spell
	BattleStance    *core.Spell
	DefensiveStance *core.Spell
	BerserkerStance *core.Spell

	MortalStrike                    *core.Spell
	DeepWounds                      *core.Spell
	ShieldSlam                      *core.Spell
	SweepingStrikesNormalizedAttack *core.Spell

	sharedShoutsCD   *core.Timer
	sharedHSCleaveCD *core.Timer

	BattleStanceAura    *core.Aura
	DefensiveStanceAura *core.Aura
	BerserkerStanceAura *core.Aura

	InciteAura          *core.Aura
	UltimatumAura       *core.Aura
	SweepingStrikesAura *core.Aura
	EnrageAura          *core.Aura
	BerserkerRageAura   *core.Aura
	ShieldBlockAura     *core.Aura
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
}

func (warrior *Warrior) GetCharacter() *core.Character {
	return &warrior.Character
}

func (warrior *Warrior) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {

}

func (warrior *Warrior) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (warrior *Warrior) Initialize() {
	warrior.sharedHSCleaveCD = warrior.NewTimer()
	warrior.sharedShoutsCD = warrior.NewTimer()

	warrior.WeakenedArmorAuras = warrior.NewEnemyAuraArray(core.WeakenedArmorAura)

	// warrior.registerStances()
	// warrior.registerShouts()
	warrior.registerPassives()

	// warrior.registerBerserkerRage()
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
	// warrior.Stance = StanceNone
}

func (warrior *Warrior) OnEncounterStart(sim *core.Simulation) {
	warrior.PrePullChargeGain = 0
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
	})

	warrior.PseudoStats.CanParry = true

	warrior.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])
	warrior.AddStatDependency(stats.Strength, stats.AttackPower, 2)

	// Base strength to Parry is not affected by Diminishing Returns
	baseStrength := warrior.GetBaseStats()[stats.Strength]
	warrior.PseudoStats.BaseParryChance += baseStrength * core.StrengthToParryPercent
	warrior.AddStat(stats.ParryRating, -baseStrength*core.StrengthToParryRating)
	warrior.AddStatDependency(stats.Strength, stats.ParryRating, core.StrengthToParryRating)
	warrior.AddStatDependency(stats.Agility, stats.DodgeRating, 0.1/10000.0/100.0)
	warrior.AddStatDependency(stats.BonusArmor, stats.Armor, 1)
	// warrior.MultiplyStat(stats.HasteRating, 1.5)

	// Base dodge unaffected by Diminishing Returns
	warrior.PseudoStats.BaseDodgeChance += 0.03
	warrior.PseudoStats.BaseParryChance += 0.03
	warrior.PseudoStats.BaseBlockChance += 0.03
	warrior.CriticalBlockChance = append(warrior.CriticalBlockChance, 0.0, 0.0)

	warrior.HeroicStrikeCleaveCostMod = warrior.AddDynamicMod(core.SpellModConfig{
		ClassMask:  SpellMaskHeroicStrike | SpellMaskCleave,
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -2,
	})

	return warrior
}

func (warrior *Warrior) GetCriticalBlockChance() float64 {
	return warrior.CriticalBlockChance[0] + warrior.CriticalBlockChance[1]
}

func (warrior *Warrior) CastNormalizedSweepingStrikesAttack(results core.SpellResultSlice, sim *core.Simulation) {
	if warrior.SweepingStrikesAura != nil && warrior.SweepingStrikesAura.IsActive() {
		for _, result := range results {
			if result.Landed() {
				warrior.SweepingStrikesNormalizedAttack.Cast(sim, warrior.Env.NextActiveTargetUnit(result.Target))
				break
			}
		}
	}
}

// Agent is a generic way to access underlying warrior on any of the agents.
type WarriorAgent interface {
	GetWarrior() *Warrior
}
