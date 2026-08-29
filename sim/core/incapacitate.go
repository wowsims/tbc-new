package core

// Incapacitate effects take control away from a unit. Fear is the only form
// modelled so far; Sap, Incapacitate and friends belong here too.

import (
	"time"
)

// FearAuraTag marks auras that represent a Fear effect, so that immunity effects
// such as Berserker Rage can find and break them.
const FearAuraTag = "Fear"

// RegisterFearAura registers a Fear effect on the unit. While it is active the
// unit is uncontrollable: anything in progress is interrupted, and no swings,
// rotation actions or casts happen until the Fear ends. Spells flagged with
// SpellFlagCastableWhileIncapacitated remain usable, which is how effects like
// Berserker Rage break out.
//
// Apply it with ApplyFear (or AuraArray.ApplyFearToAllPlayers) so that Fear
// immunity is respected, and remove it early with Unit.BreakFear.
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

// ApplyFear activates a Fear aura unless its unit is immune to Fear. Returns
// whether the Fear landed.
func ApplyFear(sim *Simulation, fearAura *Aura) bool {
	if fearAura == nil || fearAura.Unit.PseudoStats.FearImmune {
		return false
	}

	fearAura.Activate(sim)
	return true
}

// ApplyFearToAllPlayers applies the Fear auras in the array to every enabled
// player that is not immune to Fear.
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

// BreakFear removes any active Fear effect from the unit, returning control.
func (unit *Unit) BreakFear(sim *Simulation) {
	for _, aura := range unit.GetAurasWithTag(FearAuraTag) {
		if aura.IsActive() {
			aura.Deactivate(sim)
		}
	}
}
