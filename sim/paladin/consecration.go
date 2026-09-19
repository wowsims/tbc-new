package paladin

import (
	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

func (paladin *Paladin) getConsecrationTimer() *core.Timer {
	if paladin.consecrationTimer == nil {
		paladin.consecrationTimer = paladin.NewTimer()
	}
	return paladin.consecrationTimer
}

var ConsecrationRankMap = genRanks.Consecration

// Consecration
// https://www.wowhead.com/tbc/spell=26573
//
// Consecrates the land beneath the Paladin, doing X Holy damage over 8 sec to enemies who enter the area.
func (paladin *Paladin) registerConsecration(rankConfig shared.SpellRank) {
	tick := rankConfig.Periodic.(shared.SpellRankPeriodic)

	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := shared.SpellRankMin(rankConfig.Periodic)
	coefficient := rankConfig.Periodic.BonusCoefficient()

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskConsecration,
		Rank:           rankConfig.Rank,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: 8,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: rankConfig.GCD,
			},
			CD: core.Cooldown{
				Timer:    paladin.getConsecrationTimer(),
				Duration: rankConfig.Cooldown,
			},
		},

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				ActionID: core.ActionID{SpellID: spellID},
				Label:    "Consecration" + paladin.Label + " " + rankConfig.GetRankLabel(),
			},
			NumberOfTicks:    7, // the table says 8; the sim adds an immediate tick below
			TickLength:       tick.TickLength,
			BonusCoefficient: coefficient,
			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				dot.Spell.CalcAndDealPeriodicAoeDamage(sim, minDamage, dot.OutcomeTickMagicHit)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Consecration does one hit check on cast but the ground effect will still be applied
			// meaning it's only needed to proc things like Eye of Magtheridon (procs on resist)
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)

			dot := spell.AOEDot()
			dot.Apply(sim)
			dot.Spell.CalcAndDealPeriodicAoeDamage(sim, minDamage, dot.OutcomeTickMagicHit)
		},
	})
}
