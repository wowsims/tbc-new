package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

func (shaman *Shaman) registerLightningBoltSpell() {
	shared.SpellRankMap{
		{Rank: 1, SpellID: 403, Cost: 15, MinDamage: 15, MaxDamage: 17, Coefficient: 0.13699999452},
		{Rank: 2, SpellID: 529, Cost: 25, MinDamage: 28, MaxDamage: 33, Coefficient: 0.34900000691},
		{Rank: 3, SpellID: 548, Cost: 40, MinDamage: 48, MaxDamage: 57, Coefficient: 0.61599999666},
		{Rank: 4, SpellID: 915, Cost: 70, MinDamage: 88, MaxDamage: 100, Coefficient: 0.79400002956},
		{Rank: 5, SpellID: 943, Cost: 95, MinDamage: 131, MaxDamage: 149, Coefficient: 0.79400002956},
		{Rank: 6, SpellID: 6041, Cost: 125, MinDamage: 179, MaxDamage: 202, Coefficient: 0.79400002956},
		{Rank: 7, SpellID: 10391, Cost: 150, MinDamage: 235, MaxDamage: 264, Coefficient: 0.79400002956},
		{Rank: 8, SpellID: 10392, Cost: 175, MinDamage: 291, MaxDamage: 326, Coefficient: 0.79400002956},
		{Rank: 9, SpellID: 15207, Cost: 210, MinDamage: 357, MaxDamage: 400, Coefficient: 0.79400002956},
		{Rank: 10, SpellID: 15208, Cost: 240, MinDamage: 431, MaxDamage: 479, Coefficient: 0.79400002956},
		{Rank: 11, SpellID: 25448, Cost: 275, MinDamage: 505, MaxDamage: 576, Coefficient: 0.79400002956},
		{Rank: 12, SpellID: 25449, Cost: 300, MinDamage: 571, MaxDamage: 652, Coefficient: 0.79400002956},
	}.RegisterAll(func(config shared.SpellRankConfig) {
		shaman.LightningBolts = append(shaman.LightningBolts, shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(config, false)))
		shaman.LightningBoltOverloads = append(shaman.LightningBoltOverloads, shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(config, true)))
	})
}

func (shaman *Shaman) newLightningBoltSpellConfig(config shared.SpellRankConfig, isElementalOverload bool) core.SpellConfig {
	shamConfig := ShamSpellConfig{
		ActionID:            core.ActionID{SpellID: config.SpellID},
		IsElementalOverload: isElementalOverload,
		BaseFlatCost:        config.Cost,
		BonusCoefficient:    config.Coefficient,
		BaseCastTime:        time.Millisecond * 2500,
	}
	spellConfig := shaman.newElectricSpellConfig(shamConfig)

	spellConfig.ClassSpellMask = core.TernaryInt64(isElementalOverload, SpellMaskLightningBoltOverload, SpellMaskLightningBolt)
	spellConfig.MissileSpeed = 20

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := shaman.CalcAndRollDamageRange(sim, config.MinDamage, config.MaxDamage)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			if !isElementalOverload && result.Landed() && sim.Proc(shaman.GetOverloadChance(), "Lightning Bolt Elemental Overload") {
				shaman.LightningBoltOverloads[config.Rank-1].Cast(sim, target)
			}

			spell.DealDamage(sim, result)
		})
	}

	return spellConfig
}
