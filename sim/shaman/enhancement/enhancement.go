package enhancement

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/shaman"
)

func RegisterEnhancementShaman() {
	core.RegisterAgentFactory(
		proto.Player_EnhancementShaman{},
		proto.Spec_SpecEnhancementShaman,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
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
		ShieldProcrate: enhOptions.ClassOptions.ShieldProcrate,
		ImbueMH:        enhOptions.ClassOptions.ImbueMh,
		ImbueOH:        enhOptions.ImbueOh,
		ImbueMHSwap:    enhOptions.ClassOptions.ImbueMhSwap,
		ImbueOHSwap:    enhOptions.ImbueOhSwap,
	}

	enh := &EnhancementShaman{
		Shaman: shaman.NewShaman(character, options.TalentsString, selfBuffs),
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
}

func (enh *EnhancementShaman) GetShaman() *shaman.Shaman {
	return enh.Shaman
}

func (enh *EnhancementShaman) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	enh.Shaman.AddRaidBuffs(raidBuffs)
}

func (enh *EnhancementShaman) ApplyTalents() {
	enh.Shaman.ApplyTalents()
}

func (enh *EnhancementShaman) Initialize() {
	enh.Shaman.Initialize()
	// In the Initialize due to frost brand adding the aura to the enemy
	enh.RegisterFrostbrandImbue(enh.GetImbueProcMask(proto.ShamanImbue_FrostbrandWeapon))
	enh.RegisterFlametongueImbue(enh.GetImbueProcMask(proto.ShamanImbue_FlametongueWeapon))
	enh.RegisterWindfuryImbue(enh.GetImbueProcMask(proto.ShamanImbue_WindfuryWeapon))
}

func (enh *EnhancementShaman) Reset(sim *core.Simulation) {
	enh.Shaman.Reset(sim)
}

func (enh *EnhancementShaman) ApplySyncType(syncType proto.ShamanSyncType) {
	const FlurryICD = time.Millisecond * 500

	switch syncType {
	case proto.ShamanSyncType_Auto:
		enh.AutoSyncWeapons(FlurryICD)
	case proto.ShamanSyncType_SyncMainhandOffhandSwings:
		enh.SyncMainhandOffhandSwings(FlurryICD)
	case proto.ShamanSyncType_DelayOffhandSwings:
		enh.DelayOffhandSwing(FlurryICD)
	default:
		enh.AutoAttacks.SetReplaceMHSwing(nil)
	}
}

// Automatically sync weapons to 'None' sync type for mismatched weapon speeds or to
// delayed offhand sync type for matching weapon speeds
func (enh *EnhancementShaman) AutoSyncWeapons(FlurryICD time.Duration) {
	enh.AutoAttacks.SetReplaceMHSwing(func(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
		if mh, oh := enh.MainHand(), enh.OffHand(); mh.SwingSpeed != oh.SwingSpeed {

			return mhSwingSpell
		}
		return delayOffhandSwing(enh, sim, FlurryICD, mhSwingSpell)
	})
}

func (enh *EnhancementShaman) SyncMainhandOffhandSwings(FlurryICD time.Duration) {
	enh.AutoAttacks.SetReplaceMHSwing(func(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
		return syncMainhandOffhandSwings(enh, sim, FlurryICD, mhSwingSpell)
	})
}

func syncMainhandOffhandSwings(enh *EnhancementShaman, sim *core.Simulation, FlurryICD time.Duration, mhSwingSpell *core.Spell) *core.Spell {
	if aa := &enh.AutoAttacks; aa.OffhandSwingAt()-sim.CurrentTime > FlurryICD {
		if nextMHSwingAt := sim.CurrentTime + aa.MainhandSwingSpeed(); nextMHSwingAt != aa.OffhandSwingAt() {
			aa.SetOffhandSwingAt(nextMHSwingAt)
			if sim.Log != nil {
				enh.Unit.Log(sim, "(Weapon Sync for %s/%s) Syncing OH with MH, setting next OH swing to %s",
					aa.MainhandSwingSpeed().Truncate(time.Millisecond), aa.OffhandSwingSpeed().Truncate(time.Millisecond), nextMHSwingAt)
			}
		}
	}
	return mhSwingSpell
}

func (enh *EnhancementShaman) DelayOffhandSwing(FlurryICD time.Duration) {
	enh.AutoAttacks.SetReplaceMHSwing(func(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
		return delayOffhandSwing(enh, sim, FlurryICD, mhSwingSpell)
	})
}

func delayOffhandSwing(enh *EnhancementShaman, sim *core.Simulation, FlurryICD time.Duration, mhSwingSpell *core.Spell) *core.Spell {
	if aa := &enh.AutoAttacks; aa.OffhandSwingAt()-sim.CurrentTime > FlurryICD {
		if nextMHSwingAt := sim.CurrentTime + aa.MainhandSwingSpeed() + 100*time.Millisecond; nextMHSwingAt > aa.OffhandSwingAt() {
			aa.SetOffhandSwingAt(nextMHSwingAt)
			if sim.Log != nil {
				enh.Unit.Log(sim, "(Weapon Sync for %s/%s) Delaying OH swing, setting next OH swing to %s",
					aa.MainhandSwingSpeed().Truncate(time.Millisecond), aa.OffhandSwingSpeed().Truncate(time.Millisecond), nextMHSwingAt)
			}
		}
	}
	return mhSwingSpell
}
