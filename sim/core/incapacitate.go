package core

import (
	"time"
)

const FearAuraTag = "Fear"
const StunAuraTag = "Stun"

// incapacitateKind holds the parts that differ between crowd control effects.
// Everything else (aura skeleton, immunity handling, breaking) is shared.
type incapacitateKind struct {
	tag string

	// setActive toggles the PseudoStats fields owned by this kind.
	setActive func(unit *Unit, active bool)

	isImmune  func(unit *Unit) bool
	setImmune func(unit *Unit, immune bool)
}

func (kind *incapacitateKind) registerAura(unit *Unit, label string, actionID ActionID, duration time.Duration) *Aura {
	return unit.RegisterAura(Aura{
		Label:    label,
		ActionID: actionID,
		Tag:      kind.tag,
		Duration: duration,

		OnGain: func(aura *Aura, sim *Simulation) {
			kind.setActive(aura.Unit, true)
			aura.Unit.Interrupt(sim)
			aura.Unit.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+duration)
			aura.Unit.AutoAttacks.StopRangedUntil(sim, sim.CurrentTime+duration)
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			kind.setActive(aura.Unit, false)

			// Restart the swing timers, which matters when the effect was broken
			// early. On a natural expiry these are no-ops.
			aura.Unit.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime)
			aura.Unit.AutoAttacks.StopRangedUntil(sim, sim.CurrentTime)
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

func (kind *incapacitateKind) applyToAllPlayers(sim *Simulation, auras AuraArray) {
	if auras == nil {
		return
	}

	for _, playerUnit := range sim.Raid.AllPlayerUnits {
		if playerUnit.IsEnabled() {
			kind.apply(sim, auras.Get(playerUnit))
		}
	}
}

func (kind *incapacitateKind) attachImmunity(aura *Aura) *Aura {
	if aura == nil {
		return nil
	}

	return aura.ApplyOnGain(func(aura *Aura, sim *Simulation) {
		kind.setImmune(aura.Unit, true)
		kind.breakAll(aura.Unit, sim)
	}).ApplyOnExpire(func(aura *Aura, _ *Simulation) {
		kind.setImmune(aura.Unit, false)
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
	tag: FearAuraTag,
	setActive: func(unit *Unit, active bool) {
		unit.PseudoStats.Incapacitated = active
	},
	isImmune:  func(unit *Unit) bool { return unit.PseudoStats.FearImmune },
	setImmune: func(unit *Unit, immune bool) { unit.PseudoStats.FearImmune = immune },
}

func (unit *Unit) RegisterFearAura(label string, actionID ActionID, duration time.Duration) *Aura {
	return fearKind.registerAura(unit, label, actionID, duration)
}

func ApplyFear(sim *Simulation, fearAura *Aura) bool {
	return fearKind.apply(sim, fearAura)
}

func (auras AuraArray) ApplyFearToAllPlayers(sim *Simulation) {
	fearKind.applyToAllPlayers(sim, auras)
}

func (aura *Aura) AttachFearImmunity() *Aura {
	return fearKind.attachImmunity(aura)
}

func (unit *Unit) BreakFear(sim *Simulation) {
	fearKind.breakAll(unit, sim)
}

var stunKind = &incapacitateKind{
	tag: StunAuraTag,
	setActive: func(unit *Unit, active bool) {
		unit.PseudoStats.Incapacitated = active
		unit.PseudoStats.Stunned = active
	},
	isImmune:  func(unit *Unit) bool { return unit.PseudoStats.StunImmune },
	setImmune: func(unit *Unit, immune bool) { unit.PseudoStats.StunImmune = immune },
}

func (unit *Unit) RegisterStunAura(label string, actionID ActionID, duration time.Duration) *Aura {
	return stunKind.registerAura(unit, label, actionID, duration)
}

func ApplyStun(sim *Simulation, stunAura *Aura) bool {
	return stunKind.apply(sim, stunAura)
}

func (auras AuraArray) ApplyStunToAllPlayers(sim *Simulation) {
	stunKind.applyToAllPlayers(sim, auras)
}

func (aura *Aura) AttachStunImmunity() *Aura {
	return stunKind.attachImmunity(aura)
}

func (unit *Unit) BreakStun(sim *Simulation) {
	stunKind.breakAll(unit, sim)
}
