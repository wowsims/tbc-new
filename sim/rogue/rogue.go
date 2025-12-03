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

	MasteryBaseValue  float64
	MasteryMultiplier float64

	SliceAndDiceBonusFlat    float64 // The flat bonus Attack Speed bonus before Mastery is applied
	AdditiveEnergyRegenBonus float64

	sliceAndDiceDurations [6]time.Duration

	Backstab         *core.Spell
	BladeFlurry      *core.Spell
	DeadlyPoison     *core.Spell
	FanOfKnives      *core.Spell
	Feint            *core.Spell
	Garrote          *core.Spell
	Ambush           *core.Spell
	Hemorrhage       *core.Spell
	GhostlyStrike    *core.Spell
	HungerForBlood   *core.Spell
	WoundPoison      *core.Spell
	Mutilate         *core.Spell
	Dispatch         *core.Spell
	MutilateMH       *core.Spell
	MutilateOH       *core.Spell
	Shiv             *core.Spell
	SinisterStrike   *core.Spell
	TricksOfTheTrade *core.Spell
	Shadowstep       *core.Spell
	Preparation      *core.Spell
	Premeditation    *core.Spell
	ShadowDance      *core.Spell
	ColdBlood        *core.Spell
	Vanish           *core.Spell
	VenomousWounds   *core.Spell
	Vendetta         *core.Spell
	RevealingStrike  *core.Spell
	KillingSpree     *core.Spell
	AdrenalineRush   *core.Spell
	Gouge            *core.Spell
	ShadowBlades     *core.Spell

	Envenom           *core.Spell
	Eviscerate        *core.Spell
	ExposeArmor       *core.Spell
	Rupture           *core.Spell
	SliceAndDice      *core.Spell
	CrimsonTempest    *core.Spell
	CrimsonTempestDoT *core.Spell

	deadlyPoisonPPHM *core.DynamicProcManager
	woundPoisonPPHM  *core.DynamicProcManager

	AdrenalineRushAura   *core.Aura
	BladeFlurryAura      *core.Aura
	EnvenomAura          *core.Aura
	ExposeArmorAuras     core.AuraArray
	HungerForBloodAura   *core.Aura
	KillingSpreeAura     *core.Aura
	SliceAndDiceAura     *core.Aura
	MasterOfSubtletyAura *core.Aura
	ShadowstepAura       *core.Aura
	ShadowDanceAura      *core.Aura
	DirtyDeedsAura       *core.Aura
	HonorAmongThieves    *core.Aura
	StealthAura          *core.Aura
	SubterfugeAura       *core.Aura
	BanditsGuileAura     *core.Aura
	AnticipationAura     *core.Aura
	ShadowBladesAura     *core.Aura

	NightstalkerMod *core.SpellMod
	ShadowFocusMod  *core.SpellMod

	MasterPoisonerDebuffAuras core.AuraArray
	SavageCombatDebuffAuras   core.AuraArray
	WoundPoisonDebuffAuras    core.AuraArray

	Has2PT15      bool
	T16EnergyAura *core.Aura
	T16SpecMod    *core.SpellMod

	ruthlessnessMetrics      *core.ResourceMetrics
	relentlessStrikesMetrics *core.ResourceMetrics
}

