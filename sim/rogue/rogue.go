package rogue

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

const (
	SpellFlagBuilder  = core.SpellFlagAgentReserved2
	SpellFlagFinisher = core.SpellFlagAgentReserved3
	SpellFlagSealFate = core.SpellFlagAgentReserved4
)

var TalentTreeSizes = [3]int{21, 24, 22}

const RogueBleedTag = "RogueBleed"

type Rogue struct {
	core.Character

	ClassSpellScaling float64

	Talents              *proto.RogueTalents
	Options              *proto.RogueOptions
	AssassinationOptions *proto.Rogue_Options

	SliceAndDiceBonusFlat    float64 // The flat bonus Attack Speed bonus before Mastery is applied
	AdditiveEnergyRegenBonus float64
	ExposeArmorModifier      float64

	sliceAndDiceDurations [6]time.Duration

	Backstab       *core.Spell
	BladeFlurry    *core.Spell
	DeadlyPoison   *core.Spell
	Feint          *core.Spell
	Garrote        *core.Spell
	Ambush         *core.Spell
	Hemorrhage     *core.Spell
	GhostlyStrike  *core.Spell
	WoundPoison    *core.Spell
	Mutilate       *core.Spell
	MutilateMH     *core.Spell
	MutilateOH     *core.Spell
	Shiv           *core.Spell
	SinisterStrike *core.Spell
	Shadowstep     *core.Spell
	Preparation    *core.Spell
	Premeditation  *core.Spell
	ColdBlood      *core.Spell
	Vanish         *core.Spell
	AdrenalineRush *core.Spell
	Gouge          *core.Spell

	Envenom      *core.Spell
	Eviscerate   *core.Spell
	ExposeArmor  *core.Spell
	Rupture      *core.Spell
	SliceAndDice *core.Spell

	deadlyPoisonPPHM  *core.DynamicProcManager
	woundPoisonPPHM   *core.DynamicProcManager
	instantPoisonPPHM *core.DynamicProcManager

	AdrenalineRushAura   *core.Aura
	BladeFlurryAura      *core.Aura
	ExposeArmorAuras     core.AuraArray
	SliceAndDiceAura     *core.Aura
	MasterOfSubtletyAura *core.Aura
	ShadowstepAura       *core.Aura
	StealthAura          *core.Aura

	WoundPoisonDebuffAuras core.AuraArray

	ruthlessnessMetrics      *core.ResourceMetrics
	relentlessStrikesMetrics *core.ResourceMetrics

	HasPvpEnergy              bool
	DeathmantleBonus          float64
	SliceAndDiceBonusDuration float64
}

// ApplyTalents implements core.Agent.
func (rogue *Rogue) ApplyTalents() {
	rogue.registerAssassinationTalents()
	rogue.registerCombatTalents()
	rogue.registerSubtletyTalents()
}

func (rogue *Rogue) GetCharacter() *core.Character {
	return &rogue.Character
}

func (rogue *Rogue) GetRogue() *Rogue {
	return rogue
}

func (rogue *Rogue) AddRaidBuffs(_ *proto.RaidBuffs)   {}
func (rogue *Rogue) AddPartyBuffs(_ *proto.PartyBuffs) {}

// Apply the effect of successfully casting a finisher to combo points
func (rogue *Rogue) ApplyFinisher(sim *core.Simulation, spell *core.Spell) {
	numPoints := rogue.ComboPoints()
	rogue.SpendComboPoints(sim, spell.ComboPointMetrics())

	// Relentless Strikes
	if rogue.Talents.RelentlessStrikes && sim.Proc(0.2*float64(numPoints), "Relentless Strikes") {
		rogue.AddEnergy(sim, 25, rogue.relentlessStrikesMetrics)
	}

	// Ruthlessness
	if rogue.Talents.Ruthlessness > 0 && sim.Proc(0.2*float64(rogue.Talents.Ruthlessness), "Ruthlessness") {
		rogue.AddComboPoints(sim, 1, rogue.ruthlessnessMetrics)
	}
}

func (rogue *Rogue) GetBaseDamageFromCoefficient(c float64) float64 {
	return c * rogue.ClassSpellScaling
}

func (rogue *Rogue) Initialize() {
	// Update auto crit multipliers now that we have the targets.
	rogue.AutoAttacks.MHConfig().CritMultiplier = rogue.DefaultMeleeCritMultiplier()
	rogue.AutoAttacks.OHConfig().CritMultiplier = rogue.DefaultMeleeCritMultiplier()
	rogue.AutoAttacks.RangedConfig().CritMultiplier = rogue.DefaultMeleeCritMultiplier()

	rogue.registerAmbushSpell()
	rogue.registerBackstabSpell()
	rogue.registerEnvenom()
	rogue.registerEviscerate()
	rogue.registerExposeArmorSpell()
	rogue.registerGarrote()
	rogue.registerDeadlyPoisonSpell()
	rogue.registerInstantPoisonSpell()
	rogue.registerWoundPoisonSpell()
	rogue.registerRupture()
	rogue.registerSinisterStrikeSpell()
	rogue.registerSliceAndDice()
	rogue.registerVanishSpell()

	rogue.ruthlessnessMetrics = rogue.NewComboPointMetrics(core.ActionID{SpellID: 14161})
	rogue.relentlessStrikesMetrics = rogue.NewEnergyMetrics(core.ActionID{SpellID: 14179})
}

func (rogue *Rogue) ApplyAdditiveEnergyRegenBonus(sim *core.Simulation, increment float64) {
	oldBonus := rogue.AdditiveEnergyRegenBonus
	newBonus := oldBonus + increment
	rogue.AdditiveEnergyRegenBonus = newBonus
	rogue.MultiplyEnergyRegenSpeed(sim, (1.0+newBonus)/(1.0+oldBonus))
}

