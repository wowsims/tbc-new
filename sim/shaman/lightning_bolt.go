package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var LightningBoltRankMap = genRanks.LightningBolt

func (shaman *Shaman) registerLightningBoltSpell() {
	shaman.LightningBoltOverloads = make(map[int32]*core.Spell, len(LightningBoltRankMap))
	LightningBoltRankMap.RegisterAll(func(config shared.SpellRank) {
		shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(config, false))
		shaman.LightningBoltOverloads[config.Rank] = shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(config, true))
	})
}

func (shaman *Shaman) newLightningBoltSpellConfig(config shared.SpellRank, isElementalOverload bool) core.SpellConfig {
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
