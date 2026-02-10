package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

// Consecration
// https://www.wowhead.com/tbc/spell=26573
//
// Consecrates the land beneath the Paladin, doing X Holy damage over 8 sec to enemies who enter the area.
func (paladin *Paladin) registerConsecration() {
	var ranks = []struct{
		level int32
		spellID int32
		manaCost int32
		value float64
		coeff float64
	}{
		{},
		{level: 20, spellID: 26573, manaCost: 120, value: 8, coeff: 0.119},
		{level: 30, spellID: 20116, manaCost: 205, value: 15, coeff: 0.119},
		{level: 40, spellID: 20922, manaCost: 290, value: 24, coeff: 0.119},
		{level: 50, spellID: 20923, manaCost: 390, value: 35, coeff: 0.119},
		{level: 60, spellID: 20924, manaCost: 505, value: 48, coeff: 0.119},
		{level: 70, spellID: 27173, manaCost: 660, value: 64, coeff: 0.119},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: 8 * time.Second,
	}

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		consecration := paladin.RegisterSpell(core.SpellConfig{
			ActionID:       core.ActionID{SpellID: ranks[rank].spellID},
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskSpellDamage,
			Flags:          core.SpellFlagAPL,
			ClassSpellMask: SpellMaskConsecration,
	
			MaxRange: 8,
	
			ManaCost: core.ManaCostOptions{
				FlatCost: ranks[rank].manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: cd,
			},
	
			Dot: core.DotConfig{
				IsAOE: true,
				Aura: core.Aura{
					ActionID: core.ActionID{SpellID: ranks[rank].spellID},
					Label:    "Consecration" + paladin.Label,
				},
				NumberOfTicks: 8,
				TickLength:    time.Second * 1,
				BonusCoefficient: ranks[rank].coeff,
				OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
					if dot.RemainingTicks() == 8 { // The first tick can be resisted
						dot.Spell.CalcPeriodicAoeDamage(sim, ranks[rank].value, dot.Spell.OutcomeMagicHit)
						dot.Spell.DealBatchedPeriodicDamage(sim)
					} else {
						dot.Spell.CalcPeriodicAoeDamage(sim, ranks[rank].value, dot.Spell.OutcomeAlwaysHit)
						dot.Spell.DealBatchedPeriodicDamage(sim)
					}
				},
			},
	
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.AOEDot().Apply(sim)
			},
		})

		paladin.Consecrations = append(paladin.Consecrations, consecration)
	}
}
