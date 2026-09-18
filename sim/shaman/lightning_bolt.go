package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var LightningBoltRankMap = shared.RankTable{
	{Rank: 1, SpellID: 403, Cost: 15, Direct: &shared.Amount{Min: 15, Max: 17, Coef: 0.13699999452}},
	{Rank: 2, SpellID: 529, Cost: 25, Direct: &shared.Amount{Min: 28, Max: 33, Coef: 0.34900000691}},
	{Rank: 3, SpellID: 548, Cost: 40, Direct: &shared.Amount{Min: 48, Max: 57, Coef: 0.61599999666}},
	{Rank: 4, SpellID: 915, Cost: 70, Direct: &shared.Amount{Min: 88, Max: 100, Coef: 0.79400002956}},
	{Rank: 5, SpellID: 943, Cost: 95, Direct: &shared.Amount{Min: 131, Max: 149, Coef: 0.79400002956}},
	{Rank: 6, SpellID: 6041, Cost: 125, Direct: &shared.Amount{Min: 179, Max: 202, Coef: 0.79400002956}},
	{Rank: 7, SpellID: 10391, Cost: 150, Direct: &shared.Amount{Min: 235, Max: 264, Coef: 0.79400002956}},
	{Rank: 8, SpellID: 10392, Cost: 175, Direct: &shared.Amount{Min: 291, Max: 326, Coef: 0.79400002956}},
	{Rank: 9, SpellID: 15207, Cost: 210, Direct: &shared.Amount{Min: 357, Max: 400, Coef: 0.79400002956}},
	{Rank: 10, SpellID: 15208, Cost: 240, Direct: &shared.Amount{Min: 431, Max: 479, Coef: 0.79400002956}},
	{Rank: 11, SpellID: 25448, Cost: 275, Direct: &shared.Amount{Min: 505, Max: 576, Coef: 0.79400002956}},
	{Rank: 12, SpellID: 25449, Cost: 300, Direct: &shared.Amount{Min: 571, Max: 652, Coef: 0.79400002956}},
}

func (shaman *Shaman) registerLightningBoltSpell() {
	shaman.LightningBoltOverloads = make(map[int32]*core.Spell, len(LightningBoltRankMap))
	LightningBoltRankMap.RegisterAll(func(config shared.RankRow) {
		shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(config, false))
		shaman.LightningBoltOverloads[config.Rank] = shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(config, true))
	})
}

func (shaman *Shaman) newLightningBoltSpellConfig(config shared.RankRow, isElementalOverload bool) core.SpellConfig {
	shamConfig := ShamSpellConfig{
		ActionID:            core.ActionID{SpellID: config.SpellID},
		Rank:                config.Rank,
		IsElementalOverload: isElementalOverload,
		BaseFlatCost:        config.Cost,
		BonusCoefficient:    config.Direct.Coef,
		BaseCastTime:        time.Millisecond * 2500,
	}
	spellConfig := shaman.newElectricSpellConfig(shamConfig)

	spellConfig.ClassSpellMask = core.TernaryInt64(isElementalOverload, SpellMaskLightningBoltOverload, SpellMaskLightningBolt)
	spellConfig.MissileSpeed = 20

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := shaman.CalcAndRollDamageRange(sim, config.Direct.Min, config.Direct.Max)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			if !isElementalOverload && result.Landed() && sim.Proc(shaman.GetOverloadChance(), "Lightning Bolt Elemental Overload") {
				shaman.LightningBoltOverloads[config.Rank].Cast(sim, target)
			}

			spell.DealDamage(sim, result)
		})
	}

	return spellConfig
}
