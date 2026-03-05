package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (warlock *Warlock) registerCurseOfElements() {
	warlock.CurseOfElementsAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.CurseOfElementsAura(target, warlock.Talents.Malediction)
	})
	warlock.CurseOfElements = warlock.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27228},
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: WarlockSpellCurseOfElements,

		ManaCost: core.ManaCostOptions{
			FlatCost: 260,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCDMin: time.Millisecond * 500,
				GCD:    core.GCDMin,
			},
		},

		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				warlock.DeactivateOtherCurses(sim, target)
				warlock.CurseOfElementsAuras.Get(target).Activate(sim)
			}

			spell.DealOutcome(sim, result)
		},
	})
}
