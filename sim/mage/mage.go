package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

var TalentTreeSizes = [3]int{23, 22, 22}

type Mage struct {
	core.Character

	waterElemental *WaterElemental

	ClassSpellScaling float64

	Talents *proto.MageTalents
	Options *proto.MageOptions
	// ArcaneOptions *proto.ArcaneMage_Options
	// FireOptions   *proto.FireMage_Options
	// FrostOptions  *proto.FrostMage_Options

	ArcaneBlast        *core.Spell
	ArcaneChargesAura  *core.Aura
	ClearCasting       *core.Aura
	PresenceOfMindAura *core.Aura
	ArcanePowerAura    *core.Aura

	ImprovedScorchAuras core.AuraArray
	SlowAuras           core.AuraArray

	Ignite               *core.Spell
	FireBlast            *core.Spell
	FlameOrbExplode      *core.Spell
	Flamestrike          *core.Spell
	FlamestrikeBW        *core.Spell
	FrostfireOrb         *core.Spell
	Pyroblast            *core.Spell
	SummonWaterElemental *core.Spell
	IcyVeins             *core.Spell

	IcyVeinsAura *core.Aura

	//T15_4PC_FrostboltProcChance float64
	//T15_4PC_ArcaneChargeEffect  float64
	//Icicles                     []float64

	// Item sets
	//T16_4pc *core.Aura
}

func (mage *Mage) GetCharacter() *core.Character {
	return &mage.Character
}

func (mage *Mage) GetMage() *Mage {
	return mage
}

func RegisterMage() {
	core.RegisterAgentFactory(
		proto.Player_Mage{},
		proto.Spec_SpecMage,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewMage(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Mage)
			if !ok {
				panic("Invalid spec value for Survival Hunter!")
			}
			player.Spec = playerSpec
		},
	)
}

func (mage *Mage) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	raidBuffs.ArcaneBrilliance = true
}

func (mage *Mage) AddPartyBuffs(partyBuffs *proto.PartyBuffs) {
}

func (mage *Mage) Initialize() {

	mage.ImprovedScorchAuras = mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ImprovedScorchAura(target, 0)
	})

	mage.SlowAuras = mage.NewEnemyAuraArray(core.SlowAura)

	mage.registerPassives()
	mage.registerSpells()
}

func (mage *Mage) registerPassives() {
	mage.registerArcaneCharges()
}

func (mage *Mage) registerSpells() {
	mage.registerArcaneBlastSpell()
	mage.registerArcaneExplosionSpell()
	mage.registerArcaneMissilesSpell()
	mage.registerArmorSpells()
	mage.registerBlizzardSpell()
	mage.registerConeOfColdSpell()
	mage.registerFrostboltSpell()
	mage.registerEvocation()
	mage.registerFireballSpell()
	mage.registerFireBlastSpell()
	mage.registerFlamestrikeSpell()
	mage.registerFrostNovaSpell()
	mage.registerManaGems()
	mage.registerScorchSpell()

	//TalentSpells
	mage.registerPresenceOfMindSpell()
	mage.registerArcanePowerSpell()
	mage.registerSlowSpell()

	mage.registerBlastWaveSpell()
	mage.registerPyroblastSpell()
	mage.registerCombustionSpell()
	mage.registerDragonsBreathSpell()

	mage.registerColdSnapSpell()
	mage.registerSummonWaterElementalSpell()

	//Hotfixes will go here
	mage.registerHotfixes()
}

func (mage *Mage) Reset(sim *core.Simulation) {
}

func (mage *Mage) OnEncounterStart(sim *core.Simulation) {
}

func NewMage(character *core.Character, options *proto.Player) *Mage {
	mageOptions := options.GetMage().Options.ClassOptions
	mage := &Mage{
		Character: *character,
		Talents:   &proto.MageTalents{},
		Options:   mageOptions,
	}

	core.FillTalentsProto(mage.Talents.ProtoReflect(), options.TalentsString, TalentTreeSizes)

	mage.EnableManaBar()

	return mage
}

// Agent is a generic way to access underlying mage on any of the agents.
type MageAgent interface {
	GetMage() *Mage
}

const (
	FireSpellMaxTimeUntilResult       = 750 * time.Millisecond
	MageSpellFlagNone           int64 = 0
	MageSpellArcaneBlast        int64 = 1 << iota
	MageSpellArcaneExplosion
	MageSpellArcanePower
	MageSpellArcaneMissilesCast
	MageSpellArcaneMissilesTick
	MageSpellBlastWave
	MageSpellBlizzard
	MageSpellColdSnap
	MageSpellConeOfCold
	MageSpellDragonsBreath
	MageSpellEvocation
	MageSpellFireBlast
	MageSpellFireball
	MageSpellFlamestrike
	MageSpellFlamestrikeDot
	MageSpellFrostArmor
	MageSpellFrostbolt
	MageSpellFrostNova
	MageSpellIceBarrier
	MageSpellIceBlock
	MageSpellIceLance
	MageSpellIcyVeins
	MageSpellIgnite
	MageSpellMageArmor
	MageSpellManaGems
	MageSpellMoltenArmor
	MageSpellPresenceOfMind
	MageSpellPyroblast
	MageSpellPyroblastDot
	MageSpellScorch
	MageSpellSlow
	MageSpellCombustion
	MageWaterElementalSpellWaterBolt
	MageSpellLast
	MageSpellsAll  = MageSpellLast<<1 - 1
	MageSpellFrost = MageSpellFrostbolt | MageSpellBlizzard | MageSpellFrostNova | MageSpellConeOfCold | MageSpellIceLance
	MageSpellFire  = MageSpellDragonsBreath | MageSpellFireball | MageSpellCombustion |
		MageSpellFireBlast | MageSpellFlamestrike | MageSpellIgnite | MageSpellPyroblast | MageSpellScorch
	MageSpellsAllDamaging = MageSpellArcaneBlast | MageSpellArcaneExplosion | MageSpellArcaneMissilesTick | MageSpellBlizzard |
		MageSpellDragonsBreath | MageSpellFireBlast | MageSpellFireball | MageSpellFlamestrike | MageSpellFrostbolt |
		MageSpellIceLance | MageSpellPyroblast | MageSpellPyroblastDot | MageSpellScorch
	MageSpellInstantCast = MageSpellArcaneMissilesCast | MageSpellArcaneMissilesTick | MageSpellFireBlast | MageSpellArcaneExplosion | MageSpellPyroblastDot |
		MageSpellCombustion | MageSpellConeOfCold | MageSpellDragonsBreath | MageSpellIceLance | MageSpellManaGems | MageSpellPresenceOfMind
	MageSpellExtraResult = MageSpellArcaneMissilesTick | MageSpellBlizzard
	FireSpellIgnitable   = MageSpellFireball | MageSpellScorch | MageSpellPyroblast
)
