package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

type Stance uint8

const (
	StanceNone          = 0
	BattleStance Stance = 1 << iota
	DefensiveStance
	BerserkerStance
)

const stanceEffectCategory = "Stance"

func (warrior *Warrior) StanceMatches(other Stance) bool {
	return (warrior.Stance & other) != 0
}

func (warrior *Warrior) makeStanceSpell(stance Stance, aura *core.Aura, stanceCD *core.Timer) *core.Spell {
	maxRetainedRage := 10.0 + 5*float64(warrior.Talents.TacticalMastery)
	actionID := aura.ActionID
	rageMetrics := warrior.NewRageMetrics(actionID)

	return warrior.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    stanceCD,
				Duration: time.Millisecond * 1500,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warrior.Stance != stance
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			if warrior.WarriorInputs.StanceSnapshot {
				// Delayed, so same-GCD casts are affected by the current aura.
				//  Alternatively, those casts could just (artificially) happen before the stance change.
				pa := sim.GetConsumedPendingActionFromPool()
				pa.NextActionAt = sim.CurrentTime + 10*time.Millisecond
				pa.OnAction = aura.Activate
				sim.AddPendingAction(pa)
			} else {
				aura.Activate(sim)
			}

			if warrior.CurrentRage() > maxRetainedRage {
				warrior.SpendRage(sim, warrior.CurrentRage()-maxRetainedRage, rageMetrics)
			}

			warrior.Stance = stance
		},
	})
}

func (warrior *Warrior) registerBattleStanceAura() {
	actionID := core.ActionID{SpellID: 2457}

	warrior.BattleStanceAura = warrior.GetOrRegisterAura(core.Aura{
		Label:    "Battle Stance",
		ActionID: actionID,
		Duration: core.NeverExpires,
	}).AttachMultiplicativePseudoStatBuff(&warrior.PseudoStats.ThreatMultiplier, 0.8)

	warrior.BattleStanceAura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})
}

func (warrior *Warrior) registerDefensiveStanceAura() {
	actionID := core.ActionID{SpellID: 71}
	threatMultiplier := 1.3 * (1 + 0.05*float64(warrior.Talents.Defiance))
	impDefStanceMultiplier := 1 - 0.02*float64(warrior.Talents.ImprovedDefensiveStance)

	warrior.DefensiveStanceAura = warrior.GetOrRegisterAura(core.Aura{
		Label:    "Defensive Stance",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexArcane] *= impDefStanceMultiplier
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire] *= impDefStanceMultiplier
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFrost] *= impDefStanceMultiplier
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexHoly] *= impDefStanceMultiplier
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexNature] *= impDefStanceMultiplier
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow] *= impDefStanceMultiplier
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexArcane] /= impDefStanceMultiplier
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire] /= impDefStanceMultiplier
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFrost] /= impDefStanceMultiplier
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexHoly] /= impDefStanceMultiplier
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexNature] /= impDefStanceMultiplier
			warrior.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow] /= impDefStanceMultiplier
		},
	}).AttachMultiplicativePseudoStatBuff(
		&warrior.PseudoStats.ThreatMultiplier, threatMultiplier,
	).AttachMultiplicativePseudoStatBuff(
		&warrior.PseudoStats.DamageTakenMultiplier, 0.9,
	).AttachMultiplicativePseudoStatBuff(
		&warrior.PseudoStats.DamageDealtMultiplier, 0.9,
	)

	warrior.DefensiveStanceAura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})
}

func (warrior *Warrior) registerBerserkerStanceAura() {
	actionId := core.ActionID{SpellID: 2458}
	threatMultiplier := 0.8 - 0.02*float64(warrior.Talents.ImprovedBerserkerStance)

	warrior.BerserkerStanceAura = warrior.GetOrRegisterAura(core.Aura{
		Label:    "Berserker Stance",
		ActionID: actionId,
		Duration: core.NeverExpires,
	}).AttachMultiplicativePseudoStatBuff(
		&warrior.PseudoStats.ThreatMultiplier, threatMultiplier,
	).AttachStatBuff(stats.PhysicalCritPercent, 3)

	warrior.BerserkerStanceAura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})
}

func (warrior *Warrior) registerStances() {
	stanceCD := warrior.NewTimer()
	warrior.registerBattleStanceAura()
	warrior.registerDefensiveStanceAura()
	warrior.registerBerserkerStanceAura()
	warrior.BattleStance = warrior.makeStanceSpell(BattleStance, warrior.BattleStanceAura, stanceCD)
	warrior.DefensiveStance = warrior.makeStanceSpell(DefensiveStance, warrior.DefensiveStanceAura, stanceCD)
	warrior.BerserkerStance = warrior.makeStanceSpell(BerserkerStance, warrior.BerserkerStanceAura, stanceCD)
}
