package feralcat

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/druid"
)

func RegisterFeralCatDruid() {
	core.RegisterAgentFactory(
		proto.Player_FeralCatDruid{},
		proto.Spec_SpecFeralCatDruid,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewFeralCatDruid(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_FeralCatDruid)
			if !ok {
				panic("Invalid spec value for Feral Druid!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewFeralCatDruid(character *core.Character, options *proto.Player) *FeralDruid {
	feralOptions := options.GetFeralCatDruid()
	selfBuffs := druid.SelfBuffs{}

	cat := &FeralDruid{
		Druid: druid.New(character, druid.Cat, selfBuffs, options.TalentsString),
	}

	cat.CannotShredTarget = feralOptions.Options.CannotShredTarget

	cat.EnableEnergyBar(core.EnergyBarOptions{
		MaxComboPoints: 5,
		MaxEnergy:      100.0,
		UnitClass:      proto.Class_ClassDruid,
	})
	cat.EnableRageBar(core.RageBarOptions{BaseRageMultiplier: 2.5})

	cat.EnableAutoAttacks(cat, core.AutoAttackOptions{
		// Base paw weapon.
		MainHand:       cat.GetCatWeapon(),
		AutoSwingMelee: true,
	})

	cat.RegisterCatFormAura()
	cat.RegisterBearFormAura()

	return cat
}

type FeralDruid struct {
	*druid.Druid

	Rotation FeralDruidRotation

	readyToShift   bool
	waitingForTick bool
}

func (cat *FeralDruid) GetDruid() *druid.Druid {
	return cat.Druid
}

func (cat *FeralDruid) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
}

// AddPartyBuffs auto-applies Leader of the Pack from the druid's own talent.
// Windfury Totem is stripped so cats never register the WF proc aura, but
// TotemTwisting is preserved so Grace of Air gets the correct reduced uptime.
func (cat *FeralDruid) AddPartyBuffs(partyBuffs *proto.PartyBuffs) {
	if cat.Talents.LeaderOfThePack {
		// Idol of the Raven Goddess (32387) upgrades the LotP party buff to Improved (+2% crit).
		// The ImprovedLeaderOfThePack talent provides healing-on-crit only, NOT extra crit.
		if cat.HasItemEquipped(32387, []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}) {
			if partyBuffs.LeaderOfThePack < proto.TristateEffect_TristateEffectImproved {
				partyBuffs.LeaderOfThePack = proto.TristateEffect_TristateEffectImproved
			}
		} else {
			if partyBuffs.LeaderOfThePack < proto.TristateEffect_TristateEffectRegular {
				partyBuffs.LeaderOfThePack = proto.TristateEffect_TristateEffectRegular
			}
		}
	}

	// Feral cats do not proc Windfury Totem. Strip the aura so the sim never
	// registers WF procs, while keeping TotemTwisting intact so that Grace of
	// Air receives the correct ~90% uptime when twisting is enabled.
	partyBuffs.WindfuryTotem = proto.TristateEffect_TristateEffectMissing
}

func (cat *FeralDruid) Initialize() {
	cat.Druid.Initialize()
	cat.RegisterFeralCatSpells()
}

func (cat *FeralDruid) ApplyTalents() {
	cat.Druid.ApplyTalents()
}

// DrumsOfBattleActionID is the player's own Drums of Battle consumable. The
// party-provided version shares this spell id but carries tag -1.
var DrumsOfBattleActionID = core.ActionID{SpellID: 35476}

func (cat *FeralDruid) Reset(sim *core.Simulation) {
	cat.Druid.Reset(sim)
	cat.Druid.ClearForm(sim)
	cat.CatFormAura.Activate(sim)
	cat.readyToShift = false
	cat.waitingForTick = false

	drums := cat.majorCooldown(DrumsOfBattleActionID)
	if drums == nil {
		drums = cat.majorCooldown(DrumsOfBattleActionID.WithTag(-1))
	}

	cat.scheduleRecurringMCD(sim, drums, 0)
	cat.scheduleFixedMCD(sim, cat.majorCooldownBySpellID(core.BloodlustActionID.SpellID), 5*time.Second)
}

// majorCooldown finds the MCD for an exact ActionID.
func (cat *FeralDruid) majorCooldown(actionID core.ActionID) *core.MajorCooldown {
	for _, mcd := range cat.GetMajorCooldowns() {
		if mcd.Spell.ActionID.SameAction(actionID) {
			return mcd
		}
	}

	return nil
}

func (cat *FeralDruid) majorCooldownBySpellID(spellID int32) *core.MajorCooldown {
	for _, mcd := range cat.GetMajorCooldowns() {
		if mcd.Spell.ActionID.SpellID == spellID {
			return mcd
		}
	}

	return nil
}

// scheduleFixedMCD schedules a one-shot firing of the given MCD at the given sim time.
func (cat *FeralDruid) scheduleFixedMCD(sim *core.Simulation, mcd *core.MajorCooldown, fireAt time.Duration) {
	if mcd == nil {
		return
	}

	pa := sim.GetConsumedPendingActionFromPool()
	pa.NextActionAt = fireAt
	pa.OnAction = func(sim *core.Simulation) {
		if mcd.IsReady(sim) {
			mcd.TryActivate(sim, &cat.Character)
		}
	}

	sim.AddPendingAction(pa)
}

// scheduleRecurringMCD fires the given MCD at fireAt, then re-arms itself so it
// is used again every time the cooldown comes back up.
func (cat *FeralDruid) scheduleRecurringMCD(sim *core.Simulation, mcd *core.MajorCooldown, fireAt time.Duration) {
	if mcd == nil {
		return
	}

	pa := sim.GetConsumedPendingActionFromPool()
	pa.NextActionAt = fireAt
	pa.OnAction = func(sim *core.Simulation) {
		cast := mcd.IsReady(sim) && mcd.TryActivate(sim, &cat.Character)

		nextAt := sim.CurrentTime + mcd.TimeToNextCast(sim)
		if !cast {
			nextAt = max(nextAt, sim.CurrentTime+core.GCDDefault)
		}

		if nextAt < sim.Duration {
			pa.NextActionAt = nextAt
			sim.AddPendingAction(pa)
		}
	}

	sim.AddPendingAction(pa)
}
