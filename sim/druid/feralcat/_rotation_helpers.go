package feral

import (
	"math"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/druid"
)

func (cat *FeralDruid) tfExpectedBefore(sim *core.Simulation, futureTime time.Duration) bool {
	if !cat.TigersFury.IsReady(sim) {
		return cat.TigersFury.ReadyAt() < futureTime
	}
	if cat.BerserkCatAura.IsActive() {
		return cat.BerserkCatAura.ExpiresAt() < futureTime
	}
	return true
}

func (rotation *FeralDruidRotation) WaitUntil(sim *core.Simulation, nextEvaluation time.Duration) {
	rotation.nextActionAt = nextEvaluation
	rotation.agent.WaitUntil(sim, nextEvaluation)
}

func (cat *FeralDruid) calcTfEnergyThresh() float64 {
	delayTime := cat.ReactionTime + core.TernaryDuration(cat.ClearcastingAura.IsActive(), time.Second, 0)
	return 40.0 - delayTime.Seconds()*cat.EnergyRegenPerSecond()
}

func (rotation *FeralDruidRotation) TryTigersFury(sim *core.Simulation) {
	cat := rotation.agent

	if !cat.TigersFury.IsReady(sim) {
		return
	}

	// Don't over-cap Energy with TF, unless the next special is a DoC Rip.
	tfEnergyThresh := core.TernaryFloat64(rotation.UseHealingTouch && cat.DreamOfCenariusAura.IsActive() && (cat.ComboPoints() == 5), 100, cat.calcTfEnergyThresh())
	tfNow := (cat.CurrentEnergy() < tfEnergyThresh) && !cat.BerserkCatAura.IsActive()

	if tfNow {
		cat.TigersFury.Cast(sim, nil)
		rotation.WaitUntil(sim, sim.CurrentTime+cat.ReactionTime)
	}
}

func (rotation *FeralDruidRotation) TryBerserk(sim *core.Simulation) {
	// Berserk algorithm: time Berserk for just after a Tiger's Fury
	// *unless* we'll lose Berserk uptime by waiting for Tiger's Fury to
	// come off cooldown.
	cat := rotation.agent
	simTimeRemain := sim.GetRemainingDuration()
	tfCdRemain := cat.TigersFury.TimeToReady(sim)
	waitForTf := !cat.TigersFuryAura.IsActive() && (tfCdRemain+cat.ReactionTime < simTimeRemain-cat.BerserkCatAura.Duration)
	berserkNow := rotation.UseBerserk && cat.Berserk.IsReady(sim) && !waitForTf && !cat.ClearcastingAura.IsActive() && (cat.CurrentEnergy() > 60)

	if berserkNow && (simTimeRemain/cat.Berserk.CD.Duration == 0) && !sim.IsExecutePhase25() {
		projectedExecuteStart := core.DurationFromSeconds((1.0 - sim.Encounter.ExecuteProportion_25) * sim.Duration.Seconds())

		if (sim.CurrentTime+tfCdRemain < projectedExecuteStart) && (tfCdRemain+cat.ReactionTime < simTimeRemain-cat.BerserkCatAura.Duration) {
			allProcsReady := true

			for _, aura := range rotation.itemProcAuras {
				if !aura.IsActive() && !aura.Icd.IsReady(sim) {
					allProcsReady = false
					break
				}
			}

			if !allProcsReady {
				return
			}
		}
	}

	if berserkNow {
		cat.Berserk.Cast(sim, nil)

		if (cat.Incarnation != nil) && cat.Incarnation.IsReady(sim) && !cat.ClearcastingAura.IsActive() && (cat.CurrentEnergy()+cat.EnergyRegenPerSecond() < 100) {
			cat.Incarnation.Cast(sim, nil)
		}

		cat.UpdateMajorCooldowns()
		rotation.WaitUntil(sim, sim.CurrentTime+cat.ReactionTime)
	}
}

