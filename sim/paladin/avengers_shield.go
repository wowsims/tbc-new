package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

// Avenger's Shield (Talent)
// https://www.wowhead.com/tbc/spell=31935
//
// Hurls a holy shield at the enemy, dealing Holy damage, dazing them and
// then jumping to additional nearby enemies. Affects 3 total targets.
func (paladin *Paladin) registerAvengersShield() {
	var ranks = []struct {
		level        int32
		spellID      int32
		manaCost     int32
		minValue     float64
		maxValue     float64
		coeff        float64
	}{
		{},
		{level: 50, spellID: 31935, manaCost: 500, minValue: 270, maxValue: 330, coeff: 0.193},
		{level: 60, spellID: 32699, manaCost: 615, minValue: 370, maxValue: 452, coeff: 0.193},
		{level: 70, spellID: 32700, manaCost: 780, minValue: 494, maxValue: 602, coeff: 0.193},
	}

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Second * 30,
	}

	targets := min(int32(3), paladin.Env.TotalTargetCount())

	for rank := 1; rank < len(ranks); rank++ {
		if paladin.Level < ranks[rank].level {
			break
		}

		actionID := core.ActionID{SpellID: ranks[rank].spellID}
		manaCost := ranks[rank].manaCost
		minValue := ranks[rank].minValue
		maxValue := ranks[rank].maxValue
		coeff := ranks[rank].coeff

		paladin.AvengersShields = append(paladin.AvengersShields, paladin.RegisterSpell(core.SpellConfig{
			ActionID:       actionID,
			SpellSchool:    core.SpellSchoolHoly,
			ProcMask:       core.ProcMaskSpellDamage,
			Flags:          core.SpellFlagAPL,
			ClassSpellMask: SpellMaskAvengersShield,

			MaxRange:     30,
			MissileSpeed: 35,

			ManaCost: core.ManaCostOptions{
				FlatCost: manaCost,
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: time.Second,
					CastTime: time.Second,
				},
				CD: cd,
			},

			BonusCoefficient: coeff,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				// TODO: Verify if its rolled once for all targets or per target
				damage := sim.Roll(minValue, maxValue)
				for range targets {
					spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
					target = sim.Environment.NextActiveTargetUnit(target)
				}
			},
		}))
	}
}
