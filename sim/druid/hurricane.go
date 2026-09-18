package druid

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

var hurricaneRank = genRanks.Hurricane.BySpellID(27012)

func (druid *Druid) registerHurricaneSpell() {
	druid.Hurricane = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: hurricaneRank.SpellID},
		SpellSchool:    core.SpellSchoolNature,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagChanneled | core.SpellFlagAPL,
		ClassSpellMask: DruidSpellHurricane,
		MaxRange:       hurricaneRank.MaxRange,

		ManaCost: core.ManaCostOptions{
			FlatCost: hurricaneRank.Cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: hurricaneRank.GCD,
			},
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: hurricaneRank.Cooldown,
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
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: DruidSpellHurricane,
		// 42230 is the tick the channel triggers, a proc rather than a cast.
		Flags: core.SpellFlagProc,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: hurricaneRank.Direct.BonusCoefficient(),

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.CalcAndDealAoeDamage(sim, hurricaneRank.Direct.Damage(sim), spell.OutcomeMagicHit)
		},
	})
}
