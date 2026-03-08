package core

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
)

type OnFocusGain func(*Simulation, float64)

type focusBar struct {
	unit *Unit

	maxFocus          float64
	currentFocus      float64
	focusRegenPerTick float64
	focusTickInterval float64
	focusTickDuration time.Duration
	nextFocusTick     time.Duration

	regenMetrics       *ResourceMetrics
	focusRefundMetrics *ResourceMetrics
}

func (unit *Unit) EnableFocusBar(focusRegenMultiplier float64) {
	unit.SetCurrentPowerBar(FocusBar)

	unit.focusBar = focusBar{
		unit:               unit,
		maxFocus:           100.0,
		focusRegenPerTick:  25.0 * focusRegenMultiplier,
		focusTickInterval:  5,
		focusTickDuration:  time.Second * 5,
		regenMetrics:       unit.NewFocusMetrics(ActionID{OtherID: proto.OtherAction_OtherActionFocusRegen}),
		focusRefundMetrics: unit.NewFocusMetrics(ActionID{OtherID: proto.OtherAction_OtherActionRefund}),
	}
}

func (unit *Unit) HasFocusBar() bool {
	return unit.focusBar.unit != nil
}

func (fb *focusBar) AddFocus(sim *Simulation, amount float64, metrics *ResourceMetrics) {
	if amount < 0 {
		panic("Trying to add negative focus!")
	}
	newFocus := min(fb.currentFocus+amount, fb.maxFocus)
	if (fb.currentFocus != newFocus) && sim.Log != nil {
		fb.unit.Log(sim, "Gained %0.3f focus from %s (%0.3f --> %0.3f) of %0.0f total.", amount, metrics.ActionID, fb.currentFocus, newFocus, fb.maxFocus)
	}

	fb.currentFocus = newFocus
}

func (fb *focusBar) SpendFocus(sim *Simulation, amount float64, metrics *ResourceMetrics) {
	if amount < 0 {
		panic("Trying to spend negative focus!")
	}

	newFocus := fb.currentFocus - amount
	metrics.AddEvent(-amount, -amount)

	if sim.Log != nil {
		fb.unit.Log(sim, "Spent %0.3f focus from %s (%0.3f --> %0.3f) of %0.0f total.", amount, metrics.ActionID, fb.currentFocus, newFocus, fb.maxFocus)
	}

	fb.currentFocus = newFocus
}

func (fb *focusBar) IsTicking(sim *Simulation) bool {
	return (fb.nextFocusTick != 0) && (sim.CurrentTime <= fb.nextFocusTick) && (fb.nextFocusTick-sim.CurrentTime <= fb.focusTickDuration)
}

func (fb *focusBar) RunTask(sim *Simulation) time.Duration {
	if sim.CurrentTime < fb.nextFocusTick {
		return fb.nextFocusTick
	}
	fb.AddFocus(sim, fb.focusRegenPerTick, fb.regenMetrics)
	fb.nextFocusTick = sim.CurrentTime + fb.focusTickDuration
	return fb.nextFocusTick
}

func (fb *focusBar) reset(sim *Simulation) {
	if fb.unit == nil {
		return
	}

	fb.currentFocus = fb.maxFocus

	if fb.unit.Type != PetUnit {
		fb.enable(sim, sim.Environment.PrepullStartTime())
	}
}

func (fb *focusBar) enable(sim *Simulation, startAt time.Duration) {
	sim.AddTask(fb)
	fb.nextFocusTick = startAt + time.Duration(sim.RandomFloat("Focus Tick")*float64(fb.focusTickDuration))
	sim.RescheduleTask(fb.nextFocusTick)
}

func (fb *focusBar) disable(sim *Simulation) {
	fb.nextFocusTick = NeverExpires
	sim.RemoveTask(fb)
}

type FocusCostOptions struct {
	Cost int32

	Refund        float64
	RefundMetrics *ResourceMetrics // Optional, will default to unit.FocusRefundMetrics if not supplied.
}

type FocusCost struct {
	Refund          float64
	RefundMetrics   *ResourceMetrics
	ResourceMetrics *ResourceMetrics
}

func newFocusCost(spell *Spell, options FocusCostOptions) *SpellCost {
	if options.Refund > 0 && options.RefundMetrics == nil {
		options.RefundMetrics = spell.Unit.focusRefundMetrics
	}

	return &SpellCost{
		spell:           spell,
		BaseCost:        options.Cost,
		PercentModifier: 1,
		ResourceCostImpl: &FocusCost{
			Refund:          options.Refund,
			RefundMetrics:   options.RefundMetrics,
			ResourceMetrics: spell.Unit.NewFocusMetrics(spell.ActionID),
		},
	}
}

func (ec *FocusCost) MeetsRequirement(_ *Simulation, spell *Spell) bool {
	spell.CurCast.Cost = spell.Cost.GetCurrentCost()
	return spell.Unit.currentFocus >= spell.CurCast.Cost
}

func (ec *FocusCost) CostFailureReason(_ *Simulation, spell *Spell) string {
	return fmt.Sprintf("not enough focus (Current Focus = %0.03f, Focus Cost = %0.03f)", spell.Unit.currentFocus, spell.CurCast.Cost)
}
func (ec *FocusCost) SpendCost(sim *Simulation, spell *Spell) {
	spell.Unit.SpendFocus(sim, spell.CurCast.Cost, ec.ResourceMetrics)
}
func (ec *FocusCost) IssueRefund(sim *Simulation, spell *Spell) {
	if ec.Refund > 0 && spell.CurCast.Cost > 0 {
		spell.Unit.AddFocus(sim, ec.Refund*spell.CurCast.Cost, ec.RefundMetrics)
	}
}
