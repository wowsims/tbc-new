package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/common/shared"
	"github.com/wowsims/tbc/sim/core"
)

func (paladin *Paladin) registerHealingSpells() {
	HolyLightRankMap.RegisterAll(paladin.registerHolyLight)
	FlashOfLightRankMap.RegisterAll(paladin.registerFlashOfLight)
	LayOnHandsRankMap.RegisterAll(paladin.registerLayOnHands)
}

var HolyLightRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 635, Cost: 35, MinDamage: 42, MaxDamage: 51, Coefficient: 0.205},
	{Rank: 2, SpellID: 639, Cost: 60, MinDamage: 81, MaxDamage: 96, Coefficient: 0.339},
	{Rank: 3, SpellID: 647, Cost: 110, MinDamage: 167, MaxDamage: 196, Coefficient: 0.554},
	{Rank: 4, SpellID: 1026, Cost: 190, MinDamage: 322, MaxDamage: 368, Coefficient: 0.714},
	{Rank: 5, SpellID: 1042, Cost: 275, MinDamage: 506, MaxDamage: 569, Coefficient: 0.714},
	{Rank: 6, SpellID: 3472, Cost: 365, MinDamage: 717, MaxDamage: 799, Coefficient: 0.714},
	{Rank: 7, SpellID: 10328, Cost: 465, MinDamage: 968, MaxDamage: 1076, Coefficient: 0.714},
	{Rank: 8, SpellID: 10329, Cost: 580, MinDamage: 1272, MaxDamage: 1414, Coefficient: 0.714},
	{Rank: 9, SpellID: 25292, Cost: 660, MinDamage: 1619, MaxDamage: 1799, Coefficient: 0.714},
	{Rank: 10, SpellID: 27135, Cost: 710, MinDamage: 1773, MaxDamage: 1971, Coefficient: 0.714},
	{Rank: 11, SpellID: 27136, Cost: 840, MinDamage: 2196, MaxDamage: 2446, Coefficient: 0.714},
}

// Holy Light
// https://www.wowhead.com/tbc/spell=27136
//
// Heals a friendly target for a large amount.
func (paladin *Paladin) registerHolyLight(rankConfig shared.SpellRankConfig) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minHealing := rankConfig.MinDamage
	maxHealing := rankConfig.MaxDamage
	coefficient := rankConfig.Coefficient

	holyLight := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskHolyLight,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: 40,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 2500,
			},
		},

		BonusCoefficient: coefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, sim.Roll(minHealing, maxHealing), spell.OutcomeHealingCrit)
		},
	})

	paladin.HolyLights = append(paladin.HolyLights, holyLight)
}

var FlashOfLightRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 19750, Cost: 35, MinDamage: 67, MaxDamage: 77, Coefficient: 0.429},
	{Rank: 2, SpellID: 19939, Cost: 50, MinDamage: 102, MaxDamage: 117, Coefficient: 0.429},
	{Rank: 3, SpellID: 19940, Cost: 70, MinDamage: 153, MaxDamage: 171, Coefficient: 0.429},
	{Rank: 4, SpellID: 19941, Cost: 90, MinDamage: 206, MaxDamage: 231, Coefficient: 0.429},
	{Rank: 5, SpellID: 19942, Cost: 115, MinDamage: 278, MaxDamage: 310, Coefficient: 0.429},
	{Rank: 6, SpellID: 19943, Cost: 140, MinDamage: 356, MaxDamage: 396, Coefficient: 0.429},
	{Rank: 7, SpellID: 27137, Cost: 180, MinDamage: 458, MaxDamage: 513, Coefficient: 0.429},
}

// Flash of Light
// https://www.wowhead.com/tbc/spell=27137
//
// Heals a friendly target for a small amount.
func (paladin *Paladin) registerFlashOfLight(rankConfig shared.SpellRankConfig) {
	spellID := rankConfig.SpellID
	cost := rankConfig.Cost
	minHealing := rankConfig.MinDamage
	maxHealing := rankConfig.MaxDamage
	coefficient := rankConfig.Coefficient

	flashOfLight := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskFlashOfLight,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: 40,

		ManaCost: core.ManaCostOptions{
			FlatCost: cost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 1500,
			},
		},

		BonusCoefficient: coefficient,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, sim.Roll(minHealing, maxHealing), spell.OutcomeHealingCrit)
		},
	})

	paladin.FlashOfLights = append(paladin.FlashOfLights, flashOfLight)
}

var LayOnHandsRankMap = shared.SpellRankMap{
	{Rank: 1, SpellID: 633, Cost: 0, MinDamage: 0},
	{Rank: 2, SpellID: 2800, Cost: 0, MinDamage: 250},
	{Rank: 3, SpellID: 10310, Cost: 0, MinDamage: 550},
	{Rank: 4, SpellID: 27154, Cost: 0, MinDamage: 900},
}

// Lay on Hands
// https://www.wowhead.com/tbc/spell=27154
//
// Heals a friendly target for an amount equal to the Paladin's maximum health
// and restores mana to the target. Causes Forbearance for 1 min.
func (paladin *Paladin) registerLayOnHands(rankConfig shared.SpellRankConfig) {
	spellID := rankConfig.SpellID
	manaRestore := rankConfig.MinDamage

	cd := core.Cooldown{
		Timer:    paladin.NewTimer(),
		Duration: time.Hour,
	}

	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: spellID})

	paladin.LayOnHands = append(paladin.LayOnHands, paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolHoly,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskLayOnHands,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: 40,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: cd,
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Drain all of the caster's mana
			spell.Unit.AddMana(sim, -spell.Unit.CurrentMana(), manaMetrics)

			// Restore mana and health to the target
			target.AddMana(sim, manaRestore, manaMetrics)
			spell.CalcAndDealHealing(sim, target, spell.Unit.MaxHealth(), spell.OutcomeHealingCrit)
		},
	}))
}
