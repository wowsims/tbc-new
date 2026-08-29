package core

import (
	"time"
)

const FearAuraTag = "Fear"

func (unit *Unit) RegisterFearAura(label string, actionID ActionID, duration time.Duration) *Aura {
	return unit.RegisterAura(Aura{
		Label:    label,
		ActionID: actionID,
		Tag:      FearAuraTag,
		Duration: duration,

		OnGain: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.Incapacitated = true
			aura.Unit.Interrupt(sim)
			aura.Unit.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+duration)
			aura.Unit.AutoAttacks.StopRangedUntil(sim, sim.CurrentTime+duration)
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			aura.Unit.PseudoStats.Incapacitated = false

			// Restart the swing timers, which matters when the Fear was broken
			// early. On a natural expiry these are no-ops.
			aura.Unit.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime)
			aura.Unit.AutoAttacks.StopRangedUntil(sim, sim.CurrentTime)
		},
	})
}

func ApplyFear(sim *Simulation, fearAura *Aura) bool {
	if fearAura == nil || fearAura.Unit.PseudoStats.FearImmune {
		return false
	}

	fearAura.Activate(sim)
	return true
}

func (auras AuraArray) ApplyFearToAllPlayers(sim *Simulation) {
	if auras == nil {
		return
	}

	for _, playerUnit := range sim.Raid.AllPlayerUnits {
		if playerUnit.IsEnabled() {
			ApplyFear(sim, auras.Get(playerUnit))
		}
	}
}

func (aura *Aura) AttachFearImmunity() *Aura {
	if aura == nil {
		return nil
	}

	return aura.ApplyOnGain(func(aura *Aura, sim *Simulation) {
		aura.Unit.PseudoStats.FearImmune = true
		aura.Unit.BreakFear(sim)
	}).ApplyOnExpire(func(aura *Aura, _ *Simulation) {
		aura.Unit.PseudoStats.FearImmune = false
	})
}

func (unit *Unit) BreakFear(sim *Simulation) {
	for _, aura := range unit.GetAurasWithTag(FearAuraTag) {
		if aura.IsActive() {
			aura.Deactivate(sim)
		}
	}
}