// ApplyTalents implements core.Agent.
func (rogue *Rogue) ApplyTalents() {
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
	rogue.AutoAttacks.MHConfig().CritMultiplier = rogue.CritMultiplier(false)
	rogue.AutoAttacks.OHConfig().CritMultiplier = rogue.CritMultiplier(false)
	rogue.AutoAttacks.RangedConfig().CritMultiplier = rogue.CritMultiplier(false)

	// rogue.registerStealthAura()
	// rogue.registerVanishSpell()
	// rogue.registerAmbushSpell()
	// rogue.registerGarrote()
	// rogue.registerRupture()
	// rogue.registerSliceAndDice()
	// rogue.registerEviscerate()
	// rogue.registerExposeArmorSpell()
	// rogue.registerFanOfKnives()
	// rogue.registerTricksOfTheTradeSpell()
	// rogue.registerDeadlyPoisonSpell()
	// rogue.registerWoundPoisonSpell()
	// rogue.registerPoisonAuras()
	// rogue.registerShadowBladesCD()
	// rogue.registerCrimsonTempest()
	// rogue.registerPreparationCD()

	rogue.ruthlessnessMetrics = rogue.NewComboPointMetrics(core.ActionID{SpellID: 14161})
	rogue.relentlessStrikesMetrics = rogue.NewEnergyMetrics(core.ActionID{SpellID: 58423})
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

func (rogue *Rogue) CritMultiplier(applyLethality bool) float64 {
	secondaryModifier := 0.0
	return rogue.GetCharacter().CritMultiplier(1.0, secondaryModifier)
}

func NewRogue(character *core.Character, options *proto.Player, talents string) *Rogue {
	rogueOptions := options.GetRogue()
	rogue := &Rogue{
		Character:         *character,
		Talents:           &proto.RogueTalents{},
		Options:           rogueOptions.Options.ClassOptions,
		ClassSpellScaling: core.GetClassSpellScalingCoefficient(proto.Class_ClassRogue),
	}

	core.FillTalentsProto(rogue.Talents.ProtoReflect(), talents, TalentTreeSizes)

	// Passive rogue threat reduction: https://wotlk.wowhead.com/spell=21184/rogue-passive-dnd
	rogue.PseudoStats.ThreatMultiplier *= 0.71
	rogue.PseudoStats.CanParry = true

	maxEnergy := 100.0

	if rogue.Talents.Vigor {
		maxEnergy += 10
	}

	rogue.EnableEnergyBar(core.EnergyBarOptions{
		MaxComboPoints:        5,
		MaxEnergy:             maxEnergy,
		UnitClass:             proto.Class_ClassRogue,
		HasHasteRatingScaling: true,
	})

	rogue.EnableAutoAttacks(rogue, core.AutoAttackOptions{
		MainHand:       rogue.WeaponFromMainHand(0), // Set crit multiplier later when we have targets.
		OffHand:        rogue.WeaponFromOffHand(0),  // Set crit multiplier later when we have targets.
		AutoSwingMelee: true,
	})

	//rogue.applyPoisons()

	rogue.AddStatDependency(stats.Strength, stats.AttackPower, 1)
	rogue.AddStatDependency(stats.Agility, stats.AttackPower, 2)
	rogue.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])

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
	RogueSpellFanOfKnives
	RogueSpellFeint
	RogueSpellGarrote
	RogueSpellGouge
	RogueSpellRecuperate
	RogueSpellRupture
	RogueSpellCrimsonTempest
	RogueSpellCrimsonTempestDoT
	RogueSpellShiv
	RogueSpellSinisterStrike
	RogueSpellSliceAndDice
	RogueSpellStealth
	RogueSpellTricksOfTheTrade
	RogueSpellTricksOfTheTradeThreat
	RogueSpellVanish
	RogueSpellHemorrhage
	RogueSpellPremeditation
	RogueSpellPreparation
	RogueSpellShadowDance
	RogueSpellShadowstep
	RogueSpellAdrenalineRush
	RogueSpellBladeFlurry
	RogueSpellKillingSpree
	RogueSpellKillingSpreeHit
	RogueSpellMainGauche
	RogueSpellRevealingStrike
	RogueSpellColdBlood
	RogueSpellMutilate
	RogueSpellMutilateHit
	RogueSpellDispatch
	RogueSpellVendetta
	RogueSpellVenomousWounds
	RogueSpellWoundPoison
	RogueSpellDeadlyPoison
	RogueSpellShadowBlades
	RogueSpellShadowBladesHit
	RogueSpellMarkedForDeath

	RogueSpellLast
	RogueSpellsAll = RogueSpellLast<<1 - 1

	RogueSpellPoisons          = RogueSpellVenomousWounds | RogueSpellWoundPoison | RogueSpellDeadlyPoison
	RogueSpellGenerator        = RogueSpellBackstab | RogueSpellHemorrhage | RogueSpellSinisterStrike | RogueSpellRevealingStrike | RogueSpellMutilate | RogueSpellDispatch | RogueSpellAmbush | RogueSpellGarrote | RogueSpellFanOfKnives
	RogueSpellDamagingFinisher = RogueSpellEnvenom | RogueSpellEviscerate | RogueSpellRupture | RogueSpellCrimsonTempest
	RogueSpellWeightedBlades   = RogueSpellSinisterStrike | RogueSpellRevealingStrike
	RogueSpellActives          = RogueSpellGenerator | RogueSpellDamagingFinisher | RogueSpellSliceAndDice
)
