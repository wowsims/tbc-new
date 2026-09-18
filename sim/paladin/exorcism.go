package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

var ExorcismRankMap = shared.RankTable{
	{Rank: 1, SpellID: 879, Cost: 70, Direct: &shared.Amount{Min: 90, Max: 102, Coef: 0.429}},
	{Rank: 2, SpellID: 5614, Cost: 115, Direct: &shared.Amount{Min: 160, Max: 180, Coef: 0.429}},
	{Rank: 3, SpellID: 5615, Cost: 155, Direct: &shared.Amount{Min: 227, Max: 255, Coef: 0.429}},
	{Rank: 4, SpellID: 10312, Cost: 200, Direct: &shared.Amount{Min: 316, Max: 354, Coef: 0.429}},
	{Rank: 5, SpellID: 10313, Cost: 240, Direct: &shared.Amount{Min: 453, Max: 507, Coef: 0.429}},
	{Rank: 6, SpellID: 10314, Cost: 295, Direct: &shared.Amount{Min: 521, Max: 579, Coef: 0.429}},
	{Rank: 7, SpellID: 27138, Cost: 340, Direct: &shared.Amount{Min: 626, Max: 698, Coef: 0.429}},
}

func (paladin *Paladin) getExorcismTimer() *core.Timer {
	if paladin.exorcismTimer == nil {
		paladin.exorcismTimer = paladin.NewTimer()
	}
	return paladin.exorcismTimer
}

// Exorcism
// https://www.wowhead.com/tbc/spell=10314
//
// Causes X to Y Holy damage to an Undead or Demon target.
func (paladin *Paladin) registerExorcism(rankConfig shared.RankRow) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := rankConfig.Direct.Min
	maxDamage := rankConfig.Direct.Max
	coefficient := rankConfig.Direct.Coef

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		Rank:           rankConfig.Rank,
		ClassSpellMask: SpellMaskExorcism,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: 30,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.getExorcismTimer(),
				Duration: time.Second * 15,
			},
		},

		BonusCoefficient: coefficient,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return target.MobType == proto.MobType_MobTypeUndead || target.MobType == proto.MobType_MobTypeDemon
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, sim.Roll(minDamage, maxDamage), spell.OutcomeMagicHitAndCrit)
		},
	})
}
