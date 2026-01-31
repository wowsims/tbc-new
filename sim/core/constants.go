package core

import (
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
)

const CharacterLevel = 70
const MinIlvl = 60
const MaxIlvl = 600

const GCDMin = time.Second * 1
const GCDDefault = time.Millisecond * 1500
const BossGCD = time.Millisecond * 1620
const MaxSpellQueueWindow = time.Millisecond * 400
const SpellBatchWindow = time.Millisecond * 10
const PetUpdateInterval = time.Millisecond * 5250
const RppmLastCheckCap = time.Second * 10
const RppmLastProcCap = time.Second * 1000
const RppmLastProcResetValue = time.Second * 120
const MaxMeleeRange = 5.0 // in yards

const DefaultAttackPowerPerDPS = 14.0

const ArmorPenPerPercentArmor = 5.92
const MissDodgeParryBlockCritChancePerDefense = 0.04
const ResilienceRatingPerCritReductionChance = 39.4231
const ResilienceRatingPerCritDamageReductionPercent = 39.4231 / 2
const DefenseRatingToChanceReduction = (1.0 / DefenseRatingPerDefenseLevel) * MissDodgeParryBlockCritChancePerDefense / 100

const EnemyAutoAttackAPCoefficient = 0.000649375

// IDs for items used in core
// const ()

type Hand bool

const MainHand Hand = true
const OffHand Hand = false

const CombatTableCoverageCap = 1.024 // 102.4% chance to avoid an attack

const NumItemSlots = proto.ItemSlot_ItemSlotRanged + 1

func TrinketSlots() []proto.ItemSlot {
	return []proto.ItemSlot{proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2}
}

func AllWeaponSlots() []proto.ItemSlot {
	return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand, proto.ItemSlot_ItemSlotRanged}
}

func AllMeleeWeaponSlots() []proto.ItemSlot {
	return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}
}

func ArmorSpecializationSlots() []proto.ItemSlot {
	return []proto.ItemSlot{
		proto.ItemSlot_ItemSlotHead,
		proto.ItemSlot_ItemSlotShoulder,
		proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotWrist,
		proto.ItemSlot_ItemSlotHands,
		proto.ItemSlot_ItemSlotWaist,
		proto.ItemSlot_ItemSlotLegs,
		proto.ItemSlot_ItemSlotFeet,
	}
}
