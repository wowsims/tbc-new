package feralcat

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/druid"
)

// Ported from https://github.com/NerdEgghead/TBC_cat_sim

const BiteTrickCP = int32(2)
const BiteTrickMax = 39.0
const BiteTime = time.Duration(0)
const RipTrickMin = 52.0
const RipEndThresh = time.Second * 10
const MaxWaitTime = time.Second * 1

type FeralDruidRotation struct {
	RipCP          int32
	BiteCP         int32
	RipTrickCP     int32
	UseBite        bool
	BiteOverRip    bool
	UseMangleTrick bool
	UseRipTrick    bool
	UseRakeTrick   bool
	Wolfshead      bool

	MaintainFaerieFire bool
}

func (cat *FeralDruid) setupRotation(rotation *proto.FeralCatDruid_Rotation) {
	useBite := (rotation.Biteweave && rotation.FinishingMove == proto.FeralCatDruid_Rotation_Rip) ||
		rotation.FinishingMove == proto.FeralCatDruid_Rotation_Bite
	ripCP := rotation.RipMinComboPoints

	if rotation.FinishingMove != proto.FeralCatDruid_Rotation_Rip {
		ripCP = 6
	}

	cat.Rotation = FeralDruidRotation{
		RipCP:          ripCP,
		BiteCP:         rotation.BiteMinComboPoints,
		RipTrickCP:     rotation.RipMinComboPoints,
		UseBite:        useBite,
		BiteOverRip:    useBite && rotation.FinishingMove != proto.FeralCatDruid_Rotation_Rip,
		UseMangleTrick: rotation.MangleTrick,
		UseRipTrick:    rotation.Ripweave,
		UseRakeTrick:   rotation.RakeTrick,
		Wolfshead:      cat.HasItemEquipped(8345, []proto.ItemSlot{proto.ItemSlot_ItemSlotHead}),

		MaintainFaerieFire: rotation.MaintainFaerieFire,
	}
}

func (cat *FeralDruid) OnGCDReady(_ *core.Simulation) {}

func (cat *FeralDruid) shift(sim *core.Simulation) bool {
	cat.waitingForTick = false

	// If we have just now decided to shift, then we do not execute the
	// shift immediately, but instead trigger an input delay for realism.
	if !cat.readyToShift {
		cat.readyToShift = true
		return false
	}

	cat.readyToShift = false

	// Drop form, fire all ready cooldowns (potions, sappers, runes) while out
	// of form so their form-cancelling side-effects are irrelevant, then
	// immediately reshift. Mirrors PowerShiftCat from the old TBC sim.
	cat.ClearForm(sim)
	for _, mcd := range cat.GetMajorCooldowns() {
		if mcd.IsReady(sim) {
			mcd.TryActivate(sim, &cat.Character)
		}
	}

	if !cat.GCD.IsReady(sim) {
		return true
	}
	return cat.CatForm.Cast(sim, nil)
}

