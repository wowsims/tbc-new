package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

var TalentTreeSizes = [3]int{20, 21, 20}

// Start looking to refresh 5 minute totems at 4:55.
const TotemRefreshTime5M = time.Second * 295

// Damage Done By Caster setup
const (
	DDBC_FrostbrandWeapon int = iota
	DDBC_UnleashedFury
	DDBC_2PT16

	DDBC_Total
)

const (
	SpellFlagShamanSpell = core.SpellFlagAgentReserved1
	SpellFlagShock       = core.SpellFlagAgentReserved2
	SpellFlagInstant     = core.SpellFlagAgentReserved3
	SpellFlagFocusable   = core.SpellFlagAgentReserved4
)

func NewShaman(character *core.Character, talents string, selfBuffs SelfBuffs) *Shaman {
	shaman := &Shaman{
		Character: *character,
		Talents:   &proto.ShamanTalents{},
		Totems:    &proto.ShamanTotems{},
		SelfBuffs: selfBuffs,
	}
	// shaman.waterShieldManaMetrics = shaman.NewManaMetrics(core.ActionID{SpellID: 57960})

	core.FillTalentsProto(shaman.Talents.ProtoReflect(), talents, TalentTreeSizes)

	// Add Shaman stat dependencies
	shaman.AddStatDependency(stats.BonusArmor, stats.Armor, 1)
	shaman.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[shaman.Class])
	shaman.AddStatDependency(stats.Agility, stats.DodgeRating, 1.0/25*core.DodgeRatingPerDodgePercent)
	shaman.EnableManaBarWithModifier()

	shaman.AddStatDependency(stats.Agility, stats.AttackPower, 2.0)
	shaman.AddStat(stats.AttackPower, -20)

	shaman.AddStatDependency(stats.Strength, stats.AttackPower, 1.0)
	shaman.AddStat(stats.AttackPower, -10)

	if selfBuffs.Shield == proto.ShamanShield_WaterShield {
		shaman.AddStat(stats.MP5, 2138)
	}

	shaman.FireElemental = shaman.NewFireElemental()
	//shaman.EarthElemental = shaman.NewEarthElemental()

	return shaman
}

func (shaman *Shaman) GetImbueProcMask(imbue proto.ShamanImbue) core.ProcMask {
	var mask core.ProcMask
	if shaman.SelfBuffs.ImbueMH == imbue || shaman.SelfBuffs.ImbueMHSwap == imbue {
		mask |= core.ProcMaskMeleeMH
	}
	if shaman.SelfBuffs.ImbueOH == imbue {
		mask |= core.ProcMaskMeleeOH
	}
	return mask
}

// Which buffs this shaman is using.
type SelfBuffs struct {
	Shield      proto.ShamanShield
	ImbueMH     proto.ShamanImbue
	ImbueOH     proto.ShamanImbue
	ImbueMHSwap proto.ShamanImbue
	ImbueOHSwap proto.ShamanImbue
}

// Indexes into NextTotemDrops for self buffs
const (
	AirTotem int = iota
	EarthTotem
	FireTotem
	WaterTotem
)

// Shaman represents a shaman character.
type Shaman struct {
	core.Character

	Talents   *proto.ShamanTalents
	SelfBuffs SelfBuffs

	Totems *proto.ShamanTotems

	// The expiration time of each totem (earth, air, fire, water).
	TotemExpirations [4]time.Duration

	LightningBolt         *core.Spell
	LightningBoltOverload *core.Spell

	ChainLightning          *core.Spell
	ChainLightningOverloads []*core.Spell

	Stormstrike           *core.Spell
	StormstrikeCastResult *core.SpellResult

	LightningShield       *core.Spell
	LightningShieldDamage *core.Spell
	LightningShieldAura   *core.Aura

	EarthShock *core.Spell
	FlameShock *core.Spell
	FrostShock *core.Spell

	FireElementalTotem *core.Spell
	FireElemental      *FireElemental

	EarthElementalTotem *core.Spell
	EarthElemental      *EarthElemental

	StormStrikeDebuffAuras core.AuraArray

	ElementalSharedCDTimer *core.Timer

	MagmaTotem         *core.Spell
	HealingStreamTotem *core.Spell
	SearingTotem       *core.Spell
	TremorTotem        *core.Spell
	FireNovaTotemPA    *core.PendingAction

	waterShieldManaMetrics *core.ResourceMetrics
}

