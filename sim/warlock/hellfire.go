package warlock

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

const hellFireCoeff = 0.095

func (warlock *Warlock) RegisterHellfire() *core.Spell {
	hellfireActionID := core.ActionID{SpellID: 1949}
	manaMetric := warlock.NewManaMetrics(hellfireActionID)

	manaCost := int32(1665)
	warlock.Hellfire = warlock.RegisterSpell(core.SpellConfig{
		ActionID:         hellfireActionID,
		SpellSchool:      core.SpellSchoolFire,
		Flags:            core.SpellFlagChanneled | core.SpellFlagAPL,
		ProcMask:         core.ProcMaskSpellDamage,
		ClassSpellMask:   WarlockSpellHellfire,
		ThreatMultiplier: 1,
		DamageMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},
		ManaCost: core.ManaCostOptions{FlatCost: manaCost},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Hellfire",
			},

			IsAOE:                true,
			TickLength:           time.Second,
			NumberOfTicks:        14,
			HasteReducesDuration: true,
			AffectedByCastSpeed:  true,
			BonusCoefficient:     hellFireCoeff,

			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				dot.Spell.CalcAndDealPeriodicAoeDamage(sim, 308, dot.Spell.OutcomeMagicHit)
				warlock.RemoveHealth(sim, 308)

			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		},
	})

	return warlock.Hellfire
}