func (rotation *FeralDruidRotation) ShiftBearCat(sim *core.Simulation) {
	rotation.readyToShift = false
	rotation.lastShiftAt = sim.CurrentTime
	cat := rotation.agent

	if cat.InForm(druid.Cat) {
		cat.BearForm.Cast(sim, nil)
	} else {
		cat.CatForm.Cast(sim, nil)

		if cat.ItemSwap.IsEnabled() {
			cat.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Main, false)
		}

		// Reset swing timer with Albino Snake when advantageous
		if rotation.SnekWeave && (cat.AutoAttacks.NextAttackAt()-sim.CurrentTime > cat.AutoAttacks.MainhandSwingSpeed()) {
			cat.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime)
		}
	}
}

func (cat *FeralDruid) calcBleedRefreshTime(sim *core.Simulation, bleedSpell *druid.DruidSpell, bleedDot *core.Dot, isExecutePhase bool, isRip bool) time.Duration {
	if !bleedDot.IsActive() {
		return sim.CurrentTime - cat.ReactionTime
	}

	// DoC takes priority over other logic.
	if cat.DreamOfCenariusAura.IsActive() && (bleedSpell.NewSnapshotPower > bleedSpell.CurrentSnapshotPower+0.001) {
		return sim.CurrentTime - cat.ReactionTime
	}

	// If we're not gaining a stronger snapshot, then use the standard 1
	// tick refresh window.
	bleedEnd := bleedDot.ExpiresAt()
	standardRefreshTime := bleedEnd - bleedDot.BaseTickLength

	if !cat.tempSnapshotAura.IsActive() {
		return standardRefreshTime
	}

	// For Rip specifically, also bypass clipping calculations if CP count
	// is too low for the calculation to be relevant.
	if isRip && (cat.ComboPoints() < 5) {
		return standardRefreshTime
	}

	// Likewise, if the existing buff will still be up at the start of the normal
	// window, then don't clip unnecessarily. For long buffs that cover a full bleed
	// duration, project "buffEnd" forward in time such that we block clips if we are
	// already maxing out the number of full durations we can snapshot.
	buffRemains := cat.tempSnapshotAura.RemainingDuration(sim)
	maxTickCount := core.TernaryInt32(isRip, cat.RipMaxNumTicks, bleedDot.BaseTickCount)
	maxBleedDur := bleedDot.BaseTickLength * time.Duration(maxTickCount)
	numCastsCovered := buffRemains / maxBleedDur
	buffEnd := cat.tempSnapshotAura.ExpiresAt() - numCastsCovered*maxBleedDur

	if buffEnd > standardRefreshTime+cat.ReactionTime {
		return standardRefreshTime
	}

	// Potential clips for a buff snapshot should be done as late as possible
	latestPossibleSnapshot := buffEnd - cat.ReactionTime*time.Duration(2)
	numClippedTicks := (bleedEnd - latestPossibleSnapshot) / bleedDot.BaseTickLength
	targetClipTime := standardRefreshTime - numClippedTicks*bleedDot.BaseTickLength

	// Since the clip can cost us 30-35 Energy, we need to determine whether the damage gain is worth the
	// spend. First calculate the maximum number of buffed bleed ticks we can get out before the fight
	// ends.
	buffedTickCount := min(maxTickCount, int32((sim.Duration-targetClipTime)/bleedDot.BaseTickLength))

	// Perform a DPE comparison vs. Shred
	expectedDamageGain := (bleedSpell.NewSnapshotPower - bleedSpell.CurrentSnapshotPower) * float64(buffedTickCount)

	// For Rake specifically, we get 1 free "tick" immediately upon cast.
	if !isRip {
		expectedDamageGain += bleedSpell.NewSnapshotPower
	}

	shredDpc := cat.Shred.ExpectedInitialDamage(sim, cat.CurrentTarget)
	energyEquivalent := expectedDamageGain / shredDpc * cat.Shred.DefaultCast.Cost

	// Finally, discount the effective Energy cost of the clip based on the number of clipped ticks.
	discountedRefreshCost := core.TernaryFloat64(isRip, float64(numClippedTicks)/float64(maxTickCount), 1.0) * bleedSpell.DefaultCast.Cost

	if sim.Log != nil {
		cat.Log(sim, "%s buff snapshot is worth %.1f Energy, discounted refresh cost is %.1f Energy.", bleedSpell.ShortName, energyEquivalent, discountedRefreshCost)
	}

	if cat.BerserkCatAura.IsActive() && (cat.BerserkCatAura.ExpiresAt() > targetClipTime+cat.ReactionTime) {
		return core.TernaryDuration(expectedDamageGain > shredDpc, targetClipTime, standardRefreshTime)
	} else {
		return core.TernaryDuration(energyEquivalent > discountedRefreshCost, targetClipTime, standardRefreshTime)
	}
}

