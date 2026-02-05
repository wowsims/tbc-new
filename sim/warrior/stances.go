package warrior

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
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

func (warrior *Warrior) makeStanceSpell(stance Stance, mask int64, aura *core.Aura, stanceCD *core.Timer) *core.Spell {
	maxRetainedRage := 10.0 + 5*float64(warrior.Talents.TacticalMastery)
	actionID := aura.ActionID
	rageMetrics := warrior.NewRageMetrics(actionID)

	return warrior.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		ClassSpellMask: mask,
		Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    stanceCD,
				Duration: time.Second * 1,
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

		RelatedSelfBuff: aura,
	})
}

func (warrior *Warrior) registerBattleStanceAura() *core.Aura {
	actionID := core.ActionID{SpellID: 2457}

	aura := warrior.GetOrRegisterAura(core.Aura{
		Label:      "Battle Stance",
		ActionID:   actionID,
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(warrior.DefaultStance == proto.WarriorStance_WarriorStanceBattle, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone),
	}).AttachMultiplicativePseudoStatBuff(&warrior.PseudoStats.ThreatMultiplier, 0.8)

	aura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})

	return aura
}

func (warrior *Warrior) registerDefensiveStanceAura() *core.Aura {
	actionID := core.ActionID{SpellID: 71}

	aura := warrior.GetOrRegisterAura(core.Aura{
		Label:      "Defensive Stance",
		ActionID:   actionID,
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(warrior.DefaultStance == proto.WarriorStance_WarriorStanceDefensive, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone),
	}).AttachMultiplicativePseudoStatBuff(
		&warrior.PseudoStats.ThreatMultiplier, 1.3,
	).AttachMultiplicativePseudoStatBuff(
		&warrior.PseudoStats.DamageTakenMultiplier, 0.9,
	).AttachMultiplicativePseudoStatBuff(
		&warrior.PseudoStats.DamageDealtMultiplier, 0.9,
	)

	aura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})

	return aura
}

func (warrior *Warrior) registerBerserkerStanceAura() *core.Aura {
	actionId := core.ActionID{SpellID: 2458}
	threatMultiplier := 0.8 - 0.02*float64(warrior.Talents.ImprovedBerserkerStance)

	aura := warrior.GetOrRegisterAura(core.Aura{
		Label:      "Berserker Stance",
		ActionID:   actionId,
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(warrior.DefaultStance == proto.WarriorStance_WarriorStanceBerserker, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone),
	}).AttachMultiplicativePseudoStatBuff(
		&warrior.PseudoStats.ThreatMultiplier, threatMultiplier,
	).AttachStatBuff(stats.PhysicalCritPercent, 3)

	aura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})

	return aura
}

func (warrior *Warrior) registerStances() {
	stanceCD := warrior.NewTimer()
	battleStanceAura := warrior.registerBattleStanceAura()
	defensiveStanceAura := warrior.registerDefensiveStanceAura()
	berserkerStanceAura := warrior.registerBerserkerStanceAura()
	warrior.BattleStance = warrior.makeStanceSpell(BattleStance, SpellMaskBattleStance, battleStanceAura, stanceCD)
	warrior.DefensiveStance = warrior.makeStanceSpell(DefensiveStance, SpellMaskDefensiveStance, defensiveStanceAura, stanceCD)
	warrior.BerserkerStance = warrior.makeStanceSpell(BerserkerStance, SpellMaskBerserkerStance, berserkerStanceAura, stanceCD)
}
