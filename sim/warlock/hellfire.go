package warlock

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var hellfireRank = genRanks.Hellfire.BySpellID(27213)
var hellfireTick = hellfireRank.Periodic.(shared.SpellRankPeriodic)
var hellFireCoeff = hellfireTick.Coef

func (warlock *Warlock) registerHellfire() *core.Spell {
	hellfireActionID := core.ActionID{SpellID: 27213}

	manaCost := hellfireRank.Cost
	warlock.Hellfire = warlock.RegisterSpell(core.SpellConfig{
		ActionID:         hellfireActionID,
		SpellSchool:      core.SpellSchoolFire,
		DefenseType:      core.DefenseTypeMagic,
		Flags:            core.SpellFlagChanneled | core.SpellFlagAPL,
		ProcMask:         core.ProcMaskSpellDamage,
		ClassSpellMask:   WarlockSpellHellfire,
		ThreatMultiplier: 1,
		DamageMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: hellfireRank.GCD,
			},
		},
		ManaCost: core.ManaCostOptions{FlatCost: manaCost},

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "Hellfire",
			},

			IsAOE:                true,
			TickLength:           hellfireTick.Period,
			NumberOfTicks:        hellfireTick.Ticks,
			HasteReducesDuration: true,
			AffectedByCastSpeed:  true,
			BonusCoefficient:     hellFireCoeff,

			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				resultSlice := dot.Spell.CalcPeriodicAoeDamage(sim, hellfireTick.Tick, dot.Spell.OutcomeTickMagicHitNoHitCounter)
				if resultSlice[0].Damage > warlock.CurrentHealth() {
					dot.Deactivate(sim)
				}

				dot.Spell.DealBatchedPeriodicDamage(sim)
				warlock.RemoveHealth(sim, hellfireTick.Tick)

			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		},
	})

	return warlock.Hellfire
}