// Determine whether Tiger's Fury will be usable soon enough for the snapshot to
// outweigh the lost Rip/Rake ticks from delaying a refresh.
func (cat *FeralDruid) shouldDelayBleedRefreshForTf(sim *core.Simulation, bleedDot *core.Dot, isRip bool) bool {
	if cat.TigersFuryAura.IsActive() || cat.BerserkCatAura.IsActive() || cat.DreamOfCenariusAura.IsActive() {
		return false
	}

	finalTickLeeway := core.TernaryDuration(bleedDot.IsActive(), bleedDot.TimeUntilNextTick(sim), 0)
	maxTickCount := core.TernaryInt32(isRip, cat.RipMaxNumTicks, bleedDot.BaseTickCount)
	buffedTickCount := min(maxTickCount, int32((sim.GetRemainingDuration()-finalTickLeeway)/bleedDot.BaseTickLength))
	delayBreakpoint := finalTickLeeway + core.DurationFromSeconds(0.15*float64(buffedTickCount)*bleedDot.BaseTickLength.Seconds())

	if !cat.tfExpectedBefore(sim, sim.CurrentTime+delayBreakpoint) {
		return false
	}

	if isRip && cat.tempSnapshotAura.IsActive() && (cat.tempSnapshotAura.RemainingDuration(sim) <= delayBreakpoint) {
		return false
	}

	delaySeconds := delayBreakpoint.Seconds()
	energyToDump := cat.CurrentEnergy() + delaySeconds*cat.EnergyRegenPerSecond() - cat.calcTfEnergyThresh()
	secondsToDump := math.Ceil(energyToDump / cat.Shred.DefaultCast.Cost)
	return secondsToDump < delaySeconds
}

func (cat *FeralDruid) calcRoarRefreshTime(sim *core.Simulation, ripRefreshTime time.Duration, ripLeeway time.Duration, minRoarOffset time.Duration) time.Duration {
	roarBuff := cat.SavageRoarBuff
	ripDot := cat.Rip.CurDot()

	if !roarBuff.IsActive() {
		return sim.CurrentTime - cat.ReactionTime
	}

	// If we're not proactively offsetting the Roar, then use the standard 1
	// tick refresh window, unless there is a Rip conflict.
	roarEnd := roarBuff.ExpiresAt()

	if !ripDot.IsActive() || (ripRefreshTime < roarEnd+cat.ReactionTime) {
		return roarEnd
	}

	if cat.ComboPoints() == 0 {
		return roarEnd
	}

	standardRefreshTime := core.TernaryDuration(cat.ComboPoints() < 5, roarEnd, roarEnd-roarBuff.BaseTickLength)

	// Project Rip end time assuming full Bloodletting extensions
	remainingExtensions := cat.RipMaxNumTicks - ripDot.BaseTickCount
	ripEnd := ripDot.ExpiresAt() + time.Duration(remainingExtensions)*ripDot.BaseTickLength
	fightEnd := sim.Duration

	if roarEnd > (ripEnd + ripLeeway) {
		return standardRefreshTime
	}

	if roarEnd >= fightEnd {
		return standardRefreshTime
	}

	// Potential clips for offsetting timers should be done just after a
	// Roar "tick" in order to exploit the Pandemic behavior in MoP.
	targetClipTime := roarBuff.NextTickAt()

	// Calculate when Roar would end if refreshed at the optimal clip time.
	newRoarDur := cat.SavageRoarDurationTable[cat.ComboPoints()]
	newRoarEnd := targetClipTime + newRoarDur + roarBuff.BaseTickLength

	// If a fresh Roar cast would cover us to the end of the fight, then
	// clip at the next tick for maximum CP efficiency.
	if newRoarEnd >= fightEnd {
		return targetClipTime
	}

	// Outside of Execute, use offset rule to determine whether to clip.
	if !sim.IsExecutePhase25() {
		return core.TernaryDuration(newRoarEnd >= ripEnd+minRoarOffset, targetClipTime, standardRefreshTime)
	}

	// Under Execute conditions, ignore the offset rule and instead optimize
	// for as few Roar casts as possible.
	if cat.ComboPoints() < 5 {
		return standardRefreshTime
	}

	minRoarsPossible := (fightEnd - roarEnd) / newRoarDur
	projectedRoarCasts := (fightEnd-newRoarEnd)/newRoarDur + 1
	return core.TernaryDuration(projectedRoarCasts == minRoarsPossible, targetClipTime, standardRefreshTime)
}

