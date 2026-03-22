package warlock

import (
	"github.com/wowsims/tbc/sim/core"
)

func (warlock *Warlock) registerCurseOfRecklessness() {
	warlock.CurseOfRecklessnessAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.CurseOfRecklessnessAura(target, 1)
	})
	warlock.CurseOfRecklessness = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27226},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellCurseOfRecklessness,

		ManaCost: core.ManaCostOptions{
			FlatCost: 160,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				warlock.DeactivateOtherCurses(sim, spell, target)
				warlock.CurseOfRecklessnessAuras.Get(target).Activate(sim)
			}

			spell.DealOutcome(sim, result)
		},

		RelatedAuraArrays: warlock.CurseOfRecklessnessAuras.ToMap(),
	})
}