func (cat *FeralDruid) doRotation(sim *core.Simulation) bool {
	if !cat.GCD.IsReady(sim) {
		return false
	}

	// If we're out of form always shift back in.
	if !cat.InForm(druid.Cat) {
		return cat.CatForm.Cast(sim, nil)
	}

	// If we previously decided to shift, execute now after input delay.
	if cat.readyToShift {
		return cat.shift(sim)
	}

	rotation := &cat.Rotation

	// Maintain Faerie Fire (Feral) before other decisions.
	if rotation.MaintainFaerieFire && cat.FaerieFireFeral != nil &&
		cat.FaerieFireAuras != nil &&
		!cat.FaerieFireAuras.Get(cat.CurrentTarget).IsActive() &&
		cat.FaerieFireFeral.CanCast(sim, cat.CurrentTarget) {
		return cat.FaerieFireFeral.Cast(sim, cat.CurrentTarget)
	}

	energy := cat.CurrentEnergy()
	cp := cat.ComboPoints()
	ripDot := cat.Rip.CurDot()
	ripDebuff := ripDot.IsActive()
	ripEnd := core.NeverExpires
	if ripDebuff {
		ripEnd = ripDot.ExpiresAt()
	}
	mangleAura := cat.MangleAuras.Get(cat.CurrentTarget)
	mangleDebuff := mangleAura.IsActive()
	mangleEnd := core.NeverExpires
	if mangleDebuff {
		mangleEnd = mangleAura.ExpiresAt()
	}
	rakeDebuff := cat.Rake.CurDot().IsActive()
	nextTick := cat.NextEnergyTickAt()
	shiftCost := cat.CatForm.Cost.GetCurrentCost()
	omenProc := cat.ClearcastingAura.IsActive()

	ripCost := cat.CurrentRipCost()
	biteCost := cat.CurrentFerociousBiteCost()
	shredCost := cat.CurrentShredCost()
	mangleCost := 45.0 // sentinel when Mangle is not talented
	if cat.MangleCat != nil {
		mangleCost = cat.CurrentMangleCatCost()
	}

	ripNow := cp >= rotation.RipCP && !ripDebuff
	ripweaveNow := rotation.UseRipTrick &&
		cp >= rotation.RipTrickCP &&
		!ripDebuff &&
		energy >= RipTrickMin

	remainingDuration := sim.GetRemainingDuration()
	ripNow = (ripNow || ripweaveNow) && remainingDuration >= RipEndThresh

	biteAtEnd := cp >= rotation.BiteCP &&
		(remainingDuration < RipEndThresh ||
			(ripDebuff && sim.Duration-ripEnd < RipEndThresh))

	mangleNow := cat.MangleCat != nil && !ripNow && !mangleDebuff

	biteBeforeRip := ripDebuff && rotation.UseBite &&
		ripEnd-sim.CurrentTime >= BiteTime

	biteNow := (biteBeforeRip || rotation.BiteOverRip) &&
		cp >= rotation.BiteCP

	ripNext := (ripNow || (cp >= rotation.RipCP && ripEnd <= nextTick)) &&
		sim.Duration-nextTick >= RipEndThresh

	mangleNext := !ripNext && (mangleNow || mangleEnd <= nextTick)

	waitToMangle := mangleNext || (!rotation.Wolfshead && mangleCost <= 38)

	biteBeforeRipNext := biteBeforeRip && ripEnd-nextTick >= BiteTime

	prioBiteOverMangle := rotation.BiteOverRip || !mangleNow

	timeToNextTick := nextTick - sim.CurrentTime
	cat.waitingForTick = true
	markOOM := false

	if cat.CurrentMana() < shiftCost {
		// No-shift rotation when OOM.
		if ripNow && (energy >= ripCost || omenProc) {
			cat.Metrics.MarkOOM(sim)
			return cat.Rip.Cast(sim, cat.CurrentTarget)
		} else if mangleNow && (energy >= mangleCost || omenProc) {
			cat.Metrics.MarkOOM(sim)
			return cat.MangleCat.Cast(sim, cat.CurrentTarget)
		} else if biteNow && (energy >= biteCost || omenProc) {
			cat.Metrics.MarkOOM(sim)
			return cat.FerociousBite.Cast(sim, cat.CurrentTarget)
		} else if energy >= shredCost || omenProc {
			cat.Metrics.MarkOOM(sim)
			return cat.Shred.Cast(sim, cat.CurrentTarget)
		} else {
			markOOM = true
		}
	} else if energy < 10 {
		cat.shift(sim)
	} else if ripNow {
		if energy >= ripCost || omenProc {
			cat.Rip.Cast(sim, cat.CurrentTarget)
			cat.waitingForTick = false
		} else if timeToNextTick > MaxWaitTime {
			cat.shift(sim)
		}
	} else if (biteNow || biteAtEnd) && prioBiteOverMangle {
		cutoffMod := 20.0
		if timeToNextTick <= time.Second {
			cutoffMod = 0.0
		}
		if energy >= shredCost+15.0+cutoffMod || (energy >= 15+cutoffMod && omenProc) {
			return cat.Shred.Cast(sim, cat.CurrentTarget)
		}
		if energy >= biteCost {
			return cat.FerociousBite.Cast(sim, cat.CurrentTarget)
		}
		wait := false
		if energy >= 22 && biteBeforeRip && !biteBeforeRipNext {
			wait = true
		} else if energy >= 15 && (!biteBeforeRip || biteBeforeRipNext || biteAtEnd) {
			wait = true
		} else if !ripNext && (energy < 20 || !mangleNext) {
			wait = false
			cat.shift(sim)
		} else {
			wait = true
		}
		if wait && timeToNextTick > MaxWaitTime {
			cat.shift(sim)
		}
	} else if energy >= biteCost && energy <= BiteTrickMax &&
		rotation.UseRakeTrick &&
		timeToNextTick > cat.ReactionTime &&
		!omenProc &&
		cp >= BiteTrickCP {
		return cat.FerociousBite.Cast(sim, cat.CurrentTarget)
	} else if energy >= biteCost && energy < mangleCost &&
		rotation.UseRakeTrick &&
		timeToNextTick > time.Second+cat.ReactionTime &&
		!rakeDebuff &&
		!omenProc {
		return cat.Rake.Cast(sim, cat.CurrentTarget)
	} else if mangleNow {
		if energy < mangleCost-20 && !ripNext {
			cat.shift(sim)
		} else if energy >= mangleCost || omenProc {
			return cat.MangleCat.Cast(sim, cat.CurrentTarget)
		} else if timeToNextTick > MaxWaitTime {
			cat.shift(sim)
		}
	} else if energy >= 22 {
		if omenProc {
			return cat.Shred.Cast(sim, cat.CurrentTarget)
		}
		// Mangle trick: if energy is in range to fit two Mangles instead of
		// Shred + shift on the current cycle (relevant for no-Wolfshead rotations).
		if cat.MangleCat != nil && energy >= 2*mangleCost-20 && energy < 22+mangleCost &&
			timeToNextTick <= time.Second &&
			rotation.UseMangleTrick &&
			(!rotation.UseRakeTrick || mangleCost == 35) {
			return cat.MangleCat.Cast(sim, cat.CurrentTarget)
		}
		if energy >= shredCost {
			return cat.Shred.Cast(sim, cat.CurrentTarget)
		}
		if cat.MangleCat != nil && energy >= mangleCost && timeToNextTick > time.Second+cat.ReactionTime {
			return cat.MangleCat.Cast(sim, cat.CurrentTarget)
		}
		if timeToNextTick > MaxWaitTime {
			cat.shift(sim)
		}
	} else if !ripNext && (energy < mangleCost-20 || !waitToMangle) {
		cat.shift(sim)
	} else if timeToNextTick > MaxWaitTime {
		cat.shift(sim)
	}

	// Model input latency: delay GCD trigger after a power shift or energy tick.
	if cat.readyToShift {
		cat.WaitUntil(sim, sim.CurrentTime+cat.ReactionTime)
	} else if cat.waitingForTick {
		cat.WaitUntil(sim, sim.CurrentTime+timeToNextTick+cat.ReactionTime)
		if markOOM {
			cat.Metrics.MarkOOM(sim)
		}
	}

	return false
}