// Implemented by each Shaman spec.
type ShamanAgent interface {
	core.Agent

	// The Shaman controlled by this Agent.
	GetShaman() *Shaman
}

func (shaman *Shaman) GetCharacter() *core.Character {
	return &shaman.Character
}

func (shaman *Shaman) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
}

func (shaman *Shaman) Initialize() {
	shaman.registerChainLightningSpell()
	shaman.registerFireElementalTotem()
	//shaman.registerEarthElementalTotem()
	shaman.registerLightningBoltSpell()
	shaman.registerLightningShieldSpell()
	shaman.registerMagmaTotemSpell()
	shaman.registerSearingTotemSpell()
	shaman.registerFireNovaTotemSpell()
	shaman.registerShocks()
	shaman.registerStormstrikeSpell()
	shaman.registerBloodlustCD()
}

func (shaman *Shaman) ApplyTalents() {
	shaman.ApplyElementalTalents()
	shaman.ApplyEnhancementTalents()
	shaman.ApplyRestorationTalents()
}

func (shaman *Shaman) Reset(sim *core.Simulation) {
}

func (shaman *Shaman) OnEncounterStart(sim *core.Simulation) {
}

func (shaman *Shaman) GetOverloadChance() float64 {
	if shaman.Talents.LightningOverload == 0 {
		return 0.0
	}
	return 0.04 * float64(shaman.Talents.LightningOverload)
}

const (
	SpellMaskNone               int64 = 0
	SpellMaskFireElementalTotem int64 = 1 << iota
	SpellMaskEarthElementalTotem
	SpellMaskFireElementalMelee
	SpellMaskFlameShockDirect
	SpellMaskFlameShockDot
	SpellMaskLightningBolt
	SpellMaskLightningBoltOverload
	SpellMaskChainLightning
	SpellMaskChainLightningOverload
	SpellMaskEarthShock
	SpellMaskLightningShield
	SpellMaskMagmaTotem
	SpellMaskSearingTotem
	SpellMaskFireNovaTotem
	SpellMaskFlametongueTotem
	SpellMaskStormstrikeCast
	SpellMaskStormstrikeDamage
	SpellMaskEarthShield
	SpellMaskFrostShock
	SpellMaskFlametongueWeapon
	SpellMaskWindfuryWeapon
	SpellMaskFrostbrandWeapon
	SpellMaskRockbiterWeapon
	SpellMaskElementalMastery
	SpellMaskShamanisticRage
	SpellMaskBloodlust

	SpellMaskStormstrike  = SpellMaskStormstrikeCast | SpellMaskStormstrikeDamage
	SpellMaskFlameShock   = SpellMaskFlameShockDirect | SpellMaskFlameShockDot
	SpellMaskFire         = SpellMaskFlameShock
	SpellMaskNature       = SpellMaskLightningBolt | SpellMaskLightningBoltOverload | SpellMaskChainLightning | SpellMaskChainLightningOverload | SpellMaskEarthShock
	SpellMaskFrost        = SpellMaskFrostShock
	SpellMaskOverload     = SpellMaskLightningBoltOverload | SpellMaskChainLightningOverload
	SpellMaskShock        = SpellMaskFlameShock | SpellMaskEarthShock | SpellMaskFrostShock
	SpellMaskFireTotem    = SpellMaskMagmaTotem | SpellMaskSearingTotem | SpellMaskFireNovaTotem
	SpellMaskTotem        = SpellMaskFireTotem | SpellMaskFireElementalTotem | SpellMaskEarthElementalTotem
	SpellMaskInstantSpell = SpellMaskBloodlust
	SpellMaskImbue        = SpellMaskFrostbrandWeapon | SpellMaskWindfuryWeapon | SpellMaskFlametongueWeapon | SpellMaskRockbiterWeapon
)
