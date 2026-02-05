package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

// Judgement
// https://www.wowhead.com/tbc/spell=20271
//
// Unleashes the energy of a Seal to judge an enemy for 20 sec.
// The effect depends on which Seal is active.
func (paladin *Paladin) registerJudgement() {
	// Judgement functions as a dummy spell in TBC.
	// It rolls on the spell hit table and can only miss or hit.
	// Individual seals have their own effects that this spell triggers,
	// that are handled in the implementations of the seal auras.
	paladin.Judgement = paladin.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 20271},
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		ClassSpellMask: SpellMaskJudgement,

		Cast: core.CastConfig{
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Second * 10,
			},
		},

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 5,
		},

		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return paladin.CurrentSeal.IsActive() || paladin.PreviousSeal.IsActive()
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			// The oldest seal is consumed, if active. This only matters for if the paladin judges during a twist.
			if paladin.PreviousSeal.IsActive() {
				paladin.PreviousJudgement.Cast(sim, target)
				paladin.PreviousSeal.Deactivate(sim)
			} else {
				paladin.CurrentJudgement.Cast(sim, target)
				paladin.CurrentSeal.Deactivate(sim)
			}
		},
	})
}