func (cat *FeralDruid) canBearWeave(sim *core.Simulation, furorCap float64, regenRate float64, currentEnergy float64, excessEnergy float64, upcomingTimers *PoolingActions) bool {
	if cat.ClearcastingAura.IsActive() || cat.BerserkCatAura.IsActive() {
		return false
	}

	// If we can Shred now and then weave on the next GCD, prefer that.
	if excessEnergy > cat.Shred.DefaultCast.Cost {
		return false
	}

	// Calculate effective Energy cap for out-of-form pooling.
	targetWeaveDuration := core.GCDDefault*3 + cat.ReactionTime*2
	maxStartingEnergy := furorCap - targetWeaveDuration.Seconds()*regenRate

	if currentEnergy > maxStartingEnergy {
		return false
	}

	// Prioritize all timers over weaving.
	earliestWeaveEnd := sim.CurrentTime + core.GCDDefault*3 + cat.ReactionTime*2
	isPooling, nextRefresh := upcomingTimers.nextRefreshTime()

	if isPooling && (nextRefresh < earliestWeaveEnd) {
		return false
	}

	// Mana check
	if cat.CurrentMana() < cat.CatForm.DefaultCast.Cost*2 {
		cat.Metrics.MarkOOM(sim)
		return false
	}

	// Also add a condition to make sure we can spend down our Energy
	// post-weave before the encounter ends or TF is ready.
	energyToDump := currentEnergy + (earliestWeaveEnd-sim.CurrentTime).Seconds()*regenRate
	timeToDump := earliestWeaveEnd + core.DurationFromSeconds(math.Floor(energyToDump/cat.Shred.DefaultCast.Cost))
	return (timeToDump < sim.Duration) && !cat.tfExpectedBefore(sim, timeToDump)
}

func (rotation *FeralDruidRotation) shouldTerminateBearWeave(sim *core.Simulation, isClearcast bool, currentEnergy float64, furorCap float64, regenRate float64, upcomingTimers *PoolingActions) bool {
	// Shift back early if a bear auto resulted in an Omen proc.
	if isClearcast && (sim.CurrentTime-rotation.lastShiftAt > core.GCDDefault) {
		return true
	}

	// Check Energy pooling leeway.
	cat := rotation.agent
	smallestWeaveExtension := core.GCDDefault + cat.ReactionTime
	finalEnergy := currentEnergy + smallestWeaveExtension.Seconds()*regenRate

	if finalEnergy > furorCap {
		return true
	}

	// Check timer leeway.
	earliestWeaveEnd := sim.CurrentTime + smallestWeaveExtension + core.GCDDefault
	isPooling, nextRefresh := upcomingTimers.nextRefreshTime()

	if isPooling && (nextRefresh < earliestWeaveEnd) {
		return true
	}

	// Also add a condition to prevent extending a weave if we don't have
	// enough time to spend the pooled Energy thus far.
	energyToDump := finalEnergy + 1.5*regenRate // need to include Cat Form GCD here
	timeToDump := earliestWeaveEnd + core.DurationFromSeconds(math.Floor(energyToDump/cat.Shred.DefaultCast.Cost))
	return (timeToDump >= sim.Duration) || cat.tfExpectedBefore(sim, timeToDump)
}

