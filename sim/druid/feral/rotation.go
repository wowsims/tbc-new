package feral

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func (cat *FeralDruid) NewAPLAction(_ *core.APLRotation, config *proto.APLAction) core.APLActionImpl {
	switch config.Action.(type) {
	case *proto.APLAction_CatOptimalRotationAction:
		return cat.newActionCatOptimalRotationAction(config.GetCatOptimalRotationAction())
	default:
		return nil
	}
}

func (cat *FeralDruid) newActionCatOptimalRotationAction(config *proto.APLActionCatOptimalRotationAction) core.APLActionImpl {
	rotation := &FeralDruidRotation{
		APLActionCatOptimalRotationAction: config,
		agent:                             cat,
	}

	// Process rotation parameters
	rotation.ForceMangleFiller = cat.PseudoStats.InFrontOfTarget || cat.CannotShredTarget
	rotation.UseBerserk = (rotation.RotationType == proto.FeralDruid_Rotation_SingleTarget) || rotation.AllowAoeBerserk
	rotation.UseHealingTouch = cat.Talents.DreamOfCenarius

	if rotation.ManualParams {
		rotation.BiteTime = core.DurationFromSeconds(config.BiteTime)
		rotation.BerserkBiteTime = core.DurationFromSeconds(config.BerserkBiteTime)
		rotation.MinRoarOffset = core.DurationFromSeconds(config.MinRoarOffset)
		rotation.RipLeeway = core.DurationFromSeconds(config.RipLeeway)
	} else {
		rotation.UseBite = (rotation.RotationType == proto.FeralDruid_Rotation_SingleTarget)
		rotation.BiteTime = core.TernaryDuration(rotation.UseHealingTouch, time.Second*4, time.Second*7)
		rotation.BerserkBiteTime = core.TernaryDuration(rotation.UseHealingTouch, time.Second*3, time.Second*7)
		rotation.MinRoarOffset = time.Second * 40
		rotation.RipLeeway = core.TernaryDuration(rotation.UseHealingTouch, time.Second*2, time.Second*5)
	}

	// Pre-allocate PoolingActions
	rotation.pendingPool = &PoolingActions{}
	rotation.pendingPool.create(4)
	rotation.pendingPoolWeaves = &PoolingActions{}
	rotation.pendingPoolWeaves.create(3)

	// Store relevant proc auras for snapshot timing.
	rotation.itemProcAuras = cat.GetMatchingItemProcAuras([]stats.Stat{stats.Agility, stats.AttackPower, stats.MasteryRating}, time.Second*30)

	return rotation
}

type FeralDruidRotation struct {
	*proto.APLActionCatOptimalRotationAction

	// Overwritten parameters
	BiteTime          time.Duration
	BerserkBiteTime   time.Duration
	MinRoarOffset     time.Duration
	RipLeeway         time.Duration
	ForceMangleFiller bool
	UseBerserk        bool
	UseHealingTouch   bool

	// Bookkeeping fields
	agent             *FeralDruid
	lastActionAt      time.Duration
	nextActionAt      time.Duration
	pendingPool       *PoolingActions
	pendingPoolWeaves *PoolingActions
	readyToShift      bool
	lastShiftAt       time.Duration
	itemProcAuras     []*core.StatBuffAura
}

func (rotation *FeralDruidRotation) Finalize(_ *core.APLRotation)                     {}
func (rotation *FeralDruidRotation) GetAPLValues() []core.APLValue                    { return nil }
func (rotation *FeralDruidRotation) GetInnerActions() []*core.APLAction               { return nil }
func (rotation *FeralDruidRotation) GetNextAction(_ *core.Simulation) *core.APLAction { return nil }
func (rotation *FeralDruidRotation) PostFinalize(_ *core.APLRotation)                 {}

func (rotation *FeralDruidRotation) IsReady(sim *core.Simulation) bool {
	return sim.CurrentTime > rotation.lastActionAt
}

func (rotation *FeralDruidRotation) Reset(_ *core.Simulation) {
	rotation.lastActionAt = -core.NeverExpires
	rotation.nextActionAt = -core.NeverExpires
	rotation.readyToShift = false
	rotation.lastShiftAt = -core.NeverExpires
}

