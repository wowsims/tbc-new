package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
)

var ExorcismRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 879, Cost: 70, MinDamage: 90, MaxDamage: 102, Coefficient: 0.429},
	{Rank: 2, SpellID: 5614, Cost: 115, MinDamage: 160, MaxDamage: 180, Coefficient: 0.429},
	{Rank: 3, SpellID: 5615, Cost: 155, MinDamage: 227, MaxDamage: 255, Coefficient: 0.429},
	{Rank: 4, SpellID: 10312, Cost: 200, MinDamage: 316, MaxDamage: 354, Coefficient: 0.429},
	{Rank: 5, SpellID: 10313, Cost: 240, MinDamage: 453, MaxDamage: 507, Coefficient: 0.429},
	{Rank: 6, SpellID: 10314, Cost: 295, MinDamage: 521, MaxDamage: 579, Coefficient: 0.429},
	{Rank: 7, SpellID: 27138, Cost: 340, MinDamage: 626, MaxDamage: 698, Coefficient: 0.429},
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
func (paladin *Paladin) registerExorcism(rankConfig shared.SpellRankConfig) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minDamage := rankConfig.MinDamage
	maxDamage := rankConfig.MaxDamage
	coefficient := rankConfig.Coefficient

	exorcism := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		Rank:           rankConfig.Rank,
		ClassSpellMask: SpellMaskExorcism,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.DefaultSpellCritMultiplier(),

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

	paladin.Exorcisms = append(paladin.Exorcisms, exorcism)
}