func (rotation *FeralDruidRotation) shouldWrathWeave(sim *core.Simulation) bool {
	if !rotation.WrathWeave {
		return false
	}

	cat := rotation.agent
	remainingGCD := cat.GCD.TimeToReady(sim)
	maxWrathCastTime := cat.Wrath.DefaultCast.CastTime

	if !cat.HeartOfTheWildAura.IsActive() || (cat.HeartOfTheWildAura.RemainingDuration(sim) <= maxWrathCastTime+remainingGCD) {
		return false
	}

	if cat.ClearcastingAura.IsActive() {
		return false
	}

	regenRate := cat.EnergyRegenPerSecond()
	furorCap := 100.0 - 1.5*regenRate
	startingEnergy := cat.CurrentEnergy() + remainingGCD.Seconds()*regenRate
	curCp := cat.ComboPoints()

	if (curCp < 3) && (startingEnergy+maxWrathCastTime.Seconds()*2*regenRate > furorCap) {
		return false
	}

	ripDot := cat.Rip.CurDot()
	timeToNextCatSpecial := remainingGCD + maxWrathCastTime + cat.ReactionTime + core.GCDDefault

	if !ripDot.IsActive() || ((curCp == 5) && (ripDot.RemainingDuration(sim) < timeToNextCatSpecial)) {
		return false
	}

	rakeDot := cat.Rake.CurDot()

	if !rakeDot.IsActive() || (rakeDot.RemainingDuration(sim) < timeToNextCatSpecial) {
		return false
	}

	return true
}

func (rotation *FeralDruidRotation) ProcessNextPlannedAction(sim *core.Simulation, nextActionAt time.Duration) {
	// Also schedule an action right at Energy cap to make sure we never
	// accidentally over-cap while waiting on other timers.
	cat := rotation.agent
	timeToCap := core.DurationFromSeconds((cat.MaximumEnergy() - cat.CurrentEnergy()) / cat.EnergyRegenPerSecond())
	nextActionAt = min(nextActionAt, sim.CurrentTime+timeToCap)

	// Offset the ideal evaluation time by player latency.
	nextActionAt += cat.ReactionTime

	if nextActionAt <= sim.CurrentTime {
		panic("nextActionAt in the past!")
	} else {
		rotation.WaitUntil(sim, nextActionAt)
	}
}

func (rotation *FeralDruidRotation) shouldAoeRake(sim *core.Simulation, roarNow bool, shouldSingleTargetRake bool) (bool, *core.Unit) {
	if roarNow {
		return false, nil
	}

	cat := rotation.agent

	if rotation.RotationType == proto.FeralDruid_Rotation_SingleTarget {
		return shouldSingleTargetRake, cat.CurrentTarget
	}

	if cat.ClearcastingAura.IsActive() || !cat.ThrashCat.CurDot().IsActive() {
		return false, nil
	}

	var shouldRake bool
	var rakeTarget *core.Unit
	var rakeDot *core.Dot

	for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
		rakeDot = cat.Rake.Dot(aoeTarget)

		if !rakeDot.IsActive() || (rakeDot.RemainingDuration(sim) < rakeDot.BaseTickLength) {
			shouldRake = true
			rakeTarget = aoeTarget
			break
		}
	}

	if !shouldRake {
		return false, nil
	}

	// Compare DPE versus Swipe to see if it's worth casting
	potentialRakeTicks := min(rakeDot.BaseTickCount, int32(sim.GetRemainingDuration()/rakeDot.BaseTickLength))
	expectedRakeDamage := cat.Rake.ExpectedInitialDamage(sim, rakeTarget) + cat.Rake.ExpectedTickDamage(sim, rakeTarget)*float64(potentialRakeTicks)
	rakeDPE := expectedRakeDamage / cat.Rake.DefaultCast.Cost

	var expectedSwipeDamage float64

	for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
		expectedSwipeDamage += cat.SwipeCat.ExpectedInitialDamage(sim, aoeTarget)
	}

	swipeDPE := expectedSwipeDamage / cat.SwipeCat.DefaultCast.Cost
	shouldRake = core.Ternary(cat.BerserkCatAura.IsActive(), expectedRakeDamage > expectedSwipeDamage, rakeDPE > swipeDPE)

	return shouldRake, rakeTarget
}