func (rotation *FeralDruidRotation) Execute(sim *core.Simulation) {
	rotation.lastActionAt = sim.CurrentTime
	cat := rotation.agent

	// If a melee swing resulted in an Omen proc, then schedule the next
	// player decision based on latency.
	ccRefreshTime := cat.ClearcastingAura.ExpiresAt() - cat.ClearcastingAura.Duration

	if ccRefreshTime >= sim.CurrentTime-cat.ReactionTime {
		rotation.WaitUntil(sim, max(cat.NextGCDAt(), ccRefreshTime+cat.ReactionTime))
	}

	// Keep up Sunder debuff if not provided externally. Do this here since
	// FF can be cast while moving.
	for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
		if cat.ShouldFaerieFire(sim, aoeTarget) {
			cat.FaerieFire.CastOrQueue(sim, aoeTarget)
		}
	}

	// Off-GCD Maul check
	if cat.BearFormAura.IsActive() && !cat.ClearcastingAura.IsActive() && !cat.DreamOfCenariusAura.IsActive() && cat.Maul.CanCast(sim, cat.CurrentTarget) {
		cat.Maul.Cast(sim, cat.CurrentTarget)
	}

	// Handle movement before any rotation logic
	if cat.Moving || (cat.Hardcast.Expires > sim.CurrentTime) {
		return
	}

	if cat.DistanceFromTarget > core.MaxMeleeRange {
		// Try leaping if no boots
		if !cat.GetAura("Nitro Boosts").IsActive() && cat.Talents.WildCharge && cat.CatCharge.CanCast(sim, cat.CurrentTarget) {
			cat.CatCharge.Cast(sim, cat.CurrentTarget)
		} else {
			if sim.Log != nil {
				cat.Log(sim, "Out of melee range (%.6fy) and cannot charge or teleport, initiating manual run-in...", cat.DistanceFromTarget)
			}

			cat.MoveTo(core.MaxMeleeRange-1, sim) // movement aura is discretized in 1 yard intervals, so need to overshoot to guarantee melee range
			return
		}
	}

	if !cat.GCD.IsReady(sim) {
		cat.WaitUntil(sim, cat.NextGCDAt())
		return
	}

	rotation.TryTigersFury(sim)
	rotation.TryBerserk(sim)

	if rotation.UseHealingTouch && rotation.UseNs && cat.NaturesSwiftness.IsReady(sim) && (cat.ComboPoints() == 5) && !cat.DreamOfCenariusAura.IsActive() && !cat.PredatorySwiftnessAura.IsActive() {
		cat.NaturesSwiftness.Cast(sim, &cat.Unit)
	}

	if sim.CurrentTime < rotation.nextActionAt {
		cat.WaitUntil(sim, rotation.nextActionAt)
	} else if rotation.readyToShift {
		rotation.ShiftBearCat(sim)
	} else {
		rotation.PickGCDAction(sim, rotation.RotationType != proto.FeralDruid_Rotation_SingleTarget)

		if !cat.GCD.IsReady(sim) && rotation.WrathWeave && cat.HeartOfTheWild.IsReady(sim) && cat.BerserkCatAura.IsActive() && (cat.BerserkCatAura.RemainingDuration(sim) < core.GCDMin*5) {
			cat.HeartOfTheWild.Cast(sim, nil)
			cat.UpdateMajorCooldowns()
		}

		if !cat.GCD.IsReady(sim) && cat.ItemSwap.IsEnabled() && rotation.shouldWrathWeave(sim) && cat.CatFormAura.IsActive() {
			cat.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
		}
	}
}