func (rogue *Rogue) Reset(sim *core.Simulation) {
	for _, mcd := range rogue.GetMajorCooldowns() {
		mcd.Disable()
	}

	rogue.MultiplyEnergyRegenSpeed(sim, 1.0+rogue.AdditiveEnergyRegenBonus)
}

func (rogue *Rogue) OnEncounterStart(sim *core.Simulation) {
}

func NewRogue(character *core.Character, options *proto.Player, talents string) *Rogue {
	rogueOptions := options.GetRogue()
	rogue := &Rogue{
		Character: *character,
		Talents:   &proto.RogueTalents{},
		Options:   rogueOptions.Options.ClassOptions,
	}

	core.FillTalentsProto(rogue.Talents.ProtoReflect(), talents, TalentTreeSizes)

	// Passive rogue threat reduction: https://wotlk.wowhead.com/spell=21184/rogue-passive-dnd
	rogue.PseudoStats.ThreatMultiplier *= 0.71
	rogue.PseudoStats.CanParry = true

	maxEnergy := 100.0

	if rogue.Talents.Vigor {
		maxEnergy += 10
	}
	if rogue.HasPvpEnergy {
		maxEnergy += 10
	}

	rogue.EnableEnergyBar(core.EnergyBarOptions{
		MaxComboPoints: 5,
		MaxEnergy:      maxEnergy,
		UnitClass:      proto.Class_ClassRogue,
	})

	rogue.EnableAutoAttacks(rogue, core.AutoAttackOptions{
		MainHand:       rogue.WeaponFromMainHand(0), // Set crit multiplier later when we have targets.
		OffHand:        rogue.WeaponFromOffHand(0),  // Set crit multiplier later when we have targets.
		AutoSwingMelee: true,
	})

	//rogue.applyPoisons()

	rogue.AddStatDependency(stats.Strength, stats.AttackPower, 1)
	rogue.AddStatDependency(stats.Agility, stats.AttackPower, 1)
	rogue.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])
	rogue.AddStatDependency(stats.Agility, stats.DodgeRating, 1/20*core.DodgeRatingPerDodgePercent)

	return rogue
}

// Deactivate Stealth if it is active. This must be added to all abilities that cause Stealth to fade.
func (rogue *Rogue) BreakStealth(sim *core.Simulation) {
	if rogue.StealthAura.IsActive() {
		rogue.StealthAura.Deactivate(sim)
		rogue.AutoAttacks.EnableAutoSwing(sim)
	}
}

// Does the rogue have a dagger equipped in the specified hand (main or offhand)?
func (rogue *Rogue) HasDagger(hand core.Hand) bool {
	if hand == core.MainHand && rogue.MainHand() != nil {
		return rogue.MainHand().WeaponType == proto.WeaponType_WeaponTypeDagger
	}

	if rogue.OffHand() != nil {
		return rogue.OffHand().WeaponType == proto.WeaponType_WeaponTypeDagger
	}

	return false
}

// Does the rogue have a thrown weapon equipped in the ranged slot?
func (rogue *Rogue) HasThrown() bool {
	weapon := rogue.Ranged()
	return weapon != nil && weapon.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeThrown
}

// Check if the rogue is considered in "stealth" for the purpose of casting abilities
func (rogue *Rogue) IsStealthed() bool {
	return rogue.StealthAura.IsActive()
}

func RegisterRogue() {
	core.RegisterAgentFactory(
		proto.Player_Rogue{},
		proto.Spec_SpecRogue,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewRogue(character, options, options.TalentsString)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Rogue)
			if !ok {
				panic("Invalid spec value for Combat Rogue!")
			}
			player.Spec = playerSpec
		},
	)
}

// Agent is a generic way to access underlying rogue on any of the agents.
type RogueAgent interface {
	GetRogue() *Rogue
}

const (
	RogueSpellFlagNone int64 = 0
	RogueSpellAmbush   int64 = 1 << iota
	RogueSpellBackstab
	RogueSpellEnvenom
	RogueSpellEviscerate
	RogueSpellExposeArmor
	RogueSpellFeint
	RogueSpellGarrote
	RogueSpellGouge
	RogueSpellRupture
	RogueSpellShiv
	RogueSpellSinisterStrike
	RogueSpellSliceAndDice
	RogueSpellStealth
	RogueSpellVanish
	RogueSpellHemorrhage
	RogueSpellPremeditation
	RogueSpellPreparation
	RogueSpellShadowstep
	RogueSpellAdrenalineRush
	RogueSpellBladeFlurry
	RogueSpellColdBlood
	RogueSpellMutilate
	RogueSpellMutilateHit
	RogueSpellInstantPoison
	RogueSpellWoundPoison
	RogueSpellDeadlyPoison
	RogueSpellGhostlyStrike

	RogueSpellLast
	RogueSpellsAll = RogueSpellLast<<1 - 1

	RogueSpellPoisons        = RogueSpellWoundPoison | RogueSpellDeadlyPoison | RogueSpellInstantPoison
	RogueSpellLethality      = RogueSpellSinisterStrike | RogueSpellGouge | RogueSpellBackstab | RogueSpellGhostlyStrike | RogueSpellMutilateHit | RogueSpellShiv | RogueSpellHemorrhage
	RogueSpellDirectFinisher = RogueSpellEnvenom | RogueSpellEviscerate
	RogueSpellFinisher       = RogueSpellDirectFinisher | RogueSpellSliceAndDice | RogueSpellRupture | RogueSpellExposeArmor
	RogueSpellCanCrit        = RogueSpellLethality | RogueSpellDirectFinisher
)
