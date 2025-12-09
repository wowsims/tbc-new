package enhancement

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
	"github.com/wowsims/tbc/sim/shaman"
)

func RegisterEnhancementShaman() {
	core.RegisterAgentFactory(
		proto.Player_EnhancementShaman{},
		proto.Spec_SpecEnhancementShaman,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewEnhancementShaman(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_EnhancementShaman)
			if !ok {
				panic("Invalid spec value for Enhancement Shaman!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewEnhancementShaman(character *core.Character, options *proto.Player) *EnhancementShaman {
	enhOptions := options.GetEnhancementShaman().Options

	selfBuffs := shaman.SelfBuffs{
		Shield:      enhOptions.ClassOptions.Shield,
		ImbueMH:     enhOptions.ClassOptions.ImbueMh,
		ImbueOH:     enhOptions.ImbueOh,
		ImbueMHSwap: enhOptions.ClassOptions.ImbueMhSwap,
		ImbueOHSwap: enhOptions.ImbueOhSwap,
	}

	enh := &EnhancementShaman{
		Shaman: shaman.NewShaman(character, options.TalentsString, selfBuffs, true, enhOptions.ClassOptions.FeleAutocast),
	}

	// Enable Auto Attacks for this spec
	enh.EnableAutoAttacks(enh, core.AutoAttackOptions{
		MainHand:       enh.WeaponFromMainHand(enh.DefaultMeleeCritMultiplier()),
		OffHand:        enh.WeaponFromOffHand(enh.DefaultMeleeCritMultiplier()),
		AutoSwingMelee: true,
	})

	enh.ApplySyncType(enhOptions.SyncType)

	if !enh.HasMHWeapon() {
		enh.SelfBuffs.ImbueMH = proto.ShamanImbue_NoImbue
	}

	if !enh.HasOHWeapon() {
		enh.SelfBuffs.ImbueOH = proto.ShamanImbue_NoImbue
	}

	enh.PseudoStats.CanParry = true

	return enh
}

type EnhancementShaman struct {
	*shaman.Shaman

	StormStrikeDebuffAuras core.AuraArray
}

func (enh *EnhancementShaman) GetShaman() *shaman.Shaman {
	return enh.Shaman
}

func (enh *EnhancementShaman) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	enh.Shaman.AddRaidBuffs(raidBuffs)
}

func (enh *EnhancementShaman) ApplyTalents() {
	// enh.ApplyEnhancementTalents()
	enh.Shaman.ApplyTalents()
	enh.ApplyArmorSpecializationEffect(stats.Agility, proto.ArmorType_ArmorTypeMail, 86529)
}

func (enh *EnhancementShaman) Initialize() {
	enh.Shaman.Initialize()
	// In the Initialize due to frost brand adding the aura to the enemy
	// enh.RegisterFrostbrandImbue(enh.GetImbueProcMask(proto.ShamanImbue_FrostbrandWeapon))
	// enh.RegisterFlametongueImbue(enh.GetImbueProcMask(proto.ShamanImbue_FlametongueWeapon))
	// enh.RegisterWindfuryImbue(enh.GetImbueProcMask(proto.ShamanImbue_WindfuryWeapon))

	if enh.ItemSwap.IsEnabled() {
		enh.RegisterItemSwapCallback(core.AllWeaponSlots(), func(_ *core.Simulation, slot proto.ItemSlot) {
			enh.ApplySyncType(proto.ShamanSyncType_Auto)
		})
	}

	//Mental Quickness
	enh.GetSpellDamageValue = func(spell *core.Spell) float64 {
		if spell.SpellID == 8024 {
			// Flametongue weapon damage scales with AP for enh
			return spell.MeleeAttackPower()
		}
		return spell.MeleeAttackPower() * 0.65
	}

	// enh.registerStormstrikeSpell()
}

func (enh *EnhancementShaman) Reset(sim *core.Simulation) {
	enh.Shaman.Reset(sim)
}

func (enh *EnhancementShaman) AutoSyncWeapons() proto.ShamanSyncType {
	if mh, oh := enh.MainHand(), enh.OffHand(); mh.SwingSpeed != oh.SwingSpeed {
		return proto.ShamanSyncType_NoSync
	}
	return proto.ShamanSyncType_SyncMainhandOffhandSwings
}

func (enh *EnhancementShaman) ApplySyncType(syncType proto.ShamanSyncType) {
	const FlurryICD = time.Millisecond * 500

	if syncType == proto.ShamanSyncType_Auto {
		syncType = enh.AutoSyncWeapons()
	}

	switch syncType {
	case proto.ShamanSyncType_SyncMainhandOffhandSwings:
		enh.AutoAttacks.SetReplaceMHSwing(func(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
			if aa := &enh.AutoAttacks; aa.OffhandSwingAt()-sim.CurrentTime > FlurryICD {
				if nextMHSwingAt := sim.CurrentTime + aa.MainhandSwingSpeed(); nextMHSwingAt != aa.OffhandSwingAt() {
					aa.SetOffhandSwingAt(nextMHSwingAt)
				}
			}
			return mhSwingSpell
		})
	case proto.ShamanSyncType_DelayOffhandSwings:
		enh.AutoAttacks.SetReplaceMHSwing(func(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
			if aa := &enh.AutoAttacks; aa.OffhandSwingAt()-sim.CurrentTime > FlurryICD {
				if nextMHSwingAt := sim.CurrentTime + aa.MainhandSwingSpeed() + 100*time.Millisecond; nextMHSwingAt > aa.OffhandSwingAt() {
					aa.SetOffhandSwingAt(nextMHSwingAt)
				}
			}
			return mhSwingSpell
		})
	default:
		enh.AutoAttacks.SetReplaceMHSwing(nil)
	}
}
