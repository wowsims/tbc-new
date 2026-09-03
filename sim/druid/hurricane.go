package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

func (druid *Druid) registerHurricaneSpell() {
	druid.Hurricane = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 27012},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagChanneled | core.SpellFlagAPL,
		ClassSpellMask: DruidSpellHurricane,

		ManaCost: core.ManaCostOptions{
			FlatCost: 1905,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Second * 60,
			},
		},
		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: "Hurricane (Aura)",
			},
			NumberOfTicks:       10,
			TickLength:          time.Second * 1,
			AffectedByCastSpeed: true,
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				druid.Hurricane.RelatedDotSpell.Cast(sim, target)
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		},
	})

	druid.Hurricane.RelatedDotSpell = druid.Unit.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 42230},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskSpellProc,
		ClassSpellMask: DruidSpellHurricane,

		CritMultiplier:   druid.DefaultSpellCritMultiplier(),
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 0.129,

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			// Cannot crit, the same way Blizzard cannot. 1472 Hurricane ticks across TBC logs
			// land as plain hits.
			spell.CalcAndDealAoeDamage(sim, 206, spell.OutcomeMagicHit)
		},
	})
}
