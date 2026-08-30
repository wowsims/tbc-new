package core

import (
	"time"
)

const FearAuraTag = "Fear"
const StunAuraTag = "Stun"

// Tag on the tank hardcast aura, which suppresses avoidance the same way a stun
// does and so shares the Stunned flag with the stun kind.
const ReducedAvoidanceAuraTag = "Reduced Avoidance"

// incapacitateKind holds the parts that differ between crowd control effects.
// Everything else (aura skeleton, immunity handling, breaking) is shared.
type incapacitateKind struct {
	tag string

	// setImmune writes the PseudoStats flag this kind's immunity controls.
	setImmune func(unit *Unit, immune bool)

	isImmune func(unit *Unit) bool
}

// Several effects can own the same PseudoStats flag at once: two fears, a fear
// overlapping a stun, or a stun landing while a tank is hardcasting. Writing the
// flags absolutely lets whichever effect ends first clear a flag the others
// still need, so they are derived from whatever is active instead.
func (unit *Unit) refreshIncapacitateState() {
	feared := unit.HasActiveAuraWithTag(FearAuraTag)
	stunned := unit.HasActiveAuraWithTag(StunAuraTag)

	unit.PseudoStats.Incapacitated = feared || stunned
	unit.PseudoStats.Stunned = stunned || unit.HasActiveAuraWithTag(ReducedAvoidanceAuraTag)
}

// Immunity is granted by arbitrary auras rather than by a shared tag, so each
// kind tracks the ones attached to it and derives the flag the same way.
func (kind *incapacitateKind) refreshImmunity(unit *Unit) {
	for _, aura := range unit.ccImmunityAuras[kind.tag] {
		if aura.IsActive() {
			kind.setImmune(unit, true)
			return
		}
	}

	kind.setImmune(unit, false)
}

func (kind *incapacitateKind) registerAura(unit *Unit, label string, actionID ActionID, duration time.Duration) *Aura {
	return unit.RegisterAura(Aura{
		Label:    label,
		ActionID: actionID,
		Tag:      kind.tag,
		Duration: duration,

		OnGain: func(aura *Aura, sim *Simulation) {
			// Interrupt first: cancelling a hardcast drops the tank avoidance
			// aura, which has to settle before the flags are recomputed.
			aura.Unit.Interrupt(sim)
			aura.Unit.refreshIncapacitateState()

			// The swing timer keeps running while incapacitated
			aura.Unit.AutoAttacks.PauseMeleeBy(sim, duration)
			aura.Unit.AutoAttacks.DelayRangedUntil(sim, sim.CurrentTime+duration)
		},

		OnExpire: func(aura *Aura, _ *Simulation) {
			aura.Unit.refreshIncapacitateState()
		},
	})
}

func (kind *incapacitateKind) apply(sim *Simulation, aura *Aura) bool {
	if aura == nil || kind.isImmune(aura.Unit) {
		return false
	}

	aura.Activate(sim)
	return true
}

func (kind *incapacitateKind) applyToAllUnits(sim *Simulation, auras AuraArray) {
	if auras == nil {
		return
	}

	for _, unit := range sim.Raid.AllUnits {
		if unit.IsEnabled() {
			kind.apply(sim, auras.Get(unit))
		}
	}
}

func (kind *incapacitateKind) attachImmunity(aura *Aura) *Aura {
	if aura == nil {
		return nil
	}

	if aura.Unit.ccImmunityAuras == nil {
		aura.Unit.ccImmunityAuras = map[string][]*Aura{}
	}
	aura.Unit.ccImmunityAuras[kind.tag] = append(aura.Unit.ccImmunityAuras[kind.tag], aura)

	return aura.ApplyOnGain(func(aura *Aura, sim *Simulation) {
		kind.refreshImmunity(aura.Unit)
		kind.breakAll(aura.Unit, sim)
	}).ApplyOnExpire(func(aura *Aura, _ *Simulation) {
		// Another immunity aura may still be running, so re-derive rather than
		// clearing outright.
		kind.refreshImmunity(aura.Unit)
	})
}

func (kind *incapacitateKind) breakAll(unit *Unit, sim *Simulation) {
	for _, aura := range unit.GetAurasWithTag(kind.tag) {
		if aura.IsActive() {
			aura.Deactivate(sim)
		}
	}
}

var fearKind = &incapacitateKind{
	tag:       FearAuraTag,
	isImmune:  func(unit *Unit) bool { return unit.PseudoStats.FearImmune },
	setImmune: func(unit *Unit, immune bool) { unit.PseudoStats.FearImmune = immune },
}

func (unit *Unit) RegisterFearAura(label string, actionID ActionID, duration time.Duration) *Aura {
	return fearKind.registerAura(unit, label, actionID, duration)
}

func ApplyFear(sim *Simulation, fearAura *Aura) bool {
	return fearKind.apply(sim, fearAura)
}

func (auras AuraArray) ApplyFearToAllUnits(sim *Simulation) {
	fearKind.applyToAllUnits(sim, auras)
}

func (aura *Aura) AttachFearImmunity() *Aura {
	return fearKind.attachImmunity(aura)
}

func (unit *Unit) BreakFear(sim *Simulation) {
	fearKind.breakAll(unit, sim)
}

var stunKind = &incapacitateKind{
	tag:       StunAuraTag,
	isImmune:  func(unit *Unit) bool { return unit.PseudoStats.StunImmune },
	setImmune: func(unit *Unit, immune bool) { unit.PseudoStats.StunImmune = immune },
}

func (unit *Unit) RegisterStunAura(label string, actionID ActionID, duration time.Duration) *Aura {
	return stunKind.registerAura(unit, label, actionID, duration)
}

func ApplyStun(sim *Simulation, stunAura *Aura) bool {
	return stunKind.apply(sim, stunAura)
}

func (auras AuraArray) ApplyStunToAllUnits(sim *Simulation) {
	stunKind.applyToAllUnits(sim, auras)
}

func (aura *Aura) AttachStunImmunity() *Aura {
	return stunKind.attachImmunity(aura)
}

func (unit *Unit) BreakStun(sim *Simulation) {
	stunKind.breakAll(unit, sim)
}
