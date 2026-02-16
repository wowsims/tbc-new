package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var ConsecrationRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 26573, Cost: 120, MinDamage: 8, MaxDamage: 8, Coefficient: 0.119},
	{Rank: 2, SpellID: 20116, Cost: 205, MinDamage: 15, MaxDamage: 15, Coefficient: 0.119},
	{Rank: 3, SpellID: 20922, Cost: 290, MinDamage: 24, MaxDamage: 24, Coefficient: 0.119},
	{Rank: 4, SpellID: 20923, Cost: 390, MinDamage: 35, MaxDamage: 35, Coefficient: 0.119},
	{Rank: 5, SpellID: 20924, Cost: 505, MinDamage: 48, MaxDamage: 48, Coefficient: 0.119},
	{Rank: 6, SpellID: 27173, Cost: 660, MinDamage: 64, MaxDamage: 64, Coefficient: 0.119},
}

// Consecration
// https://www.wowhead.com/tbc/spell=26573
//
// Consecrates the land beneath the Paladin, doing X Holy damage over 8 sec to enemies who enter the area.
func (paladin *Paladin) registerConsecration(rankConfig shared.SpellRankConfig) {
	rank := rankConfig.Rank
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := rankConfig.MinDamage
	coefficient := rankConfig.Coefficient

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: 8 * time.Second,
	}

	consecration := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskConsecration,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: 8,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
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
				ActionID: core.ActionID{SpellID: spellID},
				Label:    "Consecration" + paladin.Label + " Rank " + strconv.Itoa(int(rank)),
			},
			NumberOfTicks:    8,
			TickLength:       time.Second * 1,
			BonusCoefficient: coefficient,
			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				dot.Spell.CalcAndDealPeriodicAoeDamage(sim, minDamage, dot.Spell.OutcomeAlwaysHit)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		},
	})

	paladin.Consecrations = append(paladin.Consecrations, consecration)
}