func (rotation *FeralDruidRotation) PickGCDAction(sim *core.Simulation, isAoe bool) {
	// Store state variables for re-use
	cat := rotation.agent
	curEnergy := cat.CurrentEnergy()
	curCp := cat.ComboPoints()
	regenRate := cat.EnergyRegenPerSecond()
	isExecutePhase := sim.IsExecutePhase25()
	isClearcast := cat.ClearcastingAura.IsActive()
	isBerserk := cat.BerserkCatAura.IsActive()
	anyBleedActive := cat.AssumeBleedActive || (cat.BleedsActive[cat.CurrentTarget] > 0)
	fightDur := sim.GetRemainingDuration()
	ripDot := cat.Rip.CurDot()
	ripDur := ripDot.RemainingDuration(sim)
	roarBuff := cat.SavageRoarBuff
	roarDur := roarBuff.RemainingDuration(sim)
	rakeDot := cat.Rake.CurDot()
	rakeDur := rakeDot.RemainingDuration(sim)
	thrashDot := cat.ThrashCat.CurDot()

	// Rip logic
	ripRefreshTime := cat.calcBleedRefreshTime(sim, cat.Rip, ripDot, isExecutePhase, true)
	ripNow := (curCp >= 5) && (!ripDot.IsActive() || ((sim.CurrentTime > ripRefreshTime) && (!isExecutePhase || (cat.Rip.NewSnapshotPower > cat.Rip.CurrentSnapshotPower+0.001))) || (!isExecutePhase && (roarDur < rotation.RipLeeway) && (ripDot.ExpiresAt() < roarBuff.ExpiresAt()+rotation.RipLeeway))) && (fightDur > ripDot.BaseTickLength) && (!isClearcast || !anyBleedActive || cat.DreamOfCenariusAura.IsActive()) && !cat.shouldDelayBleedRefreshForTf(sim, ripDot, true)

	// Roar logic
	newRoarDur := cat.SavageRoarDurationTable[curCp]
	roarRefreshTime := cat.calcRoarRefreshTime(sim, ripRefreshTime, rotation.RipLeeway, rotation.MinRoarOffset)
	roarNow := (newRoarDur > 0) && (!roarBuff.IsActive() || (sim.CurrentTime > roarRefreshTime))

	// Bite logic
	biteTime := core.TernaryDuration(isBerserk, rotation.BerserkBiteTime, rotation.BiteTime)
	shouldBite := (curCp >= 5) && ripDot.IsActive() && roarBuff.IsActive() && ((rotation.UseBite && (min(ripRefreshTime, roarRefreshTime)-sim.CurrentTime >= biteTime)) || isExecutePhase) && !isClearcast
	shouldEmergencyBite := isExecutePhase && ripDot.IsActive() && (ripDur < ripDot.BaseTickLength) && (curCp >= 1)
	biteNow := shouldBite || shouldEmergencyBite

	// Rake logic
	rakeRefreshTime := cat.calcBleedRefreshTime(sim, cat.Rake, rakeDot, isExecutePhase, false)
	rakeNow := (!rakeDot.IsActive() || (sim.CurrentTime > rakeRefreshTime)) && (fightDur > rakeDot.BaseTickLength) && (!isClearcast || !rakeDot.IsActive() || (rakeDur < time.Second) || cat.DreamOfCenariusAura.IsActive()) && !cat.shouldDelayBleedRefreshForTf(sim, rakeDot, false) && roarBuff.IsActive()
	rakeNow, rakeTarget := rotation.shouldAoeRake(sim, roarNow, rakeNow)

	// Pooling calcs
	ripRefreshPending := ripDot.IsActive() && (ripDur < fightDur-ripDot.BaseTickLength) && (curCp >= core.TernaryInt32(isExecutePhase, 1, 5)) && !isAoe
	rakeRefreshPending := rakeDot.IsActive() && (rakeDur < fightDur-rakeDot.BaseTickLength) && !isAoe
	roarRefreshPending := roarBuff.IsActive() && (roarDur < fightDur-cat.ReactionTime) && (newRoarDur > 0)
	rotation.pendingPool.reset()
	rotation.pendingPoolWeaves.reset()

	if ripRefreshPending && (sim.CurrentTime < ripRefreshTime) {
		ripRefreshCost := core.Ternary(isExecutePhase, cat.FerociousBite.DefaultCast.Cost, cat.Rip.DefaultCast.Cost)
		rotation.pendingPool.addAction(ripRefreshTime, ripRefreshCost)
		rotation.pendingPoolWeaves.addAction(ripRefreshTime, ripRefreshCost)
	}

	if rakeRefreshPending && (sim.CurrentTime < rakeRefreshTime) {
		rotation.pendingPool.addAction(rakeRefreshTime, cat.Rake.DefaultCast.Cost)
		rotation.pendingPoolWeaves.addAction(rakeRefreshTime, cat.Rake.DefaultCast.Cost)
	}

	if roarRefreshPending && (sim.CurrentTime < roarRefreshTime) {
		rotation.pendingPool.addAction(roarRefreshTime, cat.SavageRoar.DefaultCast.Cost)
	}

	if rotation.UseHealingTouch && cat.PredatorySwiftnessAura.IsActive() && (cat.PredatorySwiftnessAura.RemainingDuration(sim) > cat.ReactionTime*2) {
		rotation.pendingPool.addAction(cat.PredatorySwiftnessAura.ExpiresAt()-cat.ReactionTime*2, 0)
		rotation.pendingPoolWeaves.addAction(cat.PredatorySwiftnessAura.ExpiresAt()-cat.ReactionTime*2, 0)
	}

	rotation.pendingPool.sort()
	rotation.pendingPoolWeaves.sort()
	floatingEnergy := rotation.pendingPool.calcFloatingEnergy(cat, sim)
	excessE := curEnergy - floatingEnergy

	// Check bear-weaving conditions.
	furorCap := 100.0 - 1.5*regenRate
	bearWeaveNow := rotation.BearWeave && cat.canBearWeave(sim, furorCap, regenRate, curEnergy, excessE, rotation.pendingPoolWeaves)

	// Check Wrath-weaving conditions.
	wrathWeaveNow := rotation.shouldWrathWeave(sim)

	// Main decision tree starts here.
	var timeToNextAction time.Duration

	if cat.BearFormAura.IsActive() {
		if rotation.shouldTerminateBearWeave(sim, isClearcast, curEnergy, furorCap, regenRate, rotation.pendingPoolWeaves) {
			rotation.readyToShift = true
		} else if cat.ThrashBear.CanCast(sim, cat.CurrentTarget) {
			cat.ThrashBear.Cast(sim, cat.CurrentTarget)
		} else if isAoe && cat.SwipeBear.CanCast(sim, cat.CurrentTarget) {
			cat.SwipeBear.Cast(sim, cat.CurrentTarget)
		} else if cat.MangleBear.CanCast(sim, cat.CurrentTarget) {
			cat.MangleBear.Cast(sim, cat.CurrentTarget)
		} else if cat.Lacerate.CanCast(sim, cat.CurrentTarget) {
			cat.Lacerate.Cast(sim, cat.CurrentTarget)
		} else {
			rotation.readyToShift = true
		}

		// Last second Maul check if we are about to shift back.
		if rotation.readyToShift && !isClearcast && cat.Maul.CanCast(sim, cat.CurrentTarget) {
			cat.Maul.Cast(sim, cat.CurrentTarget)
		}

		if !rotation.readyToShift {
			timeToNextAction = cat.ReactionTime
		}
	} else if !cat.CatFormAura.IsActive() {
		if !cat.HeartOfTheWildAura.IsActive() || (cat.HeartOfTheWildAura.RemainingDuration(sim) <= cat.Wrath.DefaultCast.CastTime) || !ripDot.IsActive() || (ripRefreshPending && (ripDot.ExpiresAt() <= sim.CurrentTime+cat.Wrath.DefaultCast.CastTime+core.GCDDefault)) || (isAoe && (curEnergy+cat.Wrath.DefaultCast.CastTime.Seconds()*regenRate > furorCap)) {
			rotation.readyToShift = true
		} else {
			cat.Wrath.Cast(sim, cat.CurrentTarget)
			return
		}
	} else if roarNow {
		if cat.SavageRoar.CanCast(sim, cat.CurrentTarget) {
			cat.SavageRoar.Cast(sim, nil)
			return
		}

		timeToNextAction = core.DurationFromSeconds((cat.CurrentSavageRoarCost() - curEnergy) / regenRate)
	} else if rotation.UseHealingTouch && (cat.PredatorySwiftnessAura.IsActive() || cat.NaturesSwiftness.RelatedSelfBuff.IsActive()) && ((curCp >= 4) || (cat.PredatorySwiftnessAura.RemainingDuration(sim) <= time.Second)) && (!isBerserk || (curCp == 5)) {
		cat.HealingTouch.Cast(sim, &cat.Unit)
		return
	} else if ripNow {
		if cat.Rip.CanCast(sim, cat.CurrentTarget) {
			cat.Rip.Cast(sim, cat.CurrentTarget)
			return
		}

		timeToNextAction = core.DurationFromSeconds((cat.CurrentRipCost() - curEnergy) / regenRate)
	} else if biteNow && ((curEnergy >= cat.CurrentFerociousBiteCost()) || !bearWeaveNow) {
		if cat.FerociousBite.CanCast(sim, cat.CurrentTarget) {
			cat.FerociousBite.Cast(sim, cat.CurrentTarget)
			return
		}

		timeToNextAction = core.DurationFromSeconds((cat.CurrentFerociousBiteCost() - curEnergy) / regenRate)
	} else if rakeNow && (!isAoe || !bearWeaveNow || (curEnergy >= cat.CurrentRakeCost())) {
		if cat.Rake.CanCast(sim, rakeTarget) {
			cat.Rake.Cast(sim, rakeTarget)
			return
		}

		timeToNextAction = core.DurationFromSeconds((cat.CurrentRakeCost() - curEnergy) / regenRate)
	} else if wrathWeaveNow {
		cat.Wrath.Cast(sim, cat.CurrentTarget)
		return
	} else if bearWeaveNow {
		rotation.readyToShift = true
	} else if (isClearcast || isBerserk || isAoe) && (!thrashDot.IsActive() || (thrashDot.RemainingDuration(sim) < thrashDot.BaseTickLength) || (cat.DreamOfCenariusAura.IsActive() && ((curCp < 3) || (curCp == 5)))) {
		cat.ThrashCat.Cast(sim, cat.CurrentTarget)
		return
	} else if isClearcast || !ripRefreshPending || !cat.tempSnapshotAura.IsActive() || (ripRefreshTime+cat.ReactionTime-sim.CurrentTime > core.GCDMin) {
		fillerSpell := core.Ternary(rotation.ForceMangleFiller || ((curCp < 5) && !isClearcast && !isBerserk), cat.MangleCat, cat.Shred)

		// Force Shred usage in opener.
		if !rotation.ForceMangleFiller && cat.Berserk.IsReady(sim) && (sim.CurrentTime < cat.Berserk.CD.Duration) {
			fillerSpell = cat.Shred
		}

		if cat.IncarnationAura.IsActive() || cat.StampedeAura.IsActive() {
			fillerSpell = cat.Ravage
		}

		fillerDpc := fillerSpell.ExpectedInitialDamage(sim, cat.CurrentTarget)
		rakeDpc := cat.Rake.ExpectedInitialDamage(sim, cat.CurrentTarget)

		if ((fillerDpc < rakeDpc) || (!isBerserk && !isClearcast && (fillerDpc/fillerSpell.DefaultCast.Cost < rakeDpc/cat.Rake.DefaultCast.Cost))) && (cat.Rake.NewSnapshotPower > cat.Rake.CurrentSnapshotPower-0.001) && (!ripDot.IsActive() || (ripDur >= rotation.RipLeeway) || (ripDot.BaseTickCount == cat.RipMaxNumTicks)) {
			fillerSpell = cat.Rake
		}

		if isAoe {
			fillerSpell = cat.SwipeCat
		}

		// Force filler on Clearcasts or when about to Energy cap.
		if isClearcast || (curEnergy > cat.MaximumEnergy()-regenRate*cat.ReactionTime.Seconds()) {
			fillerSpell.Cast(sim, cat.CurrentTarget)
			return
		}

		fillerCost := fillerSpell.Cost.GetCurrentCost()
		energyForCalc := core.TernaryFloat64(isBerserk, curEnergy, excessE)

		if energyForCalc >= fillerCost {
			fillerSpell.Cast(sim, cat.CurrentTarget)
			return
		}

		timeToNextAction = core.DurationFromSeconds((fillerCost - energyForCalc) / regenRate)
	}

	nextActionAt := sim.CurrentTime + timeToNextAction
	isPooling, nextRefresh := rotation.pendingPool.nextRefreshTime()

	if isPooling {
		nextActionAt = min(nextActionAt, nextRefresh)
	}

	rotation.ProcessNextPlannedAction(sim, nextActionAt)
}

func (action *FeralDruidRotation) ReResolveVariableRefs(*core.APLRotation, map[string]*proto.APLValue) {
}
