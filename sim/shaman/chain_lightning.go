package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

var ChainLightningRankMap = shared.RankTable{
	{Rank: 1, SpellID: 421, Cost: 255, Direct: &shared.Amount{Min: 200, Max: 227, Coef: 0.65100002289}},
	{Rank: 2, SpellID: 930, Cost: 345, Direct: &shared.Amount{Min: 288, Max: 323, Coef: 0.65100002289}},
	{Rank: 3, SpellID: 2860, Cost: 445, Direct: &shared.Amount{Min: 391, Max: 438, Coef: 0.65100002289}},
	{Rank: 4, SpellID: 10605, Cost: 550, Direct: &shared.Amount{Min: 508, Max: 567, Coef: 0.65100002289}},
	{Rank: 5, SpellID: 25439, Cost: 650, Direct: &shared.Amount{Min: 620, Max: 705, Coef: 0.65100002289}},
	{Rank: 6, SpellID: 25442, Cost: 760, Direct: &shared.Amount{Min: 734, Max: 838, Coef: 0.65100002289}},
}

func (shaman *Shaman) registerChainLightningSpell() {
	maxHits := min(3, shaman.Env.TotalTargetCount())
	sharedCDTimer := shaman.NewTimer()
	shaman.ChainLightningOverloads = make(map[int32][]*core.Spell, len(ChainLightningRankMap))
	ChainLightningRankMap.RegisterAll(func(config shared.RankRow) {
		shaman.newChainLightningSpell(config, false, sharedCDTimer)
		for range maxHits {
			shaman.ChainLightningOverloads[config.Rank] = append(shaman.ChainLightningOverloads[config.Rank], shaman.newChainLightningSpell(config, true, nil))
		}
	})

}

func (shaman *Shaman) newChainLightningSpell(config shared.RankRow, isElementalOverload bool, sharedCDTimer *core.Timer) *core.Spell {
	shamConfig := ShamSpellConfig{
		ActionID:            core.ActionID{SpellID: config.SpellID},
		Rank:                config.Rank,
		IsElementalOverload: isElementalOverload,
		BaseFlatCost:        config.Cost,
		BonusCoefficient:    config.Direct.Coef,
		SpellSchool:         core.SpellSchoolNature,
		Overloads:           shaman.ChainLightningOverloads,
		BounceReduction:     0.7 + core.TernaryFloat64(shaman.CouldHaveSetBonus(ItemSetTidefuryRaiment, 2), 0.13, 0),
		ClassSpellMask:      core.TernaryInt64(isElementalOverload, SpellMaskChainLightningOverload, SpellMaskChainLightning),
		BaseCastTime:        time.Second * 2,
	}
	spellConfig := shaman.newElectricSpellConfig(shamConfig)
	if !isElementalOverload {
		spellConfig.Cast.CD = core.Cooldown{
			Timer:    sharedCDTimer,
			Duration: time.Second * 6,
		}
	}
	maxHits := int32(3)
	maxHits = min(maxHits, shaman.Env.TotalTargetCount())

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		curTarget := target

		// Damage calculation and DealDamage are in separate loops so that e.g. a spell power proc
		// can't proc on the first target and apply to the second
		numHits := min(maxHits, shaman.Env.ActiveTargetCount())
		results := make([]*core.SpellResult, numHits)
		for hitIndex := range numHits {
			baseDamage := shaman.CalcAndRollDamageRange(sim, config.Direct.Min, config.Direct.Max)
			results[hitIndex] = spell.CalcDamage(sim, curTarget, baseDamage, spell.OutcomeMagicHitAndCrit)

			curTarget = sim.Environment.NextActiveTargetUnit(curTarget)
			spell.DamageMultiplier *= shamConfig.BounceReduction
		}

		for hitIndex := range numHits {
			if !isElementalOverload && results[hitIndex].Landed() && sim.Proc(shaman.GetOverloadChance()/3, "Chain Lightning Elemental Overload") {
				shamConfig.Overloads[config.Rank][hitIndex].Cast(sim, results[hitIndex].Target)
			}
			spell.DealDamage(sim, results[hitIndex])
			spell.DamageMultiplier /= shamConfig.BounceReduction
		}
	}

	return shaman.RegisterSpell(spellConfig)
}
